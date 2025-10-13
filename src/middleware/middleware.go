package middleware

import (
	"connection-service/src/config"
	"fmt"
	"log"
	"log/slog"
	"time"

	"github.com/streadway/amqp"
)

type Middleware struct {
	conn          *amqp.Connection
	channel       *amqp.Channel
	confirms_chan chan amqp.Confirmation
	config        config.GlobalConfig
}

const MAX_RETRIES = 5

// Publisher interface for compatibility with existing services
type Publisher interface {
	Publish(exchange string, body []byte) error
}

func NewMiddleware(config config.GlobalConfig) (*Middleware, error) {
	url := fmt.Sprintf("amqp://%s:%s@%s:%d/",
		config.RabbitUser, config.RabbitPass, config.RabbitHost, config.RabbitPort)

	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	if err := ch.Confirm(false); err != nil {
		return nil, err
	}

	confirms_chan := ch.NotifyPublish(make(chan amqp.Confirmation, 1))

	if err := ch.Qos(1, 0, false); err != nil {
		return nil, err
	}

	slog.Info("Connected to RabbitMQ", "host", config.RabbitHost, "port", config.RabbitPort, "user", config.RabbitUser)

	middleware := &Middleware{
		conn:          conn,
		channel:       ch,
		confirms_chan: confirms_chan,
		config:        config,
	}

	// Setup connection queues automatically
	if err := middleware.setupConnectionQueues(); err != nil {
		middleware.Close()
		return nil, fmt.Errorf("failed to setup connection queues: %w", err)
	}

	return middleware, nil
}

func (m *Middleware) DeclareQueue(queueName string, durable bool) error {
	_, err := m.channel.QueueDeclare(
		queueName, // name
		durable,   // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	return err
}

func (m *Middleware) DeclareExchange(exchangeName string, exchangeType string, durable bool) error {
	return m.channel.ExchangeDeclare(
		exchangeName,
		exchangeType,
		durable, // durable
		false,   // autoDelete
		false,   // internal
		false,   // noWait
		nil,     // arguments
	)
}

func (m *Middleware) BindQueue(queueName, exchangeName, routingKey string) error {
	return m.channel.QueueBind(
		queueName,
		routingKey,
		exchangeName,
		false,
		nil,
	)
}

// Publish method compatible with Publisher interface
func (m *Middleware) Publish(exchange string, body []byte) error {
	return m.PublishWithRouting("", body, exchange)
}

// PublishWithRouting publishes with specific routing key
func (m *Middleware) PublishWithRouting(routingKey string, message []byte, exchangeName string) error {
	for attempt := 1; attempt <= MAX_RETRIES; attempt++ {
		err := m.channel.Publish(
			exchangeName,
			routingKey,
			false, // mandatory
			false, // immediate
			amqp.Publishing{
				DeliveryMode: amqp.Persistent,
				ContentType:  "application/json",
				Body:         message,
			},
		)

		if err != nil {
			slog.Error("Failed to publish message to exchange", "routing_key", routingKey, "exchange", exchangeName, "attempt", attempt)
			time.Sleep(time.Second * time.Duration(attempt))
			continue
		}

		confirmed := <-m.confirms_chan

		if !confirmed.Ack {
			slog.Error("Failed to publish message to exchange - not acknowledged", "routing_key", routingKey, "exchange", exchangeName, "attempt", attempt)
			time.Sleep(time.Second * time.Duration(attempt))
			continue
		}

		slog.Debug("Published message to exchange", "routing_key", routingKey, "exchange", exchangeName)

		return nil
	}
	return fmt.Errorf("failed to publish message to exchange %s after %d attempts", exchangeName, MAX_RETRIES)
}

func (m *Middleware) BasicConsume(queueName string, callback func(amqp.Delivery)) error {
	msgs, err := m.channel.Consume(
		queueName,
		"",    // consumer
		false, // autoAck
		false, // exclusive
		false, // noLocal
		false, // noWait
		nil,   // args
	)
	if err != nil {
		return err
	}

	go func() {
		for msg := range msgs {
			func(m amqp.Delivery) {
				defer func() {
					if r := recover(); r != nil {
						log.Printf("action: rabbitmq_callback | result: fail | error: %v\n", r)
						_ = m.Nack(false, true)
					}
				}()
				callback(m)
				_ = m.Ack(false)
			}(msg)
		}
	}()

	return nil
}

func (m *Middleware) Close() {
	if err := m.channel.Close(); err != nil {
		log.Printf("action: rabbitmq_channel_close | result: fail | error: %v", err)
	}
	if err := m.conn.Close(); err != nil {
		log.Printf("action: rabbitmq_connection_close | result: fail | error: %v", err)
	}
}

func (m *Middleware) setupConnectionQueues() error {
	exchangeName := config.CONNECTION_EXCHANGE
	queues := []string{config.DATA_DISPATCHER_CONNECTION, config.CALIBRATION_SERVICE_CONNECTION}

	slog.Info("Setting up RabbitMQ queues and bindings", "exchange", exchangeName, "queues", queues)

	// Declare the exchange (fanout type for broadcasting)
	if err := m.DeclareExchange(exchangeName, "fanout", false); err != nil {
		return err
	}

	// Declare queues and bind them to the exchange
	for _, queueName := range queues {
		if err := m.DeclareQueue(queueName, false); err != nil {
			return err
		}

		if err := m.BindQueue(queueName, exchangeName, ""); err != nil {
			return err
		}

		slog.Info("Queue created and bound to exchange", "queue", queueName, "exchange", exchangeName)
	}

	return nil
}

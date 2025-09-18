package rabbitmq

import (
	"auth-gateway/src/config"
	"fmt"

	"github.com/streadway/amqp"
)

// Publisher defines the interface for publishing messages to RabbitMQ.
type Publisher interface {
	Publish(exchange string, body []byte) error
}

type AMQPPublisher struct {
	conn    *amqp.Connection
	channel *amqp.Channel
}

// NewAMQPPublisherFromConfig creates a new AMQPPublisher using configuration.
func NewAMQPPublisherFromConfig(cfg config.GlobalConfig) (*AMQPPublisher, error) {
	amqpURL := fmt.Sprintf("amqp://%s:%s@%s:%d/", cfg.RabbitUser, cfg.RabbitPass, cfg.RabbitHost, cfg.RabbitPort)
	return NewAMQPPublisher(amqpURL)
}

// NewAMQPPublisher creates a new AMQPPublisher and connects to RabbitMQ.
func NewAMQPPublisher(amqpURL string) (*AMQPPublisher, error) {
	conn, err := amqp.Dial(amqpURL)
	if err != nil {
		return nil, err
	}
	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, err
	}
	return &AMQPPublisher{conn: conn, channel: ch}, nil
}

// Publish publishes a message to the given exchange.
func (p *AMQPPublisher) Publish(exchange string, body []byte) error {
	err := p.channel.ExchangeDeclare(
		exchange,
		"fanout",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}
	return p.channel.Publish(
		exchange,
		"",
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
}

// Close closes the RabbitMQ connection and channel.
func (p *AMQPPublisher) Close() {
	if p.channel != nil {
		p.channel.Close()
	}
	if p.conn != nil {
		p.conn.Close()
	}
}

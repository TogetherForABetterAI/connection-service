package rabbitmq

import "github.com/streadway/amqp"

// Publisher defines the interface for publishing messages to RabbitMQ.
type Publisher interface {
	Publish(exchange string, body []byte) error
}

type AMQPPublisher struct {
	conn    *amqp.Connection
	channel *amqp.Channel
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

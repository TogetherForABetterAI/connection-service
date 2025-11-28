package middleware

import (
	"bytes"
	"connection-service/src/config"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/streadway/amqp"
)

type Middleware struct {
	conn          *amqp.Connection
	channel       *amqp.Channel
	confirms_chan chan amqp.Confirmation
	config        *config.GlobalConfig
}

const MAX_RETRIES = 5

// Publisher interface for compatibility with existing services
type Publisher interface {
	Publish(exchange string, body []byte) error
}

func NewMiddleware(config *config.GlobalConfig) (*Middleware, error) {
	middlewareConfig := config.GetMiddlewareConfig()
	url := fmt.Sprintf("amqp://%s:%s@%s:%d/",
		middlewareConfig.GetUsername(),
		middlewareConfig.GetPassword(),
		middlewareConfig.GetHost(),
		middlewareConfig.GetPort())

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

	slog.Info("Connected to RabbitMQ",
		"host", middlewareConfig.GetHost(),
		"port", middlewareConfig.GetPort(),
		"user", middlewareConfig.GetUsername())

	middleware := &Middleware{
		conn:          conn,
		channel:       ch,
		confirms_chan: confirms_chan,
		config:        config,
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

// DeleteQueue deletes a RabbitMQ queue
func (m *Middleware) DeleteQueue(queueName string) error {
	_, err := m.channel.QueueDelete(
		queueName, // name
		false,     // ifUnused
		false,     // ifEmpty
		false,     // noWait
	)

	if err != nil {
		return fmt.Errorf("failed to delete queue %s: %w", queueName, err)
	}

	slog.Info("Deleted Queue", "queue", queueName)
	return nil
}

//
// HTTP Management API Methods
//

// GetAdminAPIURL returns the RabbitMQ Management API URL
func (m *Middleware) GetAdminAPIURL() string {
	middlewareConfig := m.config.GetMiddlewareConfig()
	return fmt.Sprintf("http://%s:15672/api", middlewareConfig.GetHost())
}

// GetAdminCredentials returns the admin username and password
func (m *Middleware) GetAdminCredentials() (string, string) {
	middlewareConfig := m.config.GetMiddlewareConfig()
	return middlewareConfig.GetUsername(), middlewareConfig.GetPassword()
}

// CreateUser creates a new RabbitMQ user using HTTP Management API
func (m *Middleware) CreateUser(username, password string) error {
	adminAPIURL := m.GetAdminAPIURL()
	adminUser, adminPass := m.GetAdminCredentials()

	url := fmt.Sprintf("%s/users/%s", adminAPIURL, username)

	userData := map[string]interface{}{
		"password": password,
		"tags":     "", // No admin tags for client users
	}

	jsonData, err := json.Marshal(userData)
	if err != nil {
		return fmt.Errorf("failed to marshal user data: %w", err)
	}

	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.SetBasicAuth(adminUser, adminPass)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// 201: User was created successfully
	// 204: User already exists and was updated (not an error)
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to create user, status code: %d, response: %s", resp.StatusCode, string(body))
	}

	if resp.StatusCode == http.StatusCreated {
		slog.Info("Created new RabbitMQ user", "username", username)
	} else {
		slog.Info("RabbitMQ user already exists, credentials updated", "username", username)
	}

	return nil
}

// SetPermissions sets permissions for a user on a vhost using HTTP Management API
func (m *Middleware) SetPermissions(vhost, username, configurePattern, writePattern, readPattern string) error {
	adminAPIURL := m.GetAdminAPIURL()
	adminUser, adminPass := m.GetAdminCredentials()

	encodedVhost := strings.ReplaceAll(vhost, "/", "%2F")
	url := fmt.Sprintf("%s/permissions/%s/%s", adminAPIURL, encodedVhost, username)

	permissions := map[string]interface{}{
		"configure": configurePattern,
		"write":     writePattern,
		"read":      readPattern,
	}

	jsonData, err := json.Marshal(permissions)
	if err != nil {
		return fmt.Errorf("failed to marshal permissions: %w", err)
	}

	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.SetBasicAuth(adminUser, adminPass)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("failed to set permissions, status code: %d", resp.StatusCode)
	}

	slog.Info("Set Permissions", "vhost", vhost, "username", username)
	return nil
}

// DeleteUser deletes a RabbitMQ user using HTTP Management API
func (m *Middleware) DeleteUser(username string) error {
	adminAPIURL := m.GetAdminAPIURL()
	adminUser, adminPass := m.GetAdminCredentials()

	url := fmt.Sprintf("%s/users/%s", adminAPIURL, username)

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.SetBasicAuth(adminUser, adminPass)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusNotFound {
		return fmt.Errorf("failed to delete user, status code: %d", resp.StatusCode)
	}

	slog.Info("Deleted User", "username", username)
	return nil
}

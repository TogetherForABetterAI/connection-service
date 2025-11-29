package config

import (
	"fmt"
	"os"
	"strconv"
)

const (
	CONNECTION_EXCHANGE             = "new_connections_exchange"
	DISPATCHER_TO_CLIENT_QUEUE      = "%s_dispatcher_queue"
	CLIENT_TO_CALIBRATION_QUEUE     = "%s_outputs_cal_queue"
	DISPATCHER_TO_CALIBRATION_QUEUE = "%s_inputs_cal_queue"
)

// Interface defines the configuration contract
type Interface interface {
	GetLogLevel() string
	GetPodName() string
	GetHost() string
	GetPort() string
	GetMiddlewareConfig() *MiddlewareConfig
	GetDatabaseConfig() *DatabaseConfig
}

// GlobalConfig holds all service configuration
type GlobalConfig struct {
	logLevel         string
	podName          string
	host             string
	port             string
	middlewareConfig *MiddlewareConfig
	databaseConfig   *DatabaseConfig
}

// DatabaseConfig holds PostgreSQL connection configuration
type DatabaseConfig struct {
	host     string
	port     int32
	user     string
	password string
	dbname   string
}

// MiddlewareConfig holds RabbitMQ connection configuration
type MiddlewareConfig struct {
	host       string
	port       int32
	username   string
	password   string
	maxRetries int
}

// Getters for GlobalConfig
func (c *GlobalConfig) GetLogLevel() string {
	return c.logLevel
}

func (c *GlobalConfig) GetPodName() string {
	return c.podName
}

func (c *GlobalConfig) GetHost() string {
	return c.host
}

func (c *GlobalConfig) GetPort() string {
	return c.port
}

func (c *GlobalConfig) GetMiddlewareConfig() *MiddlewareConfig {
	return c.middlewareConfig
}

func (c *GlobalConfig) GetDatabaseConfig() *DatabaseConfig {
	return c.databaseConfig
}

// Getters for DatabaseConfig
func (d *DatabaseConfig) GetHost() string {
	return d.host
}

func (d *DatabaseConfig) GetPort() int32 {
	return d.port
}

func (d *DatabaseConfig) GetUser() string {
	return d.user
}

func (d *DatabaseConfig) GetPassword() string {
	return d.password
}

func (d *DatabaseConfig) GetDBName() string {
	return d.dbname
}

// Getters for MiddlewareConfig
func (m *MiddlewareConfig) GetHost() string {
	return m.host
}

func (m *MiddlewareConfig) GetPort() int32 {
	return m.port
}

func (m *MiddlewareConfig) GetUsername() string {
	return m.username
}

func (m *MiddlewareConfig) GetPassword() string {
	return m.password
}

func (m *MiddlewareConfig) GetMaxRetries() int {
	return m.maxRetries
}

// GetRabbitMQHost returns the RabbitMQ host from config
func (c *GlobalConfig) GetRabbitMQHost() string {
	return c.middlewareConfig.GetHost()
}

// GetRabbitMQPort returns the RabbitMQ port from config
func (c *GlobalConfig) GetRabbitMQPort() int32 {
	return c.middlewareConfig.GetPort()
}

func NewConfig() (*GlobalConfig, error) {
	// Get RabbitMQ connection details from environment
	rabbitHost := os.Getenv("RABBITMQ_HOST")
	if rabbitHost == "" {
		return nil, fmt.Errorf("RABBITMQ_HOST environment variable is required")
	}

	rabbitPortStr := os.Getenv("RABBITMQ_PORT")
	if rabbitPortStr == "" {
		return nil, fmt.Errorf("RABBITMQ_PORT environment variable is required")
	}
	rabbitPort, err := strconv.ParseInt(rabbitPortStr, 10, 32)
	if err != nil {
		return nil, fmt.Errorf("RABBITMQ_PORT must be a valid integer: %w", err)
	}

	rabbitUser := os.Getenv("RABBITMQ_USER")
	if rabbitUser == "" {
		return nil, fmt.Errorf("RABBITMQ_USER environment variable is required")
	}

	rabbitPass := os.Getenv("RABBITMQ_PASS")
	if rabbitPass == "" {
		return nil, fmt.Errorf("RABBITMQ_PASS environment variable is required")
	}

	// Set log level from environment
	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "info" // default value
	}

	// Get pod name from environment (optional for local development)
	podName := os.Getenv("POD_NAME")
	if podName == "" {
		podName = "connection-service-local"
	}

	host := os.Getenv("HOST")
	if host == "" {
		return nil, fmt.Errorf("HOST environment variable is required")
	}

	port := os.Getenv("PORT")
	if port == "" {
		return nil, fmt.Errorf("PORT environment variable is required")
	}

	// Get PostgreSQL connection details from environment
	postgresHost := os.Getenv("POSTGRES_HOST")
	if postgresHost == "" {
		return nil, fmt.Errorf("POSTGRES_HOST environment variable is required")
	}

	postgresPortStr := os.Getenv("POSTGRES_PORT")
	if postgresPortStr == "" {
		return nil, fmt.Errorf("POSTGRES_PORT environment variable is required")
	}
	postgresPort, err := strconv.ParseInt(postgresPortStr, 10, 32)
	if err != nil {
		return nil, fmt.Errorf("POSTGRES_PORT must be a valid integer: %w", err)
	}

	postgresUser := os.Getenv("POSTGRES_USER")
	if postgresUser == "" {
		return nil, fmt.Errorf("POSTGRES_USER environment variable is required")
	}

	postgresPass := os.Getenv("POSTGRES_PASS")
	if postgresPass == "" {
		return nil, fmt.Errorf("POSTGRES_PASS environment variable is required")
	}

	postgresDB := os.Getenv("POSTGRES_DB")
	if postgresDB == "" {
		return nil, fmt.Errorf("POSTGRES_DB environment variable is required")
	}

	// Create middleware config
	middlewareConfig := &MiddlewareConfig{
		host:       rabbitHost,
		port:       int32(rabbitPort),
		username:   rabbitUser,
		password:   rabbitPass,
		maxRetries: 5, // default max retries
	}

	// Create database config
	databaseConfig := &DatabaseConfig{
		host:     postgresHost,
		port:     int32(postgresPort),
		user:     postgresUser,
		password: postgresPass,
		dbname:   postgresDB,
	}

	return &GlobalConfig{
		logLevel:         logLevel,
		podName:          podName,
		host:             host,
		port:             port,
		middlewareConfig: middlewareConfig,
		databaseConfig:   databaseConfig,
	}, nil
}

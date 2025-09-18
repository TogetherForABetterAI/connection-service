package config

import (
	"fmt"
	"os"
	"strconv"
)

type GlobalConfig struct {
	LogLevel                  string
	RabbitHost                string
	RabbitPort                int32
	RabbitUser                string
	RabbitPass                string
	DatasetAddr               string
	CalibrationServiceAddr    string
	DataDispatcherServiceAddr string
	Host                      string
	Port                      string
	AppPort                   string
}

func NewConfig() (GlobalConfig, error) {
	// Get RabbitMQ connection details from environment
	rabbitHost := os.Getenv("RABBITMQ_HOST")
	if rabbitHost == "" {
		return GlobalConfig{}, fmt.Errorf("RABBITMQ_HOST environment variable is required")
	}

	rabbitPortStr := os.Getenv("RABBITMQ_PORT")
	if rabbitPortStr == "" {
		return GlobalConfig{}, fmt.Errorf("RABBITMQ_PORT environment variable is required")
	}
	rabbitPort, err := strconv.ParseInt(rabbitPortStr, 10, 32)
	if err != nil {
		return GlobalConfig{}, fmt.Errorf("RABBITMQ_PORT must be a valid integer: %w", err)
	}

	rabbitUser := os.Getenv("RABBITMQ_USER")
	if rabbitUser == "" {
		return GlobalConfig{}, fmt.Errorf("RABBITMQ_USER environment variable is required")
	}

	rabbitPass := os.Getenv("RABBITMQ_PASS")
	if rabbitPass == "" {
		return GlobalConfig{}, fmt.Errorf("RABBITMQ_PASS environment variable is required")
	}

	// Set log level from environment
	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		return GlobalConfig{}, fmt.Errorf("LOG_LEVEL environment variable is required")
	}

	// Get calibration service address from environment
	calibrationAddr := os.Getenv("CALIBRATION_SERVICE_ADDR")
	if calibrationAddr == "" {
		return GlobalConfig{}, fmt.Errorf("CALIBRATION_SERVICE_ADDR environment variable is required")
	}

	// Get data dispatcher service address from environment
	dataDispatcherAddr := os.Getenv("DATA_DISPATCHER_SERVICE_ADDR")
	if dataDispatcherAddr == "" {
		return GlobalConfig{}, fmt.Errorf("DATA_DISPATCHER_SERVICE_ADDR environment variable is required")
	}

	appPort := os.Getenv("APP_PORT")
	if appPort == "" {
		return GlobalConfig{}, fmt.Errorf("APP_PORT environment variable is required")
	}

	host := os.Getenv("HOST")
	if host == "" {
		return GlobalConfig{}, fmt.Errorf("HOST environment variable is required")
	}

	port := os.Getenv("PORT")
	if port == "" {
		return GlobalConfig{}, fmt.Errorf("PORT environment variable is required")
	}

	return GlobalConfig{
		LogLevel:                  logLevel,
		RabbitHost:                rabbitHost,
		RabbitPort:                int32(rabbitPort),
		RabbitUser:                rabbitUser,
		RabbitPass:                rabbitPass,
		CalibrationServiceAddr:    calibrationAddr,
		DataDispatcherServiceAddr: dataDispatcherAddr,
		Host:                      host,
		Port:                      port,
		AppPort:                   appPort,
	}, nil
}

var Config GlobalConfig

func init() {
	config, err := NewConfig()
	if err != nil {
		panic(fmt.Sprintf("Failed to load configuration: %v", err))
	}
	Config = config
}

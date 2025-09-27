package config

import (
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
	AppPort                   string
	SupabaseUrl               string
	SupabaseKey               string
}

func InitializeConfig() GlobalConfig {
	// Get RabbitMQ connection details from environment
	rabbitHost := os.Getenv("RABBITMQ_HOST")
	if rabbitHost == "" {
		rabbitHost = "localhost"
	}

	rabbitPort := int32(5672) // default RabbitMQ port
	if portStr := os.Getenv("RABBITMQ_PORT"); portStr != "" {
		if parsed, err := strconv.ParseInt(portStr, 10, 32); err == nil {
			rabbitPort = int32(parsed)
		}
	}

	rabbitUser := os.Getenv("RABBITMQ_USER")
	if rabbitUser == "" {
		rabbitUser = "guest"
	}

	rabbitPass := os.Getenv("RABBITMQ_PASS")
	if rabbitPass == "" {
		rabbitPass = "guest"
	}

	// Set log level from environment
	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "info"
	}

	// Get calibration service address from environment or use default
	calibrationAddr := os.Getenv("CALIBRATION_SERVICE_ADDR")
	if calibrationAddr == "" {
		calibrationAddr = "calibration-service:50052"
	}

	// Get data dispatcher service address from environment or use default
	dataDispatcherAddr := os.Getenv("DATA_DISPATCHER_SERVICE_ADDR")
	if dataDispatcherAddr == "" {
		dataDispatcherAddr = "data-dispatcher-service:50058"
	}

	appPort := os.Getenv("APP_PORT")
	if appPort == "" {
		appPort = "8080"
	}

	supabase_url := os.Getenv("SUPABASE_URL")
	supabase_key := os.Getenv("SUPABASE_KEY")

	return GlobalConfig{
		LogLevel:                  logLevel,
		RabbitHost:                rabbitHost,
		RabbitPort:                rabbitPort,
		RabbitUser:                rabbitUser,
		RabbitPass:                rabbitPass,
		CalibrationServiceAddr:    calibrationAddr,
		DataDispatcherServiceAddr: dataDispatcherAddr,
		AppPort:                   appPort,
		SupabaseUrl:               supabase_url,
		SupabaseKey:               supabase_key,
	}
}

var Config GlobalConfig

func init() {
	Config = InitializeConfig()
}

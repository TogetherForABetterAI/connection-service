package middleware

import (
	"connection-service/src/config"
	"fmt"
	"log/slog"
)

// RabbitMQTopologyManager manages RabbitMQ topology for client isolation
type RabbitMQTopologyManager struct {
	config     *config.GlobalConfig
	middleware *Middleware
}

// NewTopologyManager creates a new topology manager using an existing middleware instance
func NewTopologyManager(cfg *config.GlobalConfig, middleware *Middleware) *RabbitMQTopologyManager {
	return &RabbitMQTopologyManager{
		config:     cfg,
		middleware: middleware,
	}
}

// SetUpTopologyFor creates the RabbitMQ topology for a client in the SHARED VHost ('/')
// This includes: User, Client Queues, Exchange, Bindings, and Permissions
func (tm *RabbitMQTopologyManager) SetUpTopologyFor(UserID string, password string) error {
	const vhost = "/"

	slog.Info("Setting up RabbitMQ topology for client",
		"user_id", UserID,
		"vhost", vhost,
		"username", UserID)

	if err := tm.middleware.CreateUser(UserID, password); err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	dispatcherToClientQueue := fmt.Sprintf(config.DISPATCHER_TO_CLIENT_QUEUE, UserID)
	clientToCalibrationQueue := fmt.Sprintf(config.CLIENT_TO_CALIBRATION_QUEUE, UserID)

	slog.Info("Middleware channel", "channel", tm.middleware.channel)

	if err := tm.middleware.DeclareQueue(dispatcherToClientQueue, true); err != nil {
		return fmt.Errorf("failed to create dispatcher queue: %w", err)
	}
	if err := tm.middleware.DeclareQueue(clientToCalibrationQueue, true); err != nil {
		return fmt.Errorf("failed to create calibration queue: %w", err)
	}

	readPattern := fmt.Sprintf("^%s$", dispatcherToClientQueue)
	writePattern := fmt.Sprintf("^(%s|amq\\.default)$", clientToCalibrationQueue)
	configurePattern := ""

	if err := tm.middleware.SetPermissions(vhost, UserID, configurePattern, writePattern, readPattern); err != nil { //
		return fmt.Errorf("failed to set permissions for user %s: %w", UserID, err)
	}

	slog.Info("Successfully set up RabbitMQ topology for client", "user_id", UserID)
	return nil
}

// DeleteTopologyFor removes all RabbitMQ resources for a client (useful for cleanup)
func (tm *RabbitMQTopologyManager) DeleteTopologyFor(UserID string) error {
	username := UserID
	dispatcherQueue := fmt.Sprintf(config.DISPATCHER_TO_CLIENT_QUEUE, UserID)
	calibrationQueue := fmt.Sprintf(config.CLIENT_TO_CALIBRATION_QUEUE, UserID)
	dispatcherToCalibrationQueue := fmt.Sprintf(config.DISPATCHER_TO_CALIBRATION_QUEUE, UserID)

	slog.Info("Deleting RabbitMQ topology for client", "user_id", UserID)

	if err := tm.middleware.DeleteQueue(dispatcherQueue); err != nil {
		slog.Error("Failed to delete dispatcher queue", "queue", dispatcherQueue, "error", err)
	}
	if err := tm.middleware.DeleteQueue(calibrationQueue); err != nil {
		slog.Error("Failed to delete calibration queue", "queue", calibrationQueue, "error", err)
	}

	if err := tm.middleware.DeleteQueue(dispatcherToCalibrationQueue); err != nil {
		slog.Error("Failed to delete calibration queue", "queue", calibrationQueue, "error", err)
	}

	if err := tm.middleware.DeleteUser(username); err != nil {
		slog.Error("Failed to delete user", "username", username, "error", err)
		return fmt.Errorf("failed to complete topology deletion for %s", UserID)
	}

	slog.Info("Successfully deleted RabbitMQ topology for client", "user_id", UserID)
	return nil
}

func (tm *RabbitMQTopologyManager) GetMiddleware() *Middleware {
	return tm.middleware
}

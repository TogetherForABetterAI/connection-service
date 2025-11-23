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
func (tm *RabbitMQTopologyManager) SetUpTopologyFor(clientID string, password string) error {
	const vhost = "/"
	username := fmt.Sprintf("%s_user", clientID)

	slog.Info("Setting up RabbitMQ topology for client",
		"client_id", clientID,
		"vhost", vhost,
		"username", username)

	if err := tm.middleware.CreateUser(username, password); err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	dispatcherToClientQueue := fmt.Sprintf("%s_dispatcher_queue", clientID)
	clientToCalibrationQueue := fmt.Sprintf("%s_calibration_queue", clientID)

	if err := tm.middleware.DeclareQueue(dispatcherToClientQueue, true); err != nil {
		return fmt.Errorf("failed to create dispatcher queue: %w", err)
	}
	if err := tm.middleware.DeclareQueue(clientToCalibrationQueue, true); err != nil {
		return fmt.Errorf("failed to create calibration queue: %w", err)
	}

	readPattern := fmt.Sprintf("^%s$", dispatcherToClientQueue)
	writePattern := fmt.Sprintf("^%s$", clientToCalibrationQueue)
	configurePattern := ""

	if err := tm.middleware.SetPermissions(vhost, username, configurePattern, writePattern, readPattern); err != nil { //
		return fmt.Errorf("failed to set permissions for user %s: %w", username, err)
	}

	slog.Info("Successfully set up RabbitMQ topology for client", "client_id", clientID)
	return nil
}

// DeleteTopologyFor removes all RabbitMQ resources for a client (useful for cleanup)
func (tm *RabbitMQTopologyManager) DeleteTopologyFor(clientID string) error {
	username := fmt.Sprintf("%s_user", clientID)
	dispatcherQueue := fmt.Sprintf("%s_dispatcher_queue", clientID)
	calibrationQueue := fmt.Sprintf("%s_calibration_queue", clientID)
	dispatcherToCalibrationQueue := fmt.Sprintf("%s_labeled_queue", clientID)

	slog.Info("Deleting RabbitMQ topology for client", "client_id", clientID)

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
		return fmt.Errorf("failed to complete topology deletion for %s", clientID)
	}

	slog.Info("Successfully deleted RabbitMQ topology for client", "client_id", clientID)
	return nil
}


func (tm *RabbitMQTopologyManager) GetMiddleware() *Middleware {
	return tm.middleware
}
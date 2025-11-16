package middleware

import (
	"fmt"
	"log/slog"
	"os/exec"

	"connection-service/src/config"
)

// RabbitMQTopologyManager manages RabbitMQ topology for client isolation
type RabbitMQTopologyManager struct {
	config     *config.GlobalConfig
	middleware *Middleware
}

// NewRabbitMQTopologyManager creates a new topology manager
func NewRabbitMQTopologyManager(cfg *config.GlobalConfig) *RabbitMQTopologyManager {
	// Create middleware instance for admin operations
	middleware, err := NewMiddleware(cfg)
	if err != nil {
		slog.Error("Failed to create middleware for topology manager", "error", err)
		return nil
	}

	return &RabbitMQTopologyManager{
		config:     cfg,
		middleware: middleware,
	}
}

// SetUpTopologyFor creates the complete RabbitMQ topology for a client
// This includes: VHost, User, Client Queues, Permissions, Bridge Queues, and Shovels
func (tm *RabbitMQTopologyManager) SetUpTopologyFor(clientID string, password string) error {
	vhost := fmt.Sprintf("%s_vhost", clientID)
	username := fmt.Sprintf("%s_user", clientID)

	slog.Info("Setting up RabbitMQ topology for client",
		"client_id", clientID,
		"vhost", vhost,
		"username", username)

	// Step 1: Create VHost
	if err := tm.middleware.CreateVHost(vhost); err != nil {
		return fmt.Errorf("failed to create vhost: %w", err)
	}

	// Step 2: Create User
	if err := tm.middleware.CreateUser(username, password); err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	// Step 3: Create Client Queues (in client's VHost)
	dispatcherQueue := fmt.Sprintf("%s_dispatcher_queue", clientID)
	calibrationQueue := fmt.Sprintf("%s_calibration_queue", clientID)

	if err := tm.middleware.CreateQueueHTTP(vhost, dispatcherQueue); err != nil {
		return fmt.Errorf("failed to create dispatcher queue: %w", err)
	}

	if err := tm.middleware.CreateQueueHTTP(vhost, calibrationQueue); err != nil {
		return fmt.Errorf("failed to create calibration queue: %w", err)
	}

	// Step 4: Set Permissions (user can read from dispatcher_queue and write to calibration_queue)
	// Configure: allow access to all resources
	// Write: allow publishing to calibration queue and default exchange
	// Read: allow consuming from dispatcher queue
	configurePattern := ".*"
	writePattern := fmt.Sprintf("(%s|amq\\.default)", calibrationQueue)
	readPattern := fmt.Sprintf("(%s)", dispatcherQueue)

	if err := tm.middleware.SetPermissions(vhost, username, configurePattern, writePattern, readPattern); err != nil {
		return fmt.Errorf("failed to set permissions: %w", err)
	}

	// Step 5: Create Bridge Queues (in main VHost "/")
	// Using DeclareQueue since middleware is connected to VHost "/"
	bridgeToDispatcher := fmt.Sprintf("bridge_to_%s_dispatcher", clientID)
	bridgeFromCalibration := fmt.Sprintf("bridge_from_%s_calibration", clientID)

	if err := tm.middleware.DeclareQueue(bridgeToDispatcher, true); err != nil {
		return fmt.Errorf("failed to create bridge to dispatcher queue: %w", err)
	}

	if err := tm.middleware.DeclareQueue(bridgeFromCalibration, true); err != nil {
		return fmt.Errorf("failed to create bridge from calibration queue: %w", err)
	}

	// Step 6: Create Shovels
	// Shovel 1: Dispatcher -> Client (from "/" bridge to client vhost)
	shovel1Name := fmt.Sprintf("shovel_to_%s_dispatcher", clientID)
	if err := tm.middleware.CreateShovel(shovel1Name, "/", bridgeToDispatcher, vhost, dispatcherQueue); err != nil {
		return fmt.Errorf("failed to create shovel to dispatcher: %w", err)
	}

	// Shovel 2: Client -> Calibration (from client vhost to "/" bridge)
	shovel2Name := fmt.Sprintf("shovel_from_%s_calibration", clientID)
	if err := tm.middleware.CreateShovel(shovel2Name, vhost, calibrationQueue, "/", bridgeFromCalibration); err != nil {
		return fmt.Errorf("failed to create shovel from calibration: %w", err)
	}

	slog.Info("Successfully set up RabbitMQ topology for client", "client_id", clientID)
	return nil
}

// DeleteTopologyFor removes all RabbitMQ resources for a client (useful for cleanup)
func (tm *RabbitMQTopologyManager) DeleteTopologyFor(clientID string) error {
	vhost := fmt.Sprintf("%s_vhost", clientID)
	username := fmt.Sprintf("%s_user", clientID)

	slog.Info("Deleting RabbitMQ topology for client", "client_id", clientID)

	// Delete bridge queues from main vhost
	bridgeToDispatcher := fmt.Sprintf("bridge_to_%s_dispatcher", clientID)
	bridgeFromCalibration := fmt.Sprintf("bridge_from_%s_calibration", clientID)

	_ = tm.middleware.DeleteQueueHTTP("/", bridgeToDispatcher)
	_ = tm.middleware.DeleteQueueHTTP("/", bridgeFromCalibration)

	// Delete vhost (this will delete all queues and shovels in that vhost)
	if err := tm.middleware.DeleteVHost(vhost); err != nil {
		slog.Warn("Failed to delete vhost", "vhost", vhost, "error", err)
	}

	// Delete user
	if err := tm.middleware.DeleteUser(username); err != nil {
		slog.Warn("Failed to delete user", "username", username, "error", err)
	}

	slog.Info("Deleted RabbitMQ topology for client", "client_id", clientID)
	return nil
}

// Legacy helper function for backwards compatibility
// This can be used by existing code that expects rabbitmqctl commands
func executeRabbitMQCtlCommand(args ...string) error {
	cmd := exec.Command("rabbitmqctl", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("rabbitmqctl command failed: %s, output: %s", err, string(output))
	}
	return nil
}

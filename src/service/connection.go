package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"connection-service/src/config"
	"connection-service/src/middleware"
	"connection-service/src/models"
	"connection-service/src/repository"
)

type ConnectionService struct {
	Publisher         middleware.Publisher
	TopologyManager   *middleware.RabbitMQTopologyManager
	Config            *config.GlobalConfig
	SessionRepository *repository.SessionRepository
}

func NewConnectionService(publisher middleware.Publisher, topologyManager *middleware.RabbitMQTopologyManager, cfg *config.GlobalConfig, sessionRepo *repository.SessionRepository) *ConnectionService {
	return &ConnectionService{
		Publisher:         publisher,
		TopologyManager:   topologyManager,
		Config:            cfg,
		SessionRepository: sessionRepo,
	}
}

func (s *ConnectionService) NotifyNewConnection(UserID, sessionId, inputsFormat, outputsFormat, modelType string) error {

	exchangeName := config.CONNECTION_EXCHANGE

	notification := models.ConnectNotification{
		UserID:        UserID,
		SessionId:     sessionId,
		InputsFormat:  inputsFormat,
		OutputsFormat: outputsFormat,
		ModelType:     modelType,
	}
	body, err := json.Marshal(notification)
	if err != nil {
		return fmt.Errorf("failed to marshal notification: %w", err)
	}
	return s.Publisher.Publish(exchangeName, body)
}

// HandleClientConnection manages the entire client connection flow
func (s *ConnectionService) HandleClientConnection(ctx context.Context, UserID string, token string) (*models.ConnectResponse, error) {
	// Step 1: Validate Token
	if err := s.validateToken(token, UserID); err != nil {
		return nil, fmt.Errorf("token validation failed: %w", err)
	}

	// Step 2: Query Database for Active Session
	activeSession, err := s.SessionRepository.GetActiveSession(ctx, UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to query database: %w", err)
	}

	// Prepare credentials (deterministic naming)
	credentials := s.generateCredentials(UserID)

	// Step 3: Decide Flow based on Session Status
	if activeSession != nil {
		// CASE A: Active Session Found - Client is reconnecting
		slog.Info("Client reconnecting to existing session",
			"user_id", UserID,
			"session_id", activeSession.SessionID)

		// Get user data for inputs_format
		userData, err := s.getUserData(UserID)
		if err != nil {
			return nil, fmt.Errorf("failed to get user data: %w", err)
		}

		return &models.ConnectResponse{
			Status:       "success",
			Message:      "Client reconnected to existing session",
			Credentials:  credentials,
			InputsFormat: userData.InputsFormat,
		}, nil
	}

	// CASE B: No Active Session - New Client or Completed/Timeout Session
	slog.Info("Creating new session for client", "user_id", UserID)

	// Action 1: Create new session in database
	newSession, err := s.SessionRepository.CreateSession(ctx, UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	slog.Info("Created new session", "user_id", UserID, "session_id", newSession.SessionID)

	// Action 2: Set up RabbitMQ topology
	if err := s.TopologyManager.SetUpTopologyFor(UserID, credentials.Password); err != nil {
		slog.Error("Failed to setup RabbitMQ topology", "user_id", UserID, "error", err)
		return nil, fmt.Errorf("failed to setup RabbitMQ topology: %w", err)
	}

	// Action 3: Get user data and notify dispatcher service
	userData, err := s.getUserData(UserID)
	slog.Info("Fetched user data", "user_id", UserID, "user_data", userData)
	if err != nil {
		return nil, fmt.Errorf("failed to get user data: %w", err)
	}

	if err := s.NotifyNewConnection(userData.UserID, newSession.SessionID, userData.InputsFormat, userData.OutputsFormat, userData.ModelType); err != nil {
		return nil, fmt.Errorf("failed to notify new connection: %w", err)
	}

	// Action 4: Return success response with credentials
	return &models.ConnectResponse{
		Status:       "success",
		Message:      "Client connected successfully with new session",
		Credentials:  credentials,
		InputsFormat: userData.InputsFormat,
	}, nil
}

// generateCredentials creates RabbitMQ credentials for a client
func (s *ConnectionService) generateCredentials(UserID string) *models.RabbitMQCredentials {
	return &models.RabbitMQCredentials{
		Username: UserID,
		Password: "123",
		Host:     s.Config.GetRabbitMQHost(),
		Port:     s.Config.GetRabbitMQPort(),
	}
}

// validateToken validates a client token with the users-service
func (s *ConnectionService) validateToken(token, userID string) error {
	postBody, err := json.Marshal(map[string]string{"token": token, "user_id": userID})
	if err != nil {
		return fmt.Errorf("failed to marshal token validation request: %w", err)
	}

	resp, err := http.Post("http://users-service:8000/tokens/validate", "application/json", bytes.NewBuffer(postBody))
	if err != nil {
		return fmt.Errorf("failed to make request to users-service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("token validation failed with status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	var tokenResp models.TokenValidateResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return fmt.Errorf("failed to unmarshal token response: %w", err)
	}

	if !tokenResp.IsValid {
		return fmt.Errorf("invalid token")
	}

	return nil
}

// getUserData fetches user data from users-service
func (s *ConnectionService) getUserData(UserID string) (*models.GetUserDataResponse, error) {
	url := fmt.Sprintf("http://users-service:8000/users/%s", UserID)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to make request to users-service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("users-service returned status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var userData models.GetUserDataResponse
	if err := json.Unmarshal(body, &userData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal user data: %w", err)
	}

	return &userData, nil
}

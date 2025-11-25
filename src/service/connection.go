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
// Returns (response, statusCode, detail, error) for proper error propagation
func (s *ConnectionService) HandleClientConnection(ctx context.Context, UserID string, token string) (*models.ConnectResponse, *models.ServiceError, error) {
	// Step 1: Validate Connection
	serviceErr, err := s.validateConnection(token, UserID)
	if serviceErr != nil {
		return nil, serviceErr, nil
	}
	if err != nil {
		return nil, nil, fmt.Errorf("connection validation failed: %w", err)
	}

	// Step 2: Query Database for Active Session
	activeSession, err := s.SessionRepository.GetActiveSession(ctx, UserID)
	if err != nil {
		return nil, models.NewServiceError(http.StatusInternalServerError, "", "Failed to query database"), fmt.Errorf("failed to query database: %w", err)
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
			return nil, models.NewServiceError(http.StatusInternalServerError, "", "Failed to get user data"), fmt.Errorf("failed to get user data: %w", err)
		}

		return &models.ConnectResponse{
			Status:       "success",
			Message:      "Client reconnected to existing session",
			Credentials:  credentials,
			InputsFormat: userData.InputsFormat,
		}, nil, nil
	}

	// CASE B: No Active Session - New Client or Completed/Timeout Session
	slog.Info("Creating new session for client", "user_id", UserID)

	// Action 1: Create new session in database
	newSession, err := s.SessionRepository.CreateSession(ctx, UserID)
	if err != nil {
		return nil, models.NewServiceError(http.StatusInternalServerError, "", "Failed to create session"), fmt.Errorf("failed to create session: %w", err)
	}

	slog.Info("Created new session", "user_id", UserID, "session_id", newSession.SessionID)

	// Action 2: Set up RabbitMQ topology
	if err := s.TopologyManager.SetUpTopologyFor(UserID, credentials.Password); err != nil {
		slog.Error("Failed to setup RabbitMQ topology", "user_id", UserID, "error", err)
		return nil, models.NewServiceError(http.StatusInternalServerError, "", "Failed to setup RabbitMQ topology"), fmt.Errorf("failed to setup RabbitMQ topology: %w", err)
	}

	// Action 3: Get user data and notify dispatcher service
	userData, err := s.getUserData(UserID)
	slog.Info("Fetched user data", "user_id", UserID, "user_data", userData)
	if err != nil {
		return nil, models.NewServiceError(http.StatusInternalServerError, "", "Failed to get user data"), fmt.Errorf("failed to get user data: %w", err)
	}

	if err := s.NotifyNewConnection(userData.UserID, newSession.SessionID, userData.InputsFormat, userData.OutputsFormat, userData.ModelType); err != nil {
		return nil, models.NewServiceError(http.StatusInternalServerError, "", "Failed to notify new connection"), fmt.Errorf("failed to notify new connection: %w", err)
	}

	// Action 4: Return success response with credentials
	return &models.ConnectResponse{
		Status:       "success",
		Message:      "Client connected successfully with new session",
		Credentials:  credentials,
		InputsFormat: userData.InputsFormat,
	}, nil, nil
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

// validateConnection validates a client connection with the users-service
// Returns (*ServiceError, error) where ServiceError contains statusCode and detail from users-service
func (s *ConnectionService) validateConnection(token, userID string) (*models.ServiceError, error) {
	postBody, err := json.Marshal(map[string]string{"token": token, "user_id": userID})
	if err != nil {
		return models.NewServiceError(http.StatusInternalServerError, "", "Failed to marshal request"), fmt.Errorf("failed to marshal connection validation request: %w", err)
	}

	resp, err := http.Post("http://users-service:8000/sessions/validate-connection", "application/json", bytes.NewBuffer(postBody))
	if err != nil {
		return models.NewServiceError(http.StatusInternalServerError, "", "Failed to connect to users-service"), fmt.Errorf("failed to make request to users-service: %w", err)
	}
	defer resp.Body.Close()

	// If status is 200, connection is valid
	if resp.StatusCode == http.StatusOK {
		slog.Info("Connection validated successfully", "user_id", userID)
		return nil, nil
	}

	// For any non-200 status, read the body and extract detail
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Warn("Failed to read error response body", "error", err, "status", resp.StatusCode)
		return models.NewServiceError(resp.StatusCode, "", "Failed to read error response"), fmt.Errorf("failed to read error response body: %w", err)
	}

	// Parse JSON to extract "detail" field
	var errorResponse struct {
		Detail string `json:"detail"`
	}

	if err := json.Unmarshal(body, &errorResponse); err != nil {
		// If can't parse, use raw body as detail
		slog.Warn("Failed to parse error response", "error", err, "body", string(body))
		return models.NewServiceError(resp.StatusCode, string(body), "validation failed"), nil
	}

	// Log and return status code and detail from users-service
	slog.Warn("Connection validation failed",
		"user_id", userID,
		"status", resp.StatusCode,
		"detail", errorResponse.Detail)

	return models.NewServiceError(resp.StatusCode, errorResponse.Detail, "validation failed"), nil
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

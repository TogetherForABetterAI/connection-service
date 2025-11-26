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
	"connection-service/src/repository"
	"connection-service/src/schemas"
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

	notification := schemas.NotifyNewConnection{
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
// Returns (response, error) following idiomatic Go error handling
func (s *ConnectionService) HandleClientConnection(ctx context.Context, UserID string, token string) (*schemas.ConnectResponse, error) {
	// Step 1: Validate Connection y obtener datos del usuario
	userData, err := s.validateConnection(token, UserID)
	if err != nil {
		return nil, err
	}

	// Step 2: Query Database for Active Session
	activeSession, err := s.SessionRepository.GetActiveSession(ctx, UserID)
	if err != nil {
		return nil, schemas.NewInternalError(
			fmt.Sprintf("failed to query database: %v", err),
			"/users/connect",
		)
	}

	// Prepare credentials (deterministic naming)
	credentials := s.generateCredentials(UserID)

	// Step 3: Check Active Session
	if activeSession != nil {
		// CASE A: Active Session Found - Client is reconnecting
		slog.Info("Client reconnecting to existing session",
			"user_id", UserID,
			"session_id", activeSession.SessionID)

		return &schemas.ConnectResponse{
			Status:        "success",
			Message:       "Client reconnected to existing session",
			Credentials:   credentials,
			InputsFormat:  userData.InputsFormat,
			OutputsFormat: userData.OutputsFormat,
			ModelType:     userData.ModelType,
		}, nil
	}

	// CASE B: No Active Session - New Client Connection (New Session)
	slog.Info("Creating new session for client", "user_id", UserID)

	// Action 1: Create new session in database
	newSession, err := s.SessionRepository.CreateSession(ctx, UserID)
	if err != nil {
		return nil, schemas.NewInternalError(
			fmt.Sprintf("failed to create session: %v", err),
			"/users/connect",
		)
	}

	slog.Info("Created new session", "user_id", UserID, "session_id", newSession.SessionID)

	// Action 2: Set up RabbitMQ topology
	if err := s.TopologyManager.SetUpTopologyFor(UserID, credentials.Password); err != nil {
		slog.Error("Failed to setup RabbitMQ topology", "user_id", UserID, "error", err)
		return nil, schemas.NewInternalError(
			fmt.Sprintf("failed to setup RabbitMQ topology: %v", err),
			"/users/connect",
		)
	}

	// Action 3: Notificar dispatcher service usando userData
	slog.Info("Fetched user data", "user_id", UserID, "user_data", userData)

	if err := s.NotifyNewConnection(userData.UserID, newSession.SessionID, userData.InputsFormat, userData.OutputsFormat, userData.ModelType); err != nil {
		return nil, schemas.NewInternalError(
			fmt.Sprintf("failed to notify new connection: %v", err),
			"/users/connect",
		)
	}

	// Action 4: Return success response with credentials
	return &schemas.ConnectResponse{
		Status:        "success",
		Message:       "Client connected successfully with new session",
		Credentials:   credentials,
		InputsFormat:  userData.InputsFormat,
		OutputsFormat: userData.OutputsFormat,
		ModelType:     userData.ModelType,
	}, nil
}

// generateCredentials creates RabbitMQ credentials for a client
func (s *ConnectionService) generateCredentials(UserID string) *schemas.RabbitMQCredentials {
	return &schemas.RabbitMQCredentials{
		Username: UserID,
		Password: "123",
		Host:     s.Config.GetRabbitMQHost(),
		Port:     s.Config.GetRabbitMQPort(),
	}
}

// validateConnection validates a client connection with the users-service
// Ahora devuelve los datos del usuario si la validación es exitosa
func (s *ConnectionService) validateConnection(token, userID string) (*schemas.UserInfo, error) {
	postBody, err := json.Marshal(map[string]string{"token": token, "user_id": userID})
	if err != nil {
		return nil, schemas.NewInternalError(
			fmt.Sprintf("failed to marshal connection validation request: %v", err),
			"/users/connect",
		)
	}

	resp, err := http.Post("http://users-service:8000/sessions/validate-connection", "application/json", bytes.NewBuffer(postBody))
	if err != nil {
		// Network error - return 502 Bad Gateway
		return nil, schemas.NewBadGatewayError(
			fmt.Sprintf("failed to connect to users-service: %v", err),
			"/users/connect",
		)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, schemas.NewBadGatewayError(
			"failed to read response from users-service",
			"/users/connect",
		)
	}

	if resp.StatusCode == http.StatusOK {
		var userData schemas.UserInfo
		if err := json.Unmarshal(body, &userData); err != nil {
			return nil, schemas.NewBadGatewayError(
				fmt.Sprintf("failed to parse user data: %v", err),
				"/users/connect",
			)
		}
		slog.Info("Connection validated successfully", "user_id", userID)
		return &userData, nil
	}

	// Handle 4xx client errors - propagate them
	if resp.StatusCode >= 400 && resp.StatusCode < 500 {

		var legacyError struct {
			Detail string `json:"detail"`
		}
		if err := json.Unmarshal(body, &legacyError); err == nil && legacyError.Detail != "" {
			return nil, &schemas.ErrorResponse{
				Type:     "https://connection-service.com/external-service-error",
				Title:    "Validation Failed",
				Status:   resp.StatusCode,
				Detail:   legacyError.Detail,
				Instance: "/users/connect",
			}
		}

		// Could not decode - return a generic error with the status code
		slog.Warn("Connection validation failed", "user_id", userID, "status", resp.StatusCode, "body", string(body))
		return nil, &schemas.ErrorResponse{
			Type:     "https://connection-service.com/external-service-error",
			Title:    "Validation Failed",
			Status:   resp.StatusCode,
			Detail:   string(body),
			Instance: "/users/connect",
		}
	}

	// Handle 5xx server errors - return 502 Bad Gateway
	slog.Warn("Users-service returned server error", "status", resp.StatusCode, "body", string(body))
	return nil, schemas.NewBadGatewayError(
		fmt.Sprintf("users-service returned status %d: %s", resp.StatusCode, string(body)),
		"/users/connect",
	)
}

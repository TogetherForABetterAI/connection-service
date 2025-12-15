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

func (s *ConnectionService) NotifyNewConnection(UserID, sessionId, email, inputsFormat, outputsFormat, modelType string) error {

	exchangeName := config.CONNECTION_EXCHANGE

	notification := schemas.NotifyNewConnection{
		UserID:        UserID,
		SessionId:     sessionId,
		Email:         email,
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
	userData, tokenID, err := s.validateConnection(token, UserID)
	if err != nil {
		return nil, err
	}

	// Step 2: Query Database for Active Session
	activeSession, err := s.SessionRepository.GetActiveSession(ctx, UserID)
	if err != nil {
		return nil, schemas.NewInternalError(
			fmt.Sprintf("failed to query database: %v", err),
			"/sessions/start",
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
	newSession, err := s.SessionRepository.CreateSession(ctx, UserID, tokenID)
	if err != nil {
		return nil, schemas.NewInternalError(
			fmt.Sprintf("failed to create session: %v", err),
			"/sessions/start",
		)
	}

	slog.Info("Created new session", "user_id", UserID, "session_id", newSession.SessionID)

	// Action 2: Set up RabbitMQ topology
	if err := s.TopologyManager.SetUpTopologyFor(UserID, credentials.Password); err != nil {
		slog.Error("Failed to setup RabbitMQ topology", "user_id", UserID, "error", err)
		s.SessionRepository.DeleteSession(ctx, newSession.SessionID)
		return nil, schemas.NewInternalError(
			fmt.Sprintf("failed to setup RabbitMQ topology: %v", err),
			"/sessions/start",
		)
	}

	// Action 3: Notificar dispatcher service usando userData
	slog.Info("Fetched user data", "user_id", UserID, "user_data", userData)

	if err := s.NotifyNewConnection(userData.ID, newSession.SessionID, userData.Email, userData.InputsFormat, userData.OutputsFormat, userData.ModelType); err != nil {
		slog.Error("Failed to notify new connection", "user_id", UserID, "session_id", newSession.SessionID, "error", err)
		s.TopologyManager.DeleteTopologyFor(UserID)
		s.SessionRepository.DeleteSession(ctx, newSession.SessionID)
		return nil, schemas.NewInternalError(
			fmt.Sprintf("failed to notify new connection: %v", err),
			"/sessions/start",
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
		Host:     s.Config.GetRabbitPublicIp(),
		Port:     s.Config.GetRabbitMQPort(),
	}
}

// validateConnection validates a client connection with the users-service
// Ahora devuelve los datos del usuario si la validación es exitosa
func (s *ConnectionService) validateConnection(token, userID string) (*schemas.UserInfo, string, error) {

	if userID == "" || token == "" {
		return nil, "", &schemas.ErrorResponse{
			Type:     "https://connection-service.com/invalid-request",
			Title:    "Invalid Request",
			Status:   http.StatusBadRequest,
			Detail:   "UserID and Token must be provided",
			Instance: "/sessions/start",
		}
	}

	// User Validation
	userResp, err := http.Get(fmt.Sprintf("%s/users/%s", s.Config.GetUsersServiceURL(), userID))
	if err != nil {
		return nil, "", schemas.NewBadGatewayError(
			fmt.Sprintf("failed to connect to users-service: %v", err),
			"/sessions/start",
		)
	}
	defer userResp.Body.Close()

	if userResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(userResp.Body)
		slog.Warn("Failed to fetch user data", "user_id", userID, "status", userResp.StatusCode, "body", string(body))
		return nil, "", &schemas.ErrorResponse{
			Type:     "https://connection-service.com/external-service-error",
			Title:    "User Not Found",
			Status:   userResp.StatusCode,
			Detail:   string(body),
			Instance: "/sessions/start",
		}
	}
	userInfo := schemas.UserInfo{}
	if err := json.NewDecoder(userResp.Body).Decode(&userInfo); err != nil {
		return nil, "", schemas.NewInternalError(
			fmt.Sprintf("failed to decode user info response: %v", err),
			"/sessions/start",
		)
	}

	if !userInfo.IsAuthorized {
		return nil, "", &schemas.ErrorResponse{
			Type:     "https://connection-service.com/unauthorized",
			Title:    "User Not Authorized",
			Status:   http.StatusForbidden,
			Detail:   "User is not authorized to connect",
			Instance: "/sessions/start",
		}
	}

	// Token Validation
	postBody, err := json.Marshal(map[string]string{"token": token, "user_id": userID})
	if err != nil {
		return nil, "", schemas.NewInternalError(
			fmt.Sprintf("failed to marshal connection validation request: %v", err),
			"/sessions/start",
		)
	}

	validateTokenResp, err := http.Post(fmt.Sprintf("%s/tokens/validate", s.Config.GetUsersServiceURL()), "application/json", bytes.NewBuffer(postBody))
	if err != nil {
		return nil, "", schemas.NewBadGatewayError(
			fmt.Sprintf("failed to connect to users-service: %v", err),
			"/sessions/start",
		)
	}
	defer validateTokenResp.Body.Close()

	body, err := io.ReadAll(validateTokenResp.Body)
	if err != nil {
		return nil, "", schemas.NewBadGatewayError(
			"failed to read response from users-service",
			"/sessions/start",
		)
	}
	if validateTokenResp.StatusCode >= 500 {
		// Handle 5xx server errors - return 502 Bad Gateway
		slog.Warn("Users-service returned server error", "status", validateTokenResp.StatusCode, "body", string(body))
		return nil, "", schemas.NewBadGatewayError(
			fmt.Sprintf("users-service returned status %d: %s", validateTokenResp.StatusCode, string(body)),
			"/sessions/start",
		)
	}

	// Handle 4xx client errors - propagate them
	if validateTokenResp.StatusCode >= 400 {

		var legacyError struct {
			Detail string `json:"detail"`
		}
		if err := json.Unmarshal(body, &legacyError); err == nil && legacyError.Detail != "" {
			return nil, "", &schemas.ErrorResponse{
				Type:     "https://connection-service.com/external-service-error",
				Title:    "Validation Failed",
				Status:   validateTokenResp.StatusCode,
				Detail:   legacyError.Detail,
				Instance: "/sessions/start",
			}
		}

		// Could not decode - return a generic error with the status code
		slog.Warn("Connection validation failed", "user_id", userID, "status", validateTokenResp.StatusCode, "body", string(body))
		return nil, "", &schemas.ErrorResponse{
			Type:     "https://connection-service.com/external-service-error",
			Title:    "Validation Failed",
			Status:   validateTokenResp.StatusCode,
			Detail:   string(body),
			Instance: "/sessions/start",
		}
	}
	tokenInfo := schemas.TokenInfo{}
	if err := json.NewDecoder(bytes.NewReader(body)).Decode(&tokenInfo); err != nil {
		return nil, "", schemas.NewInternalError(
			fmt.Sprintf("failed to decode token info response: %v", err),
			"/sessions/start",
		)
	}

	return &userInfo, tokenInfo.TokenID, nil
}

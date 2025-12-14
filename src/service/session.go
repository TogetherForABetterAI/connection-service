package service

import (
	"connection-service/src/config"
	"connection-service/src/middleware"
	"connection-service/src/models"
	"connection-service/src/repository"
	"connection-service/src/schemas"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type SessionService struct {
	repo   *repository.SessionRepository
	tm     *middleware.RabbitMQTopologyManager
	config *config.GlobalConfig
}

func NewSessionService(repo *repository.SessionRepository, tm *middleware.RabbitMQTopologyManager, cfg *config.GlobalConfig) *SessionService {
	return &SessionService{
		repo:   repo,
		tm:     tm,
		config: cfg,
	}
}

// SetSessionStatusToCompleted sets the session status to COMPLETED and revokes user authorization
func (s *SessionService) SetSessionStatusToCompleted(ctx context.Context, sessionID string) error {
	// Check if session exists and is in progress
	session, err := s.repo.GetSessionByID(ctx, sessionID)
	if err != nil {
		if errors.Is(err, models.ErrSessionNotFound) {
			return schemas.NewNotFoundError(
				fmt.Sprintf("session with ID %s not found", sessionID),
				"/sessions/"+sessionID+"/status/completed",
			)
		}
		return schemas.NewInternalError(
			fmt.Sprintf("failed to get session status: %v", err),
			"/sessions/"+sessionID+"/status/completed",
		)
	}

	// Check if session is in progress
	if session.SessionStatus != models.StatusInProgress {
		return schemas.SessionNotInProgressError(
			"cannot update status: session is not IN_PROGRESS",
			"/sessions/"+sessionID+"/status/completed",
		)
	}

	// Update session status to COMPLETED and set completed_at
	err = s.repo.SetSessionStatusToCompleted(ctx, sessionID)
	if err != nil {
		if errors.Is(err, models.ErrSessionNotFound) {
			return schemas.NewNotFoundError(
				fmt.Sprintf("session with ID %s not found", sessionID),
				"/sessions/"+sessionID+"/status/completed",
			)
		}
		return schemas.NewInternalError(
			fmt.Sprintf("failed to update session status to COMPLETED: %v", err),
			"/sessions/"+sessionID+"/status/completed",
		)
	}

	// Revoke user authorization
	if err := s.RevokeAuthorization(session.UserID, sessionID); err != nil {
		return err
	}

	if err := s.RevokeToken(session.UserID, session.TokenID); err != nil {
		return err
	}

	s.tm.DeleteTopologyFor(session.UserID)

	return nil
}

// SetSessionStatusToTimeout sets the session status to TIMEOUT
func (s *SessionService) SetSessionStatusToTimeout(ctx context.Context, sessionID string) error {
	// Check if session exists and is in progress
	session, err := s.repo.GetSessionByID(ctx, sessionID)
	if err != nil {
		// Translate repository errors to schema errors
		if errors.Is(err, models.ErrSessionNotFound) {
			return schemas.NewNotFoundError(
				fmt.Sprintf("session with ID %s not found", sessionID),
				"/sessions/"+sessionID+"/status/timeout",
			)
		}
		return schemas.NewInternalError(
			fmt.Sprintf("failed to get session status: %v", err),
			"/sessions/"+sessionID+"/status/timeout",
		)
	}

	// Check if session is in progress
	if session.SessionStatus != models.StatusInProgress {
		return schemas.SessionNotInProgressError(
			"cannot update status: session is not IN_PROGRESS",
			"/sessions/"+sessionID+"/status/timeout",
		)
	}

	// Update session status to TIMEOUT
	err = s.repo.UpdateSessionStatus(ctx, sessionID, models.StatusTimeout)
	if err != nil {
		if errors.Is(err, models.ErrSessionNotFound) {
			return schemas.NewNotFoundError(
				fmt.Sprintf("session with ID %s not found", sessionID),
				"/sessions/"+sessionID+"/status/timeout",
			)
		}
		return schemas.NewInternalError(
			fmt.Sprintf("failed to update session status to TIMEOUT: %v", err),
			"/sessions/"+sessionID+"/status/timeout",
		)
	}

	s.tm.DeleteTopologyFor(session.UserID)

	return nil
}


func (s *SessionService) RevokeToken(userID, tokenID string) error {
	url := fmt.Sprintf("%s/tokens/revoke/%s?user_id=%s", s.config.GetUsersServiceURL(), tokenID, userID)
	instance := "/tokens/revoke/" + tokenID

	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		// Network error - return 502 Bad Gateway
		return schemas.NewBadGatewayError(
			fmt.Sprintf("failed to connect to users-service: %v", err),
			instance,
		)
	}

	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body for error details
	bodyBytes, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return schemas.NewBadGatewayError(
			"failed to read response from users-service",
			instance,
		)
	}

	// Success case
	if resp.StatusCode == http.StatusNoContent {
		return nil
	}

	// Handle 4xx client errors - propagate them
	if resp.StatusCode >= 400 && resp.StatusCode < 500 {
		// Try to decode the error response from the remote service
		var remoteError schemas.ErrorResponse
		if err := json.Unmarshal(bodyBytes, &remoteError); err == nil {
			// Successfully decoded remote error - propagate it
			return &remoteError
		}
		// Could not decode - return a generic error with the status code
		return &schemas.ErrorResponse{
			Type:     "https://connection-service.com/external-service-error",
			Title:    "External Service Error",
			Status:   resp.StatusCode,
			Detail:   fmt.Sprintf("users-service returned error: %s", string(bodyBytes)),
			Instance: instance,
		}
	}

	// Handle 5xx server errors - return 502 Bad Gateway
	return schemas.NewBadGatewayError(
		fmt.Sprintf("users-service returned status %d: %s", resp.StatusCode, string(bodyBytes)),
		instance,
	)
}

// RevokeAuthorization revokes user authorization in users-service
// Implements smart propagation: 4xx errors are propagated, 5xx/network errors return 502
func (s *SessionService) RevokeAuthorization(userID, sessionID string) error {
	url := fmt.Sprintf("%s/users/%s/status", s.config.GetUsersServiceURL(), userID)
	body := `{"is_authorized": false}`
	instance := "/sessions/" + sessionID + "/status/completed"

	resp, err := httpPatchJSON(url, body)
	if err != nil {
		// Network error - return 502 Bad Gateway
		return schemas.NewBadGatewayError(
			fmt.Sprintf("failed to connect to users-service: %v", err),
			instance,
		)
	}
	defer resp.Body.Close()

	// Read response body for error details
	bodyBytes, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return schemas.NewBadGatewayError(
			"failed to read response from users-service",
			instance,
		)
	}

	// Success case
	if resp.StatusCode == http.StatusOK {
		return nil
	}

	// Handle 4xx client errors - propagate them
	if resp.StatusCode >= 400 && resp.StatusCode < 500 {
		// Try to decode the error response from the remote service
		var remoteError schemas.ErrorResponse
		if err := json.Unmarshal(bodyBytes, &remoteError); err == nil {
			// Successfully decoded remote error - propagate it
			return &remoteError
		}
		// Could not decode - return a generic error with the status code
		return &schemas.ErrorResponse{
			Type:     "https://connection-service.com/external-service-error",
			Title:    "External Service Error",
			Status:   resp.StatusCode,
			Detail:   fmt.Sprintf("users-service returned error: %s", string(bodyBytes)),
			Instance: instance,
		}
	}

	// Handle 5xx server errors - return 502 Bad Gateway
	return schemas.NewBadGatewayError(
		fmt.Sprintf("users-service returned status %d: %s", resp.StatusCode, string(bodyBytes)),
		instance,
	)
}

// httpPatchJSON performs a PATCH request with JSON body and returns the response
func httpPatchJSON(url, body string) (*http.Response, error) {
	req, err := http.NewRequest("PATCH", url, strings.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	return client.Do(req)
}

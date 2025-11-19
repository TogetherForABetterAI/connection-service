package service

import (
	"connection-service/src/repository"
	"context"
	"fmt"
)

type SessionService struct {
	repo *repository.SessionRepository
}

func NewSessionService(repo *repository.SessionRepository) *SessionService {
	return &SessionService{
		repo: repo,
	}
}

// UpdateSessionStatus updates the status of a session
func (s *SessionService) UpdateSessionStatus(ctx context.Context, sessionID string, status string) error {
	// Validate the status
	var sessionStatus repository.SessionStatus
	switch status {
	case "COMPLETED":
		sessionStatus = repository.StatusCompleted
	case "TIMEOUT":
		sessionStatus = repository.StatusTimeout
	case "IN_PROGRESS":
		sessionStatus = repository.StatusInProgress
	default:
		return fmt.Errorf("invalid session status: %s. Valid values are: IN_PROGRESS, COMPLETED, TIMEOUT", status)
	}

	// Update the session status in the repository
	err := s.repo.UpdateSessionStatus(ctx, sessionID, sessionStatus)
	if err != nil {
		return fmt.Errorf("failed to update session status: %w", err)
	}

	return nil
}



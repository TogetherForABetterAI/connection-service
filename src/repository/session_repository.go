package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"connection-service/src/db"
	"connection-service/src/models"

	"github.com/google/uuid"
)

// SessionRepository handles all database operations for sessions
type SessionRepository struct {
	db *db.DB
}

// NewSessionRepository creates a new session repository
func NewSessionRepository(database *db.DB) *SessionRepository {
	return &SessionRepository{
		db: database,
	}
}

func (r *SessionRepository) GetSessionByID(ctx context.Context, sessionID string) (*models.Session, error) {
	query := `
		SELECT session_id, user_id, token_id, session_status, dispatcher_status, 
		       created_at, completed_at
		FROM client_sessions
		WHERE session_id = $1
	`

	var session models.Session
	err := r.db.GetConnection().QueryRowContext(ctx, query, sessionID).Scan(
		&session.SessionID,
		&session.UserID,
		&session.TokenID,
		&session.SessionStatus,
		&session.DispatcherStatus,
		&session.CreatedAt,
		&session.CompletedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("session with ID %s: %w", sessionID, models.ErrSessionNotFound)
		}
		return nil, fmt.Errorf("failed to get session by ID: %w", err)
	}

	return &session, nil
}

// GetActiveSession retrieves an active session for a given User ID
func (r *SessionRepository) GetActiveSession(ctx context.Context, UserID string) (*models.Session, error) {
	query := `
		SELECT session_id, user_id, token_id, session_status, dispatcher_status, 
		       created_at, completed_at
		FROM client_sessions
		WHERE user_id = $1 AND session_status = $2
		ORDER BY created_at DESC
		LIMIT 1
	`

	var session models.Session
	err := r.db.GetConnection().QueryRowContext(ctx, query, UserID, models.StatusInProgress).Scan(
		&session.SessionID,
		&session.UserID,
		&session.SessionStatus,
		&session.DispatcherStatus,
		&session.CreatedAt,
		&session.CompletedAt,
	)

	if err == sql.ErrNoRows {
		// No active session found - this is not an error, just means no session exists
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get active session: %w", err)
	}

	slog.Info("Found active session",
		"user_id", UserID,
		"session_id", session.SessionID)

	return &session, nil
}

// CreateSession creates a new session for a client
func (r *SessionRepository) CreateSession(ctx context.Context, UserID string, tokenID string) (*models.Session, error) {
	sessionID := uuid.New().String()
	now := time.Now()

	query := `
		INSERT INTO client_sessions 
		(session_id, user_id, token_id, session_status, dispatcher_status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING session_id, user_id, token_id, session_status, dispatcher_status, 
		          created_at, completed_at
	`

	var session models.Session
	err := r.db.GetConnection().QueryRowContext(
		ctx,
		query,
		sessionID,
		UserID,
		tokenID,
		models.StatusInProgress,
		"PENDING", // dispatcher_status
		now,       // created_at
	).Scan(
		&session.SessionID,
		&session.UserID,
		&session.TokenID,
		&session.SessionStatus,
		&session.DispatcherStatus,
		&session.CreatedAt,
		&session.CompletedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	slog.Info("Created new session",
		"user_id", UserID,
		"session_id", session.SessionID)

	return &session, nil
}

// UpdateSessionStatus updates the status of a session
func (r *SessionRepository) UpdateSessionStatus(ctx context.Context, sessionID string, status models.SessionStatus) error {
	query := `
		UPDATE client_sessions
		SET session_status = $1
		WHERE session_id = $2
	`

	result, err := r.db.GetConnection().ExecContext(ctx, query, status, sessionID)
	if err != nil {
		return fmt.Errorf("failed to update session status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("update session %s: %w", sessionID, models.ErrSessionNotFound)
	}

	slog.Info("Updated session status",
		"session_id", sessionID,
		"status", status)

	return nil
}

// UpdateDispatcherStatus updates the dispatcher status of a session
func (r *SessionRepository) UpdateDispatcherStatus(ctx context.Context, sessionID string, status string) error {
	query := `
		UPDATE client_sessions
		SET dispatcher_status = $1
		WHERE session_id = $2
	`

	result, err := r.db.GetConnection().ExecContext(ctx, query, status, sessionID)
	if err != nil {
		return fmt.Errorf("failed to update dispatcher status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("update dispatcher status for session %s: %w", sessionID, models.ErrSessionNotFound)
	}

	return nil
}

// GetSessionStatus retrieves the session status for the given session ID
// Returns the session status and error if not found
func (r *SessionRepository) GetSessionStatus(ctx context.Context, sessionID string) (models.SessionStatus, error) {
	query := `SELECT session_status FROM client_sessions WHERE session_id = $1`
	var status models.SessionStatus
	err := r.db.GetConnection().QueryRowContext(ctx, query, sessionID).Scan(&status)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", fmt.Errorf("get status for session %s: %w", sessionID, models.ErrSessionNotFound)
		}
		return "", fmt.Errorf("failed to get session status: %w", err)
	}
	return status, nil
}

// DeleteSession deletes a session by session ID
// Just used in case of rollback during connection setup
func (r *SessionRepository) DeleteSession(ctx context.Context, sessionID string) error {
	query := `DELETE FROM client_sessions WHERE session_id = $1`

	result, err := r.db.GetConnection().ExecContext(ctx, query, sessionID)
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		slog.Warn("Attempted to delete non-existent session", "session_id", sessionID)
		return nil // Don't fail if session doesn't exist (already deleted or never created)
	}

	slog.Info("Deleted session", "session_id", sessionID)
	return nil
}

package repository

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	"connection-service/src/db"

	"github.com/google/uuid"
)

// SessionStatus represents the status of a client session
type SessionStatus string

const (
	StatusInProgress SessionStatus = "IN_PROGRESS"
	StatusCompleted  SessionStatus = "COMPLETED"
	StatusTimeout    SessionStatus = "TIMEOUT"
)

// Session represents a client session in the database
type Session struct {
	SessionID        string        `json:"session_id"`
	UserID           string        `json:"user_id"`
	SessionStatus    SessionStatus `json:"session_status"`
	DispatcherStatus string        `json:"dispatcher_status"`
	CreatedAt        time.Time     `json:"created_at"`
	CompletedAt      *time.Time    `json:"completed_at,omitempty"`
}

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

// GetActiveSession retrieves an active session for a given User ID
func (r *SessionRepository) GetActiveSession(ctx context.Context, UserID string) (*Session, error) {
	query := `
		SELECT session_id, user_id, session_status, dispatcher_status, 
		       created_at, completed_at
		FROM client_sessions
		WHERE user_id = $1 AND session_status = $2
		ORDER BY created_at DESC
		LIMIT 1
	`

	var session Session
	err := r.db.GetConnection().QueryRowContext(ctx, query, UserID, StatusInProgress).Scan(
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
func (r *SessionRepository) CreateSession(ctx context.Context, UserID string) (*Session, error) {
	sessionID := uuid.New().String()
	now := time.Now()

	query := `
		INSERT INTO client_sessions 
		(session_id, user_id, session_status, dispatcher_status, created_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING session_id, user_id, session_status, dispatcher_status, 
		          created_at, completed_at
	`

	var session Session
	err := r.db.GetConnection().QueryRowContext(
		ctx,
		query,
		sessionID,
		UserID,
		StatusInProgress,
		"PENDING", // dispatcher_status
		now,       // created_at
	).Scan(
		&session.SessionID,
		&session.UserID,
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
func (r *SessionRepository) UpdateSessionStatus(ctx context.Context, sessionID string, status SessionStatus) error {
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
		return fmt.Errorf("session not found: %s", sessionID)
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
		return fmt.Errorf("session not found: %s", sessionID)
	}

	return nil
}

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
	SessionID         string        `json:"session_id"`
	ClientID          string        `json:"client_id"`
	SessionStatus     SessionStatus `json:"session_status"`
	DispatcherStatus  string        `json:"dispatcher_status"`
	CalibrationStatus string        `json:"calibration_status"`
	VecScores         []byte        `json:"vec_scores"`
	CreatedAt         time.Time     `json:"created_at"`
	CompletedAt       *time.Time    `json:"completed_at,omitempty"`
	LastActivityAt    time.Time     `json:"last_activity_at"`
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

// GetActiveSession retrieves an active session for a given client ID
func (r *SessionRepository) GetActiveSession(ctx context.Context, clientID string) (*Session, error) {
	query := `
		SELECT session_id, client_id, session_status, dispatcher_status, 
		       calibration_status, vec_scores, created_at, completed_at, last_activity_at
		FROM client_sessions
		WHERE client_id = $1 AND session_status = $2
		ORDER BY created_at DESC
		LIMIT 1
	`

	var session Session
	err := r.db.GetConnection().QueryRowContext(ctx, query, clientID, StatusInProgress).Scan(
		&session.SessionID,
		&session.ClientID,
		&session.SessionStatus,
		&session.DispatcherStatus,
		&session.CalibrationStatus,
		&session.VecScores,
		&session.CreatedAt,
		&session.CompletedAt,
		&session.LastActivityAt,
	)

	if err == sql.ErrNoRows {
		// No active session found - this is not an error, just means no session exists
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get active session: %w", err)
	}

	slog.Info("Found active session",
		"client_id", clientID,
		"session_id", session.SessionID)

	return &session, nil
}

// CreateSession creates a new session for a client
func (r *SessionRepository) CreateSession(ctx context.Context, clientID string) (*Session, error) {
	sessionID := uuid.New().String()
	now := time.Now()

	query := `
		INSERT INTO client_sessions 
		(session_id, client_id, session_status, dispatcher_status, calibration_status, 
		 created_at, last_activity_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING session_id, client_id, session_status, dispatcher_status, 
		          calibration_status, vec_scores, created_at, completed_at, last_activity_at
	`

	var session Session
	err := r.db.GetConnection().QueryRowContext(
		ctx,
		query,
		sessionID,
		clientID,
		StatusInProgress,
		"PENDING", // dispatcher_status
		"PENDING", // calibration_status
		now,       // created_at
		now,       // last_activity_at
	).Scan(
		&session.SessionID,
		&session.ClientID,
		&session.SessionStatus,
		&session.DispatcherStatus,
		&session.CalibrationStatus,
		&session.VecScores,
		&session.CreatedAt,
		&session.CompletedAt,
		&session.LastActivityAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	slog.Info("Created new session",
		"client_id", clientID,
		"session_id", session.SessionID)

	return &session, nil
}

// UpdateSessionStatus updates the status of a session
func (r *SessionRepository) UpdateSessionStatus(ctx context.Context, sessionID string, status SessionStatus) error {
	query := `
		UPDATE client_sessions
		SET session_status = $1, last_activity_at = $2
		WHERE session_id = $3
	`

	result, err := r.db.GetConnection().ExecContext(ctx, query, status, time.Now(), sessionID)
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
		SET dispatcher_status = $1, last_activity_at = $2
		WHERE session_id = $3
	`

	result, err := r.db.GetConnection().ExecContext(ctx, query, status, time.Now(), sessionID)
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

// UpdateCalibrationStatus updates the calibration status of a session
func (r *SessionRepository) UpdateCalibrationStatus(ctx context.Context, sessionID string, status string) error {
	query := `
		UPDATE client_sessions
		SET calibration_status = $1, last_activity_at = $2
		WHERE session_id = $3
	`

	result, err := r.db.GetConnection().ExecContext(ctx, query, status, time.Now(), sessionID)
	if err != nil {
		return fmt.Errorf("failed to update calibration status: %w", err)
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

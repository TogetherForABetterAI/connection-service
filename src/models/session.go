package models

import "time"

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
	TokenID          string        `json:"token_id"`
	SessionStatus    SessionStatus `json:"session_status"`
	DispatcherStatus string        `json:"dispatcher_status"`
	CreatedAt        time.Time     `json:"created_at"`
	CompletedAt      *time.Time    `json:"completed_at,omitempty"`
}

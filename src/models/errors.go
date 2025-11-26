package models

import "errors"

// Domain-level sentinel errors for business logic
// These errors should not contain HTTP-specific information

var (
	// ErrSessionNotFound indicates that a session with the given ID does not exist
	ErrSessionNotFound = errors.New("session not found")

	// ErrSessionNotInProgress indicates that the session is not in IN_PROGRESS status
	ErrSessionNotInProgress = errors.New("session is not in progress")

	// ErrActiveSessionExists indicates that an active session already exists for the user
	ErrActiveSessionExists = errors.New("active session already exists")

	// ErrSessionAlreadyCompleted indicates that the session has already been completed
	ErrSessionAlreadyCompleted = errors.New("session already completed")

	// ErrInvalidSessionStatus indicates that the session status is invalid
	ErrInvalidSessionStatus = errors.New("invalid session status")
)

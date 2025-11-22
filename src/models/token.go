package models

import (
	"time"

	"github.com/google/uuid"
)

// TokenValidateResponse represents the response from token validation
type TokenValidateResponse struct {
	IsValid       bool       `json:"is_valid"`
	TokenID       *uuid.UUID `json:"token_id,omitempty"`
	ExpiresAt     *time.Time `json:"expires_at,omitempty"`
	UsageCount    *int       `json:"usage_count,omitempty"`
	MaxUses       *int       `json:"max_uses,omitempty"`
	UsesRemaining *int       `json:"uses_remaining,omitempty"`
}

// TokenCreateRequest represents a request to create a new token
type TokenCreateRequest struct {
	ClientId string `json:"client_id"`
}

// TokenCreateResponse represents the response after creating a token
type TokenCreateResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}

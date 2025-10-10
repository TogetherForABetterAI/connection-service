package models

import (
	"time"

	"github.com/google/uuid"
)

// APIError represents an error response in the API, according to RFC 7807.
type APIError struct {
	Type     string `json:"type"`
	Title    string `json:"title"`
	Status   int    `json:"status"`
	Detail   string `json:"detail"`
	Instance string `json:"instance"`
}

// ConnectRequest represents the body of a request to create a snap.
type ConnectRequest struct {
	ClientId string `json:"client_id"`
	Token    string `json:"token"`
}

type ConnectResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

// class TokenValidateResponse(BaseModel):
//     """Schema for token validation response"""

//     is_valid: bool
//     token_id: Optional[uuid.UUID] = None
//     expires_at: Optional[datetime] = None
//     usage_count: Optional[int] = None
//     max_uses: Optional[int] = None
//     uses_remaining: Optional[int] = None

type TokenValidateResponse struct {
	IsValid       bool       `json:"is_valid"`
	TokenID       *uuid.UUID `json:"token_id,omitempty"`
	ExpiresAt     *time.Time `json:"expires_at,omitempty"`
	UsageCount    *int       `json:"usage_count,omitempty"`
	MaxUses       *int       `json:"max_uses,omitempty"`
	UsesRemaining *int       `json:"uses_remaining,omitempty"`
}

type TokenCreateRequest struct {
	ClientId string `json:"client_id"`
}

type TokenCreateResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}

type UserCreateRequest struct {
	Username      string `json:"username"`
	Email         string `json:"email"`
	ModelType     string `json:"model_type"`
	InputsFormat  string `json:"inputs_format"`
	OutputsFormat string `json:"outputs_format"`
}

type UserInfo struct {
	ClientId      string    `json:"client_id"`
	Username      string    `json:"username"`
	Email         string    `json:"email"`
	ModelType     string    `json:"model_type"`
	InputsFormat  string    `json:"inputs_format"`
	OutputsFormat string    `json:"outputs_format"`
	CreatedAt     time.Time `json:"created_at"`
}

type UserCreateResponse struct {
	ClientId string `json:"client_id"`
}

type ConnectNotification struct {
	ClientId      string `json:"client_id"`
	InputsFormat  string `json:"inputs_format"`
	OutputsFormat string `json:"outputs_format"`
	ModelType     string `json:"model_type"`
}

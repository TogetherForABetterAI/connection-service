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
	UserId string `json:"user_id"`
	Token  string `json:"token"`
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
	Username string `json:"username"`
}

type TokenCreateResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}

type TokenInfo struct {
	Id         uuid.UUID `json:"id"`
	UserId     string    `json:"user_id"`
	Username   string    `json:"username"`
	Token      string    `json:"token_hash"`
	CreatedAt  time.Time `json:"created_at"`
	ExpiresAt  time.Time `json:"expires_at"`
	UsageCount int       `json:"usage_count"`
	MaxUses    int       `json:"max_uses"`
	IsActive   bool      `json:"is_active"`
}

type UserCreateRequest struct {
	Username      string `json:"username"`
	Email         string `json:"email"`
	ModelType     string `json:"model_type"`
	InputsFormat  string `json:"inputs_format"`
	OutputsFormat string `json:"outputs_format"`
}

type UserInfo struct {
	UserID        string    `json:"user_id"`
	Username      string    `json:"username"`
	Email         string    `json:"email"`
	ModelType     string    `json:"model_type"`
	InputsFormat  string    `json:"inputs_format"`
	OutputsFormat string    `json:"outputs_format"`
	CreatedAt     time.Time `json:"created_at"`
}

type UserCreateResponse struct {
	Id string `json:"id"`
}

type AdminAuth struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AdminInfo struct {
	Id        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
}

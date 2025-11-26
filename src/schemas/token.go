package schemas

type TokenValidationResponse struct {
	TokenID       string `json:"token_id"`
	IsValid       bool   `json:"is_valid"`
	ExpiresAt     string `json:"expires_at"`
	UsageCount    int    `json:"usage_count"`
	MaxUses       int    `json:"max_uses"`
	UsesRemaining int    `json:"uses_remaining"`
}

// ...existing code from models/token.go...

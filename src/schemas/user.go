package schemas

// ...existing code from models/user.go...
import "time"

// UserCreateRequest represents a request to create a new user
type UserCreateRequest struct {
	Username      string `json:"username"`
	Email         string `json:"email"`
	ModelType     string `json:"model_type"`
	InputsFormat  string `json:"inputs_format"`
	OutputsFormat string `json:"outputs_format"`
}

// UserCreateResponse represents the response after creating a user
type UserCreateResponse struct {
	UserID string `json:"user_id"`
}

// UserInfo represents user information
type UserInfo struct {
	UserID        string    `json:"user_id"`
	Username      string    `json:"username"`
	Email         string    `json:"email"`
	ModelType     string    `json:"model_type"`
	InputsFormat  string    `json:"inputs_format"`
	OutputsFormat string    `json:"outputs_format"`
	CreatedAt     time.Time `json:"created_at"`
}


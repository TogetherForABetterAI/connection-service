package schemas

// ...existing code from models/user.go...
import "time"

// UserInfo represents user information
type UserInfo struct {
	ID            string    `json:"id"`
	Username      string    `json:"username"`
	Email         string    `json:"email"`
	ModelType     string    `json:"model_type"`
	InputsFormat  string    `json:"inputs_format"`
	OutputsFormat string    `json:"outputs_format"`
	IsAuthorized  bool      `json:"is_authorized"`
	CreatedAt     time.Time `json:"created_at"`
}

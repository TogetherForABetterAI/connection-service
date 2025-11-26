package schemas

// UpdateSessionStatusRequest represents the request body for updating session status
type UpdateSessionStatusRequest struct {
	SessionID string `json:"session_id" binding:"required"`
	Status    string `json:"status" binding:"required"`
}

// UpdateSessionStatusResponse represents the response for updating session status
type UpdateSessionStatusResponse struct {
	Message   string `json:"message"`
	SessionID string `json:"session_id"`
	Status    string `json:"status"`
}

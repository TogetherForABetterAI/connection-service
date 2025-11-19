package controller

import (
	"connection-service/src/service"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

type SessionController struct {
	Service *service.SessionService
}

func NewSessionController(service *service.SessionService) *SessionController {
	return &SessionController{
		Service: service,
	}
}

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

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

// @Summary Update session status
// @Description Updates the status of a client session to COMPLETED or TIMEOUT
// @Tags sessions
// @Accept json
// @Produce json
// @Param UpdateSessionStatusRequest body UpdateSessionStatusRequest true "Update Session Status Request"
// @Success 200 {object} UpdateSessionStatusResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /sessions/status [put]
func (sc *SessionController) UpdateSessionStatus(ctx *gin.Context) {
	var req UpdateSessionStatusRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		slog.Error("Invalid request body", "error", err.Error())
		ctx.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Bad Request",
			Message: "Invalid JSON format: " + err.Error(),
		})
		return
	}

	// Update session status
	err := sc.Service.UpdateSessionStatus(ctx.Request.Context(), req.SessionID, req.Status)
	if err != nil {
		slog.Error("Failed to update session status",
			"session_id", req.SessionID,
			"status", req.Status,
			"error", err.Error())

		ctx.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Internal Server Error",
			Message: err.Error(),
		})
		return
	}

	slog.Info("Session status updated successfully",
		"session_id", req.SessionID,
		"status", req.Status)

	ctx.JSON(http.StatusOK, UpdateSessionStatusResponse{
		Message:   "Session status updated successfully",
		SessionID: req.SessionID,
		Status:    req.Status,
	})
}



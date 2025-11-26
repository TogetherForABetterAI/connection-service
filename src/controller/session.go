package controller

import (
	"errors"
	"net/http"

	"connection-service/src/schemas"
	"connection-service/src/service"

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

// SetSessionStatusToCompleted sets the session status to COMPLETED
func (sc *SessionController) SetSessionStatusToCompleted(ctx *gin.Context) {
	sessionID := ctx.Param("session_id")

	err := sc.Service.SetSessionStatusToCompleted(ctx.Request.Context(), sessionID)
	if err != nil {
		var apiError *schemas.ErrorResponse
		if errors.As(err, &apiError) {
			ctx.JSON(apiError.Status, apiError)
			return
		}
		ctx.JSON(http.StatusInternalServerError, schemas.NewInternalError(
			err.Error(),
			"/sessions/"+sessionID+"/status/completed",
		))
		return
	}

	ctx.JSON(http.StatusOK, schemas.UpdateSessionStatusResponse{
		Message:   "Session status updated to COMPLETED",
		SessionID: sessionID,
		Status:    "COMPLETED",
	})
}

// SetSessionStatusToTimeout sets the session status to TIMEOUT
func (sc *SessionController) SetSessionStatusToTimeout(ctx *gin.Context) {
	sessionID := ctx.Param("session_id")

	err := sc.Service.SetSessionStatusToTimeout(ctx.Request.Context(), sessionID)
	if err != nil {
		// Check if the error is an ErrorResponse (from schemas)
		var apiError *schemas.ErrorResponse
		if errors.As(err, &apiError) {
			ctx.JSON(apiError.Status, apiError)
			return
		}

		// Unknown error - return 500 Internal Server Error
		ctx.JSON(http.StatusInternalServerError, schemas.NewInternalError(
			err.Error(),
			"/sessions/"+sessionID+"/status/timeout",
		))
		return
	}

	ctx.JSON(http.StatusOK, schemas.UpdateSessionStatusResponse{
		Message:   "Session status updated to TIMEOUT",
		SessionID: sessionID,
		Status:    "TIMEOUT",
	})
}

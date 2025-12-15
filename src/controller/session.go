package controller

import (
	"errors"
	"log/slog"
	"net/http"

	"connection-service/src/config"
	"connection-service/src/schemas"
	"connection-service/src/service"

	"github.com/gin-gonic/gin"
)

type SessionController struct {
	Service           *service.SessionService
	ConnectionService *service.ConnectionService
	Config            *config.GlobalConfig
}

func NewSessionController(service *service.SessionService, connectionService *service.ConnectionService, config *config.GlobalConfig) *SessionController {
	return &SessionController{
		Service:           service,
		ConnectionService: connectionService,
		Config:            config,
	}
}

// Start handles the client connection
func (sc *SessionController) Start(ctx *gin.Context) {
	var reqBody schemas.ConnectRequest
	if err := ctx.ShouldBindJSON(&reqBody); err != nil {
		slog.Error("Invalid JSON format", "error", err)
		ctx.JSON(http.StatusBadRequest, schemas.NewBadRequestError(
			"Invalid JSON format: "+err.Error(),
			"/sessions/start",
		))
		return
	}

	// Delegate all business logic to the service layer
	response, err := sc.ConnectionService.HandleClientConnection(ctx.Request.Context(), reqBody.UserID, reqBody.Token)
	if err != nil {
		// Check if the error is an ErrorResponse (from schemas)
		var apiError *schemas.ErrorResponse
		if errors.As(err, &apiError) {
			slog.Error("Connection failed", "error", apiError, "user_id", reqBody.UserID, "status", apiError.Status)
			ctx.JSON(apiError.Status, apiError)
			return
		}

		// Unknown error - return 500 Internal Server Error
		slog.Error("Internal error during connection", "error", err, "user_id", reqBody.UserID)
		ctx.JSON(http.StatusInternalServerError, schemas.NewInternalError(
			err.Error(),
			"/sessions/start",
		))
		return
	}

	slog.Info("Client connected successfully", "user_id", reqBody.UserID)
	ctx.JSON(http.StatusOK, response)
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



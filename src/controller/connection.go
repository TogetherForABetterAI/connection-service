package controller

import (
	"errors"
	"log/slog"
	"net/http"

	"connection-service/src/schemas"
	"connection-service/src/service"

	"github.com/gin-gonic/gin"
)

type ConnectionController struct {
	Service *service.ConnectionService
}

func NewConnectionController(service *service.ConnectionService) *ConnectionController {
	return &ConnectionController{
		Service: service,
	}
}

func (c *ConnectionController) Connect(ctx *gin.Context) {
	var reqBody schemas.ConnectRequest
	if err := ctx.ShouldBindJSON(&reqBody); err != nil {
		slog.Error("Invalid JSON format", "error", err)
		ctx.JSON(http.StatusBadRequest, schemas.NewBadRequestError(
			"Invalid JSON format: "+err.Error(),
			"/users/connect",
		))
		return
	}

	// Delegate all business logic to the service layer
	response, err := c.Service.HandleClientConnection(ctx.Request.Context(), reqBody.UserID, reqBody.Token)
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
			"/users/connect",
		))
		return
	}

	slog.Info("Client connected successfully", "user_id", reqBody.UserID)
	ctx.JSON(http.StatusOK, response)
}

package controller

import (
	"connection-service/src/models"
	"connection-service/src/service"
	"encoding/json"
	"log/slog"
	"net/http"

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
	var reqBody models.ConnectRequest
	if err := ctx.ShouldBindJSON(&reqBody); err != nil {
		slog.Error("Invalid JSON format", "error", err)
		ctx.JSON(http.StatusBadRequest, models.APIError{
			Type:     "https://connection-service.com/validation-error",
			Title:    "Bad Request",
			Status:   http.StatusBadRequest,
			Detail:   "Invalid JSON format: " + err.Error(),
			Instance: "/users/connect",
		})
		return
	}

	// Delegate all business logic to the service layer
	response, serviceErr, err := c.Service.HandleClientConnection(ctx.Request.Context(), reqBody.UserID, reqBody.Token)
	if serviceErr != nil {
		slog.Error("Connection failed", "error", serviceErr, "user_id", reqBody.UserID, "status", serviceErr.StatusCode)
		var detailObj interface{}
		if err := json.Unmarshal([]byte(serviceErr.ResponseBody), &detailObj); err == nil {
			ctx.JSON(serviceErr.StatusCode, gin.H{"detail": detailObj})
		} else {
			ctx.JSON(serviceErr.StatusCode, gin.H{"detail": serviceErr.ResponseBody})
		}
		return
	}
	if err != nil {
		slog.Error("Internal error during connection", "error", err, "user_id", reqBody.UserID)
		ctx.JSON(http.StatusInternalServerError, gin.H{"detail": err.Error()})
		return
	}

	slog.Info("Client connected successfully", "user_id", reqBody.UserID)
	ctx.JSON(http.StatusOK, response)
}

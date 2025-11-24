package controller

import (
	"connection-service/src/models"
	"connection-service/src/service"
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

func (c *ConnectionController) sendError(ctx *gin.Context, status int, title string, detail string, errType string, instance string) {
	errorResp := models.APIError{
		Type:     errType,
		Title:    title,
		Status:   status,
		Detail:   detail,
		Instance: instance,
	}
	ctx.JSON(status, errorResp)
	slog.Error("API Error",
		slog.String("title", title),
		slog.String("detail", detail),
		slog.Int("status", status),
		slog.String("instance", instance))
}

// @Summary Connect authenticated client
// @Description Connects an authenticated client and notifies other services about the new connection
// @Tags users
// @Accept json
// @Produce json
// @Param ConnectRequest body models.ConnectRequest true "Connect Request with User ID and token"
// @Success 200 {object} models.ConnectResponse
// @Failure 400 {object} models.APIError
// @Failure 401 {object} models.APIError
// @Failure 500 {object} models.APIError
// @Router /users/connect [post]
func (c *ConnectionController) Connect(ctx *gin.Context) {
	var reqBody models.ConnectRequest
	if err := ctx.ShouldBindJSON(&reqBody); err != nil {
		c.sendError(ctx, http.StatusBadRequest, "Bad Request", "Invalid JSON format: "+err.Error(), "https://connection-service.com/validation-error", "/connect")
		return
	}

	// Delegate all business logic to the service layer
	response, err := c.Service.HandleClientConnection(ctx.Request.Context(), reqBody.UserID, reqBody.Token)
	if err != nil {
		// Determine the appropriate HTTP status code based on the error
		statusCode := http.StatusInternalServerError
		errorType := "https://connection-service.com/internal-error"

		// Check if it's a token validation error
		if err.Error() == "token validation failed: invalid token" ||
			err.Error() == "token validation failed: token validation failed with status code: 401" {
			statusCode = http.StatusUnauthorized
			errorType = "https://connection-service.com/validation-error"
		}

		c.sendError(ctx, statusCode, getErrorTitle(statusCode), err.Error(), errorType, "/connect")
		return
	}

	ctx.JSON(http.StatusOK, response)
}

// getErrorTitle returns an appropriate error title based on status code
func getErrorTitle(statusCode int) string {
	switch statusCode {
	case http.StatusBadRequest:
		return "Bad Request"
	case http.StatusUnauthorized:
		return "Unauthorized"
	case http.StatusInternalServerError:
		return "Internal Error"
	default:
		return "Error"
	}
}

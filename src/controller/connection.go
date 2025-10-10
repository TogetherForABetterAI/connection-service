package controller

import (
	"bytes"
	"connection-service/src/models"
	"connection-service/src/service"
	"encoding/json"
	"fmt"
	"io"
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
// @Param ConnectRequest body models.ConnectRequest true "Connect Request with client ID and token"
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

	_, err := ValidateToken(reqBody.Token, reqBody.ClientId)
	if err != nil {
		c.sendError(ctx, http.StatusUnauthorized, "Unauthorized", "Token validation failed: "+err.Error(), "https://connection-service.com/validation-error", "/connect")
		return
	}

	err = c.Service.NotifyNewConnection(reqBody.ClientId, "-", "-", reqBody.model_type)
	if err != nil {
		c.sendError(ctx, http.StatusInternalServerError, "Internal Error", err.Error(), "https://connection-service.com/internal-error", "/connect")
		return
	}

	ctx.JSON(http.StatusOK, models.ConnectResponse{
		Status:  "success",
		Message: "Client connected successfully",
	})
}

func ValidateToken(token, clientID string) (*models.TokenValidateResponse, error) {
	postBody, err := json.Marshal(map[string]string{"token": token, "client_id": clientID})

	if err != nil {
		return nil, fmt.Errorf("failed to validate token")
	}

	resp, err := http.Post("http://authenticator-service-app:8000/tokens/validate", "application/json", bytes.NewBuffer(postBody))
	if err != nil || resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to validate token")
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var tokenResp models.TokenValidateResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, err
	}

	if !tokenResp.IsValid {
		return nil, fmt.Errorf("invalid token")
	}

	return &tokenResp, nil
}

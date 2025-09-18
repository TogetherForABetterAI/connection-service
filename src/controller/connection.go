package controller

import (
	"auth-gateway/src/models"
	"auth-gateway/src/service"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type ConnectionController struct {
	Service *service.ConnectionService
	Logger  *logrus.Logger
}

func NewConnectionController(service *service.ConnectionService, logger *logrus.Logger) *ConnectionController {
	return &ConnectionController{
		Service: service,
		Logger:  logger,
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
	c.Logger.Error(title + ": " + detail)
}

func (c *ConnectionController) Connect(ctx *gin.Context) {
	var reqBody models.ConnectRequest
	if err := ctx.ShouldBindJSON(&reqBody); err != nil {
		c.sendError(ctx, http.StatusBadRequest, "Bad Request", "Invalid JSON format: "+err.Error(), "https://auth-gateway.com/validation-error", "/connect")
		return
	}

	_, err := ValidateToken(reqBody.Token, reqBody.ClientId)
	if err != nil {
		c.sendError(ctx, http.StatusUnauthorized, "Unauthorized", "Token validation failed: "+err.Error(), "https://auth-gateway.com/validation-error", "/connect")
		return
	}

	err = c.Service.NotifyNewConnection(reqBody.ClientId, "-", "-")
	if err != nil {
		c.sendError(ctx, http.StatusInternalServerError, "Internal Error", err.Error(), "https://auth-gateway.com/internal-error", "/connect")
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

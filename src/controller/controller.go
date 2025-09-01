package controller

import (
	"github.com/sirupsen/logrus"
	"auth-gateway/src/models"
	"net/http"
	pb "auth-gateway/src/pb/new-client-service"
	"auth-gateway/src/config"
	"encoding/json"
	"bytes"
	"auth-gateway/src/service"
	"github.com/gin-gonic/gin"
	"io"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"context"
	"fmt"
	"time"
)


type Controller struct {
	Logger  *logrus.Logger
	Service *service.Service
}


func (c *Controller) sendError(ctx *gin.Context, status int, title string, detail string, errType string, instance string) {
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

func (c *Controller) Connect(context *gin.Context) {
	var reqBody models.ConnectRequest
	err := context.ShouldBindJSON(&reqBody)

	if err != nil {
		c.sendError(context, http.StatusBadRequest, "Bad Request", "Invalid JSON format: " + err.Error(), "https://auth-gateway.com/validation-error", "/connect")
		return
	}

	_, err = ValidateToken(reqBody.Token, reqBody.ClientId)
	if err != nil {
		c.sendError(context, http.StatusUnauthorized, "Unauthorized", "Token validation failed: "+err.Error(), "https://auth-gateway.com/validation-error", "/connect")
		return
	}

	newClientRequest := &pb.NewClientRequest{
		ClientId: reqBody.ClientId,
		InputsFormat: "-",
		OutputsFormat: "-",
	}
	err = NotifyNewClient(config.Config.CalibrationServiceAddr, newClientRequest)
	if err != nil {
		c.sendError(context, http.StatusBadRequest, "Bad Request", "Failed to connect to calibration service: "+err.Error(), "https://auth-gateway.com/validation-error", "/connect")
		return
	}

	err = NotifyNewClient(config.Config.DataDispatcherServiceAddr, newClientRequest)
	if err != nil {
		c.sendError(context, http.StatusBadRequest, "Bad Request", "Failed to connect to data dispatcher service: "+err.Error(), "https://auth-gateway.com/validation-error", "/connect")
		return
	}

	successResponse := models.ConnectResponse{
		Status:  "success",
		Message: "Client connected successfully",
	}

	context.JSON(http.StatusOK, successResponse)
}

func (c *Controller) CreateToken(context *gin.Context) {
	var reqBody models.TokenCreateRequest
	err := context.ShouldBindJSON(&reqBody)

	if err != nil {
		c.sendError(context, http.StatusBadRequest, "Bad Request", "Invalid JSON format: " + err.Error(), "https://auth-gateway.com/validation-error", "/tokens/create")
		return
	}
	
	postBody, err := json.Marshal(map[string]string{"client_id": reqBody.ClientId})
	if err != nil {
		c.sendError(context, http.StatusInternalServerError, "Internal Error", "Failed to marshal request body: "+err.Error(), "https://auth-gateway.com/internal-error", "/tokens/create")
		return
	}

	resp, err := http.Post("http://authenticator-service-app:8000/tokens/create", "application/json", bytes.NewBuffer(postBody))
	if err != nil {
		c.sendError(context, http.StatusInternalServerError, "Internal Error", "Failed to send request: "+err.Error(), "https://auth-gateway.com/internal-error", "/tokens/create")
		return
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.sendError(context, http.StatusInternalServerError, "Internal Error", "Failed to read response body", "https://auth-gateway.com/internal-error", "/tokens/create")
		return
	}
	var tokenCreateResp models.TokenCreateResponse
	if err := json.Unmarshal(body, &tokenCreateResp); err != nil {
		c.sendError(context, http.StatusInternalServerError, "Internal Error", "Failed to unmarshal response body: "+err.Error(), "https://auth-gateway.com/internal-error", "/tokens/create")
		return
	}

	context.JSON(resp.StatusCode, tokenCreateResp)
}

func (c *Controller) CreateUser(context *gin.Context) {
	var reqBody models.UserCreateRequest
	err := context.ShouldBindJSON(&reqBody)

	if err != nil {
		c.sendError(context, http.StatusBadRequest, "Bad Request", "Invalid JSON format: "+err.Error(), "https://auth-gateway.com/validation-error", "/users/create")
		return
	}

	postBody, err := json.Marshal(map[string]interface{}{
		"username":      reqBody.Username,
		"email":         reqBody.Email,
		"inputs_format": reqBody.InputsFormat,
		"outputs_format": reqBody.OutputsFormat,
	})

	if err != nil {
		c.sendError(context, http.StatusInternalServerError, "Internal Error", "Failed to marshal request body: "+err.Error(), "https://auth-gateway.com/internal-error", "/tokens/create")
		return
	}

	resp, err := http.Post("http://authenticator-service-app:8000/users/create", "application/json", bytes.NewBuffer(postBody))
	if err != nil {
		c.sendError(context, http.StatusInternalServerError, "Internal Error", "Failed to send request: "+err.Error(), "https://auth-gateway.com/internal-error", "/tokens/create")
		return
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.sendError(context, http.StatusInternalServerError, "Internal Error", "Failed to read response body: "+err.Error(), "https://auth-gateway.com/internal-error", "/tokens/create")
		return
	}

	var userCreateResp models.UserCreateResponse
	if err := json.Unmarshal(body, &userCreateResp); err != nil {
		c.sendError(context, http.StatusInternalServerError, "Internal Error", "Failed to unmarshal response body: "+err.Error(), "https://auth-gateway.com/internal-error", "/tokens/create")
		return
	}

	context.JSON(resp.StatusCode, userCreateResp)
}

func NotifyNewClient(serviceAddr string, newClientRequest *pb.NewClientRequest) error {
	conn, err := grpc.Dial(serviceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("failed to connect to service: %w", err)
	}
	defer conn.Close()
	
	client := pb.NewClientNotificationServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	if _, err := client.NotifyNewClient(ctx, newClientRequest); err != nil {
		return err
	}
	return nil
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

    return &tokenResp, nil
}

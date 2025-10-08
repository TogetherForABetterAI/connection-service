package controller

import (
	"auth-gateway/src/config"
	"auth-gateway/src/models"
	pb "auth-gateway/src/pb/new-client-service"
	"auth-gateway/src/service"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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

// @BasePath /
// Connect godoc
// @Summary connect
// @Param ConnectRequest body models.ConnectRequest true "Connect Request"
// @Schemes
// @Description connect
// @Tags users
// @Accept json
// @Produce json
// @Success 200 {object} models.ConnectResponse
// @Failure 400 {object} models.APIError
// @Failure 404 {object} models.APIError
// @Failure 500 {object} models.APIError
// @Router /users/connect [post]
func (c *Controller) Connect(context *gin.Context) {
	var reqBody models.ConnectRequest
	err := context.ShouldBindJSON(&reqBody)

	c.Logger.Infof("Client %s requested to connect", reqBody.ClientId)

	if err != nil {
		c.sendError(context, http.StatusBadRequest, "Bad Request", "Invalid JSON format: "+err.Error(), "https://auth-gateway.com/validation-error", "/connect")
		return
	}

	_, err = ValidateToken(reqBody.Token, reqBody.ClientId)
	if err != nil {
		c.sendError(context, http.StatusUnauthorized, "Unauthorized", "Token validation failed: "+err.Error(), "https://auth-gateway.com/validation-error", "/connect")
		return
	}

	// Fetch user info to get model_type and formats
	userInfo, err := GetUserInfo(reqBody.ClientId)
	if err != nil {
		c.sendError(context, http.StatusInternalServerError, "Internal Error", "Failed to fetch user information: "+err.Error(), "https://auth-gateway.com/internal-error", "/connect")
		return
	}

	newClientRequest := &pb.NewClientRequest{
		ClientId:      reqBody.ClientId, // Also used as routing key
		InputsFormat:  userInfo.InputsFormat,
		OutputsFormat: userInfo.OutputsFormat,
		ModelType:     userInfo.ModelType,
	}
	err = NotifyNewClient(config.Config.DataDispatcherServiceAddr, newClientRequest)
	if err != nil {
		c.sendError(context, http.StatusBadRequest, "Bad Request", "Failed to connect to data dispatcher service: "+err.Error(), "https://auth-gateway.com/validation-error", "/connect")
		return
	}

	err = NotifyNewClient(config.Config.CalibrationServiceAddr, newClientRequest)
	if err != nil {
		c.sendError(context, http.StatusBadRequest, "Bad Request", "Failed to connect to calibration service: "+err.Error(), "https://auth-gateway.com/validation-error", "/connect")
		return
	}

	c.Logger.Infof("Client %s connected successfully", reqBody.ClientId)

	successResponse := models.ConnectResponse{
		Status:  "success",
		Message: "Client connected successfully",
	}

	context.JSON(http.StatusOK, successResponse)
}

// @BasePath /

// CreateToken godoc
// @Summary create token
// @Param client_id body string true "Client ID"
// @Schemes
// @Description create token
// @Tags tokens
// @Accept json
// @Produce json
// @Success 200 {object} models.TokenCreateResponse
// @Failure 400 {object} models.APIError
// @Failure 404 {object} models.APIError
// @Failure 500 {object} models.APIError
// @Router /tokens/create [post]
func (c *Controller) CreateToken(context *gin.Context) {
	var reqBody models.TokenCreateRequest
	err := context.ShouldBindJSON(&reqBody)

	if err != nil {
		c.sendError(context, http.StatusBadRequest, "Bad Request", "Invalid JSON format: "+err.Error(), "https://auth-gateway.com/validation-error", "/tokens/create")
		return
	}

	postBody, err := json.Marshal(map[string]string{"client_id": reqBody.ClientId})
	if err != nil {
		c.sendError(context, http.StatusInternalServerError, "Internal Error", "Failed to marshal request body: "+err.Error(), "https://auth-gateway.com/internal-error", "/tokens/create")
		return
	}

	resp, err := http.Post("http://users-service-app:8000/tokens/create", "application/json", bytes.NewBuffer(postBody))
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

// @BasePath /

// @Summary create user
// @Param CreateUserRequest body models.UserCreateRequest true "User Create Request"
// @Schemes
// @Description create user
// @Tags users
// @Accept json
// @Produce json
// @Success 200 {object} models.UserCreateResponse
// @Failure 400 {object} models.APIError
// @Failure 500 {object} models.APIError
// @Router /users/create [post]
func (c *Controller) CreateUser(context *gin.Context) {
	var reqBody models.UserCreateRequest
	err := context.ShouldBindJSON(&reqBody)

	if err != nil {
		c.sendError(context, http.StatusBadRequest, "Bad Request", "Invalid JSON format: "+err.Error(), "https://auth-gateway.com/validation-error", "/users/create")
		return
	}

	postBody, err := json.Marshal(map[string]interface{}{
		"username":       reqBody.Username,
		"email":          reqBody.Email,
		"model_type":     reqBody.ModelType,
		"inputs_format":  reqBody.InputsFormat,
		"outputs_format": reqBody.OutputsFormat,
	})

	if err != nil {
		c.sendError(context, http.StatusInternalServerError, "Internal Error", "Failed to marshal request body: "+err.Error(), "https://auth-gateway.com/internal-error", "/tokens/create")
		return
	}

	resp, err := http.Post("http://users-service-app:8000/users/create", "application/json", bytes.NewBuffer(postBody))
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

// @BasePath /
// @Summary get users
// @Schemes
// @Description get users
// @Tags users
// @Accept json
// @Produce json
// @Success 200 {array} models.UserInfo
// @Failure 500 {object} models.APIError
// @Router /users/ [get]
func (c *Controller) GetUsers(context *gin.Context) {
	resp, err := http.Get("http://users-service-app:8000/users/")
	if err != nil {
		c.sendError(context, http.StatusInternalServerError, "Internal Error", "Failed to send request: "+err.Error(), "https://auth-gateway.com/internal-error", "/users/")
		return
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.sendError(context, http.StatusInternalServerError, "Internal Error", "Failed to read response body: "+err.Error(), "https://auth-gateway.com/internal-error", "/users/")
		return
	}

	var usersResp []models.UserInfo
	if err := json.Unmarshal(body, &usersResp); err != nil {
		c.sendError(context, http.StatusInternalServerError, "Internal Error", "Failed to unmarshal response body: "+err.Error(), "https://auth-gateway.com/internal-error", "/users/")
		return
	}

	context.JSON(resp.StatusCode, usersResp)
}

// @BasePath /
// @Summary get user by ID
// @Param id path string true "User ID"
// @Schemes
// @Description get user by ID
// @Tags users
// @Accept json
// @Produce json
// @Success 200 {object} models.UserInfo
// @Failure 404 {object} models.APIError
// @Failure 500 {object} models.APIError
// @Router /users/{id} [get]
func (c *Controller) GetUserByID(context *gin.Context) {
	userID := context.Param("id")
	resp, err := http.Get("http://users-service-app:8000/users/" + userID)
	if err != nil {
		c.sendError(context, http.StatusInternalServerError, "Internal Error", "Failed to send request: "+err.Error(), "https://auth-gateway.com/internal-error", "/users/"+userID)
		return
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.sendError(context, http.StatusInternalServerError, "Internal Error", "Failed to read response body: "+err.Error(), "https://auth-gateway.com/internal-error", "/users/"+userID)
		return
	}

	if resp.StatusCode == http.StatusNotFound {
		c.sendError(context, http.StatusNotFound, "Not Found", "User not found", "https://auth-gateway.com/not-found", "/users/"+userID)
		return
	}

	var userResp models.UserInfo
	if err := json.Unmarshal(body, &userResp); err != nil {
		c.sendError(context, http.StatusInternalServerError, "Internal Error", "Failed to unmarshal response body: "+err.Error(), "https://auth-gateway.com/internal-error", "/users/"+userID)
		return
	}

	context.JSON(resp.StatusCode, userResp)
}

func (c *Controller) TestWebhook(context *gin.Context) {
	c.Logger.Info("TEST WEBHOOK RECEIVED")
	context.JSON(http.StatusOK, gin.H{"status": "success", "message": "Webhook received"})
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

	resp, err := http.Post("http://users-service-app:8000/tokens/validate", "application/json", bytes.NewBuffer(postBody))
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

func GetUserInfo(clientID string) (*models.UserInfo, error) {
	resp, err := http.Get("http://users-service-app:8000/users/" + clientID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("user not found or server error (status: %d)", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var userInfo models.UserInfo
	if err := json.Unmarshal(body, &userInfo); err != nil {
		return nil, fmt.Errorf("failed to unmarshal user info: %w", err)
	}

	return &userInfo, nil
}

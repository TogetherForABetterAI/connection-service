package controller

import (
	"auth-gateway/logger"
	"auth-gateway/src/config"
	"auth-gateway/src/models"
	pb "auth-gateway/src/pb/new-client-service"
	"auth-gateway/src/utils"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type UserController struct{}

func NewUserController() *UserController {
	return &UserController{}
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
func (c *UserController) CreateUser(context *gin.Context) {
	var reqBody models.UserCreateRequest
	err := context.ShouldBindJSON(&reqBody)
	logger.Logger.Info("Creating user")
	logger.Logger.Infof("Creating user: %+v", reqBody)

	if err != nil {
		utils.SendError(context, http.StatusBadRequest, "Bad Request", "Invalid JSON format: "+err.Error(), "https://auth-gateway.com/validation-error", "/users/create")
		return
	}
	logger.Logger.Infof("Creating user: %+v", reqBody)

	postBody, err := json.Marshal(map[string]interface{}{
		"username":       reqBody.Username,
		"email":          reqBody.Email,
		"model_type":     reqBody.ModelType,
		"inputs_format":  reqBody.InputsFormat,
		"outputs_format": reqBody.OutputsFormat,
	})

	if err != nil {
		utils.SendError(context, http.StatusInternalServerError, "Internal Error", "Failed to marshal request body: "+err.Error(), "https://auth-gateway.com/internal-error", "/users/create")
		return
	}

	resp, err := http.Post("http://authenticator-service-app:8000/users/create", "application/json", bytes.NewBuffer(postBody))
	if err != nil {
		logger.Logger.Infof("ERROOOR")
		utils.SendError(context, http.StatusInternalServerError, "Internal Error", "Failed to send request: "+err.Error(), "https://auth-gateway.com/internal-error", "/users/create")
		return
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		utils.SendError(context, http.StatusInternalServerError, "Internal Error", "Failed to read response body: "+err.Error(), "https://auth-gateway.com/internal-error", "/tokens/create")
		return
	}

	var userCreateResp models.UserCreateResponse
	if err := json.Unmarshal(body, &userCreateResp); err != nil {
		utils.SendError(context, http.StatusInternalServerError, "Internal Error", "Failed to unmarshal response body: "+err.Error(), "https://auth-gateway.com/internal-error", "/tokens/create")
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
func (c *UserController) GetUsers(context *gin.Context) {
	resp, err := http.Get("http://authenticator-service-app:8000/users/")
	if err != nil {
		utils.SendError(context, http.StatusInternalServerError, "Internal Error", "Failed to send request: "+err.Error(), "https://auth-gateway.com/internal-error", "/users/")
		return
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		utils.SendError(context, http.StatusInternalServerError, "Internal Error", "Failed to read response body: "+err.Error(), "https://auth-gateway.com/internal-error", "/users/")
		return
	}

	var usersResp []models.UserInfo
	if err := json.Unmarshal(body, &usersResp); err != nil {
		utils.SendError(context, http.StatusInternalServerError, "Internal Error", "Failed to unmarshal response body: "+err.Error(), "https://auth-gateway.com/internal-error", "/users/")
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
func (c *UserController) GetUserByID(context *gin.Context) {
	userID := context.Param("id")
	resp, err := http.Get("http://authenticator-service-app:8000/users/" + userID)
	if err != nil {
		utils.SendError(context, http.StatusInternalServerError, "Internal Error", "Failed to send request: "+err.Error(), "https://auth-gateway.com/internal-error", "/users/"+userID)
		return
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		utils.SendError(context, http.StatusInternalServerError, "Internal Error", "Failed to read response body: "+err.Error(), "https://auth-gateway.com/internal-error", "/users/"+userID)
		return
	}

	if resp.StatusCode == http.StatusNotFound {
		utils.SendError(context, http.StatusNotFound, "Not Found", "User not found", "https://auth-gateway.com/not-found", "/users/"+userID)
		return
	}

	var userResp models.UserInfo
	if err := json.Unmarshal(body, &userResp); err != nil {
		utils.SendError(context, http.StatusInternalServerError, "Internal Error", "Failed to unmarshal response body: "+err.Error(), "https://auth-gateway.com/internal-error", "/users/"+userID)
		return
	}

	context.JSON(resp.StatusCode, userResp)
}

func (c *UserController) TestWebhook(context *gin.Context) {
	logger.Logger.Info("TEST WEBHOOK RECEIVED")
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
func (c *UserController) Connect(context *gin.Context) {
	var reqBody models.ConnectRequest
	err := context.ShouldBindJSON(&reqBody)

	logger.Logger.Infof("Client %s requested to connect", reqBody.ClientId)

	if err != nil {
		utils.SendError(context, http.StatusBadRequest, "Bad Request", "Invalid JSON format: "+err.Error(), "https://auth-gateway.com/validation-error", "/connect")
		return
	}

	newClientRequest := &pb.NewClientRequest{
		ClientId:      reqBody.ClientId,
		InputsFormat:  "-",
		OutputsFormat: "-",
	}
	err = NotifyNewClient(config.Config.DataDispatcherServiceAddr, newClientRequest)
	if err != nil {
		utils.SendError(context, http.StatusBadRequest, "Bad Request", "Failed to connect to data dispatcher service: "+err.Error(), "https://auth-gateway.com/validation-error", "/connect")
		return
	}

	err = NotifyNewClient(config.Config.CalibrationServiceAddr, newClientRequest)
	if err != nil {
		utils.SendError(context, http.StatusBadRequest, "Bad Request", "Failed to connect to calibration service: "+err.Error(), "https://auth-gateway.com/validation-error", "/connect")
		return
	}

	logger.Logger.Infof("Client %s connected successfully", reqBody.ClientId)

	successResponse := models.ConnectResponse{
		Status:  "success",
		Message: "Client connected successfully",
	}

	context.JSON(http.StatusOK, successResponse)
}

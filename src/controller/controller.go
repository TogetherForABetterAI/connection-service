package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"auth-gateway/src/models"
	"net/http"
	pb "auth-gateway/src/pb/new-client-service"
	"auth-gateway/src/config"
	"google.golang.org/grpc" 
	"google.golang.org/grpc/credentials/insecure"
	"context"
	"fmt"
	"time"
	"io"
	"encoding/json"
	"bytes"

)

type Controller struct {
	Logger *logrus.Logger
}


func (c *Controller) notifyNewClient(serviceAddr string, newClientRequest *pb.NewClientRequest) error {
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

func (c *Controller) validateToken(token, clientID string) (*models.TokenValidateResponse, error) {
    postBody, _ := json.Marshal(map[string]string{"token": token, "client_id": clientID})

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


func (c *Controller) sendError(ctx *gin.Context, status int, title, detail string) {
	errorResp := models.APIError{
		Type:     "https://auth-gateway.com/validation-error",
		Title:    title,
		Status:   status,
		Detail:   detail,
		Instance: "/connect",
	}
	ctx.JSON(status, errorResp)
	c.Logger.Error(title + ": " + detail)
}

func (c *Controller) Connect(context *gin.Context) {
	var reqBody models.ConnectRequest
	err := context.ShouldBindJSON(&reqBody)

	if err != nil {
		errorResponse := models.APIError{
			Type:     "https://auth-gateway.com/validation-error",
			Title:    "Bad Request",
			Status:   http.StatusBadRequest,
			Detail:   "Invalid JSON format: " + err.Error(),
			Instance: "/connect",
		}
		context.JSON(http.StatusBadRequest, errorResponse)
		c.Logger.Error("Bad Request: ", err.Error())
		return
	}

	_, err = c.validateToken(reqBody.Token, reqBody.ClientId)
	if err != nil {
		c.sendError(context, http.StatusUnauthorized, "Unauthorized", "Token validation failed: "+err.Error())
		return
	}

	newClientRequest := &pb.NewClientRequest{
		ClientId: reqBody.ClientId,
		InputsFormat: reqBody.InputsFormat,
		OutputsFormat: reqBody.OutputsFormat,
	}
	err = c.notifyNewClient(config.Config.CalibrationServiceAddr, newClientRequest)
	if err != nil {
		c.sendError(context, http.StatusBadRequest, "Bad Request", "Failed to connect to calibration service: "+err.Error())
		return
	}

	err = c.notifyNewClient(config.Config.DataDispatcherServiceAddr, newClientRequest)
	if err != nil {
		c.sendError(context, http.StatusBadRequest, "Bad Request", "Failed to connect to data dispatcher service: "+err.Error())
		return
	}

	successResponse := models.ConnectResponse{
		Status:  "success",
		Message: "Client connected successfully",
	}

	context.JSON(http.StatusOK, successResponse)
}

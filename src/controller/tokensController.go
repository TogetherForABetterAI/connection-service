package controller

import (
	"auth-gateway/logger"
	"auth-gateway/src/models"
	"auth-gateway/src/utils"
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

type TokenController struct{}

func NewTokenController() *TokenController {
	return &TokenController{}
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

func (c *TokenController) CreateToken(context *gin.Context) {
	var reqBody models.TokenCreateRequest
	err := context.ShouldBindJSON(&reqBody)

	if err != nil {
		utils.SendError(context, http.StatusBadRequest, "Bad Request", "Invalid JSON format: "+err.Error(), "https://auth-gateway.com/validation-error", "/tokens/create")
		return
	}

	postBody, err := json.Marshal(map[string]string{"user_id": reqBody.UserId})
	if err != nil {
		utils.SendError(context, http.StatusInternalServerError, "Internal Error", "Failed to marshal request body: "+err.Error(), "https://auth-gateway.com/internal-error", "/tokens/create")
		return
	}

	resp, err := http.Post("http://authenticator-service-app:8000/tokens/create", "application/json", bytes.NewBuffer(postBody))
	if err != nil {
		utils.SendError(context, http.StatusInternalServerError, "Internal Error", "Failed to send request: "+err.Error(), "https://auth-gateway.com/internal-error", "/tokens/create")
		return
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		utils.SendError(context, http.StatusInternalServerError, "Internal Error", "Failed to read response body: "+err.Error(), "https://auth-gateway.com/internal-error", "/tokens/create")
		return
	}
	var tokenCreateResp models.TokenCreateResponse
	if err := json.Unmarshal(body, &tokenCreateResp); err != nil {
		utils.SendError(context, http.StatusInternalServerError, "Internal Error", "Failed to unmarshal response body: "+err.Error(), "https://auth-gateway.com/internal-error", "/tokens/create")
		return
	}

	context.JSON(resp.StatusCode, tokenCreateResp)
}

func (c *TokenController) GetTokens(context *gin.Context) {
	resp, err := http.Get("http://authenticator-service-app:8000/tokens/")
	if err != nil {
		utils.SendError(context, http.StatusInternalServerError, "Internal Error", "Failed to send request: "+err.Error(), "https://auth-gateway.com/internal-error", "/tokens/")
		return
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		utils.SendError(context, http.StatusInternalServerError, "Internal Error", "Failed to read response body: "+err.Error(), "https://auth-gateway.com/internal-error", "/tokens/")
		return
	}
	var tokens []models.TokenInfo
	if err := json.Unmarshal(body, &tokens); err != nil {
		utils.SendError(context, http.StatusInternalServerError, "Internal Error", "Failed to unmarshal response body: "+err.Error(), "https://auth-gateway.com/internal-error", "/tokens/")
		return
	}
	logger.Logger.Infof("Tokens retrieved successfully: %+v", tokens)

	context.JSON(resp.StatusCode, tokens)
}

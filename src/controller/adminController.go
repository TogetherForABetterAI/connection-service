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

type AdminController struct{}

func NewAdminController() *AdminController {
	return &AdminController{}
}

func (c *AdminController) InviteAdmin(context *gin.Context) {
	var reqBody struct {
		Email string `json:"email"`
	}
	err := context.ShouldBindJSON(&reqBody)

	if err != nil {
		utils.SendError(context, http.StatusBadRequest, "Bad Request", "Invalid JSON format: "+err.Error(), "https://auth-gateway.com/validation-error", "/admins/invite")
		return
	}
	postBody, err := json.Marshal(map[string]string{"email": reqBody.Email})

	if err != nil {
		utils.SendError(context, http.StatusInternalServerError, "Internal Error", "Failed to marshal request body: "+err.Error(), "https://auth-gateway.com/internal-error", "/admins/invite")
		return
	}
	resp, err := http.Post("http://authenticator-service-app:8000/admins/invite", "application/json", bytes.NewBuffer(postBody))

	if err != nil {
		utils.SendError(context, resp.StatusCode, err.Error(), err.Error(), "https://auth-gateway.com/internal-error", "/admins/invite")
		return
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		utils.SendError(context, http.StatusInternalServerError, "Error", "Failed to read response body: "+err.Error(), "https://auth-gateway.com/internal-error", "/admins/invite")
		return
	}

	context.Data(resp.StatusCode, "application/json", body)
}

func (c *AdminController) Signup(context *gin.Context) {
	var reqBody models.AdminAuth
	err := context.ShouldBindJSON(&reqBody)

	if err != nil {
		utils.SendError(context, http.StatusBadRequest, "Bad Request", "Invalid JSON format: "+err.Error(), "https://auth-gateway.com/validation-error", "/admins/signup")
		return
	}

	postBody, err := json.Marshal(map[string]string{
		"email":    reqBody.Email,
		"password": reqBody.Password,
	})

	if err != nil {
		utils.SendError(context, http.StatusInternalServerError, "Internal Error", "Failed to marshal request body: "+err.Error(), "https://auth-gateway.com/internal-error", "/admins/signup")
		return
	}
	resp, err := http.Post("http://authenticator-service-app:8000/admins/signup", "application/json", bytes.NewBuffer(postBody))
	logger.Logger.Info("Admin signup attempt for email wewssss")
	if err != nil {
		utils.SendError(context, http.StatusBadRequest, "Bad Request", err.Error(), "https://auth-gateway.com/internal-error", "/admins/signup")
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		utils.SendError(context, http.StatusInternalServerError, "Internal Error", "Failed to read response body: "+err.Error(), "https://auth-gateway.com/internal-error", "/admins/signup")
		return
	}
	context.Data(resp.StatusCode, "application/json", body)
}

func (c *AdminController) Login(context *gin.Context) {
	var reqBody models.AdminAuth
	err := context.ShouldBindJSON(&reqBody)

	if err != nil {
		utils.SendError(context, http.StatusBadRequest, "Bad Request", "Invalid JSON format: "+err.Error(), "https://auth-gateway.com/validation-error", "/admins/login")
		return
	}

	postBody, err := json.Marshal(map[string]string{"email": reqBody.Email, "password": reqBody.Password})

	if err != nil {
		utils.SendError(context, http.StatusInternalServerError, "Internal Error", "Failed to marshal request body: "+err.Error(), "https://auth-gateway.com/internal-error", "/admins/login")
		return
	}
	resp, err := http.Post("http://authenticator-service-app:8000/admins/login", "application/json", bytes.NewBuffer(postBody))

	if err != nil {
		utils.SendError(context, http.StatusInternalServerError, "Internal Error", "Failed to send request: "+err.Error(), "https://auth-gateway.com/internal-error", "/admins/login")
		return
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		utils.SendError(context, http.StatusInternalServerError, "Internal Error", "Failed to read response body: "+err.Error(), "https://auth-gateway.com/internal-error", "/tokens/create")
		return
	}

	var respBody struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int    `json:"expires_in"`
	}

	if err := json.Unmarshal(body, &respBody); err != nil {
		utils.SendError(context, http.StatusInternalServerError, "Internal Error", "Failed to unmarshal response body: "+err.Error(), "https://auth-gateway.com/internal-error", "/tokens/create")
		return
	}
	context.JSON(resp.StatusCode, respBody)
}

func (c *AdminController) GetAdmin(context *gin.Context) {

}

func (c *AdminController) ListAdmins(context *gin.Context) {
	var admins []models.AdminInfo

	logger.Logger.Info("Listing admins")
	resp, err := http.Get("http://authenticator-service-app:8000/admins/")

	if err != nil {
		utils.SendError(context, http.StatusInternalServerError, "Internal Error", "Failed to send request: "+err.Error(), "https://auth-gateway.com/internal-error", "/admins/")
		return
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		utils.SendError(context, http.StatusInternalServerError, "Internal Error", "Failed to read response body: "+err.Error(), "https://auth-gateway.com/internal-error", "/admins/")
		return
	}
	logger.Logger.Info("Admins response: " + string(body))

	if err := json.Unmarshal(body, &admins); err != nil {
		utils.SendError(context, http.StatusInternalServerError, "Internal Error", "Failed to unmarshal response body: "+err.Error(), "https://auth-gateway.com/internal-error", "/admins/")
		return
	}

	context.JSON(resp.StatusCode, admins)
}

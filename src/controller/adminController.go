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

// @BasePath /
// InviteAdmin godoc
// @Summary invite admin
// @Param email body string true "Admin Email"
// @Schemes
// @Description invite admin
// @Tags admins
// @Accept json
// @Produce json
// @Success 200
// @Failure 400 {object} models.APIError
// @Failure 500 {object} models.APIError
// @Router /admins/invite [post]
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

// @BasePath /
// SignUpAdmin godoc
// @Summary sign up admin
// @Param AdminAuth body models.AdminAuth true "Admin Auth"
// @Schemes
// @Description sign up admin
// @Tags admins
// @Accept json
// @Produce json
// @Success 200
// @Failure 400 {object} models.APIError
// @Failure 500 {object} models.APIError
// @Router /admins/signup [post]
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

// @BasePath /
// LoginAdmin godoc
// @Summary login admin
// @Param AdminAuth body models.AdminAuth true "Admin Auth"
// @Schemes
// @Description login admin
// @Tags admins
// @Accept json
// @Produce json
// @Success 200 {object} models.LoginResponse
// @Failure 400 {object} models.APIError
// @Failure 500 {object} models.APIError
// @Router /admins/login [post]
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
	var loginResponse models.LoginResponse

	if err := json.Unmarshal(body, &loginResponse); err != nil {
		utils.SendError(context, http.StatusInternalServerError, "Internal Error", "Failed to unmarshal response body: "+err.Error(), "https://auth-gateway.com/internal-error", "/tokens/create")
		return
	}
	context.JSON(resp.StatusCode, loginResponse)
}

func (c *AdminController) GetAdmin(context *gin.Context) {

}

// @BasePath /
// GetAdmins godoc
// @Summary get admins
// @Schemes
// @Description get admins
// @Tags admins
// @Accept json
// @Produce json
// @Success 200 {array} models.AdminInfo
// @Failure 400 {object} models.APIError
// @Failure 500 {object} models.APIError
// @Router /admins/ [get]
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

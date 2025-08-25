package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"auth-gateway/src/models"
	"net/http"
)

type Controller struct {
	Logger *logrus.Logger
}

func (c Controller) Connect(context *gin.Context) {
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
	response := models.ConnectResponse{
		Status:  "success",
		Message: "Connected successfully",
	}
	context.JSON(http.StatusOK, response)
}

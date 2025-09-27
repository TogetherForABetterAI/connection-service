package utils

import (
	"auth-gateway/src/models"

	"github.com/bytedance/gopkg/util/logger"
	"github.com/gin-gonic/gin"
)

func SendError(ctx *gin.Context, status int, title string, detail string, errType string, instance string) {
	errorResp := models.APIError{
		Type:     errType,
		Title:    title,
		Status:   status,
		Detail:   detail,
		Instance: instance,
	}
	ctx.JSON(status, errorResp)
	logger.Error("Error: ", detail)
}

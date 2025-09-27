package middleware

import (
	"auth-gateway/src/models"
	"auth-gateway/src/utils"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func ValidateToken(token, userID string) (*models.TokenValidateResponse, error) {
	postBody, err := json.Marshal(map[string]string{"token": token, "user_id": userID})

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

func UserAuthRequiredMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			utils.SendError(c, http.StatusUnauthorized, "Unauthorized", "Authorization header missing", "https://auth-gateway.com/validation-error", c.FullPath())
			c.Abort()
			return
		}
		userID := c.GetHeader("UserID")

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			utils.SendError(c, http.StatusUnauthorized, "Unauthorized", "Invalid authorization header format", "https://auth-gateway.com/validation-error", c.FullPath())
			c.Abort()
			return
		}

		token := parts[1]
		tokenResp, err := ValidateToken(token, userID)
		if err != nil {
			utils.SendError(c, http.StatusUnauthorized, "Unauthorized", err.Error(), "https://auth-gateway.com/validation-error", c.FullPath())
			c.Abort()
			return
		}

		c.Set("token_info", tokenResp)
		c.Next()
	}
}

func AdminAuthRequiredMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			utils.SendError(c, http.StatusUnauthorized, "Unauthorized", "Authorization header missing", "https://auth-gateway.com/validation-error", c.FullPath())
			c.Abort()
			return
		}

		req, err := http.NewRequest("POST", "http://authenticator-service-app:8000/admins/authorize", nil)
		if err != nil {
			utils.SendError(c, http.StatusInternalServerError, "Internal Error", "Failed to create request: "+err.Error(), "https://auth-gateway.com/internal-error", c.FullPath())
			c.Abort()
			return
		}
		req.Header.Set("Authorization", authHeader)

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			utils.SendError(c, http.StatusInternalServerError, "Internal Error", "Failed to send request: "+err.Error(), "https://auth-gateway.com/internal-error", c.FullPath())
			c.Abort()
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			utils.SendError(c, http.StatusUnauthorized, "Unauthorized", "Admin authentication failed", "https://auth-gateway.com/validation-error", c.FullPath())
			c.Abort()
			return
		}

		c.Next()
	}
}

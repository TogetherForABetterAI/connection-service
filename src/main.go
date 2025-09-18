package main

import (
	"auth-gateway/src/config"
	"auth-gateway/src/router"
	"fmt"
	"log"

	_ "auth-gateway/src/docs"

	_ "github.com/swaggo/files"
	_ "github.com/swaggo/gin-swagger"
)

// @title Auth Gateway API
// @version 1.0
// @description API Gateway for authentication and authorization

// @contact.name   Auth Gateway Team
// @contact.url    https://github.com/your-org/auth-gateway
// @contact.email  auth-gateway@example.com

func main() {
	config, err := config.NewConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}
	r := router.NewRouter(config)
	if err := r.Run(fmt.Sprintf("%s:%s", config.Host, config.Port)); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

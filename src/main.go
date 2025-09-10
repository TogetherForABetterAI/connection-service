package main

import (
	"auth-gateway/logger"
	"auth-gateway/src/config"
	"auth-gateway/src/router"
	"fmt"
)

func main() {
	logger.Init()
	router := router.Router{logger.Logger}
	r, err_router := router.SetUpRouter()
	if err_router != nil {
		logger.Logger.Fatal("Error while setting up router: ", err_router.Error())
	}
	if err_router = r.Run(fmt.Sprintf("0.0.0.0:%s", config.Config.AppPort)); err_router != nil {
		logger.Logger.Fatal("Error while running the server: ", err_router.Error())
	}
}

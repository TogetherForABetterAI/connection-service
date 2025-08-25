package main

import (
    "auth-gateway/src/router"
    "auth-gateway/logger"
	"fmt"
	"os"
)

func main() {
    logger.Init()
    router := router.Router{logger.Logger}
    r, err_router := router.SetUpRouter()
    if err_router != nil {
        logger.Logger.Fatal("Error while setting up router: ", err_router.Error())
    }
    port := os.Getenv("APP_PORT")
    if err_router = r.Run(fmt.Sprintf("0.0.0.0:%s", port)); err_router != nil {
        logger.Logger.Fatal("Error while running the server: ", err_router.Error())
	}
}
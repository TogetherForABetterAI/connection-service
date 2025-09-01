package router

import (
    "github.com/gin-gonic/gin"
    "github.com/sirupsen/logrus"
    "auth-gateway/src/controller"
    "auth-gateway/src/service"
)

type Router struct {
    Logger *logrus.Logger
}

// SetUpRouter sets up the router for the api-gateway.
// It creates a new gin.Engine, initializes the necessary controllers and routes,
// and returns the router and any error encountered.
func (r Router) SetUpRouter() (*gin.Engine, error) {
    router := gin.Default()
    controller := controller.Controller{
        Logger: r.Logger,
        Service: service.NewService(r.Logger),
    }
    router.POST("/connect", controller.Connect)
    router.POST("/tokens/create", controller.CreateToken)
    router.POST("/users/create", controller.CreateUser)
    return router, nil
}
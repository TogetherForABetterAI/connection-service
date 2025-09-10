package router

import (
	"auth-gateway/src/controller"
	"auth-gateway/src/service"

	docs "auth-gateway/src/docs"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type Router struct {
	Logger *logrus.Logger
}

// @title           Swagger Auth Gateway API
// @version         1.0
// @description     This is a sample server celler server.
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:80
// @BasePath  /

// @securityDefinitions.basic  BasicAuth

// @externalDocs.description  OpenAPI
// @externalDocs.url          https://swagger.io/resources/open-api/

// SetUpRouter sets up the router for the auth-gateway.
// It creates a new gin.Engine, initializes the necessary controllers and routes,
// and returns the router and any error encountered.
func (r Router) SetUpRouter() (*gin.Engine, error) {
	router := gin.Default()
	controller := controller.Controller{
		Logger:  r.Logger,
		Service: service.NewService(r.Logger),
	}
	docs.SwaggerInfo.BasePath = "/"
	tokens_group := router.Group("/tokens")
	{
		tokens_group.POST("/create", controller.CreateToken)

	}
	users_group := router.Group("/users")
	{
		users_group.POST("/connect", controller.Connect)
		users_group.POST("/create", controller.CreateUser)
	}

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	router.POST("/test-webhook", controller.TestWebhook)
	return router, nil
}

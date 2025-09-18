package router

import (
	"auth-gateway/src/config"
	"auth-gateway/src/controller"
	"auth-gateway/src/rabbitmq"
	"auth-gateway/src/service"
	"log"
	"log/slog"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title           Auth Gateway API
// @version         1.0
// @description     API Gateway for authentication and authorization services
// @termsOfService  http://swagger.io/terms/

// @contact.name   Auth Gateway Team
// @contact.url    https://github.com/your-org/auth-gateway
// @contact.email  auth-gateway@example.com

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8080
// @BasePath  /

// @securityDefinitions.basic  BasicAuth

// @externalDocs.description  OpenAPI
// @externalDocs.url          https://swagger.io/resources/open-api/

func createRouterFromConfig(config config.GlobalConfig) *gin.Engine {
	if config.LogLevel == "production" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	r := gin.Default()
	return r
}

func InitializeUserRoutes(r *gin.Engine, connectionController *controller.ConnectionController) {
	usersGroup := r.Group("/users")
	{
		usersGroup.POST("/connect", connectionController.Connect)
	}
}

func NewRouter(config config.GlobalConfig) *gin.Engine {
	r := createRouterFromConfig(config)

	// Initialize structured logger
	slog.Info("Initializing Auth Gateway router")

	// Initialize RabbitMQ publisher
	publisher, err := rabbitmq.NewAMQPPublisherFromConfig(config)
	if err != nil {
		log.Fatalf("Failed to create RabbitMQ publisher: %v", err)
	}

	// Initialize services
	connectionService := service.NewConnectionService(publisher)

	// Initialize controllers (no logger dependency needed)
	connectionController := controller.NewConnectionController(connectionService)

	// Initialize all routes
	InitializeRoutes(r, connectionController)

	// Swagger documentation
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	slog.Info("Auth Gateway router initialized successfully")
	return r
}

func InitializeRoutes(
	r *gin.Engine,
	connectionController *controller.ConnectionController,
) {
	InitializeUserRoutes(r, connectionController)
}

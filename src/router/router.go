package router

import (
	"connection-service/src/config"
	"connection-service/src/controller"
	"connection-service/src/middleware"
	"connection-service/src/service"
	"log"
	"log/slog"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title           Connection Service API
// @version         1.0
// @description     Connection Service for managing user connections
// @termsOfService  http://swagger.io/terms/

// @contact.name   Connection Service Team
// @contact.url    https://github.com/your-org/connection-service
// @contact.email  connection-service@example.com

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
	slog.Info("Initializing Connection Service router")

	// Initialize RabbitMQ middleware
	middleware, err := middleware.NewMiddleware(config)
	if err != nil {
		log.Fatalf("Failed to create RabbitMQ middleware: %v", err)
	}

	// Initialize services
	connectionService := service.NewConnectionService(middleware)

	// Initialize controllers (no logger dependency needed)
	connectionController := controller.NewConnectionController(connectionService)

	// Initialize all routes
	InitializeRoutes(r, connectionController)

	// Swagger documentation
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	slog.Info("Connection Service router initialized successfully")
	return r
}

func InitializeRoutes(
	r *gin.Engine,
	connectionController *controller.ConnectionController,
) {
	InitializeUserRoutes(r, connectionController)
}

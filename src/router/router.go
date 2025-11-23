package router

import (
	"connection-service/src/config"
	"connection-service/src/controller"
	"connection-service/src/db"
	"connection-service/src/middleware"
	"connection-service/src/repository"
	"connection-service/src/service"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"log"
	"log/slog"
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

func createRouterFromConfig(config *config.GlobalConfig) *gin.Engine {
	if config.GetLogLevel() == "production" {
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

func InitializeSessionRoutes(r *gin.Engine, sessionController *controller.SessionController) {
	sessionsGroup := r.Group("/sessions")
	{
		sessionsGroup.PUT("/status", sessionController.UpdateSessionStatus)
	}
}

func NewRouter(cfg *config.GlobalConfig, database *db.DB) *gin.Engine {
	r := createRouterFromConfig(cfg)

	slog.Info("Initializing Connection Service router")

	// Initialize RabbitMQ middleware
	rabbitmqMiddleware, err := middleware.NewMiddleware(cfg)
	if err != nil {
		log.Fatalf("Failed to create RabbitMQ middleware: %v", err)
	}

	// Initialize RabbitMQ topology manager
	tm := middleware.NewTopologyManager(cfg, rabbitmqMiddleware)

	tm.GetMiddleware().DeclareExchange(config.CONNECTION_EXCHANGE, "fanout", true)

	// Initialize session repository
	sessionRepository := repository.NewSessionRepository(database)

	// Initialize connection service
	connectionService := service.NewConnectionService(rabbitmqMiddleware, tm, cfg, sessionRepository)

	// Initialize session service
	sessionService := service.NewSessionService(sessionRepository)

	// Initialize connection controller
	connectionController := controller.NewConnectionController(connectionService)

	// Initialize session controller
	sessionController := controller.NewSessionController(sessionService)

	// Initialize all routes
	InitializeRoutes(r, connectionController, sessionController)

	// Swagger documentation
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	slog.Info("Connection Service router initialized successfully")
	return r
}

func InitializeRoutes(
	r *gin.Engine,
	connectionController *controller.ConnectionController,
	sessionController *controller.SessionController,
) {
	InitializeUserRoutes(r, connectionController)
	InitializeSessionRoutes(r, sessionController)
}

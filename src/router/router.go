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

	// Initialize RabbitMQ publisher
	publisher, err := rabbitmq.NewAMQPPublisherFromConfig(config)
	if err != nil {
		log.Fatalf("Failed to create RabbitMQ publisher: %v", err)
	}

	// Setup RabbitMQ queues and bindings
	if err := setupConnectionQueues(publisher); err != nil {
		log.Fatalf("Failed to setup RabbitMQ queues: %v", err)
	}

	// Initialize services
	connectionService := service.NewConnectionService(publisher)

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

func setupConnectionQueues(publisher *rabbitmq.AMQPPublisher) error {
	exchangeName := config.CONNECTION_EXCHANGE
	queues := []string{
		"data-dispatcher-connections",
		"calibration-service-connections",
	}

	slog.Info("Setting up RabbitMQ queues and bindings", "exchange", exchangeName, "queues", queues)

	// Declare the exchange (fanout type for broadcasting)
	if err := publisher.DeclareExchange(exchangeName, "fanout"); err != nil {
		return err
	}

	// Declare queues and bind them to the exchange
	for _, queueName := range queues {
		if err := publisher.DeclareQueue(queueName); err != nil {
			return err
		}

		if err := publisher.BindQueue(queueName, exchangeName); err != nil {
			return err
		}

		slog.Info("Queue created and bound to exchange", "queue", queueName, "exchange", exchangeName)
	}

	return nil
}

package router

import (
	"auth-gateway/src/controller"

	docs "auth-gateway/src/docs"

	. "auth-gateway/src/middleware"

	"github.com/gin-contrib/cors"
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
	router := gin.New()

	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept"},
		AllowCredentials: true,
	}))

	tokensCtrl := controller.NewTokenController()
	usersCtrl := controller.NewUserController()
	adminsCtrl := controller.NewAdminController()

	docs.SwaggerInfo.BasePath = "/"

	tokens_group := router.Group("/tokens")
	{
		tokens_group.POST("/create", AdminAuthRequiredMiddleware(), tokensCtrl.CreateToken)
		tokens_group.GET("/", AdminAuthRequiredMiddleware(), tokensCtrl.GetTokens)
	}
	users_group := router.Group("/users")
	{
		users_group.POST("/create", AdminAuthRequiredMiddleware(), usersCtrl.CreateUser)
		users_group.POST("/connect", UserAuthRequiredMiddleware(), usersCtrl.Connect)
		users_group.GET("/", AdminAuthRequiredMiddleware(), usersCtrl.GetUsers)
		users_group.GET("/:user_id", UserAuthRequiredMiddleware(), usersCtrl.GetUserByID)
	}
	admins_group := router.Group("/admins")
	{
		admins_group.POST("/invite", AdminAuthRequiredMiddleware(), adminsCtrl.InviteAdmin)
		admins_group.POST("/signup", adminsCtrl.Signup)
		admins_group.POST("/login", adminsCtrl.Login)
		admins_group.GET("/", AdminAuthRequiredMiddleware(), adminsCtrl.ListAdmins)
		admins_group.GET("/:admin_id", AdminAuthRequiredMiddleware(), adminsCtrl.GetAdmin)
	}

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	// router.POST("/test-webhook", controller.TestWebhook)
	return router, nil
}

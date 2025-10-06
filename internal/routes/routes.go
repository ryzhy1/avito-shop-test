package routes

import (
	"avito-shop/internal/handlers"
	"avito-shop/internal/middlewares"
	"github.com/go-openapi/runtime/middleware"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"time"
)

func InitRoutes(authHandler *handlers.AuthHandler, userHandler *handlers.UserHandler, authMiddleware *middlewares.AuthMiddleware) *gin.Engine {
	router := gin.Default()

	_ = router.SetTrustedProxies(nil)

	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:8080"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	router.StaticFile("/swagger.yaml", "./swagger.yaml")

	opts := middleware.SwaggerUIOpts{SpecURL: "/swagger.yaml"}
	sh := middleware.SwaggerUI(opts, nil)

	router.GET("/swagger/*any", func(c *gin.Context) {
		sh.ServeHTTP(c.Writer, c.Request)
	})

	api := router.Group("/api")

	// паблик роут
	api.POST("/auth", authHandler.Auth)
	api.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	// защищенные роуты
	api.Use(authMiddleware.Handle())
	{
		api.GET("/info", userHandler.GetUserInfo)
		api.POST("/sendCoins", userHandler.TransferCoins)
		api.GET("/buy/:item", userHandler.BuyMerch)
	}

	return router
}

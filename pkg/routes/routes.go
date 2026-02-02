package routes

import (
	"chalk-api/pkg/config"
	"chalk-api/pkg/handlers"

	"github.com/gin-gonic/gin"
)

// SetupRouter initializes and returns the Gin router with all routes
func SetupRouter(h *handlers.HandlersCollection, cfg config.Environment) *gin.Engine {
	router := gin.New()
	router.Use(gin.Recovery())

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "healthy",
			"version": config.DeployVersion,
		})
	})

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Add your route groups here
		// Example:
		// users := v1.Group("/users")
		// {
		// 	users.GET("", h.UserHandler.GetUsers)
		// 	users.POST("", h.UserHandler.CreateUser)
		// }
		_ = v1 // Remove this line when you add routes
	}

	return router
}

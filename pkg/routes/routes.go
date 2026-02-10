package routes

import (
	"chalk-api/pkg/config"
	"chalk-api/pkg/handlers"
	"chalk-api/pkg/middleware"

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
		auth := v1.Group("/auth")
		{
			auth.POST("/register", h.Auth.Register)
			auth.POST("/login", h.Auth.Login)
			auth.POST("/refresh", h.Auth.Refresh)
		}

		// Public invite preview endpoint for deep links before authentication.
		invites := v1.Group("/invites")
		{
			invites.GET("/:code", h.Invite.GetPreview)
		}

		protected := v1.Group("")
		protected.Use(middleware.AuthMiddleware(cfg.JWTSecret))
		{
			protected.POST("/auth/logout", h.Auth.Logout)
			protected.POST("/invites/accept", h.Invite.Accept)

			users := protected.Group("/users")
			{
				users.GET("/me", h.User.GetMe)
				users.PATCH("/me", h.User.UpdateMe)
			}

			coaches := protected.Group("/coaches")
			{
				coaches.GET("/me", h.Coach.GetMyProfile)
				coaches.PUT("/me", h.Coach.UpsertMyProfile)
				coaches.POST("/invite-codes", h.Coach.CreateInviteCode)
				coaches.GET("/invite-codes", h.Coach.ListInviteCodes)
				coaches.PATCH("/invite-codes/:id/deactivate", h.Coach.DeactivateInviteCode)
			}
		}
	}

	return router
}

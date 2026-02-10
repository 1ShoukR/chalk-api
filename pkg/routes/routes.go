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

				coaches.POST("/templates", h.Workout.CreateTemplate)
				coaches.GET("/templates", h.Workout.ListMyTemplates)
				coaches.GET("/templates/:id", h.Workout.GetMyTemplate)
				coaches.PATCH("/templates/:id", h.Workout.UpdateMyTemplate)

				coaches.POST("/workouts/assign", h.Workout.AssignWorkout)
			}

			workouts := protected.Group("/workouts")
			{
				workouts.GET("/me", h.Workout.ListMyWorkouts)
				workouts.GET("/me/:id", h.Workout.GetMyWorkout)
				workouts.POST("/me/:id/start", h.Workout.StartMyWorkout)
				workouts.POST("/me/:id/complete", h.Workout.CompleteMyWorkout)

				workouts.POST("/exercises/:id/complete", h.Workout.MarkExerciseCompleted)
				workouts.POST("/exercises/:id/skip", h.Workout.SkipExercise)
				workouts.POST("/exercises/:id/logs", h.Workout.CreateExerciseLog)
				workouts.PATCH("/logs/:id", h.Workout.UpdateWorkoutLog)
			}
		}
	}

	return router
}

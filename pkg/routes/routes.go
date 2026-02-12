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

		subscriptions := v1.Group("/subscriptions")
		{
			subscriptions.POST("/revenuecat/webhook", h.Subscription.RevenueCatWebhook)
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

				coaches.GET("/me/availability", h.Session.GetMyAvailability)
				coaches.PUT("/me/availability", h.Session.SetMyAvailability)
				coaches.POST("/me/availability-overrides", h.Session.CreateAvailabilityOverride)
				coaches.GET("/me/availability-overrides", h.Session.ListAvailabilityOverrides)
				coaches.DELETE("/me/availability-overrides/:id", h.Session.DeleteAvailabilityOverride)

				coaches.POST("/me/session-types", h.Session.CreateSessionType)
				coaches.GET("/me/session-types", h.Session.ListSessionTypes)
				coaches.PATCH("/me/session-types/:id", h.Session.UpdateSessionType)
				coaches.GET("/me/sessions", h.Session.ListCoachSessions)

				coaches.POST("/templates", h.Workout.CreateTemplate)
				coaches.GET("/templates", h.Workout.ListMyTemplates)
				coaches.GET("/templates/:id", h.Workout.GetMyTemplate)
				coaches.PATCH("/templates/:id", h.Workout.UpdateMyTemplate)

				coaches.POST("/workouts/assign", h.Workout.AssignWorkout)
				coaches.GET("/:id/bookable-slots", h.Session.GetBookableSlots)
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

			messages := protected.Group("/messages")
			{
				messages.GET("/conversations", h.Message.ListConversations)
				messages.POST("/conversations", h.Message.GetOrCreateConversation)
				messages.GET("/conversations/:id", h.Message.GetConversation)
				messages.GET("/conversations/:id/messages", h.Message.ListMessages)
				messages.POST("/conversations/:id/messages", h.Message.SendMessage)
				messages.POST("/conversations/:id/read", h.Message.MarkAsRead)
				messages.GET("/unread-count", h.Message.GetUnreadCount)
			}

			sessions := protected.Group("/sessions")
			{
				sessions.POST("/book", h.Session.BookSession)
				sessions.GET("/me", h.Session.ListMySessions)
				sessions.POST("/:id/cancel", h.Session.CancelSession)
				sessions.POST("/:id/complete", h.Session.CompleteSession)
				sessions.POST("/:id/no-show", h.Session.MarkNoShow)
			}

			protected.GET("/subscriptions/me", h.Subscription.GetMySubscription)
			protected.GET("/features/:feature/access", h.Subscription.CheckFeatureAccess)
		}
	}

	return router
}

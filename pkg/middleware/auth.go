package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware is a placeholder for authentication middleware
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Implement authentication logic
		// Example: Check for valid JWT token, API key, etc.

		// For now, just pass through
		c.Next()
	}
}

// APIKeyMiddleware validates API key from header
func APIKeyMiddleware(apiKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.GetHeader("X-API-Key")
		if key == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "API key required"})
			return
		}
		if key != apiKey {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid API key"})
			return
		}
		c.Next()
	}
}

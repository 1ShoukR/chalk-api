package utils

import (
	"crypto/rand"
	"encoding/hex"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
)

// GenerateRandomString generates a random hex string of given length
func GenerateRandomString(length int) (string, error) {
	bytes := make([]byte, length/2)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// Slugify converts a string to a URL-friendly slug
func Slugify(s string) string {
	s = strings.ToLower(s)
	s = strings.TrimSpace(s)
	reg := regexp.MustCompile("[^a-z0-9]+")
	s = reg.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	return s
}

// Contains checks if a string slice contains a value
func Contains(slice []string, val string) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}

// StringPtr returns a pointer to a string
func StringPtr(s string) *string {
	return &s
}

// IntPtr returns a pointer to an int
func IntPtr(i int) *int {
	return &i
}

// BoolPtr returns a pointer to a bool
func BoolPtr(b bool) *bool {
	return &b
}

// GetUserIDFromContext reads user_id from Gin context and converts it to uint.
func GetUserIDFromContext(c *gin.Context) (uint, bool) {
	value, exists := c.Get("user_id")
	if !exists {
		return 0, false
	}

	switch v := value.(type) {
	case uint:
		return v, true
	case int:
		if v < 0 {
			return 0, false
		}
		return uint(v), true
	case int64:
		if v < 0 {
			return 0, false
		}
		return uint(v), true
	case float64:
		if v < 0 {
			return 0, false
		}
		return uint(v), true
	default:
		return 0, false
	}
}

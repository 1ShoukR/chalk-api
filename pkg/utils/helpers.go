package utils

import (
	"crypto/rand"
	"encoding/hex"
	"regexp"
	"strings"
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

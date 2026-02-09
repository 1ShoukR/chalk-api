package stores

import "fmt"

// Cache key patterns - centralized to avoid collisions and typos
// Format: domain:entity:identifier

// User keys
func KeyUserProfile(userID uint) string {
	return fmt.Sprintf("user:profile:%d", userID)
}

func KeyUserByEmail(email string) string {
	return fmt.Sprintf("user:email:%s", email)
}

// Coach keys
func KeyCoachProfile(coachID uint) string {
	return fmt.Sprintf("coach:profile:%d", coachID)
}

func KeyCoachStats(coachID uint) string {
	return fmt.Sprintf("coach:stats:%d", coachID)
}

// Subscription keys
func KeySubscription(userID uint) string {
	return fmt.Sprintf("subscription:user:%d", userID)
}

// Exercise keys
func KeyExercise(exerciseID uint) string {
	return fmt.Sprintf("exercise:%d", exerciseID)
}

func KeyExerciseList(coachID uint, page int) string {
	return fmt.Sprintf("exercise:list:%d:%d", coachID, page)
}

func KeySystemExercises(page int) string {
	return fmt.Sprintf("exercise:system:%d", page)
}

// Nutrition / Food keys (Open Food Facts cache)
func KeyFoodByBarcode(barcode string) string {
	return fmt.Sprintf("food:barcode:%s", barcode)
}

func KeyFoodByExternalID(source, externalID string) string {
	return fmt.Sprintf("food:external:%s:%s", source, externalID)
}

func KeyFoodSearch(query string, page int) string {
	return fmt.Sprintf("food:search:%s:%d", query, page)
}

// Session/availability keys
func KeyCoachAvailability(coachID uint) string {
	return fmt.Sprintf("coach:availability:%d", coachID)
}

// Security keys - for rate limiting and attempt tracking
func KeyLoginAttempts(email string) string {
	return fmt.Sprintf("security:login:attempts:%s", email)
}

func KeyPasswordResetAttempts(email string) string {
	return fmt.Sprintf("security:reset:attempts:%s", email)
}

func KeyMagicLinkAttempts(email string) string {
	return fmt.Sprintf("security:magic:attempts:%s", email)
}

// Rate limiting keys
func KeyRateLimit(identifier, action string) string {
	return fmt.Sprintf("ratelimit:%s:%s", action, identifier)
}

func KeyAPIRateLimit(userID uint, endpoint string) string {
	return fmt.Sprintf("ratelimit:api:%d:%s", userID, endpoint)
}

// JWT blacklist (for logout)
func KeyJWTBlacklist(tokenID string) string {
	return fmt.Sprintf("jwt:blacklist:%s", tokenID)
}

// Refresh token validation
func KeyRefreshToken(tokenHash string) string {
	return fmt.Sprintf("auth:refresh:%s", tokenHash)
}

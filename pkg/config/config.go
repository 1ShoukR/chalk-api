package config

import (
	"log/slog"
	"os"

	"github.com/Netflix/go-env"
	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
)

type Environment struct {
	// Server
	Port    int    `env:"PORT,default=8080"`
	RunMode string `env:"RUN_MODE,default=local"`

	// Database - supports both individual vars and DATABASE_URL (Railway style)
	DatabaseURL string `env:"DATABASE_URL"`
	DBHost      string `env:"DB_HOST"`
	DBPort      string `env:"DB_PORT"`
	DBUser      string `env:"DB_USER"`
	DBPassword  string `env:"DB_PASSWORD"`
	DBName      string `env:"DB_NAME"`

	// Redis (optional)
	RedisURL string `env:"REDIS_URL"`

	// JWT Auth (you'll configure these later)
	JWTSecret          string `env:"JWT_SECRET"`
	JWTExpirationHours int    `env:"JWT_EXPIRATION_HOURS,default=24"`

	// OAuth (you'll configure these later)
	GoogleClientID       string `env:"GOOGLE_CLIENT_ID"`
	GoogleClientSecret   string `env:"GOOGLE_CLIENT_SECRET"`
	FacebookClientID     string `env:"FACEBOOK_CLIENT_ID"`
	FacebookClientSecret string `env:"FACEBOOK_CLIENT_SECRET"`
	AppleClientID        string `env:"APPLE_CLIENT_ID"`
	AppleTeamID          string `env:"APPLE_TEAM_ID"`
	AppleKeyID           string `env:"APPLE_KEY_ID"`

	// RevenueCat (subscription management)
	RevenueCatAPIKey               string `env:"REVENUECAT_API_KEY"`
	RevenueCatWebhookAuthorization string `env:"REVENUECAT_WEBHOOK_AUTHORIZATION"`
	// Deprecated fallback for older env naming.
	RevenueCatWebhookSecret string `env:"REVENUECAT_WEBHOOK_SECRET"`

	// Expo Push Notifications
	ExpoAccessToken string `env:"EXPO_ACCESS_TOKEN"`

	// Open Food Facts (no auth required, but we track user-agent)
	OpenFoodFactsUserAgent string `env:"OPENFOODFACTS_USER_AGENT,default=ChalkAPI/1.0"`

	// Outbox worker tuning
	OutboxPollIntervalSeconds   int `env:"OUTBOX_POLL_INTERVAL_SECONDS,default=2"`
	OutboxBatchSize             int `env:"OUTBOX_BATCH_SIZE,default=25"`
	OutboxMaxAttempts           int `env:"OUTBOX_MAX_ATTEMPTS,default=8"`
	OutboxStuckThresholdSeconds int `env:"OUTBOX_STUCK_THRESHOLD_SECONDS,default=600"`
}

var DeployVersion = "dev"

func LoadConfig() (Environment, error) {
	var cfg Environment

	// Load .env file (ignored in production, Railway sets env vars directly)
	if err := godotenv.Load(".env"); err != nil {
		slog.Debug("No .env file found, using environment variables")
	}

	_, err := env.UnmarshalFromEnviron(&cfg)
	if err != nil {
		slog.Error("Problem reading environment config", "err", err)
		return cfg, err
	}

	// Validate required fields based on environment
	if err := validateConfig(&cfg); err != nil {
		return cfg, err
	}

	return cfg, nil
}

func validateConfig(cfg *Environment) error {
	validate := validator.New()

	// If DATABASE_URL is set (Railway), we don't need individual DB vars
	if cfg.DatabaseURL == "" {
		// Validate individual DB fields
		if cfg.DBHost == "" || cfg.DBUser == "" || cfg.DBName == "" {
			slog.Warn("Database configuration incomplete - set DATABASE_URL or individual DB_* vars")
		}
	}

	return validate.Struct(cfg)
}

// IsDevelopment returns true if running in development mode
func (e *Environment) IsDevelopment() bool {
	return e.RunMode == "local" || e.RunMode == "development"
}

// IsProduction returns true if running in production mode
func (e *Environment) IsProduction() bool {
	return e.RunMode == "production"
}

// GetPort returns the port, checking for Railway's PORT env var
func GetPort() string {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	return port
}

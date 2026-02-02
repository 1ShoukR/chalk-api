package db

import (
	"chalk-api/pkg/config"
	"chalk-api/pkg/models"
	"fmt"
	"log/slog"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var db *gorm.DB

// InitializeDatabase creates and returns a database connection
func InitializeDatabase(cfg config.Environment) (*gorm.DB, error) {
	var dsn string

	// Railway provides DATABASE_URL, use it if available
	if cfg.DatabaseURL != "" {
		dsn = cfg.DatabaseURL
		slog.Info("Using DATABASE_URL for database connection")
	} else {
		// Fall back to individual connection parameters (local dev)
		sslMode := "disable"
		if cfg.IsProduction() {
			sslMode = "require"
		}
		dsn = fmt.Sprintf(
			"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
			cfg.DBHost, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBPort, sslMode,
		)
	}

	// Configure logger based on environment
	logLevel := logger.Info
	if cfg.IsProduction() {
		logLevel = logger.Warn
	}

	var err error
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	slog.Info("Database connection established")
	return db, nil
}

// CloseDatabase closes the database connection
func CloseDatabase() {
	if db != nil {
		sqlDB, err := db.DB()
		if err != nil {
			slog.Error("Failed to get underlying sql.DB", "error", err)
			return
		}
		if err := sqlDB.Close(); err != nil {
			slog.Error("Failed to close database connection", "error", err)
		}
		slog.Info("Database connection closed")
	}
}

// GetDB returns the database instance
func GetDB() *gorm.DB {
	return db
}

// RunMigrations runs all database migrations
// Add your models here as you create them
func RunMigrations(db *gorm.DB) error {
	slog.Info("Running database migrations...")

	// Auto-migrate models
	err := db.AutoMigrate(
		// User models
		&models.User{},
		&models.Profile{},
		&models.OAuthProvider{},
		&models.RefreshToken{},
		&models.DeviceToken{},
		&models.PasswordReset{},
		&models.EmailVerification{},
		&models.MagicLink{},
		// Coach models
		&models.CoachProfile{},
		&models.Certification{},
		&models.CoachLocation{},
		&models.CoachStats{},
		// Client models
		&models.ClientProfile{},
		&models.InviteCode{},
		&models.ClientIntakeForm{},
		// Subscription models
		&models.Subscription{},
		&models.SubscriptionEvent{},
	)
	if err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	// Add composite unique index for OAuth providers
	// Ensures one user can't link the same provider account twice
	if err := db.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS idx_oauth_provider_user 
		ON oauth_providers(provider, provider_user_id)
	`).Error; err != nil {
		return fmt.Errorf("failed to create oauth provider index: %w", err)
	}

	// Add composite unique index for ClientProfiles
	// Ensures one user can only be a client of a specific coach once
	if err := db.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS idx_user_coach 
		ON client_profiles(user_id, coach_id)
	`).Error; err != nil {
		return fmt.Errorf("failed to create client profile index: %w", err)
	}

	// Add indexes for efficient cleanup queries
	if err := db.Exec(`CREATE INDEX IF NOT EXISTS idx_refresh_tokens_cleanup ON refresh_tokens(expires_at, revoked)`).Error; err != nil {
		return fmt.Errorf("failed to create refresh tokens cleanup index: %w", err)
	}

	if err := db.Exec(`CREATE INDEX IF NOT EXISTS idx_password_resets_cleanup ON password_resets(expires_at, used)`).Error; err != nil {
		return fmt.Errorf("failed to create password resets cleanup index: %w", err)
	}

	if err := db.Exec(`CREATE INDEX IF NOT EXISTS idx_email_verifications_cleanup ON email_verifications(expires_at, used)`).Error; err != nil {
		return fmt.Errorf("failed to create email verifications cleanup index: %w", err)
	}

	if err := db.Exec(`CREATE INDEX IF NOT EXISTS idx_magic_links_cleanup ON magic_links(expires_at, used)`).Error; err != nil {
		return fmt.Errorf("failed to create magic links cleanup index: %w", err)
	}

	slog.Info("Database migrations completed")
	return nil
}

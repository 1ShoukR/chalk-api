package seeds

import (
	"log/slog"

	"gorm.io/gorm"
)

// InitializeSeedData seeds the database with initial data
func InitializeSeedData(db *gorm.DB) error {
	slog.Info("Initializing seed data...")

	// Add your seed logic here
	// Example:
	// if err := seedUsers(db); err != nil {
	// 	return err
	// }

	slog.Info("Seed data initialization completed")
	return nil
}

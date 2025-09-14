package infrastructure

import (
	"fmt"

	"gorm.io/gorm"
)

// MigrateAuthTables creates the authentication-related database tables
func MigrateAuthTables(db *gorm.DB) error {
	// Auto-migrate all authentication models
	err := db.AutoMigrate(
		&SessionModel{},
		&PasswordCredentialModel{},
		&PasswordResetTokenModel{},
		&ExternalIdentityModel{},
	)
	if err != nil {
		return fmt.Errorf("failed to auto-migrate authentication tables: %w", err)
	}

	// Create unique index for external identity provider + external_id combination
	err = db.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS idx_external_identity_provider_external_id 
		ON external_identity_models (provider, external_id)
	`).Error
	if err != nil {
		return fmt.Errorf("failed to create unique index for external identity: %w", err)
	}

	return nil
}

// DropAuthTables drops all authentication-related database tables (for testing)
func DropAuthTables(db *gorm.DB) error {
	err := db.Migrator().DropTable(
		&SessionModel{},
		&PasswordCredentialModel{},
		&PasswordResetTokenModel{},
		&ExternalIdentityModel{},
	)
	if err != nil {
		return fmt.Errorf("failed to drop authentication tables: %w", err)
	}

	return nil
}

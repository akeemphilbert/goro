package infrastructure

import (
	"fmt"

	"github.com/google/wire"
	"gorm.io/gorm"
)

// ProvideUserDatabase provides a GORM database instance with user models migrated
func ProvideUserDatabase(db *gorm.DB) (*gorm.DB, error) {
	// Ensure user models are migrated
	if err := MigrateUserModels(db); err != nil {
		return nil, fmt.Errorf("failed to migrate user models: %w", err)
	}

	// Seed system roles
	if err := SeedSystemRoles(db); err != nil {
		return nil, fmt.Errorf("failed to seed system roles: %w", err)
	}

	return db, nil
}

// UserInfrastructureProviderSet provides all user infrastructure dependencies
var UserInfrastructureProviderSet = wire.NewSet(
	ProvideUserDatabase,
)

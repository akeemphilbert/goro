package infrastructure

import (
	"fmt"
	"os"

	"github.com/akeemphilbert/goro/internal/user/domain"
	"github.com/google/wire"
	"gorm.io/gorm"
)

// ProvideUserDatabase provides a GORM database instance with user models migrated
func ProvideUserDatabase(db *gorm.DB) (*gorm.DB, error) {
	if db == nil {
		return nil, fmt.Errorf("database cannot be nil")
	}

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

// Repository Providers (Read-only)
func ProvideUserRepository(db *gorm.DB) (domain.UserRepository, error) {
	if db == nil {
		return nil, fmt.Errorf("database cannot be nil")
	}
	return NewGormUserRepository(db), nil
}

func ProvideAccountRepository(db *gorm.DB) (domain.AccountRepository, error) {
	if db == nil {
		return nil, fmt.Errorf("database cannot be nil")
	}
	return NewGormAccountRepository(db), nil
}

func ProvideRoleRepository(db *gorm.DB) (domain.RoleRepository, error) {
	if db == nil {
		return nil, fmt.Errorf("database cannot be nil")
	}
	return NewGormRoleRepository(db), nil
}

func ProvideAccountMemberRepository(db *gorm.DB) (domain.AccountMemberRepository, error) {
	if db == nil {
		return nil, fmt.Errorf("database cannot be nil")
	}
	return NewGormAccountMemberRepository(db), nil
}

func ProvideInvitationRepository(db *gorm.DB) (domain.InvitationRepository, error) {
	if db == nil {
		return nil, fmt.Errorf("database cannot be nil")
	}
	return NewGormInvitationRepository(db), nil
}

// Write Repository Providers
func ProvideUserWriteRepository(db *gorm.DB) (domain.UserWriteRepository, error) {
	if db == nil {
		return nil, fmt.Errorf("database cannot be nil")
	}
	return NewGormUserWriteRepository(db), nil
}

func ProvideAccountWriteRepository(db *gorm.DB) (domain.AccountWriteRepository, error) {
	if db == nil {
		return nil, fmt.Errorf("database cannot be nil")
	}
	return NewGormAccountWriteRepository(db), nil
}

func ProvideAccountMemberWriteRepository(db *gorm.DB) (domain.AccountMemberWriteRepository, error) {
	if db == nil {
		return nil, fmt.Errorf("database cannot be nil")
	}
	return NewGormAccountMemberWriteRepository(db), nil
}

func ProvideInvitationWriteRepository(db *gorm.DB) (domain.InvitationWriteRepository, error) {
	if db == nil {
		return nil, fmt.Errorf("database cannot be nil")
	}
	return NewGormInvitationWriteRepository(db), nil
}

// Service Infrastructure Providers
func ProvideWebIDGenerator(baseURL string) (domain.WebIDGenerator, error) {
	return NewWebIDGenerator(baseURL), nil
}

func ProvideFileStorage(baseDir string) (domain.FileStorage, error) {
	// Ensure base directory exists
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create base directory: %w", err)
	}

	return NewFileStorageAdapter(baseDir)
}

// Provider Sets
var UserInfrastructureProviderSet = wire.NewSet(
	ProvideUserDatabase,
	ProvideWebIDGenerator,
	ProvideFileStorage,
)

var UserRepositoryProviderSet = wire.NewSet(
	ProvideUserRepository,
	ProvideAccountRepository,
	ProvideRoleRepository,
	ProvideAccountMemberRepository,
	ProvideInvitationRepository,
)

var UserWriteRepositoryProviderSet = wire.NewSet(
	ProvideUserWriteRepository,
	ProvideAccountWriteRepository,
	ProvideAccountMemberWriteRepository,
	ProvideInvitationWriteRepository,
)

// Complete provider set for user management
var UserManagementProviderSet = wire.NewSet(
	UserInfrastructureProviderSet,
	UserRepositoryProviderSet,
	UserWriteRepositoryProviderSet,
)

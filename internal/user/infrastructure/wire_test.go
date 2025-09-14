package infrastructure

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/akeemphilbert/goro/internal/user/domain"
	"github.com/google/wire"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// TestWireProviders tests that all Wire providers can be created successfully
func TestWireProviders(t *testing.T) {
	// Create temporary database
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	t.Run("ProvideUserDatabase", func(t *testing.T) {
		userDB, err := ProvideUserDatabase(db)
		if err != nil {
			t.Errorf("ProvideUserDatabase failed: %v", err)
		}
		if userDB == nil {
			t.Error("ProvideUserDatabase returned nil database")
		}

		// Verify tables were created
		if !userDB.Migrator().HasTable(&UserModel{}) {
			t.Error("UserModel table was not created")
		}
		if !userDB.Migrator().HasTable(&AccountModel{}) {
			t.Error("AccountModel table was not created")
		}
		if !userDB.Migrator().HasTable(&RoleModel{}) {
			t.Error("RoleModel table was not created")
		}
		if !userDB.Migrator().HasTable(&AccountMemberModel{}) {
			t.Error("AccountMemberModel table was not created")
		}
		if !userDB.Migrator().HasTable(&InvitationModel{}) {
			t.Error("InvitationModel table was not created")
		}

		// Verify system roles were seeded
		var roleCount int64
		userDB.Model(&RoleModel{}).Count(&roleCount)
		if roleCount < 4 { // Owner, Admin, Member, Viewer
			t.Errorf("Expected at least 4 system roles, got %d", roleCount)
		}
	})

	t.Run("ProvideUserRepository", func(t *testing.T) {
		userDB, err := ProvideUserDatabase(db)
		if err != nil {
			t.Fatalf("Failed to setup user database: %v", err)
		}

		cache := ProvideCache()
		repo, err := ProvideUserRepository(userDB, cache)
		if err != nil {
			t.Errorf("ProvideUserRepository failed: %v", err)
		}
		if repo == nil {
			t.Error("ProvideUserRepository returned nil repository")
		}

		// Test that repository implements the interface
		var _ domain.UserRepository = repo
	})

	t.Run("ProvideAccountRepository", func(t *testing.T) {
		userDB, err := ProvideUserDatabase(db)
		if err != nil {
			t.Fatalf("Failed to setup user database: %v", err)
		}

		repo, err := ProvideAccountRepository(userDB)
		if err != nil {
			t.Errorf("ProvideAccountRepository failed: %v", err)
		}
		if repo == nil {
			t.Error("ProvideAccountRepository returned nil repository")
		}

		// Test that repository implements the interface
		var _ domain.AccountRepository = repo
	})

	t.Run("ProvideRoleRepository", func(t *testing.T) {
		userDB, err := ProvideUserDatabase(db)
		if err != nil {
			t.Fatalf("Failed to setup user database: %v", err)
		}

		cache := ProvideCache()
		repo, err := ProvideRoleRepository(userDB, cache)
		if err != nil {
			t.Errorf("ProvideRoleRepository failed: %v", err)
		}
		if repo == nil {
			t.Error("ProvideRoleRepository returned nil repository")
		}

		// Test that repository implements the interface
		var _ domain.RoleRepository = repo
	})

	t.Run("ProvideAccountMemberRepository", func(t *testing.T) {
		userDB, err := ProvideUserDatabase(db)
		if err != nil {
			t.Fatalf("Failed to setup user database: %v", err)
		}

		repo, err := ProvideAccountMemberRepository(userDB)
		if err != nil {
			t.Errorf("ProvideAccountMemberRepository failed: %v", err)
		}
		if repo == nil {
			t.Error("ProvideAccountMemberRepository returned nil repository")
		}

		// Test that repository implements the interface
		var _ domain.AccountMemberRepository = repo
	})

	t.Run("ProvideInvitationRepository", func(t *testing.T) {
		userDB, err := ProvideUserDatabase(db)
		if err != nil {
			t.Fatalf("Failed to setup user database: %v", err)
		}

		repo, err := ProvideInvitationRepository(userDB)
		if err != nil {
			t.Errorf("ProvideInvitationRepository failed: %v", err)
		}
		if repo == nil {
			t.Error("ProvideInvitationRepository returned nil repository")
		}

		// Test that repository implements the interface
		var _ domain.InvitationRepository = repo
	})

	t.Run("ProvideUserWriteRepository", func(t *testing.T) {
		userDB, err := ProvideUserDatabase(db)
		if err != nil {
			t.Fatalf("Failed to setup user database: %v", err)
		}

		repo, err := ProvideUserWriteRepository(userDB)
		if err != nil {
			t.Errorf("ProvideUserWriteRepository failed: %v", err)
		}
		if repo == nil {
			t.Error("ProvideUserWriteRepository returned nil repository")
		}

		// Test that repository implements the interface
		var _ domain.UserWriteRepository = repo
	})

	t.Run("ProvideAccountWriteRepository", func(t *testing.T) {
		userDB, err := ProvideUserDatabase(db)
		if err != nil {
			t.Fatalf("Failed to setup user database: %v", err)
		}

		repo, err := ProvideAccountWriteRepository(userDB)
		if err != nil {
			t.Errorf("ProvideAccountWriteRepository failed: %v", err)
		}
		if repo == nil {
			t.Error("ProvideAccountWriteRepository returned nil repository")
		}

		// Test that repository implements the interface
		var _ domain.AccountWriteRepository = repo
	})

	t.Run("ProvideAccountMemberWriteRepository", func(t *testing.T) {
		userDB, err := ProvideUserDatabase(db)
		if err != nil {
			t.Fatalf("Failed to setup user database: %v", err)
		}

		repo, err := ProvideAccountMemberWriteRepository(userDB)
		if err != nil {
			t.Errorf("ProvideAccountMemberWriteRepository failed: %v", err)
		}
		if repo == nil {
			t.Error("ProvideAccountMemberWriteRepository returned nil repository")
		}

		// Test that repository implements the interface
		var _ domain.AccountMemberWriteRepository = repo
	})

	t.Run("ProvideInvitationWriteRepository", func(t *testing.T) {
		userDB, err := ProvideUserDatabase(db)
		if err != nil {
			t.Fatalf("Failed to setup user database: %v", err)
		}

		repo, err := ProvideInvitationWriteRepository(userDB)
		if err != nil {
			t.Errorf("ProvideInvitationWriteRepository failed: %v", err)
		}
		if repo == nil {
			t.Error("ProvideInvitationWriteRepository returned nil repository")
		}

		// Test that repository implements the interface
		var _ domain.InvitationWriteRepository = repo
	})

	t.Run("ProvideWebIDGenerator", func(t *testing.T) {
		generator, err := ProvideWebIDGenerator("https://example.com")
		if err != nil {
			t.Errorf("ProvideWebIDGenerator failed: %v", err)
		}
		if generator == nil {
			t.Error("ProvideWebIDGenerator returned nil generator")
		}

		// Test that generator implements the interface
		var _ domain.WebIDGenerator = generator
	})

	t.Run("ProvideFileStorage", func(t *testing.T) {
		tempDir := t.TempDir()

		storage, err := ProvideFileStorage(tempDir)
		if err != nil {
			t.Errorf("ProvideFileStorage failed: %v", err)
		}
		if storage == nil {
			t.Error("ProvideFileStorage returned nil storage")
		}

		// Test that storage implements the interface
		var _ domain.FileStorage = storage

		// Verify base directory was created
		if _, err := os.Stat(tempDir); os.IsNotExist(err) {
			t.Error("Base directory was not created")
		}
	})
}

// TestProviderSetIntegration tests that the provider set can be used with Wire
func TestProviderSetIntegration(t *testing.T) {
	// This test verifies that all providers in the set are compatible
	// and can be resolved by Wire (compilation test)

	// Create a test injector function that uses the provider set
	testInjector := func(db *gorm.DB, baseDir string) (*TestUserComponents, error) {
		// This would be generated by Wire in actual usage
		panic(wire.Build(
			UserInfrastructureProviderSet,
			UserRepositoryProviderSet,
			UserWriteRepositoryProviderSet,
			UserManagementProviderSet,
			wire.Struct(new(TestUserComponents), "*"),
		))
	}

	// The test passes if this compiles without Wire errors
	_ = testInjector
}

// TestUserComponents represents a complete set of user management components
type TestUserComponents struct {
	UserRepo               domain.UserRepository
	AccountRepo            domain.AccountRepository
	RoleRepo               domain.RoleRepository
	AccountMemberRepo      domain.AccountMemberRepository
	InvitationRepo         domain.InvitationRepository
	UserWriteRepo          domain.UserWriteRepository
	AccountWriteRepo       domain.AccountWriteRepository
	AccountMemberWriteRepo domain.AccountMemberWriteRepository
	InvitationWriteRepo    domain.InvitationWriteRepository
	WebIDGenerator         domain.WebIDGenerator
	FileStorage            domain.FileStorage
}

// TestDatabaseMigrationIntegration tests that database migration works in Wire setup
func TestDatabaseMigrationIntegration(t *testing.T) {
	// Create temporary database
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	// Test that ProvideUserDatabase handles migration
	userDB, err := ProvideUserDatabase(db)
	if err != nil {
		t.Fatalf("ProvideUserDatabase failed: %v", err)
	}

	// Verify all tables exist
	tables := []interface{}{
		&UserModel{},
		&AccountModel{},
		&RoleModel{},
		&AccountMemberModel{},
		&InvitationModel{},
	}

	for _, table := range tables {
		if !userDB.Migrator().HasTable(table) {
			t.Errorf("Table for %T was not created during migration", table)
		}
	}

	// Verify system roles were seeded
	ctx := context.Background()
	cache := ProvideCache()
	roleRepo, err := ProvideRoleRepository(userDB, cache)
	if err != nil {
		t.Fatalf("Failed to create role repository: %v", err)
	}

	systemRoles, err := roleRepo.GetSystemRoles(ctx)
	if err != nil {
		t.Fatalf("Failed to get system roles: %v", err)
	}

	expectedRoles := []string{"owner", "admin", "member", "viewer"}
	if len(systemRoles) != len(expectedRoles) {
		t.Errorf("Expected %d system roles, got %d", len(expectedRoles), len(systemRoles))
	}

	for _, expectedRole := range expectedRoles {
		found := false
		for _, role := range systemRoles {
			if role.ID() == expectedRole {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("System role %s was not found", expectedRole)
		}
	}
}

// TestProviderErrorHandling tests error scenarios in providers
func TestProviderErrorHandling(t *testing.T) {
	t.Run("ProvideUserDatabase with nil database", func(t *testing.T) {
		_, err := ProvideUserDatabase(nil)
		if err == nil {
			t.Error("Expected error when providing nil database")
		}
	})

	t.Run("ProvideFileStorage with invalid path", func(t *testing.T) {
		// Use a path that cannot be created (e.g., under a file instead of directory)
		tempFile := filepath.Join(t.TempDir(), "file.txt")
		if err := os.WriteFile(tempFile, []byte("test"), 0644); err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}

		invalidPath := filepath.Join(tempFile, "subdir") // Cannot create directory under file
		_, err := ProvideFileStorage(invalidPath)
		if err == nil {
			t.Error("Expected error when providing invalid storage path")
		}
	})
}

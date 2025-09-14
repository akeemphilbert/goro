package integration

import (
	"context"
	"testing"

	"github.com/akeemphilbert/goro/internal/user/application"
	"github.com/akeemphilbert/goro/internal/user/domain"
	"github.com/akeemphilbert/goro/internal/user/infrastructure"
	pericarpdomain "github.com/akeemphilbert/pericarp/pkg/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// TestUserRegistrationWorkflow tests the complete user registration workflow
// from HTTP request to database persistence and file storage
func TestUserRegistrationWorkflow(t *testing.T) {
	// Setup test environment
	ctx := context.Background()
	tempDir := t.TempDir()

	// Setup database
	db := setupTestDatabase(t)

	// Setup file storage
	fileStorage := setupTestFileStorage(t, tempDir)

	// Setup repositories
	userRepo := infrastructure.NewGormUserRepository(db)
	userWriteRepo := infrastructure.NewGormUserWriteRepository(db)

	// Setup WebID generator
	webidGen := infrastructure.NewWebIDGenerator("https://example.com")

	// Setup event dispatcher and unit of work
	eventDispatcher := setupTestEventDispatcher(t, userWriteRepo, fileStorage)
	unitOfWorkFactory := setupTestUnitOfWorkFactory(t, eventDispatcher)

	// Setup user service
	userService := application.NewUserService(unitOfWorkFactory, webidGen, userRepo)

	// Test data
	profile := domain.UserProfile{
		Name:        "John Doe",
		Bio:         "Test user",
		Avatar:      "https://example.com/avatar.jpg",
		Preferences: map[string]interface{}{"theme": "dark"},
	}

	req := application.RegisterUserRequest{
		Email:   "john.doe@example.com",
		Profile: profile,
	}

	// Execute registration
	user, err := userService.RegisterUser(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, user)

	// Verify user entity
	assert.NotEmpty(t, user.ID())
	assert.NotEmpty(t, user.WebID)
	assert.Equal(t, req.Email, user.Email)
	assert.Equal(t, req.Profile.Name, user.Profile.Name)
	assert.Equal(t, domain.UserStatusActive, user.Status)

	// Verify database persistence
	dbUser, err := userRepo.GetByID(ctx, user.ID())
	require.NoError(t, err)
	assert.Equal(t, user.ID(), dbUser.ID())
	assert.Equal(t, user.WebID, dbUser.WebID)
	assert.Equal(t, user.Email, dbUser.Email)
	assert.Equal(t, user.Profile.Name, dbUser.Profile.Name)

	// Verify file storage
	exists, err := fileStorage.UserExists(ctx, user.ID())
	require.NoError(t, err)
	assert.True(t, exists)

	storedProfile, err := fileStorage.ReadUserProfile(ctx, user.ID())
	require.NoError(t, err)
	assert.Equal(t, profile.Name, storedProfile.Name)
	assert.Equal(t, profile.Bio, storedProfile.Bio)

	webidDoc, err := fileStorage.ReadWebIDDocument(ctx, user.ID())
	require.NoError(t, err)
	assert.Contains(t, webidDoc, user.WebID)
	assert.Contains(t, webidDoc, user.Email)
	assert.Contains(t, webidDoc, user.Profile.Name)

	// Verify WebID uniqueness
	isUnique, err := webidGen.IsUniqueWebID(ctx, user.WebID)
	require.NoError(t, err)
	assert.False(t, isUnique) // Should be false since it's now used
}

// TestUserProfileUpdateWorkflow tests the complete user profile update workflow
func TestUserProfileUpdateWorkflow(t *testing.T) {
	// Setup test environment
	ctx := context.Background()
	tempDir := t.TempDir()

	// Setup database and repositories
	db := setupTestDatabase(t)
	userRepo := infrastructure.NewGormUserRepository(db)
	userWriteRepo := infrastructure.NewGormUserWriteRepository(db)

	// Setup file storage
	fileStorage := setupTestFileStorage(t, tempDir)

	// Setup event dispatcher and unit of work
	eventDispatcher := setupTestEventDispatcher(t, userWriteRepo, fileStorage)
	unitOfWorkFactory := setupTestUnitOfWorkFactory(t, eventDispatcher)

	// Setup WebID generator
	webidGen := infrastructure.NewWebIDGenerator("https://example.com")

	// Setup user service
	userService := application.NewUserService(unitOfWorkFactory, webidGen, userRepo)

	// Create initial user
	initialProfile := domain.UserProfile{
		Name:        "John Doe",
		Bio:         "Initial bio",
		Avatar:      "https://example.com/avatar1.jpg",
		Preferences: map[string]interface{}{"theme": "light"},
	}

	req := application.RegisterUserRequest{
		Email:   "john.doe@example.com",
		Profile: initialProfile,
	}

	user, err := userService.RegisterUser(ctx, req)
	require.NoError(t, err)

	// Update profile
	updatedProfile := domain.UserProfile{
		Name:        "John Smith",
		Bio:         "Updated bio",
		Avatar:      "https://example.com/avatar2.jpg",
		Preferences: map[string]interface{}{"theme": "dark", "language": "en"},
	}

	err = userService.UpdateProfile(ctx, user.ID(), updatedProfile)
	require.NoError(t, err)

	// Verify database update
	dbUser, err := userRepo.GetByID(ctx, user.ID())
	require.NoError(t, err)
	assert.Equal(t, updatedProfile.Name, dbUser.Profile.Name)
	assert.Equal(t, updatedProfile.Bio, dbUser.Profile.Bio)
	assert.Equal(t, updatedProfile.Avatar, dbUser.Profile.Avatar)

	// Verify file storage update
	storedProfile, err := fileStorage.ReadUserProfile(ctx, user.ID())
	require.NoError(t, err)
	assert.Equal(t, updatedProfile.Name, storedProfile.Name)
	assert.Equal(t, updatedProfile.Bio, storedProfile.Bio)
	assert.Equal(t, updatedProfile.Avatar, storedProfile.Avatar)
	assert.Equal(t, "dark", storedProfile.Preferences["theme"])
	assert.Equal(t, "en", storedProfile.Preferences["language"])
}

// TestUserDeletionWorkflow tests the complete user deletion workflow
func TestUserDeletionWorkflow(t *testing.T) {
	// Setup test environment
	ctx := context.Background()
	tempDir := t.TempDir()

	// Setup database and repositories
	db := setupTestDatabase(t)
	userRepo := infrastructure.NewGormUserRepository(db)
	userWriteRepo := infrastructure.NewGormUserWriteRepository(db)

	// Setup file storage
	fileStorage := setupTestFileStorage(t, tempDir)

	// Setup event dispatcher and unit of work
	eventDispatcher := setupTestEventDispatcher(t, userWriteRepo, fileStorage)
	unitOfWorkFactory := setupTestUnitOfWorkFactory(t, eventDispatcher)

	// Setup WebID generator
	webidGen := infrastructure.NewWebIDGenerator("https://example.com")

	// Setup user service
	userService := application.NewUserService(unitOfWorkFactory, webidGen, userRepo)

	// Create user
	profile := domain.UserProfile{
		Name:        "John Doe",
		Bio:         "Test user",
		Avatar:      "https://example.com/avatar.jpg",
		Preferences: map[string]interface{}{"theme": "dark"},
	}

	req := application.RegisterUserRequest{
		Email:   "john.doe@example.com",
		Profile: profile,
	}

	user, err := userService.RegisterUser(ctx, req)
	require.NoError(t, err)

	// Verify user exists before deletion
	exists, err := fileStorage.UserExists(ctx, user.ID())
	require.NoError(t, err)
	assert.True(t, exists)

	// Delete user
	err = userService.DeleteAccount(ctx, user.ID())
	require.NoError(t, err)

	// Verify database deletion (user should be marked as deleted)
	dbUser, err := userRepo.GetByID(ctx, user.ID())
	if err == nil {
		// If user still exists in DB, it should be marked as deleted
		assert.Equal(t, domain.UserStatusDeleted, dbUser.Status)
	}

	// Verify file cleanup
	exists, err = fileStorage.UserExists(ctx, user.ID())
	require.NoError(t, err)
	assert.False(t, exists)
}

// Helper functions

func setupTestDatabase(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Run migrations
	err = infrastructure.MigrateUserModels(db)
	require.NoError(t, err)

	return db
}

func setupTestFileStorage(t *testing.T, tempDir string) *infrastructure.FileStorageAdapter {
	return infrastructure.NewFileStorageAdapter(tempDir)
}

func setupTestEventDispatcher(t *testing.T, userWriteRepo domain.UserWriteRepository, fileStorage application.FileStorage) pericarpdomain.EventDispatcher {
	dispatcher := pericarpdomain.NewInMemoryEventDispatcher()

	// Register user event handlers
	userEventHandler := application.NewUserEventHandler(userWriteRepo, fileStorage)

	// Register event handlers with dispatcher
	dispatcher.RegisterHandler("UserRegisteredEvent", func(ctx context.Context, event pericarpdomain.Event) error {
		if userEvent, ok := event.Data().(*domain.UserRegisteredEventData); ok {
			return userEventHandler.HandleUserRegistered(ctx, userEvent)
		}
		return nil
	})

	dispatcher.RegisterHandler("UserProfileUpdatedEvent", func(ctx context.Context, event pericarpdomain.Event) error {
		if userEvent, ok := event.Data().(*domain.UserProfileUpdatedEventData); ok {
			return userEventHandler.HandleUserProfileUpdated(ctx, userEvent)
		}
		return nil
	})

	dispatcher.RegisterHandler("UserDeletedEvent", func(ctx context.Context, event pericarpdomain.Event) error {
		if userEvent, ok := event.Data().(*domain.UserDeletedEventData); ok {
			return userEventHandler.HandleUserDeleted(ctx, userEvent)
		}
		return nil
	})

	dispatcher.RegisterHandler("WebIDGeneratedEvent", func(ctx context.Context, event pericarpdomain.Event) error {
		if webidEvent, ok := event.Data().(*domain.WebIDGeneratedEventData); ok {
			return userEventHandler.HandleWebIDGenerated(ctx, webidEvent)
		}
		return nil
	})

	return dispatcher
}

func setupTestUnitOfWorkFactory(t *testing.T, eventDispatcher pericarpdomain.EventDispatcher) func() pericarpdomain.UnitOfWork {
	return func() pericarpdomain.UnitOfWork {
		return pericarpdomain.NewInMemoryUnitOfWork(eventDispatcher)
	}
}

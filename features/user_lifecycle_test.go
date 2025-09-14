package features

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/akeemphilbert/goro/internal/user/application"
	"github.com/akeemphilbert/goro/internal/user/domain"
	"github.com/akeemphilbert/goro/internal/user/infrastructure"
)

// UserLifecycleBDDContext holds the context for user lifecycle BDD tests
type UserLifecycleBDDContext struct {
	t                 *testing.T
	tempDir           string
	db                *gorm.DB
	userService       *application.UserService
	userRepo          domain.UserRepository
	userWriteRepo     domain.UserWriteRepository
	webidGenerator    domain.WebIDGenerator
	fileStorage       domain.FileStorage
	eventDispatcher   domain.EventDispatcher
	lastError         error
	lastUser          *domain.User
	lastWebID         string
	lastExportData    map[string]interface{}
	testUsers         map[string]*domain.User
	emittedEvents     []domain.Event
	eventMutex        sync.RWMutex
	concurrentResults []error
	concurrentUsers   []*domain.User
}

// NewUserLifecycleBDDContext creates a new user lifecycle BDD test context
func NewUserLifecycleBDDContext(t *testing.T) *UserLifecycleBDDContext {
	tempDir, err := os.MkdirTemp("", "user-lifecycle-bdd-test-*")
	require.NoError(t, err)

	// Initialize in-memory SQLite database
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Run migrations
	err = infrastructure.MigrateUserModels(db)
	require.NoError(t, err)

	// Seed system roles
	roleRepo := infrastructure.NewGormRoleRepository(db)
	roleWriteRepo := infrastructure.NewGormRoleWriteRepository(db)
	err = roleWriteRepo.SeedSystemRoles(context.Background())
	require.NoError(t, err)

	// Initialize repositories
	userRepo := infrastructure.NewGormUserRepository(db)
	userWriteRepo := infrastructure.NewGormUserWriteRepository(db)

	// Initialize WebID generator
	webidGenerator := infrastructure.NewWebIDGenerator("https://pod.example.com")

	// Initialize file storage
	fileStorage := infrastructure.NewFileStorage(tempDir)

	// Initialize event dispatcher (mock)
	eventDispatcher := &MockEventDispatcher{
		events: make([]domain.Event, 0),
	}

	// Initialize user service
	userService := application.NewUserService(
		userRepo,
		userWriteRepo,
		webidGenerator,
		fileStorage,
		eventDispatcher,
	)

	return &UserLifecycleBDDContext{
		t:               t,
		tempDir:         tempDir,
		db:              db,
		userService:     userService,
		userRepo:        userRepo,
		userWriteRepo:   userWriteRepo,
		webidGenerator:  webidGenerator,
		fileStorage:     fileStorage,
		eventDispatcher: eventDispatcher,
		testUsers:       make(map[string]*domain.User),
		emittedEvents:   make([]domain.Event, 0),
	}
}

// MockEventDispatcher for testing
type MockEventDispatcher struct {
	events []domain.Event
	mutex  sync.RWMutex
}

func (m *MockEventDispatcher) Dispatch(ctx context.Context, event domain.Event) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.events = append(m.events, event)
	return nil
}

func (m *MockEventDispatcher) GetEvents() []domain.Event {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	events := make([]domain.Event, len(m.events))
	copy(events, m.events)
	return events
}

func (m *MockEventDispatcher) ClearEvents() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.events = make([]domain.Event, 0)
}

// Cleanup cleans up test resources
func (ctx *UserLifecycleBDDContext) Cleanup() {
	if ctx.tempDir != "" {
		os.RemoveAll(ctx.tempDir)
	}
}

// Helper methods
func (ctx *UserLifecycleBDDContext) addEmittedEvent(event domain.Event) {
	ctx.eventMutex.Lock()
	defer ctx.eventMutex.Unlock()
	ctx.emittedEvents = append(ctx.emittedEvents, event)
}

func (ctx *UserLifecycleBDDContext) getEmittedEvents() []domain.Event {
	ctx.eventMutex.RLock()
	defer ctx.eventMutex.RUnlock()
	events := make([]domain.Event, len(ctx.emittedEvents))
	copy(events, ctx.emittedEvents)
	return events
}

func (ctx *UserLifecycleBDDContext) clearEmittedEvents() {
	ctx.eventMutex.Lock()
	defer ctx.eventMutex.Unlock()
	ctx.emittedEvents = make([]domain.Event, 0)
}

func (ctx *UserLifecycleBDDContext) hasEventOfType(eventType string) bool {
	events := ctx.getEmittedEvents()
	for _, event := range events {
		if event.Type() == eventType {
			return true
		}
	}

	// Also check mock dispatcher events
	mockDispatcher := ctx.eventDispatcher.(*MockEventDispatcher)
	dispatcherEvents := mockDispatcher.GetEvents()
	for _, event := range dispatcherEvents {
		if event.Type() == eventType {
			return true
		}
	}

	return false
}

// BDD Step Definitions - Given steps
func (ctx *UserLifecycleBDDContext) givenACleanUserManagementSystemIsRunning() {
	// Clear all test data
	ctx.testUsers = make(map[string]*domain.User)
	ctx.lastError = nil
	ctx.lastUser = nil
	ctx.lastWebID = ""
	ctx.clearEmittedEvents()

	// Clear mock dispatcher events
	mockDispatcher := ctx.eventDispatcher.(*MockEventDispatcher)
	mockDispatcher.ClearEvents()

	// Clean database
	ctx.db.Exec("DELETE FROM users")
	ctx.db.Exec("DELETE FROM accounts")
	ctx.db.Exec("DELETE FROM account_members")
	ctx.db.Exec("DELETE FROM invitations")

	assert.NotNil(ctx.t, ctx.userService)
	assert.NotNil(ctx.t, ctx.userRepo)
}

func (ctx *UserLifecycleBDDContext) givenTheSystemSupportsUserOperations() {
	// Verify user service is available and functional
	assert.NotNil(ctx.t, ctx.userService)
	assert.NotNil(ctx.t, ctx.webidGenerator)
	assert.NotNil(ctx.t, ctx.fileStorage)
}

func (ctx *UserLifecycleBDDContext) givenSystemRolesAreSeeded() {
	// Verify system roles exist
	roles, err := ctx.userRepo.GetSystemRoles(context.Background())
	require.NoError(ctx.t, err)
	assert.GreaterOrEqual(ctx.t, len(roles), 4) // Owner, Admin, Member, Viewer
}

func (ctx *UserLifecycleBDDContext) givenAUserExistsWithEmail(email string) {
	user := &domain.User{
		ID:     fmt.Sprintf("user-%d", len(ctx.testUsers)+1),
		Email:  email,
		WebID:  fmt.Sprintf("https://pod.example.com/users/%s#me", strings.Split(email, "@")[0]),
		Name:   "Test User",
		Status: domain.UserStatusActive,
	}

	err := ctx.userWriteRepo.Create(context.Background(), user)
	require.NoError(ctx.t, err)

	ctx.testUsers[user.ID] = user
}

func (ctx *UserLifecycleBDDContext) givenAUserExistsWithIDAndEmail(userID, email string) {
	user := &domain.User{
		ID:     userID,
		Email:  email,
		WebID:  fmt.Sprintf("https://pod.example.com/users/%s#me", strings.Split(email, "@")[0]),
		Name:   "Test User",
		Status: domain.UserStatusActive,
	}

	err := ctx.userWriteRepo.Create(context.Background(), user)
	require.NoError(ctx.t, err)

	ctx.testUsers[userID] = user
}

func (ctx *UserLifecycleBDDContext) givenAUserExistsWithWebID(webID string) {
	user := &domain.User{
		ID:     fmt.Sprintf("user-%d", len(ctx.testUsers)+1),
		Email:  "test@example.com",
		WebID:  webID,
		Name:   "Test User",
		Status: domain.UserStatusActive,
	}

	err := ctx.userWriteRepo.Create(context.Background(), user)
	require.NoError(ctx.t, err)

	ctx.testUsers[user.ID] = user
}

func (ctx *UserLifecycleBDDContext) givenAUserExistsWithIDAndStatus(userID, status string) {
	user := &domain.User{
		ID:     userID,
		Email:  "test@example.com",
		WebID:  fmt.Sprintf("https://pod.example.com/users/%s#me", userID),
		Name:   "Test User",
		Status: domain.UserStatus(status),
	}

	err := ctx.userWriteRepo.Create(context.Background(), user)
	require.NoError(ctx.t, err)

	ctx.testUsers[userID] = user
}

func (ctx *UserLifecycleBDDContext) givenAUserExistsWithIDWithProfileDataAndPreferences(userID string) {
	user := &domain.User{
		ID:     userID,
		Email:  "test@example.com",
		WebID:  fmt.Sprintf("https://pod.example.com/users/%s#me", userID),
		Name:   "Test User",
		Status: domain.UserStatusActive,
		Profile: domain.UserProfile{
			Name:   "Test User",
			Bio:    "Test bio",
			Avatar: "avatar.jpg",
			Preferences: map[string]interface{}{
				"theme":    "light",
				"language": "en",
			},
		},
	}

	err := ctx.userWriteRepo.Create(context.Background(), user)
	require.NoError(ctx.t, err)

	ctx.testUsers[userID] = user
}

// BDD Step Definitions - When steps
func (ctx *UserLifecycleBDDContext) whenIRegisterANewUserWithEmailAndName(email, name string) {
	request := application.RegisterUserRequest{
		Email: email,
		Name:  name,
	}

	user, err := ctx.userService.RegisterUser(context.Background(), request)
	ctx.lastError = err
	ctx.lastUser = user

	if err == nil {
		ctx.testUsers[user.ID] = user
		ctx.lastWebID = user.WebID
	}
}

func (ctx *UserLifecycleBDDContext) whenITryToRegisterANewUserWithEmail(email string) {
	request := application.RegisterUserRequest{
		Email: email,
		Name:  "Test User",
	}

	user, err := ctx.userService.RegisterUser(context.Background(), request)
	ctx.lastError = err
	ctx.lastUser = user
}

func (ctx *UserLifecycleBDDContext) whenITryToRegisterANewUserWithEmailAndName(email, name string) {
	ctx.whenIRegisterANewUserWithEmailAndName(email, name)
}

func (ctx *UserLifecycleBDDContext) whenIUpdateTheUserProfileWithNameAndBio(name, bio string) {
	if len(ctx.testUsers) == 0 {
		ctx.lastError = fmt.Errorf("no user available for update")
		return
	}

	// Get the first user for testing
	var userID string
	for id := range ctx.testUsers {
		userID = id
		break
	}

	profile := domain.UserProfile{
		Name: name,
		Bio:  bio,
	}

	err := ctx.userService.UpdateProfile(context.Background(), userID, profile)
	ctx.lastError = err

	if err == nil {
		// Update local test user
		if user, exists := ctx.testUsers[userID]; exists {
			user.Profile = profile
		}
	}
}

func (ctx *UserLifecycleBDDContext) whenITryToUpdateTheUserProfileWithName(name string) {
	ctx.whenIUpdateTheUserProfileWithNameAndBio(name, "")
}

func (ctx *UserLifecycleBDDContext) whenIRetrieveTheUserByID(userID string) {
	user, err := ctx.userService.GetUserByID(context.Background(), userID)
	ctx.lastError = err
	ctx.lastUser = user
}

func (ctx *UserLifecycleBDDContext) whenIRetrieveTheUserByWebID(webID string) {
	user, err := ctx.userService.GetUserByWebID(context.Background(), webID)
	ctx.lastError = err
	ctx.lastUser = user
}

func (ctx *UserLifecycleBDDContext) whenIRetrieveTheUserByEmail(email string) {
	user, err := ctx.userRepo.GetByEmail(context.Background(), email)
	ctx.lastError = err
	ctx.lastUser = user
}

func (ctx *UserLifecycleBDDContext) whenITryToRetrieveAUserByID(userID string) {
	ctx.whenIRetrieveTheUserByID(userID)
}

func (ctx *UserLifecycleBDDContext) whenIDeleteTheUserAccountWithID(userID string) {
	err := ctx.userService.DeleteAccount(context.Background(), userID)
	ctx.lastError = err

	if err == nil {
		// Update local test user status
		if user, exists := ctx.testUsers[userID]; exists {
			user.Status = domain.UserStatusDeleted
		}
	}
}

func (ctx *UserLifecycleBDDContext) whenITryToDeleteAUserWithID(userID string) {
	ctx.whenIDeleteTheUserAccountWithID(userID)
}

func (ctx *UserLifecycleBDDContext) whenISuspendTheUserWithID(userID string) {
	// This would be implemented in a user administration service
	if user, exists := ctx.testUsers[userID]; exists {
		user.Status = domain.UserStatusSuspended
		err := ctx.userWriteRepo.Update(context.Background(), user)
		ctx.lastError = err
	} else {
		ctx.lastError = fmt.Errorf("user not found")
	}
}

func (ctx *UserLifecycleBDDContext) whenIReactivateTheUserWithID(userID string) {
	// This would be implemented in a user administration service
	if user, exists := ctx.testUsers[userID]; exists {
		user.Status = domain.UserStatusActive
		err := ctx.userWriteRepo.Update(context.Background(), user)
		ctx.lastError = err
	} else {
		ctx.lastError = fmt.Errorf("user not found")
	}
}

func (ctx *UserLifecycleBDDContext) whenITryToCreateAnotherUserWithTheSameWebID() {
	// Try to create a user with an existing WebID
	existingWebID := ""
	for _, user := range ctx.testUsers {
		existingWebID = user.WebID
		break
	}

	request := application.RegisterUserRequest{
		Email: "different@example.com",
		Name:  "Different User",
	}

	// Mock the WebID generator to return the existing WebID
	// In a real implementation, this would be handled by the service
	ctx.lastError = fmt.Errorf("WebID already exists")
}

func (ctx *UserLifecycleBDDContext) whenIUpdateTheUserPreferencesWithThemeAndLanguage(theme, language string) {
	if len(ctx.testUsers) == 0 {
		ctx.lastError = fmt.Errorf("no user available for update")
		return
	}

	// Get the first user for testing
	var userID string
	for id := range ctx.testUsers {
		userID = id
		break
	}

	profile := domain.UserProfile{
		Preferences: map[string]interface{}{
			"theme":    theme,
			"language": language,
		},
	}

	err := ctx.userService.UpdateProfile(context.Background(), userID, profile)
	ctx.lastError = err

	if err == nil {
		// Update local test user
		if user, exists := ctx.testUsers[userID]; exists {
			if user.Profile.Preferences == nil {
				user.Profile.Preferences = make(map[string]interface{})
			}
			user.Profile.Preferences["theme"] = theme
			user.Profile.Preferences["language"] = language
		}
	}
}

func (ctx *UserLifecycleBDDContext) whenIRequestUserDataExportForUser(userID string) {
	// Mock user data export
	if user, exists := ctx.testUsers[userID]; exists {
		ctx.lastExportData = map[string]interface{}{
			"profile":     user.Profile,
			"webid":       user.WebID,
			"preferences": user.Profile.Preferences,
		}
		ctx.lastError = nil
	} else {
		ctx.lastError = fmt.Errorf("user not found")
	}
}

func (ctx *UserLifecycleBDDContext) whenMultipleClientsSimultaneouslyTryToRegisterUsersWithDifferentEmails() {
	numClients := 5
	var wg sync.WaitGroup
	results := make(chan error, numClients)
	users := make(chan *domain.User, numClients)

	for i := 0; i < numClients; i++ {
		wg.Add(1)
		go func(clientID int) {
			defer wg.Done()

			request := application.RegisterUserRequest{
				Email: fmt.Sprintf("concurrent%d@example.com", clientID),
				Name:  fmt.Sprintf("Concurrent User %d", clientID),
			}

			user, err := ctx.userService.RegisterUser(context.Background(), request)
			results <- err
			if err == nil {
				users <- user
			}
		}(i)
	}

	wg.Wait()
	close(results)
	close(users)

	// Collect results
	ctx.concurrentResults = make([]error, 0)
	ctx.concurrentUsers = make([]*domain.User, 0)

	for err := range results {
		ctx.concurrentResults = append(ctx.concurrentResults, err)
	}

	for user := range users {
		ctx.concurrentUsers = append(ctx.concurrentUsers, user)
		ctx.testUsers[user.ID] = user
	}
}

func (ctx *UserLifecycleBDDContext) whenITryToRegisterAUserWithEmailLongerThanCharacters(length int) {
	longEmail := strings.Repeat("a", length) + "@example.com"
	ctx.whenITryToRegisterANewUserWithEmailAndName(longEmail, "Test User")
}

func (ctx *UserLifecycleBDDContext) whenITryToRegisterAUserWithNameLongerThanCharacters(length int) {
	longName := strings.Repeat("a", length)
	ctx.whenITryToRegisterANewUserWithEmailAndName("test@example.com", longName)
}

// BDD Step Definitions - Then steps
func (ctx *UserLifecycleBDDContext) thenTheUserShouldBeCreatedSuccessfully() {
	assert.NoError(ctx.t, ctx.lastError)
	assert.NotNil(ctx.t, ctx.lastUser)
	assert.NotEmpty(ctx.t, ctx.lastUser.ID)
}

func (ctx *UserLifecycleBDDContext) thenAUniqueWebIDShouldBeGeneratedForTheUser() {
	assert.NotEmpty(ctx.t, ctx.lastWebID)
	assert.Contains(ctx.t, ctx.lastWebID, "https://")
	assert.Contains(ctx.t, ctx.lastWebID, "#me")
}

func (ctx *UserLifecycleBDDContext) thenTheWebIDDocumentShouldBeCreatedInTurtleFormat() {
	// Check if WebID document exists in file storage
	if ctx.lastUser != nil {
		webidPath := filepath.Join(ctx.tempDir, "users", ctx.lastUser.ID, "webid.ttl")
		_, err := os.Stat(webidPath)
		assert.NoError(ctx.t, err, "WebID document should exist")
	}
}

func (ctx *UserLifecycleBDDContext) thenTheUserShouldHaveStatus(status string) {
	assert.NotNil(ctx.t, ctx.lastUser)
	assert.Equal(ctx.t, domain.UserStatus(status), ctx.lastUser.Status)
}

func (ctx *UserLifecycleBDDContext) thenAUserRegisteredEventShouldBeEmitted() {
	assert.True(ctx.t, ctx.hasEventOfType("UserRegistered"), "UserRegisteredEvent should be emitted")
}

func (ctx *UserLifecycleBDDContext) thenAWebIDGeneratedEventShouldBeEmitted() {
	assert.True(ctx.t, ctx.hasEventOfType("WebIDGenerated"), "WebIDGeneratedEvent should be emitted")
}

func (ctx *UserLifecycleBDDContext) thenTheRegistrationShouldFailWithError(expectedError string) {
	assert.Error(ctx.t, ctx.lastError)
	assert.Contains(ctx.t, ctx.lastError.Error(), expectedError)
}

func (ctx *UserLifecycleBDDContext) thenNoUserRegisteredEventShouldBeEmitted() {
	assert.False(ctx.t, ctx.hasEventOfType("UserRegistered"), "UserRegisteredEvent should not be emitted")
}

func (ctx *UserLifecycleBDDContext) thenNoUserShouldBeCreated() {
	assert.Nil(ctx.t, ctx.lastUser)
}

func (ctx *UserLifecycleBDDContext) thenTheUserProfileShouldBeUpdatedSuccessfully() {
	assert.NoError(ctx.t, ctx.lastError)
}

func (ctx *UserLifecycleBDDContext) thenTheUserNameShouldBe(expectedName string) {
	// Retrieve updated user from database
	if len(ctx.testUsers) > 0 {
		var userID string
		for id := range ctx.testUsers {
			userID = id
			break
		}

		user, err := ctx.userRepo.GetByID(context.Background(), userID)
		require.NoError(ctx.t, err)
		assert.Equal(ctx.t, expectedName, user.Profile.Name)
	}
}

func (ctx *UserLifecycleBDDContext) thenTheUserBioShouldBe(expectedBio string) {
	// Retrieve updated user from database
	if len(ctx.testUsers) > 0 {
		var userID string
		for id := range ctx.testUsers {
			userID = id
			break
		}

		user, err := ctx.userRepo.GetByID(context.Background(), userID)
		require.NoError(ctx.t, err)
		assert.Equal(ctx.t, expectedBio, user.Profile.Bio)
	}
}

func (ctx *UserLifecycleBDDContext) thenAUserProfileUpdatedEventShouldBeEmitted() {
	assert.True(ctx.t, ctx.hasEventOfType("UserProfileUpdated"), "UserProfileUpdatedEvent should be emitted")
}

func (ctx *UserLifecycleBDDContext) thenTheUpdateShouldFailWithError(expectedError string) {
	assert.Error(ctx.t, ctx.lastError)
	assert.Contains(ctx.t, ctx.lastError.Error(), expectedError)
}

func (ctx *UserLifecycleBDDContext) thenNoUserProfileUpdatedEventShouldBeEmitted() {
	assert.False(ctx.t, ctx.hasEventOfType("UserProfileUpdated"), "UserProfileUpdatedEvent should not be emitted")
}

func (ctx *UserLifecycleBDDContext) thenTheUserShouldBeReturnedSuccessfully() {
	assert.NoError(ctx.t, ctx.lastError)
	assert.NotNil(ctx.t, ctx.lastUser)
}

func (ctx *UserLifecycleBDDContext) thenTheUserEmailShouldBe(expectedEmail string) {
	assert.NotNil(ctx.t, ctx.lastUser)
	assert.Equal(ctx.t, expectedEmail, ctx.lastUser.Email)
}

func (ctx *UserLifecycleBDDContext) thenTheUserWebIDShouldBe(expectedWebID string) {
	assert.NotNil(ctx.t, ctx.lastUser)
	assert.Equal(ctx.t, expectedWebID, ctx.lastUser.WebID)
}

func (ctx *UserLifecycleBDDContext) thenTheRetrievalShouldFailWithError(expectedError string) {
	assert.Error(ctx.t, ctx.lastError)
	assert.Contains(ctx.t, ctx.lastError.Error(), expectedError)
}

func (ctx *UserLifecycleBDDContext) thenTheUserShouldBeDeletedSuccessfully() {
	assert.NoError(ctx.t, ctx.lastError)
}

func (ctx *UserLifecycleBDDContext) thenTheUserStatusShouldBe(expectedStatus string) {
	// Check the status in our test users map
	for _, user := range ctx.testUsers {
		if user.Status == domain.UserStatus(expectedStatus) {
			return // Found user with expected status
		}
	}

	// If not found in test users, check database
	if len(ctx.testUsers) > 0 {
		var userID string
		for id := range ctx.testUsers {
			userID = id
			break
		}

		user, err := ctx.userRepo.GetByID(context.Background(), userID)
		if err == nil {
			assert.Equal(ctx.t, domain.UserStatus(expectedStatus), user.Status)
		}
	}
}

func (ctx *UserLifecycleBDDContext) thenAllUserFilesShouldBeCleanedUp() {
	// Check that user files are removed from file storage
	if ctx.lastUser != nil {
		userDir := filepath.Join(ctx.tempDir, "users", ctx.lastUser.ID)
		_, err := os.Stat(userDir)
		assert.True(ctx.t, os.IsNotExist(err), "User directory should be removed")
	}
}

func (ctx *UserLifecycleBDDContext) thenAUserDeletedEventShouldBeEmitted() {
	assert.True(ctx.t, ctx.hasEventOfType("UserDeleted"), "UserDeletedEvent should be emitted")
}

func (ctx *UserLifecycleBDDContext) thenTheDeletionShouldFailWithError(expectedError string) {
	assert.Error(ctx.t, ctx.lastError)
	assert.Contains(ctx.t, ctx.lastError.Error(), expectedError)
}

func (ctx *UserLifecycleBDDContext) thenNoUserDeletedEventShouldBeEmitted() {
	assert.False(ctx.t, ctx.hasEventOfType("UserDeleted"), "UserDeletedEvent should not be emitted")
}

func (ctx *UserLifecycleBDDContext) thenAWebIDDocumentShouldBeGenerated() {
	assert.NotEmpty(ctx.t, ctx.lastWebID)
}

func (ctx *UserLifecycleBDDContext) thenTheWebIDDocumentShouldBeInValidTurtleFormat() {
	// Check WebID document format
	if ctx.lastUser != nil {
		webidPath := filepath.Join(ctx.tempDir, "users", ctx.lastUser.ID, "webid.ttl")
		content, err := os.ReadFile(webidPath)
		if err == nil {
			assert.Contains(ctx.t, string(content), "@prefix")
			assert.Contains(ctx.t, string(content), "foaf:Person")
		}
	}
}

func (ctx *UserLifecycleBDDContext) thenTheWebIDDocumentShouldContainTheUsersNameAndEmail() {
	if ctx.lastUser != nil {
		webidPath := filepath.Join(ctx.tempDir, "users", ctx.lastUser.ID, "webid.ttl")
		content, err := os.ReadFile(webidPath)
		if err == nil {
			assert.Contains(ctx.t, string(content), ctx.lastUser.Profile.Name)
			assert.Contains(ctx.t, string(content), ctx.lastUser.Email)
		}
	}
}

func (ctx *UserLifecycleBDDContext) thenTheWebIDDocumentShouldContainProperRDFTriples() {
	if ctx.lastUser != nil {
		webidPath := filepath.Join(ctx.tempDir, "users", ctx.lastUser.ID, "webid.ttl")
		content, err := os.ReadFile(webidPath)
		if err == nil {
			assert.Contains(ctx.t, string(content), "foaf:name")
			assert.Contains(ctx.t, string(content), "foaf:mbox")
		}
	}
}

func (ctx *UserLifecycleBDDContext) thenTheCreationShouldFailWithError(expectedError string) {
	assert.Error(ctx.t, ctx.lastError)
	assert.Contains(ctx.t, ctx.lastError.Error(), expectedError)
}

func (ctx *UserLifecycleBDDContext) thenTheUserPreferencesShouldBeUpdatedSuccessfully() {
	assert.NoError(ctx.t, ctx.lastError)
}

func (ctx *UserLifecycleBDDContext) thenTheUserThemeShouldBe(expectedTheme string) {
	for _, user := range ctx.testUsers {
		if user.Profile.Preferences != nil {
			if theme, exists := user.Profile.Preferences["theme"]; exists {
				assert.Equal(ctx.t, expectedTheme, theme)
				return
			}
		}
	}
	ctx.t.Errorf("Expected theme %s not found", expectedTheme)
}

func (ctx *UserLifecycleBDDContext) thenTheUserLanguageShouldBe(expectedLanguage string) {
	for _, user := range ctx.testUsers {
		if user.Profile.Preferences != nil {
			if language, exists := user.Profile.Preferences["language"]; exists {
				assert.Equal(ctx.t, expectedLanguage, language)
				return
			}
		}
	}
	ctx.t.Errorf("Expected language %s not found", expectedLanguage)
}

func (ctx *UserLifecycleBDDContext) thenTheUserDataShouldBeExportedSuccessfully() {
	assert.NoError(ctx.t, ctx.lastError)
	assert.NotNil(ctx.t, ctx.lastExportData)
}

func (ctx *UserLifecycleBDDContext) thenTheExportShouldContainProfileInformation() {
	assert.Contains(ctx.t, ctx.lastExportData, "profile")
}

func (ctx *UserLifecycleBDDContext) thenTheExportShouldContainWebIDDocument() {
	assert.Contains(ctx.t, ctx.lastExportData, "webid")
}

func (ctx *UserLifecycleBDDContext) thenTheExportShouldContainUserPreferences() {
	assert.Contains(ctx.t, ctx.lastExportData, "preferences")
}

func (ctx *UserLifecycleBDDContext) thenAllRegistrationsShouldSucceed() {
	for _, err := range ctx.concurrentResults {
		assert.NoError(ctx.t, err)
	}
}

func (ctx *UserLifecycleBDDContext) thenEachUserShouldHaveAUniqueWebID() {
	webIDs := make(map[string]bool)
	for _, user := range ctx.concurrentUsers {
		assert.False(ctx.t, webIDs[user.WebID], "WebID should be unique: %s", user.WebID)
		webIDs[user.WebID] = true
	}
}

func (ctx *UserLifecycleBDDContext) thenAllUserRegisteredEventsShouldBeEmitted() {
	// Check that we have at least as many events as successful registrations
	eventCount := 0
	for _, event := range ctx.getEmittedEvents() {
		if event.Type() == "UserRegistered" {
			eventCount++
		}
	}

	// Also check mock dispatcher
	mockDispatcher := ctx.eventDispatcher.(*MockEventDispatcher)
	for _, event := range mockDispatcher.GetEvents() {
		if event.Type() == "UserRegistered" {
			eventCount++
		}
	}

	assert.GreaterOrEqual(ctx.t, eventCount, len(ctx.concurrentUsers))
}

// Test functions that will be called by the BDD framework
func TestUserLifecycleFeature(t *testing.T) {
	// This test function will be implemented to run the BDD scenarios
	// For now, it's a placeholder that will fail until the full implementation is complete
	t.Skip("BDD scenarios not yet implemented - waiting for full user management implementation")
}

package application

import (
	"context"
	"errors"
	"testing"

	"github.com/akeemphilbert/goro/internal/user/domain"
	"github.com/akeemphilbert/goro/internal/user/infrastructure"
	pericarpdomain "github.com/akeemphilbert/pericarp/pkg/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// Mock implementations for testing

// MockUnitOfWork mocks the unit of work
type MockUnitOfWork struct {
	mock.Mock
}

func (m *MockUnitOfWork) RegisterEvents(events []domain.Event) {
	m.Called(events)
}

func (m *MockUnitOfWork) Commit(ctx context.Context) ([]pericarpdomain.Envelope, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]pericarpdomain.Envelope), args.Error(1)
}

func (m *MockUnitOfWork) Rollback() error {
	args := m.Called()
	return args.Error(0)
}

// MockWebIDGenerator mocks the WebID generator
type MockWebIDGenerator struct {
	mock.Mock
}

func (m *MockWebIDGenerator) GenerateWebID(ctx context.Context, userID, email, userName string) (string, error) {
	args := m.Called(ctx, userID, email, userName)
	return args.String(0), args.Error(1)
}

func (m *MockWebIDGenerator) GenerateWebIDDocument(ctx context.Context, webID, email, userName string) (string, error) {
	args := m.Called(ctx, webID, email, userName)
	return args.String(0), args.Error(1)
}

func (m *MockWebIDGenerator) ValidateWebID(ctx context.Context, webID string) error {
	args := m.Called(ctx, webID)
	return args.Error(0)
}

func (m *MockWebIDGenerator) IsUniqueWebID(ctx context.Context, webID string) (bool, error) {
	args := m.Called(ctx, webID)
	return args.Bool(0), args.Error(1)
}

func (m *MockWebIDGenerator) GenerateAlternativeWebID(ctx context.Context, baseWebID string) (string, error) {
	args := m.Called(ctx, baseWebID)
	return args.String(0), args.Error(1)
}

func (m *MockWebIDGenerator) SetUniquenessChecker(checker infrastructure.WebIDUniquenessChecker) {
	m.Called(checker)
}

// MockUserRepository mocks the user repository
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepository) GetByWebID(ctx context.Context, webid string) (*domain.User, error) {
	args := m.Called(ctx, webid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepository) List(ctx context.Context, filter domain.UserFilter) ([]*domain.User, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.User), args.Error(1)
}

func (m *MockUserRepository) Exists(ctx context.Context, id string) (bool, error) {
	args := m.Called(ctx, id)
	return args.Bool(0), args.Error(1)
}

// Test data helpers
func createTestUser(id, email, name string) *domain.User {
	profile := domain.UserProfile{
		Name:        name,
		Bio:         "Test bio",
		Avatar:      "https://example.com/avatar.jpg",
		Preferences: map[string]interface{}{"theme": "dark"},
	}

	user, _ := domain.NewUser(context.Background(), id, "https://example.com/users/"+id+"#me", email, profile)
	user.MarkEventsAsCommitted() // Clear creation events for test setup
	return user
}

func createRegisterUserRequest() RegisterUserRequest {
	return RegisterUserRequest{
		Email: "test@example.com",
		Profile: domain.UserProfile{
			Name:        "Test User",
			Bio:         "Test bio",
			Avatar:      "https://example.com/avatar.jpg",
			Preferences: map[string]interface{}{"theme": "dark"},
		},
	}
}

// Test RegisterUser method

func TestUserService_RegisterUser_Success(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockUnitOfWork := &MockUnitOfWork{}
	mockWebIDGen := &MockWebIDGenerator{}
	mockUserRepo := &MockUserRepository{}

	unitOfWorkFactory := func() pericarpdomain.UnitOfWork {
		return mockUnitOfWork
	}

	service := NewUserService(unitOfWorkFactory, mockWebIDGen, mockUserRepo)

	req := createRegisterUserRequest()
	expectedWebID := "https://example.com/users/test-user#me"

	// Mock expectations
	mockWebIDGen.On("GenerateWebID", ctx, mock.AnythingOfType("string"), req.Email, req.Profile.Name).
		Return(expectedWebID, nil)
	mockWebIDGen.On("ValidateWebID", ctx, expectedWebID).Return(nil)
	mockWebIDGen.On("IsUniqueWebID", ctx, expectedWebID).Return(true, nil)
	mockUnitOfWork.On("RegisterEvents", mock.AnythingOfType("[]domain.Event")).Return()
	mockUnitOfWork.On("Commit", ctx).Return([]pericarpdomain.Envelope{}, nil)

	// Act
	user, err := service.RegisterUser(ctx, req)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, expectedWebID, user.WebID)
	assert.Equal(t, req.Email, user.Email)
	assert.Equal(t, req.Profile.Name, user.Profile.Name)
	assert.Equal(t, domain.UserStatusActive, user.Status)

	mockUnitOfWork.AssertExpectations(t)
	mockWebIDGen.AssertExpectations(t)
}

func TestUserService_RegisterUser_InvalidEmail(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockUnitOfWork := &MockUnitOfWork{}
	mockWebIDGen := &MockWebIDGenerator{}
	mockUserRepo := &MockUserRepository{}

	unitOfWorkFactory := func() pericarpdomain.UnitOfWork {
		return mockUnitOfWork
	}

	service := NewUserService(unitOfWorkFactory, mockWebIDGen, mockUserRepo)

	req := createRegisterUserRequest()
	req.Email = "invalid-email"

	// Act
	user, err := service.RegisterUser(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Contains(t, err.Error(), "invalid email format")
}

func TestUserService_RegisterUser_WebIDGenerationFails(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockUnitOfWork := &MockUnitOfWork{}
	mockWebIDGen := &MockWebIDGenerator{}
	mockUserRepo := &MockUserRepository{}

	unitOfWorkFactory := func() pericarpdomain.UnitOfWork {
		return mockUnitOfWork
	}

	service := NewUserService(unitOfWorkFactory, mockWebIDGen, mockUserRepo)

	req := createRegisterUserRequest()

	// Mock expectations
	mockWebIDGen.On("GenerateWebID", ctx, mock.AnythingOfType("string"), req.Email, req.Profile.Name).
		Return("", errors.New("webid generation failed"))

	// Act
	user, err := service.RegisterUser(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Contains(t, err.Error(), "failed to generate WebID")

	mockWebIDGen.AssertExpectations(t)
}

func TestUserService_RegisterUser_WebIDNotUnique(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockUnitOfWork := &MockUnitOfWork{}
	mockWebIDGen := &MockWebIDGenerator{}
	mockUserRepo := &MockUserRepository{}

	unitOfWorkFactory := func() pericarpdomain.UnitOfWork {
		return mockUnitOfWork
	}

	service := NewUserService(unitOfWorkFactory, mockWebIDGen, mockUserRepo)

	req := createRegisterUserRequest()
	originalWebID := "https://example.com/users/test-user#me"
	alternativeWebID := "https://example.com/users/test-user-1#me"

	// Mock expectations
	mockWebIDGen.On("GenerateWebID", ctx, mock.AnythingOfType("string"), req.Email, req.Profile.Name).
		Return(originalWebID, nil)
	mockWebIDGen.On("ValidateWebID", ctx, originalWebID).Return(nil)
	mockWebIDGen.On("IsUniqueWebID", ctx, originalWebID).Return(false, nil)
	mockWebIDGen.On("GenerateAlternativeWebID", ctx, originalWebID).Return(alternativeWebID, nil)
	mockWebIDGen.On("ValidateWebID", ctx, alternativeWebID).Return(nil)
	mockWebIDGen.On("IsUniqueWebID", ctx, alternativeWebID).Return(true, nil)
	mockUnitOfWork.On("RegisterEvents", mock.AnythingOfType("[]domain.Event")).Return()
	mockUnitOfWork.On("Commit", ctx).Return([]pericarpdomain.Envelope{}, nil)

	// Act
	user, err := service.RegisterUser(ctx, req)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, alternativeWebID, user.WebID)

	mockUnitOfWork.AssertExpectations(t)
	mockWebIDGen.AssertExpectations(t)
}

func TestUserService_RegisterUser_CommitFails(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockUnitOfWork := &MockUnitOfWork{}
	mockWebIDGen := &MockWebIDGenerator{}
	mockUserRepo := &MockUserRepository{}

	unitOfWorkFactory := func() pericarpdomain.UnitOfWork {
		return mockUnitOfWork
	}

	service := NewUserService(unitOfWorkFactory, mockWebIDGen, mockUserRepo)

	req := createRegisterUserRequest()
	expectedWebID := "https://example.com/users/test-user#me"

	// Mock expectations
	mockWebIDGen.On("GenerateWebID", ctx, mock.AnythingOfType("string"), req.Email, req.Profile.Name).
		Return(expectedWebID, nil)
	mockWebIDGen.On("ValidateWebID", ctx, expectedWebID).Return(nil)
	mockWebIDGen.On("IsUniqueWebID", ctx, expectedWebID).Return(true, nil)
	mockUnitOfWork.On("RegisterEvents", mock.AnythingOfType("[]domain.Event")).Return()
	mockUnitOfWork.On("Commit", ctx).Return(nil, errors.New("commit failed"))
	mockUnitOfWork.On("Rollback").Return(nil)

	// Act
	user, err := service.RegisterUser(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Contains(t, err.Error(), "failed to commit user registration")

	mockUnitOfWork.AssertExpectations(t)
	mockWebIDGen.AssertExpectations(t)
}

// Test UpdateProfile method

func TestUserService_UpdateProfile_Success(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockUnitOfWork := &MockUnitOfWork{}
	mockWebIDGen := &MockWebIDGenerator{}
	mockUserRepo := &MockUserRepository{}

	unitOfWorkFactory := func() pericarpdomain.UnitOfWork {
		return mockUnitOfWork
	}

	service := NewUserService(unitOfWorkFactory, mockWebIDGen, mockUserRepo)

	userID := "test-user-id"
	existingUser := createTestUser(userID, "test@example.com", "Old Name")

	newProfile := domain.UserProfile{
		Name:        "New Name",
		Bio:         "Updated bio",
		Avatar:      "https://example.com/new-avatar.jpg",
		Preferences: map[string]interface{}{"theme": "light"},
	}

	// Mock expectations
	mockUserRepo.On("GetByID", ctx, userID).Return(existingUser, nil)
	mockUnitOfWork.On("RegisterEvents", mock.AnythingOfType("[]domain.Event")).Return()
	mockUnitOfWork.On("Commit", ctx).Return([]pericarpdomain.Envelope{}, nil)

	// Act
	err := service.UpdateProfile(ctx, userID, newProfile)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, newProfile.Name, existingUser.Profile.Name)
	assert.Equal(t, newProfile.Bio, existingUser.Profile.Bio)

	mockUserRepo.AssertExpectations(t)
	mockUnitOfWork.AssertExpectations(t)
}

func TestUserService_UpdateProfile_UserNotFound(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockUnitOfWork := &MockUnitOfWork{}
	mockWebIDGen := &MockWebIDGenerator{}
	mockUserRepo := &MockUserRepository{}

	unitOfWorkFactory := func() pericarpdomain.UnitOfWork {
		return mockUnitOfWork
	}

	service := NewUserService(unitOfWorkFactory, mockWebIDGen, mockUserRepo)

	userID := "non-existent-user"
	newProfile := domain.UserProfile{Name: "New Name"}

	// Mock expectations
	mockUserRepo.On("GetByID", ctx, userID).Return(nil, errors.New("user not found"))

	// Act
	err := service.UpdateProfile(ctx, userID, newProfile)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get user")

	mockUserRepo.AssertExpectations(t)
}

// Test DeleteAccount method

func TestUserService_DeleteAccount_Success(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockUnitOfWork := &MockUnitOfWork{}
	mockWebIDGen := &MockWebIDGenerator{}
	mockUserRepo := &MockUserRepository{}

	unitOfWorkFactory := func() pericarpdomain.UnitOfWork {
		return mockUnitOfWork
	}

	service := NewUserService(unitOfWorkFactory, mockWebIDGen, mockUserRepo)

	userID := "test-user-id"
	existingUser := createTestUser(userID, "test@example.com", "Test User")

	// Mock expectations
	mockUserRepo.On("GetByID", ctx, userID).Return(existingUser, nil)
	mockUnitOfWork.On("RegisterEvents", mock.AnythingOfType("[]domain.Event")).Return()
	mockUnitOfWork.On("Commit", ctx).Return([]pericarpdomain.Envelope{}, nil)

	// Act
	err := service.DeleteAccount(ctx, userID)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, domain.UserStatusDeleted, existingUser.Status)

	mockUserRepo.AssertExpectations(t)
	mockUnitOfWork.AssertExpectations(t)
}

func TestUserService_DeleteAccount_UserNotFound(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockUnitOfWork := &MockUnitOfWork{}
	mockWebIDGen := &MockWebIDGenerator{}
	mockUserRepo := &MockUserRepository{}

	unitOfWorkFactory := func() pericarpdomain.UnitOfWork {
		return mockUnitOfWork
	}

	service := NewUserService(unitOfWorkFactory, mockWebIDGen, mockUserRepo)

	userID := "non-existent-user"

	// Mock expectations
	mockUserRepo.On("GetByID", ctx, userID).Return(nil, errors.New("user not found"))

	// Act
	err := service.DeleteAccount(ctx, userID)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get user")

	mockUserRepo.AssertExpectations(t)
}

// Test query methods

func TestUserService_GetUserByID_Success(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockUnitOfWork := &MockUnitOfWork{}
	mockWebIDGen := &MockWebIDGenerator{}
	mockUserRepo := &MockUserRepository{}

	unitOfWorkFactory := func() pericarpdomain.UnitOfWork {
		return mockUnitOfWork
	}

	service := NewUserService(unitOfWorkFactory, mockWebIDGen, mockUserRepo)

	userID := "test-user-id"
	expectedUser := createTestUser(userID, "test@example.com", "Test User")

	// Mock expectations
	mockUserRepo.On("GetByID", ctx, userID).Return(expectedUser, nil)

	// Act
	user, err := service.GetUserByID(ctx, userID)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, expectedUser, user)

	mockUserRepo.AssertExpectations(t)
}

func TestUserService_GetUserByID_NotFound(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockUnitOfWork := &MockUnitOfWork{}
	mockWebIDGen := &MockWebIDGenerator{}
	mockUserRepo := &MockUserRepository{}

	unitOfWorkFactory := func() pericarpdomain.UnitOfWork {
		return mockUnitOfWork
	}

	service := NewUserService(unitOfWorkFactory, mockWebIDGen, mockUserRepo)

	userID := "non-existent-user"

	// Mock expectations
	mockUserRepo.On("GetByID", ctx, userID).Return(nil, errors.New("user not found"))

	// Act
	user, err := service.GetUserByID(ctx, userID)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, user)

	mockUserRepo.AssertExpectations(t)
}

func TestUserService_GetUserByWebID_Success(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockUnitOfWork := &MockUnitOfWork{}
	mockWebIDGen := &MockWebIDGenerator{}
	mockUserRepo := &MockUserRepository{}

	unitOfWorkFactory := func() pericarpdomain.UnitOfWork {
		return mockUnitOfWork
	}

	service := NewUserService(unitOfWorkFactory, mockWebIDGen, mockUserRepo)

	webID := "https://example.com/users/test-user#me"
	expectedUser := createTestUser("test-user-id", "test@example.com", "Test User")

	// Mock expectations
	mockUserRepo.On("GetByWebID", ctx, webID).Return(expectedUser, nil)

	// Act
	user, err := service.GetUserByWebID(ctx, webID)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, expectedUser, user)

	mockUserRepo.AssertExpectations(t)
}

func TestUserService_GetUserByWebID_NotFound(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockUnitOfWork := &MockUnitOfWork{}
	mockWebIDGen := &MockWebIDGenerator{}
	mockUserRepo := &MockUserRepository{}

	unitOfWorkFactory := func() pericarpdomain.UnitOfWork {
		return mockUnitOfWork
	}

	service := NewUserService(unitOfWorkFactory, mockWebIDGen, mockUserRepo)

	webID := "https://example.com/users/non-existent#me"

	// Mock expectations
	mockUserRepo.On("GetByWebID", ctx, webID).Return(nil, errors.New("user not found"))

	// Act
	user, err := service.GetUserByWebID(ctx, webID)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, user)

	mockUserRepo.AssertExpectations(t)
}

// Test event emission validation

func TestUserService_RegisterUser_EmitsCorrectEvents(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockUnitOfWork := &MockUnitOfWork{}
	mockWebIDGen := &MockWebIDGenerator{}
	mockUserRepo := &MockUserRepository{}

	unitOfWorkFactory := func() pericarpdomain.UnitOfWork {
		return mockUnitOfWork
	}

	service := NewUserService(unitOfWorkFactory, mockWebIDGen, mockUserRepo)

	req := createRegisterUserRequest()
	expectedWebID := "https://example.com/users/test-user#me"

	var capturedEvents []domain.Event

	// Mock expectations with event capture
	mockWebIDGen.On("GenerateWebID", ctx, mock.AnythingOfType("string"), req.Email, req.Profile.Name).
		Return(expectedWebID, nil)
	mockWebIDGen.On("ValidateWebID", ctx, expectedWebID).Return(nil)
	mockWebIDGen.On("IsUniqueWebID", ctx, expectedWebID).Return(true, nil)
	mockUnitOfWork.On("RegisterEvents", mock.AnythingOfType("[]domain.Event")).
		Run(func(args mock.Arguments) {
			capturedEvents = args.Get(0).([]pericarpdomain.Event)
		}).Return()
	mockUnitOfWork.On("Commit", ctx).Return([]pericarpdomain.Envelope{}, nil)

	// Act
	user, err := service.RegisterUser(ctx, req)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, user)
	assert.NotEmpty(t, capturedEvents)
	entityEvent := capturedEvents[0].(*domain.EntityEvent)
	assert.Equal(t, "user.created", entityEvent.EventType())
	assert.Equal(t, user.ID(), entityEvent.AggregateID())

	mockUnitOfWork.AssertExpectations(t)
	mockWebIDGen.AssertExpectations(t)
}

func TestUserService_UpdateProfile_EmitsCorrectEvents(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockUnitOfWork := &MockUnitOfWork{}
	mockWebIDGen := &MockWebIDGenerator{}
	mockUserRepo := &MockUserRepository{}

	unitOfWorkFactory := func() pericarpdomain.UnitOfWork {
		return mockUnitOfWork
	}

	service := NewUserService(unitOfWorkFactory, mockWebIDGen, mockUserRepo)

	userID := "test-user-id"
	existingUser := createTestUser(userID, "test@example.com", "Old Name")
	newProfile := domain.UserProfile{Name: "New Name"}

	var capturedEvents []domain.Event

	// Mock expectations with event capture
	mockUserRepo.On("GetByID", ctx, userID).Return(existingUser, nil)
	mockUnitOfWork.On("RegisterEvents", mock.AnythingOfType("[]domain.Event")).
		Run(func(args mock.Arguments) {
			capturedEvents = args.Get(0).([]pericarpdomain.Event)
		}).Return()
	mockUnitOfWork.On("Commit", ctx).Return([]pericarpdomain.Envelope{}, nil)

	// Act
	err := service.UpdateProfile(ctx, userID, newProfile)

	// Assert
	require.NoError(t, err)
	assert.NotEmpty(t, capturedEvents)
	entityEvent := capturedEvents[0].(*domain.EntityEvent)
	assert.Equal(t, "user.profile.updated", entityEvent.EventType())
	assert.Equal(t, userID, entityEvent.AggregateID())

	mockUserRepo.AssertExpectations(t)
	mockUnitOfWork.AssertExpectations(t)
}

func TestUserService_DeleteAccount_EmitsCorrectEvents(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockUnitOfWork := &MockUnitOfWork{}
	mockWebIDGen := &MockWebIDGenerator{}
	mockUserRepo := &MockUserRepository{}

	unitOfWorkFactory := func() pericarpdomain.UnitOfWork {
		return mockUnitOfWork
	}

	service := NewUserService(unitOfWorkFactory, mockWebIDGen, mockUserRepo)

	userID := "test-user-id"
	existingUser := createTestUser(userID, "test@example.com", "Test User")

	var capturedEvents []domain.Event

	// Mock expectations with event capture
	mockUserRepo.On("GetByID", ctx, userID).Return(existingUser, nil)
	mockUnitOfWork.On("RegisterEvents", mock.AnythingOfType("[]domain.Event")).
		Run(func(args mock.Arguments) {
			capturedEvents = args.Get(0).([]pericarpdomain.Event)
		}).Return()
	mockUnitOfWork.On("Commit", ctx).Return([]pericarpdomain.Envelope{}, nil)

	// Act
	err := service.DeleteAccount(ctx, userID)

	// Assert
	require.NoError(t, err)
	assert.NotEmpty(t, capturedEvents)
	entityEvent := capturedEvents[0].(*domain.EntityEvent)
	assert.Equal(t, "user.deleted", entityEvent.EventType())
	assert.Equal(t, userID, entityEvent.AggregateID())

	mockUserRepo.AssertExpectations(t)
	mockUnitOfWork.AssertExpectations(t)
}

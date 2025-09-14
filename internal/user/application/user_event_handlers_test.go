package application_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/akeemphilbert/goro/internal/user/domain"
)

// MockUserWriteRepository is a mock implementation of UserWriteRepository
type MockUserWriteRepository struct {
	mock.Mock
}

func (m *MockUserWriteRepository) Create(ctx context.Context, user *domain.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserWriteRepository) Update(ctx context.Context, user *domain.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserWriteRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// MockFileStorage is a mock implementation of FileStorage
type MockFileStorage struct {
	mock.Mock
}

func (m *MockFileStorage) WriteUserProfile(ctx context.Context, userID string, profile domain.UserProfile) error {
	args := m.Called(ctx, userID, profile)
	return args.Error(0)
}

func (m *MockFileStorage) WriteWebIDDocument(ctx context.Context, userID, webID, document string) error {
	args := m.Called(ctx, userID, webID, document)
	return args.Error(0)
}

func (m *MockFileStorage) DeleteUserFiles(ctx context.Context, userID string) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockFileStorage) ReadUserProfile(ctx context.Context, userID string) (domain.UserProfile, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(domain.UserProfile), args.Error(1)
}

func (m *MockFileStorage) ReadWebIDDocument(ctx context.Context, userID string) (string, error) {
	args := m.Called(ctx, userID)
	return args.String(0), args.Error(1)
}

func (m *MockFileStorage) UserExists(ctx context.Context, userID string) (bool, error) {
	args := m.Called(ctx, userID)
	return args.Bool(0), args.Error(1)
}

func TestUserEventHandler_HandleUserRegistered_Success(t *testing.T) {
	// Arrange
	mockUserRepo := &MockUserWriteRepository{}
	mockFileStorage := &MockFileStorage{}

	handler := NewUserEventHandler(mockUserRepo, mockFileStorage)

	user, _ := domain.NewUser(context.Background(), "user-123", "https://example.com/users/user-123#me", "john@example.com", domain.UserProfile{
		Name: "John Doe",
		Bio:  "Software Developer",
	})

	eventData := &domain.UserRegisteredEventData{
		BaseEventData: domain.BaseEventData{OccurredAt: time.Now()},
		UserID:        user.ID(),
		User:          user,
	}

	// Set up expectations
	mockUserRepo.On("Create", mock.Anything, user).Return(nil)
	mockFileStorage.On("WriteUserProfile", mock.Anything, user.ID(), user.Profile).Return(nil)

	// Act
	err := handler.HandleUserRegistered(context.Background(), eventData)

	// Assert
	assert.NoError(t, err)
	mockUserRepo.AssertExpectations(t)
	mockFileStorage.AssertExpectations(t)
}

func TestUserEventHandler_HandleUserRegistered_DatabaseError(t *testing.T) {
	// Arrange
	mockUserRepo := &MockUserWriteRepository{}
	mockFileStorage := &MockFileStorage{}

	handler := NewUserEventHandler(mockUserRepo, mockFileStorage)

	user, _ := domain.NewUser(context.Background(), "user-123", "https://example.com/users/user-123#me", "john@example.com", domain.UserProfile{
		Name: "John Doe",
	})

	eventData := &domain.UserRegisteredEventData{
		BaseEventData: domain.BaseEventData{OccurredAt: time.Now()},
		UserID:        user.ID(),
		User:          user,
	}

	expectedError := errors.New("database connection failed")

	// Set up expectations - database fails, file storage should not be called
	mockUserRepo.On("Create", mock.Anything, user).Return(expectedError)

	// Act
	err := handler.HandleUserRegistered(context.Background(), eventData)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to persist user registration")
	mockUserRepo.AssertExpectations(t)
	mockFileStorage.AssertNotCalled(t, "WriteUserProfile")
}

func TestUserEventHandler_HandleWebIDGenerated_Success(t *testing.T) {
	// Arrange
	mockUserRepo := &MockUserWriteRepository{}
	mockFileStorage := &MockFileStorage{}

	handler := NewUserEventHandler(mockUserRepo, mockFileStorage)

	user, _ := domain.NewUser(context.Background(), "user-123", "https://example.com/users/user-123#me", "john@example.com", domain.UserProfile{
		Name: "John Doe",
	})

	webID := "https://example.com/users/user-123#me"

	eventData := &domain.WebIDGeneratedEventData{
		BaseEventData: domain.BaseEventData{OccurredAt: time.Now()},
		UserID:        user.ID(),
		User:          user,
		WebID:         webID,
	}

	// Set up expectations
	mockFileStorage.On("WriteWebIDDocument", mock.Anything, user.ID(), webID, mock.AnythingOfType("string")).Return(nil)

	// Act
	err := handler.HandleWebIDGenerated(context.Background(), eventData)

	// Assert
	assert.NoError(t, err)
	mockFileStorage.AssertExpectations(t)
}

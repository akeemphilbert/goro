package application_test

import (
	"context"
	"testing"

	"github.com/akeemphilbert/goro/internal/auth/application"
	"github.com/akeemphilbert/goro/internal/auth/domain"
	userapplication "github.com/akeemphilbert/goro/internal/user/application"
	userdomain "github.com/akeemphilbert/goro/internal/user/domain"
	"github.com/akeemphilbert/goro/internal/user/infrastructure"
	pericarpdomain "github.com/akeemphilbert/pericarp/pkg/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockUserService is a mock implementation of UserService
type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) RegisterUser(ctx context.Context, req userapplication.RegisterUserRequest) (*userdomain.User, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userdomain.User), args.Error(1)
}

func (m *MockUserService) UpdateProfile(ctx context.Context, userID string, profile userdomain.UserProfile) error {
	args := m.Called(ctx, userID, profile)
	return args.Error(0)
}

func (m *MockUserService) DeleteAccount(ctx context.Context, userID string) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockUserService) GetUserByID(ctx context.Context, userID string) (*userdomain.User, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userdomain.User), args.Error(1)
}

func (m *MockUserService) GetUserByWebID(ctx context.Context, webID string) (*userdomain.User, error) {
	args := m.Called(ctx, webID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userdomain.User), args.Error(1)
}

// MockExternalIdentityRepository is a mock implementation of ExternalIdentityRepository
type MockExternalIdentityRepository struct {
	mock.Mock
}

func (m *MockExternalIdentityRepository) LinkIdentity(ctx context.Context, userID, provider, externalID string) error {
	args := m.Called(ctx, userID, provider, externalID)
	return args.Error(0)
}

func (m *MockExternalIdentityRepository) FindByExternalID(ctx context.Context, provider, externalID string) (string, error) {
	args := m.Called(ctx, provider, externalID)
	return args.String(0), args.Error(1)
}

func (m *MockExternalIdentityRepository) GetLinkedIdentities(ctx context.Context, userID string) ([]*domain.ExternalIdentity, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.ExternalIdentity), args.Error(1)
}

func (m *MockExternalIdentityRepository) UnlinkIdentity(ctx context.Context, userID, provider, externalID string) error {
	args := m.Called(ctx, userID, provider, externalID)
	return args.Error(0)
}

func (m *MockExternalIdentityRepository) UnlinkAllIdentities(ctx context.Context, userID string) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockExternalIdentityRepository) IsLinked(ctx context.Context, provider, externalID string) (bool, error) {
	args := m.Called(ctx, provider, externalID)
	return args.Bool(0), args.Error(1)
}

func (m *MockExternalIdentityRepository) GetByProvider(ctx context.Context, provider string) ([]*domain.ExternalIdentity, error) {
	args := m.Called(ctx, provider)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.ExternalIdentity), args.Error(1)
}

// MockWebIDGenerator is a mock implementation of WebIDGenerator
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

func TestRegistrationService_RegisterWithExternalIdentity(t *testing.T) {
	tests := []struct {
		name          string
		provider      string
		profile       domain.ExternalProfile
		setupMocks    func(*MockUserService, *MockExternalIdentityRepository, *MockWebIDGenerator)
		expectedError string
	}{
		{
			name:     "successful registration with Google",
			provider: "google",
			profile: domain.ExternalProfile{
				ID:       "google-123",
				Email:    "alice@example.com",
				Name:     "Alice Smith",
				Username: "alice.smith",
				Provider: "google",
			},
			setupMocks: func(userSvc *MockUserService, identityRepo *MockExternalIdentityRepository, webidGen *MockWebIDGenerator) {
				// External identity not found (new user)
				identityRepo.On("FindByExternalID", mock.Anything, "google", "google-123").
					Return("", domain.ErrExternalIdentityNotFound)

				// User registration succeeds
				expectedUser := &userdomain.User{
					BasicEntity: pericarpdomain.NewEntity("user-123"),
				}
				userSvc.On("RegisterUser", mock.Anything, mock.MatchedBy(func(req userapplication.RegisterUserRequest) bool {
					return req.Email == "alice@example.com" && req.Profile.Name == "Alice Smith"
				})).Return(expectedUser, nil)

				// Identity linking succeeds
				identityRepo.On("LinkIdentity", mock.Anything, "user-123", "google", "google-123").
					Return(nil)
			},
		},
		{
			name:     "invalid external profile",
			provider: "google",
			profile: domain.ExternalProfile{
				ID:       "", // Missing required field
				Email:    "alice@example.com",
				Name:     "Alice Smith",
				Provider: "google",
			},
			setupMocks:    func(*MockUserService, *MockExternalIdentityRepository, *MockWebIDGenerator) {},
			expectedError: "invalid external profile: missing required fields",
		},
		{
			name:     "external identity already linked",
			provider: "google",
			profile: domain.ExternalProfile{
				ID:       "google-123",
				Email:    "alice@example.com",
				Name:     "Alice Smith",
				Username: "alice.smith",
				Provider: "google",
			},
			setupMocks: func(userSvc *MockUserService, identityRepo *MockExternalIdentityRepository, webidGen *MockWebIDGenerator) {
				// External identity already linked to another user
				identityRepo.On("FindByExternalID", mock.Anything, "google", "google-123").
					Return("existing-user-123", nil)
			},
			expectedError: "external identity already linked",
		},
		{
			name:     "user registration fails",
			provider: "github",
			profile: domain.ExternalProfile{
				ID:       "github-456",
				Email:    "bob@example.com",
				Name:     "Bob Johnson",
				Username: "bob.johnson",
				Provider: "github",
			},
			setupMocks: func(userSvc *MockUserService, identityRepo *MockExternalIdentityRepository, webidGen *MockWebIDGenerator) {
				// External identity not found (new user)
				identityRepo.On("FindByExternalID", mock.Anything, "github", "github-456").
					Return("", domain.ErrExternalIdentityNotFound)

				// User registration fails
				userSvc.On("RegisterUser", mock.Anything, mock.Anything).
					Return(nil, assert.AnError)
			},
			expectedError: "failed to register user",
		},
		{
			name:     "identity linking fails",
			provider: "github",
			profile: domain.ExternalProfile{
				ID:       "github-456",
				Email:    "bob@example.com",
				Name:     "Bob Johnson",
				Username: "bob.johnson",
				Provider: "github",
			},
			setupMocks: func(userSvc *MockUserService, identityRepo *MockExternalIdentityRepository, webidGen *MockWebIDGenerator) {
				// External identity not found (new user)
				identityRepo.On("FindByExternalID", mock.Anything, "github", "github-456").
					Return("", domain.ErrExternalIdentityNotFound)

				// User registration succeeds
				expectedUser := &userdomain.User{
					BasicEntity: pericarpdomain.NewEntity("user-456"),
				}
				userSvc.On("RegisterUser", mock.Anything, mock.Anything).
					Return(expectedUser, nil)

				// Identity linking fails
				identityRepo.On("LinkIdentity", mock.Anything, "user-456", "github", "github-456").
					Return(assert.AnError)
			},
			expectedError: "failed to link external identity",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockUserService := &MockUserService{}
			mockIdentityRepo := &MockExternalIdentityRepository{}
			mockWebIDGen := &MockWebIDGenerator{}

			tt.setupMocks(mockUserService, mockIdentityRepo, mockWebIDGen)

			// Create service
			service := application.NewRegistrationService(mockUserService, mockIdentityRepo, mockWebIDGen)

			// Execute
			ctx := context.Background()
			user, err := service.RegisterWithExternalIdentity(ctx, tt.provider, tt.profile)

			// Assert
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, user)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, user)
			}

			// Verify all expectations
			mockUserService.AssertExpectations(t)
			mockIdentityRepo.AssertExpectations(t)
			mockWebIDGen.AssertExpectations(t)
		})
	}
}

func TestRegistrationService_LinkExternalIdentity(t *testing.T) {
	tests := []struct {
		name          string
		userID        string
		provider      string
		profile       domain.ExternalProfile
		setupMocks    func(*MockUserService, *MockExternalIdentityRepository, *MockWebIDGenerator)
		expectedError string
	}{
		{
			name:     "successful identity linking",
			userID:   "user-123",
			provider: "google",
			profile: domain.ExternalProfile{
				ID:       "google-789",
				Email:    "alice@example.com",
				Name:     "Alice Smith",
				Username: "alice.smith",
				Provider: "google",
			},
			setupMocks: func(userSvc *MockUserService, identityRepo *MockExternalIdentityRepository, webidGen *MockWebIDGenerator) {
				// User exists
				existingUser := &userdomain.User{
					BasicEntity: pericarpdomain.NewEntity("user-123"),
				}
				userSvc.On("GetUserByID", mock.Anything, "user-123").
					Return(existingUser, nil)

				// External identity not found (available for linking)
				identityRepo.On("FindByExternalID", mock.Anything, "google", "google-789").
					Return("", domain.ErrExternalIdentityNotFound)

				// Identity linking succeeds
				identityRepo.On("LinkIdentity", mock.Anything, "user-123", "google", "google-789").
					Return(nil)
			},
		},
		{
			name:          "empty user ID",
			userID:        "",
			provider:      "google",
			profile:       domain.ExternalProfile{ID: "google-789", Email: "alice@example.com", Name: "Alice", Provider: "google"},
			setupMocks:    func(*MockUserService, *MockExternalIdentityRepository, *MockWebIDGenerator) {},
			expectedError: "user ID is required",
		},
		{
			name:          "empty provider",
			userID:        "user-123",
			provider:      "",
			profile:       domain.ExternalProfile{ID: "google-789", Email: "alice@example.com", Name: "Alice", Provider: "google"},
			setupMocks:    func(*MockUserService, *MockExternalIdentityRepository, *MockWebIDGenerator) {},
			expectedError: "provider is required",
		},
		{
			name:     "invalid external profile",
			userID:   "user-123",
			provider: "google",
			profile: domain.ExternalProfile{
				ID:       "", // Missing required field
				Email:    "alice@example.com",
				Name:     "Alice Smith",
				Provider: "google",
			},
			setupMocks:    func(*MockUserService, *MockExternalIdentityRepository, *MockWebIDGenerator) {},
			expectedError: "invalid external profile: missing required fields",
		},
		{
			name:     "user not found",
			userID:   "nonexistent-user",
			provider: "google",
			profile: domain.ExternalProfile{
				ID:       "google-789",
				Email:    "alice@example.com",
				Name:     "Alice Smith",
				Provider: "google",
			},
			setupMocks: func(userSvc *MockUserService, identityRepo *MockExternalIdentityRepository, webidGen *MockWebIDGenerator) {
				// User not found
				userSvc.On("GetUserByID", mock.Anything, "nonexistent-user").
					Return(nil, assert.AnError)
			},
			expectedError: "user not found",
		},
		{
			name:     "external identity already linked to same user",
			userID:   "user-123",
			provider: "google",
			profile: domain.ExternalProfile{
				ID:       "google-789",
				Email:    "alice@example.com",
				Name:     "Alice Smith",
				Provider: "google",
			},
			setupMocks: func(userSvc *MockUserService, identityRepo *MockExternalIdentityRepository, webidGen *MockWebIDGenerator) {
				// User exists
				existingUser := &userdomain.User{
					BasicEntity: pericarpdomain.NewEntity("user-123"),
				}
				userSvc.On("GetUserByID", mock.Anything, "user-123").
					Return(existingUser, nil)

				// External identity already linked to same user
				identityRepo.On("FindByExternalID", mock.Anything, "google", "google-789").
					Return("user-123", nil)
			},
			expectedError: "external identity already linked",
		},
		{
			name:     "external identity already linked to different user",
			userID:   "user-123",
			provider: "google",
			profile: domain.ExternalProfile{
				ID:       "google-789",
				Email:    "alice@example.com",
				Name:     "Alice Smith",
				Provider: "google",
			},
			setupMocks: func(userSvc *MockUserService, identityRepo *MockExternalIdentityRepository, webidGen *MockWebIDGenerator) {
				// User exists
				existingUser := &userdomain.User{
					BasicEntity: pericarpdomain.NewEntity("user-123"),
				}
				userSvc.On("GetUserByID", mock.Anything, "user-123").
					Return(existingUser, nil)

				// External identity already linked to different user
				identityRepo.On("FindByExternalID", mock.Anything, "google", "google-789").
					Return("different-user-456", nil)
			},
			expectedError: "external identity already linked",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockUserService := &MockUserService{}
			mockIdentityRepo := &MockExternalIdentityRepository{}
			mockWebIDGen := &MockWebIDGenerator{}

			tt.setupMocks(mockUserService, mockIdentityRepo, mockWebIDGen)

			// Create service
			service := application.NewRegistrationService(mockUserService, mockIdentityRepo, mockWebIDGen)

			// Execute
			ctx := context.Background()
			err := service.LinkExternalIdentity(ctx, tt.userID, tt.provider, tt.profile)

			// Assert
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
			}

			// Verify all expectations
			mockUserService.AssertExpectations(t)
			mockIdentityRepo.AssertExpectations(t)
			mockWebIDGen.AssertExpectations(t)
		})
	}
}

func TestRegistrationService_UnlinkExternalIdentity(t *testing.T) {
	tests := []struct {
		name          string
		userID        string
		provider      string
		externalID    string
		setupMocks    func(*MockUserService, *MockExternalIdentityRepository, *MockWebIDGenerator)
		expectedError string
	}{
		{
			name:       "successful identity unlinking",
			userID:     "user-123",
			provider:   "google",
			externalID: "google-789",
			setupMocks: func(userSvc *MockUserService, identityRepo *MockExternalIdentityRepository, webidGen *MockWebIDGenerator) {
				// User exists
				existingUser := &userdomain.User{
					BasicEntity: pericarpdomain.NewEntity("user-123"),
				}
				userSvc.On("GetUserByID", mock.Anything, "user-123").
					Return(existingUser, nil)

				// Identity unlinking succeeds
				identityRepo.On("UnlinkIdentity", mock.Anything, "user-123", "google", "google-789").
					Return(nil)
			},
		},
		{
			name:          "empty user ID",
			userID:        "",
			provider:      "google",
			externalID:    "google-789",
			setupMocks:    func(*MockUserService, *MockExternalIdentityRepository, *MockWebIDGenerator) {},
			expectedError: "user ID is required",
		},
		{
			name:          "empty provider",
			userID:        "user-123",
			provider:      "",
			externalID:    "google-789",
			setupMocks:    func(*MockUserService, *MockExternalIdentityRepository, *MockWebIDGenerator) {},
			expectedError: "provider is required",
		},
		{
			name:          "empty external ID",
			userID:        "user-123",
			provider:      "google",
			externalID:    "",
			setupMocks:    func(*MockUserService, *MockExternalIdentityRepository, *MockWebIDGenerator) {},
			expectedError: "external ID is required",
		},
		{
			name:       "user not found",
			userID:     "nonexistent-user",
			provider:   "google",
			externalID: "google-789",
			setupMocks: func(userSvc *MockUserService, identityRepo *MockExternalIdentityRepository, webidGen *MockWebIDGenerator) {
				// User not found
				userSvc.On("GetUserByID", mock.Anything, "nonexistent-user").
					Return(nil, assert.AnError)
			},
			expectedError: "user not found",
		},
		{
			name:       "identity unlinking fails",
			userID:     "user-123",
			provider:   "google",
			externalID: "google-789",
			setupMocks: func(userSvc *MockUserService, identityRepo *MockExternalIdentityRepository, webidGen *MockWebIDGenerator) {
				// User exists
				existingUser := &userdomain.User{
					BasicEntity: pericarpdomain.NewEntity("user-123"),
				}
				userSvc.On("GetUserByID", mock.Anything, "user-123").
					Return(existingUser, nil)

				// Identity unlinking fails
				identityRepo.On("UnlinkIdentity", mock.Anything, "user-123", "google", "google-789").
					Return(assert.AnError)
			},
			expectedError: "failed to unlink external identity",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockUserService := &MockUserService{}
			mockIdentityRepo := &MockExternalIdentityRepository{}
			mockWebIDGen := &MockWebIDGenerator{}

			tt.setupMocks(mockUserService, mockIdentityRepo, mockWebIDGen)

			// Create service
			service := application.NewRegistrationService(mockUserService, mockIdentityRepo, mockWebIDGen)

			// Execute
			ctx := context.Background()
			err := service.UnlinkExternalIdentity(ctx, tt.userID, tt.provider, tt.externalID)

			// Assert
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
			}

			// Verify all expectations
			mockUserService.AssertExpectations(t)
			mockIdentityRepo.AssertExpectations(t)
			mockWebIDGen.AssertExpectations(t)
		})
	}
}

func TestRegistrationService_GetLinkedIdentities(t *testing.T) {
	tests := []struct {
		name          string
		userID        string
		setupMocks    func(*MockUserService, *MockExternalIdentityRepository, *MockWebIDGenerator)
		expectedError string
		expectedCount int
	}{
		{
			name:   "successful retrieval of linked identities",
			userID: "user-123",
			setupMocks: func(userSvc *MockUserService, identityRepo *MockExternalIdentityRepository, webidGen *MockWebIDGenerator) {
				// User exists
				existingUser := &userdomain.User{
					BasicEntity: pericarpdomain.NewEntity("user-123"),
				}
				userSvc.On("GetUserByID", mock.Anything, "user-123").
					Return(existingUser, nil)

				// Return linked identities
				identities := []*domain.ExternalIdentity{
					{UserID: "user-123", Provider: "google", ExternalID: "google-789"},
					{UserID: "user-123", Provider: "github", ExternalID: "github-456"},
				}
				identityRepo.On("GetLinkedIdentities", mock.Anything, "user-123").
					Return(identities, nil)
			},
			expectedCount: 2,
		},
		{
			name:          "empty user ID",
			userID:        "",
			setupMocks:    func(*MockUserService, *MockExternalIdentityRepository, *MockWebIDGenerator) {},
			expectedError: "user ID is required",
		},
		{
			name:   "user not found",
			userID: "nonexistent-user",
			setupMocks: func(userSvc *MockUserService, identityRepo *MockExternalIdentityRepository, webidGen *MockWebIDGenerator) {
				// User not found
				userSvc.On("GetUserByID", mock.Anything, "nonexistent-user").
					Return(nil, assert.AnError)
			},
			expectedError: "user not found",
		},
		{
			name:   "repository error",
			userID: "user-123",
			setupMocks: func(userSvc *MockUserService, identityRepo *MockExternalIdentityRepository, webidGen *MockWebIDGenerator) {
				// User exists
				existingUser := &userdomain.User{
					BasicEntity: pericarpdomain.NewEntity("user-123"),
				}
				userSvc.On("GetUserByID", mock.Anything, "user-123").
					Return(existingUser, nil)

				// Repository error
				identityRepo.On("GetLinkedIdentities", mock.Anything, "user-123").
					Return(nil, assert.AnError)
			},
			expectedError: "failed to get linked identities",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockUserService := &MockUserService{}
			mockIdentityRepo := &MockExternalIdentityRepository{}
			mockWebIDGen := &MockWebIDGenerator{}

			tt.setupMocks(mockUserService, mockIdentityRepo, mockWebIDGen)

			// Create service
			service := application.NewRegistrationService(mockUserService, mockIdentityRepo, mockWebIDGen)

			// Execute
			ctx := context.Background()
			identities, err := service.GetLinkedIdentities(ctx, tt.userID)

			// Assert
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, identities)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, identities)
				assert.Len(t, identities, tt.expectedCount)
			}

			// Verify all expectations
			mockUserService.AssertExpectations(t)
			mockIdentityRepo.AssertExpectations(t)
			mockWebIDGen.AssertExpectations(t)
		})
	}
}

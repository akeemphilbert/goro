package application_test

import (
	"context"
	"testing"

	"github.com/akeemphilbert/goro/internal/auth/application"
	"github.com/akeemphilbert/goro/internal/auth/domain"
	userdomain "github.com/akeemphilbert/goro/internal/user/domain"
	pericarpdomain "github.com/akeemphilbert/pericarp/pkg/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestIdentityLinkingService_LinkIdentity(t *testing.T) {
	tests := []struct {
		name          string
		userID        string
		provider      string
		profile       domain.ExternalProfile
		setupMocks    func(*MockUserService, *MockExternalIdentityRepository)
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
			setupMocks: func(userSvc *MockUserService, identityRepo *MockExternalIdentityRepository) {
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
			name:     "empty user ID",
			userID:   "",
			provider: "google",
			profile: domain.ExternalProfile{
				ID:       "google-789",
				Email:    "alice@example.com",
				Name:     "Alice Smith",
				Provider: "google",
			},
			setupMocks:    func(*MockUserService, *MockExternalIdentityRepository) {},
			expectedError: "user ID is empty",
		},
		{
			name:     "empty provider",
			userID:   "user-123",
			provider: "",
			profile: domain.ExternalProfile{
				ID:       "google-789",
				Email:    "alice@example.com",
				Name:     "Alice Smith",
				Provider: "google",
			},
			setupMocks:    func(*MockUserService, *MockExternalIdentityRepository) {},
			expectedError: "provider is empty",
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
			setupMocks:    func(*MockUserService, *MockExternalIdentityRepository) {},
			expectedError: "external identity is invalid",
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
			setupMocks: func(userSvc *MockUserService, identityRepo *MockExternalIdentityRepository) {
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
			setupMocks: func(userSvc *MockUserService, identityRepo *MockExternalIdentityRepository) {
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
			setupMocks: func(userSvc *MockUserService, identityRepo *MockExternalIdentityRepository) {
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
		{
			name:     "identity linking repository error",
			userID:   "user-123",
			provider: "google",
			profile: domain.ExternalProfile{
				ID:       "google-789",
				Email:    "alice@example.com",
				Name:     "Alice Smith",
				Provider: "google",
			},
			setupMocks: func(userSvc *MockUserService, identityRepo *MockExternalIdentityRepository) {
				// User exists
				existingUser := &userdomain.User{
					BasicEntity: pericarpdomain.NewEntity("user-123"),
				}
				userSvc.On("GetUserByID", mock.Anything, "user-123").
					Return(existingUser, nil)

				// External identity not found (available for linking)
				identityRepo.On("FindByExternalID", mock.Anything, "google", "google-789").
					Return("", domain.ErrExternalIdentityNotFound)

				// Identity linking fails
				identityRepo.On("LinkIdentity", mock.Anything, "user-123", "google", "google-789").
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

			tt.setupMocks(mockUserService, mockIdentityRepo)

			// Create service
			service := application.NewIdentityLinkingService(mockUserService, mockIdentityRepo)

			// Execute
			ctx := context.Background()
			err := service.LinkIdentity(ctx, tt.userID, tt.provider, tt.profile)

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
		})
	}
}

func TestIdentityLinkingService_UnlinkIdentity(t *testing.T) {
	tests := []struct {
		name          string
		userID        string
		provider      string
		externalID    string
		setupMocks    func(*MockUserService, *MockExternalIdentityRepository)
		expectedError string
	}{
		{
			name:       "successful identity unlinking",
			userID:     "user-123",
			provider:   "google",
			externalID: "google-789",
			setupMocks: func(userSvc *MockUserService, identityRepo *MockExternalIdentityRepository) {
				// User exists
				existingUser := &userdomain.User{
					BasicEntity: pericarpdomain.NewEntity("user-123"),
				}
				userSvc.On("GetUserByID", mock.Anything, "user-123").
					Return(existingUser, nil)

				// Identity is linked to this user
				identityRepo.On("FindByExternalID", mock.Anything, "google", "google-789").
					Return("user-123", nil)

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
			setupMocks:    func(*MockUserService, *MockExternalIdentityRepository) {},
			expectedError: "user ID is empty",
		},
		{
			name:          "empty provider",
			userID:        "user-123",
			provider:      "",
			externalID:    "google-789",
			setupMocks:    func(*MockUserService, *MockExternalIdentityRepository) {},
			expectedError: "provider is empty",
		},
		{
			name:          "empty external ID",
			userID:        "user-123",
			provider:      "google",
			externalID:    "",
			setupMocks:    func(*MockUserService, *MockExternalIdentityRepository) {},
			expectedError: "external ID is empty",
		},
		{
			name:       "user not found",
			userID:     "nonexistent-user",
			provider:   "google",
			externalID: "google-789",
			setupMocks: func(userSvc *MockUserService, identityRepo *MockExternalIdentityRepository) {
				// User not found
				userSvc.On("GetUserByID", mock.Anything, "nonexistent-user").
					Return(nil, assert.AnError)
			},
			expectedError: "user not found",
		},
		{
			name:       "external identity not found",
			userID:     "user-123",
			provider:   "google",
			externalID: "google-789",
			setupMocks: func(userSvc *MockUserService, identityRepo *MockExternalIdentityRepository) {
				// User exists
				existingUser := &userdomain.User{
					BasicEntity: pericarpdomain.NewEntity("user-123"),
				}
				userSvc.On("GetUserByID", mock.Anything, "user-123").
					Return(existingUser, nil)

				// Identity not found
				identityRepo.On("FindByExternalID", mock.Anything, "google", "google-789").
					Return("", domain.ErrExternalIdentityNotFound)
			},
			expectedError: "external identity not found",
		},
		{
			name:       "external identity linked to different user",
			userID:     "user-123",
			provider:   "google",
			externalID: "google-789",
			setupMocks: func(userSvc *MockUserService, identityRepo *MockExternalIdentityRepository) {
				// User exists
				existingUser := &userdomain.User{
					BasicEntity: pericarpdomain.NewEntity("user-123"),
				}
				userSvc.On("GetUserByID", mock.Anything, "user-123").
					Return(existingUser, nil)

				// Identity is linked to different user
				identityRepo.On("FindByExternalID", mock.Anything, "google", "google-789").
					Return("different-user-456", nil)
			},
			expectedError: "external identity is not linked to this user",
		},
		{
			name:       "identity unlinking repository error",
			userID:     "user-123",
			provider:   "google",
			externalID: "google-789",
			setupMocks: func(userSvc *MockUserService, identityRepo *MockExternalIdentityRepository) {
				// User exists
				existingUser := &userdomain.User{
					BasicEntity: pericarpdomain.NewEntity("user-123"),
				}
				userSvc.On("GetUserByID", mock.Anything, "user-123").
					Return(existingUser, nil)

				// Identity is linked to this user
				identityRepo.On("FindByExternalID", mock.Anything, "google", "google-789").
					Return("user-123", nil)

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

			tt.setupMocks(mockUserService, mockIdentityRepo)

			// Create service
			service := application.NewIdentityLinkingService(mockUserService, mockIdentityRepo)

			// Execute
			ctx := context.Background()
			err := service.UnlinkIdentity(ctx, tt.userID, tt.provider, tt.externalID)

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
		})
	}
}

func TestIdentityLinkingService_GetLinkedIdentities(t *testing.T) {
	tests := []struct {
		name          string
		userID        string
		setupMocks    func(*MockUserService, *MockExternalIdentityRepository)
		expectedError string
		expectedCount int
	}{
		{
			name:   "successful retrieval of linked identities",
			userID: "user-123",
			setupMocks: func(userSvc *MockUserService, identityRepo *MockExternalIdentityRepository) {
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
			setupMocks:    func(*MockUserService, *MockExternalIdentityRepository) {},
			expectedError: "user ID is empty",
		},
		{
			name:   "user not found",
			userID: "nonexistent-user",
			setupMocks: func(userSvc *MockUserService, identityRepo *MockExternalIdentityRepository) {
				// User not found
				userSvc.On("GetUserByID", mock.Anything, "nonexistent-user").
					Return(nil, assert.AnError)
			},
			expectedError: "user not found",
		},
		{
			name:   "repository error",
			userID: "user-123",
			setupMocks: func(userSvc *MockUserService, identityRepo *MockExternalIdentityRepository) {
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

			tt.setupMocks(mockUserService, mockIdentityRepo)

			// Create service
			service := application.NewIdentityLinkingService(mockUserService, mockIdentityRepo)

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
		})
	}
}

func TestIdentityLinkingService_UnlinkAllIdentities(t *testing.T) {
	tests := []struct {
		name          string
		userID        string
		setupMocks    func(*MockUserService, *MockExternalIdentityRepository)
		expectedError string
	}{
		{
			name:   "successful unlinking of all identities",
			userID: "user-123",
			setupMocks: func(userSvc *MockUserService, identityRepo *MockExternalIdentityRepository) {
				// User exists
				existingUser := &userdomain.User{
					BasicEntity: pericarpdomain.NewEntity("user-123"),
				}
				userSvc.On("GetUserByID", mock.Anything, "user-123").
					Return(existingUser, nil)

				// Unlink all identities succeeds
				identityRepo.On("UnlinkAllIdentities", mock.Anything, "user-123").
					Return(nil)
			},
		},
		{
			name:          "empty user ID",
			userID:        "",
			setupMocks:    func(*MockUserService, *MockExternalIdentityRepository) {},
			expectedError: "user ID is empty",
		},
		{
			name:   "user not found",
			userID: "nonexistent-user",
			setupMocks: func(userSvc *MockUserService, identityRepo *MockExternalIdentityRepository) {
				// User not found
				userSvc.On("GetUserByID", mock.Anything, "nonexistent-user").
					Return(nil, assert.AnError)
			},
			expectedError: "user not found",
		},
		{
			name:   "repository error",
			userID: "user-123",
			setupMocks: func(userSvc *MockUserService, identityRepo *MockExternalIdentityRepository) {
				// User exists
				existingUser := &userdomain.User{
					BasicEntity: pericarpdomain.NewEntity("user-123"),
				}
				userSvc.On("GetUserByID", mock.Anything, "user-123").
					Return(existingUser, nil)

				// Repository error
				identityRepo.On("UnlinkAllIdentities", mock.Anything, "user-123").
					Return(assert.AnError)
			},
			expectedError: "failed to unlink all external identities",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockUserService := &MockUserService{}
			mockIdentityRepo := &MockExternalIdentityRepository{}

			tt.setupMocks(mockUserService, mockIdentityRepo)

			// Create service
			service := application.NewIdentityLinkingService(mockUserService, mockIdentityRepo)

			// Execute
			ctx := context.Background()
			err := service.UnlinkAllIdentities(ctx, tt.userID)

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
		})
	}
}

func TestIdentityLinkingService_IsIdentityLinked(t *testing.T) {
	tests := []struct {
		name           string
		provider       string
		externalID     string
		setupMocks     func(*MockUserService, *MockExternalIdentityRepository)
		expectedError  string
		expectedLinked bool
		expectedUserID string
	}{
		{
			name:       "identity is linked",
			provider:   "google",
			externalID: "google-789",
			setupMocks: func(userSvc *MockUserService, identityRepo *MockExternalIdentityRepository) {
				// Identity is linked
				identityRepo.On("FindByExternalID", mock.Anything, "google", "google-789").
					Return("user-123", nil)
			},
			expectedLinked: true,
			expectedUserID: "user-123",
		},
		{
			name:       "identity is not linked",
			provider:   "google",
			externalID: "google-789",
			setupMocks: func(userSvc *MockUserService, identityRepo *MockExternalIdentityRepository) {
				// Identity not found
				identityRepo.On("FindByExternalID", mock.Anything, "google", "google-789").
					Return("", domain.ErrExternalIdentityNotFound)
			},
			expectedLinked: false,
			expectedUserID: "",
		},
		{
			name:          "empty provider",
			provider:      "",
			externalID:    "google-789",
			setupMocks:    func(*MockUserService, *MockExternalIdentityRepository) {},
			expectedError: "provider is empty",
		},
		{
			name:          "empty external ID",
			provider:      "google",
			externalID:    "",
			setupMocks:    func(*MockUserService, *MockExternalIdentityRepository) {},
			expectedError: "external ID is empty",
		},
		{
			name:       "repository error",
			provider:   "google",
			externalID: "google-789",
			setupMocks: func(userSvc *MockUserService, identityRepo *MockExternalIdentityRepository) {
				// Repository error
				identityRepo.On("FindByExternalID", mock.Anything, "google", "google-789").
					Return("", assert.AnError)
			},
			expectedError: "failed to check identity linking",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockUserService := &MockUserService{}
			mockIdentityRepo := &MockExternalIdentityRepository{}

			tt.setupMocks(mockUserService, mockIdentityRepo)

			// Create service
			service := application.NewIdentityLinkingService(mockUserService, mockIdentityRepo)

			// Execute
			ctx := context.Background()
			isLinked, userID, err := service.IsIdentityLinked(ctx, tt.provider, tt.externalID)

			// Assert
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.False(t, isLinked)
				assert.Empty(t, userID)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedLinked, isLinked)
				assert.Equal(t, tt.expectedUserID, userID)
			}

			// Verify all expectations
			mockUserService.AssertExpectations(t)
			mockIdentityRepo.AssertExpectations(t)
		})
	}
}

func TestIdentityLinkingService_GetIdentitiesByProvider(t *testing.T) {
	tests := []struct {
		name          string
		provider      string
		setupMocks    func(*MockUserService, *MockExternalIdentityRepository)
		expectedError string
		expectedCount int
	}{
		{
			name:     "successful retrieval by provider",
			provider: "google",
			setupMocks: func(userSvc *MockUserService, identityRepo *MockExternalIdentityRepository) {
				// Return identities for provider
				identities := []*domain.ExternalIdentity{
					{UserID: "user-123", Provider: "google", ExternalID: "google-789"},
					{UserID: "user-456", Provider: "google", ExternalID: "google-012"},
				}
				identityRepo.On("GetByProvider", mock.Anything, "google").
					Return(identities, nil)
			},
			expectedCount: 2,
		},
		{
			name:          "empty provider",
			provider:      "",
			setupMocks:    func(*MockUserService, *MockExternalIdentityRepository) {},
			expectedError: "provider is empty",
		},
		{
			name:     "repository error",
			provider: "google",
			setupMocks: func(userSvc *MockUserService, identityRepo *MockExternalIdentityRepository) {
				// Repository error
				identityRepo.On("GetByProvider", mock.Anything, "google").
					Return(nil, assert.AnError)
			},
			expectedError: "failed to get identities by provider",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockUserService := &MockUserService{}
			mockIdentityRepo := &MockExternalIdentityRepository{}

			tt.setupMocks(mockUserService, mockIdentityRepo)

			// Create service
			service := application.NewIdentityLinkingService(mockUserService, mockIdentityRepo)

			// Execute
			ctx := context.Background()
			identities, err := service.GetIdentitiesByProvider(ctx, tt.provider)

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
		})
	}
}

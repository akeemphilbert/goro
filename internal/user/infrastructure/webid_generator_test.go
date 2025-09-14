package infrastructure_test

import (
	"context"
	"strings"
	"testing"

	"github.com/akeemphilbert/goro/internal/user/infrastructure"
	"github.com/stretchr/testify/assert"
)

func TestWebIDGenerator_GenerateWebID(t *testing.T) {
	tests := []struct {
		name     string
		userID   string
		email    string
		userName string
		wantErr  bool
		validate func(t *testing.T, webID string)
	}{
		{
			name:     "valid user generates WebID URI",
			userID:   "user-123",
			email:    "john@example.com",
			userName: "John Doe",
			wantErr:  false,
			validate: func(t *testing.T, webID string) {
				assert.True(t, strings.HasPrefix(webID, "https://"), "WebID should start with https://")
				assert.Contains(t, webID, "user-123", "WebID should contain user ID")
				assert.True(t, strings.HasSuffix(webID, "#me"), "WebID should end with #me")
			},
		},
		{
			name:     "empty user ID returns error",
			userID:   "",
			email:    "john@example.com",
			userName: "John Doe",
			wantErr:  true,
		},
		{
			name:     "empty email returns error",
			userID:   "user-123",
			email:    "",
			userName: "John Doe",
			wantErr:  true,
		},
		{
			name:     "empty user name returns error",
			userID:   "user-123",
			email:    "john@example.com",
			userName: "",
			wantErr:  true,
		},
		{
			name:     "special characters in user ID are handled",
			userID:   "user-with-special@chars",
			email:    "john@example.com",
			userName: "John Doe",
			wantErr:  false,
			validate: func(t *testing.T, webID string) {
				assert.True(t, strings.HasPrefix(webID, "https://"), "WebID should start with https://")
				assert.NotContains(t, webID, "@", "WebID should not contain @ symbol")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			generator := infrastructure.NewWebIDGenerator("https://example.com")

			webID, err := generator.GenerateWebID(context.Background(), tt.userID, tt.email, tt.userName)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Empty(t, webID)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, webID)
				if tt.validate != nil {
					tt.validate(t, webID)
				}
			}
		})
	}
}

func TestWebIDGenerator_GenerateWebIDDocument(t *testing.T) {
	tests := []struct {
		name     string
		webID    string
		email    string
		userName string
		wantErr  bool
		validate func(t *testing.T, document string)
	}{
		{
			name:     "valid inputs generate Turtle document",
			webID:    "https://example.com/users/user-123#me",
			email:    "john@example.com",
			userName: "John Doe",
			wantErr:  false,
			validate: func(t *testing.T, document string) {
				assert.Contains(t, document, "@prefix foaf:", "Document should contain foaf prefix")
				assert.Contains(t, document, "@prefix solid:", "Document should contain solid prefix")
				assert.Contains(t, document, "@prefix ldp:", "Document should contain ldp prefix")
				assert.Contains(t, document, "https://example.com/users/user-123#me", "Document should contain WebID URI")
				assert.Contains(t, document, "foaf:Person", "Document should declare person type")
				assert.Contains(t, document, "John Doe", "Document should contain user name")
				assert.Contains(t, document, "mailto:john@example.com", "Document should contain email")
			},
		},
		{
			name:     "empty WebID returns error",
			webID:    "",
			email:    "john@example.com",
			userName: "John Doe",
			wantErr:  true,
		},
		{
			name:     "empty email returns error",
			webID:    "https://example.com/users/user-123#me",
			email:    "",
			userName: "John Doe",
			wantErr:  true,
		},
		{
			name:     "empty user name returns error",
			webID:    "https://example.com/users/user-123#me",
			email:    "john@example.com",
			userName: "",
			wantErr:  true,
		},
		{
			name:     "special characters in name are escaped",
			webID:    "https://example.com/users/user-123#me",
			email:    "john@example.com",
			userName: "John \"Doe\" O'Connor",
			wantErr:  false,
			validate: func(t *testing.T, document string) {
				assert.Contains(t, document, "John \\\"Doe\\\" O'Connor", "Document should escape quotes in name")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			generator := infrastructure.NewWebIDGenerator("https://example.com")

			document, err := generator.GenerateWebIDDocument(context.Background(), tt.webID, tt.email, tt.userName)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Empty(t, document)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, document)
				if tt.validate != nil {
					tt.validate(t, document)
				}
			}
		})
	}
}

func TestWebIDGenerator_ValidateWebID(t *testing.T) {
	tests := []struct {
		name    string
		webID   string
		wantErr bool
	}{
		{
			name:    "valid HTTPS WebID passes validation",
			webID:   "https://example.com/users/user-123#me",
			wantErr: false,
		},
		{
			name:    "WebID without HTTPS fails validation",
			webID:   "http://example.com/users/user-123#me",
			wantErr: true,
		},
		{
			name:    "WebID without fragment fails validation",
			webID:   "https://example.com/users/user-123",
			wantErr: true,
		},
		{
			name:    "empty WebID fails validation",
			webID:   "",
			wantErr: true,
		},
		{
			name:    "invalid URL format fails validation",
			webID:   "not-a-url",
			wantErr: true,
		},
		{
			name:    "WebID with query parameters is valid",
			webID:   "https://example.com/users/user-123?param=value#me",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			generator := infrastructure.NewWebIDGenerator("https://example.com")

			err := generator.ValidateWebID(context.Background(), tt.webID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestWebIDGenerator_IsUniqueWebID(t *testing.T) {
	tests := []struct {
		name           string
		webID          string
		existingWebIDs []string
		wantUnique     bool
		wantErr        bool
	}{
		{
			name:  "unique WebID returns true",
			webID: "https://example.com/users/user-123#me",
			existingWebIDs: []string{
				"https://example.com/users/user-456#me",
				"https://example.com/users/user-789#me",
			},
			wantUnique: true,
			wantErr:    false,
		},
		{
			name:  "duplicate WebID returns false",
			webID: "https://example.com/users/user-123#me",
			existingWebIDs: []string{
				"https://example.com/users/user-123#me",
				"https://example.com/users/user-456#me",
			},
			wantUnique: false,
			wantErr:    false,
		},
		{
			name:           "empty WebID returns error",
			webID:          "",
			existingWebIDs: []string{},
			wantUnique:     false,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			generator := infrastructure.NewWebIDGenerator("https://example.com")

			// Mock the uniqueness check by creating a simple in-memory store
			mockChecker := &MockWebIDUniquenessChecker{
				existingWebIDs: make(map[string]bool),
			}
			for _, webID := range tt.existingWebIDs {
				mockChecker.existingWebIDs[webID] = true
			}

			// Set the checker on the generator (this will require the interface to be implemented)
			generator.SetUniquenessChecker(mockChecker)

			isUnique, err := generator.IsUniqueWebID(context.Background(), tt.webID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantUnique, isUnique)
			}
		})
	}
}

func TestWebIDGenerator_GenerateAlternativeWebID(t *testing.T) {
	tests := []struct {
		name           string
		baseWebID      string
		existingWebIDs []string
		wantErr        bool
		validate       func(t *testing.T, webID string)
	}{
		{
			name:      "generates alternative when base is taken",
			baseWebID: "https://example.com/users/user-123#me",
			existingWebIDs: []string{
				"https://example.com/users/user-123#me",
			},
			wantErr: false,
			validate: func(t *testing.T, webID string) {
				assert.NotEqual(t, "https://example.com/users/user-123#me", webID)
				assert.Contains(t, webID, "user-123")
				assert.True(t, strings.HasSuffix(webID, "#me"))
			},
		},
		{
			name:      "generates multiple alternatives when needed",
			baseWebID: "https://example.com/users/user-123#me",
			existingWebIDs: []string{
				"https://example.com/users/user-123#me",
				"https://example.com/users/user-123-1#me",
				"https://example.com/users/user-123-2#me",
			},
			wantErr: false,
			validate: func(t *testing.T, webID string) {
				assert.NotContains(t, []string{
					"https://example.com/users/user-123#me",
					"https://example.com/users/user-123-1#me",
					"https://example.com/users/user-123-2#me",
				}, webID)
				assert.Contains(t, webID, "user-123")
			},
		},
		{
			name:           "empty base WebID returns error",
			baseWebID:      "",
			existingWebIDs: []string{},
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			generator := infrastructure.NewWebIDGenerator("https://example.com")

			// Mock the uniqueness check
			mockChecker := &MockWebIDUniquenessChecker{
				existingWebIDs: make(map[string]bool),
			}
			for _, webID := range tt.existingWebIDs {
				mockChecker.existingWebIDs[webID] = true
			}
			generator.SetUniquenessChecker(mockChecker)

			webID, err := generator.GenerateAlternativeWebID(context.Background(), tt.baseWebID)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Empty(t, webID)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, webID)
				if tt.validate != nil {
					tt.validate(t, webID)
				}
			}
		})
	}
}

// MockWebIDUniquenessChecker is a mock implementation for testing
type MockWebIDUniquenessChecker struct {
	existingWebIDs map[string]bool
}

func (m *MockWebIDUniquenessChecker) WebIDExists(ctx context.Context, webID string) (bool, error) {
	if webID == "" {
		return false, infrastructure.ErrInvalidWebID
	}
	return m.existingWebIDs[webID], nil
}

func TestNewWebIDGenerator(t *testing.T) {
	tests := []struct {
		name    string
		baseURL string
		wantErr bool
	}{
		{
			name:    "valid HTTPS base URL creates generator",
			baseURL: "https://example.com",
			wantErr: false,
		},
		{
			name:    "HTTP base URL returns error",
			baseURL: "http://example.com",
			wantErr: true,
		},
		{
			name:    "empty base URL returns error",
			baseURL: "",
			wantErr: true,
		},
		{
			name:    "invalid URL format returns error",
			baseURL: "not-a-url",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			generator, err := infrastructure.NewWebIDGeneratorWithValidation(tt.baseURL)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, generator)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, generator)
			}
		})
	}
}

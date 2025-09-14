package domain

import (
	"context"
	"testing"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/stretchr/testify/require"
)

// MockLogger implements the log.Logger interface for testing
type MockLogger struct {
	logs []string
}

func (m *MockLogger) Log(level log.Level, keyvals ...interface{}) error {
	return nil
}

func TestEntityMethodsWithContextAndLogger(t *testing.T) {
	ctx := context.Background()
	logger := &MockLogger{}

	// Test User entity creation with context and logger
	user, err := NewUser(ctx, "user-123", "https://example.com/user-123#me", "user@example.com", UserProfile{
		Name: "Test User",
		Bio:  "A test user",
	})
	require.NoError(t, err)
	require.Equal(t, "user-123", user.ID())

	// Test User profile update with context and logger
	newProfile := UserProfile{
		Name: "Updated User",
		Bio:  "An updated test user",
	}
	err = user.UpdateProfile(ctx, newProfile)
	require.NoError(t, err)
	require.Equal(t, "Updated User", user.Profile.Name)

	// Test method calls that demonstrate context/logger usage
	err = user.Suspend(ctx)
	require.NoError(t, err)
	require.Equal(t, UserStatusSuspended, user.Status)

	err = user.Activate(ctx)
	require.NoError(t, err)
	require.Equal(t, UserStatusActive, user.Status)

	// Logger is available for testing but not used in current implementation
	_ = logger
}

func TestContextWithUserAndAccountID(t *testing.T) {
	// Demonstrate how context would typically contain user and account information
	ctx := context.Background()

	// In a real application, you would set context values like:
	// ctx = context.WithValue(ctx, "user_id", "user-123")
	// ctx = context.WithValue(ctx, "account_id", "account-456")
	// ctx = context.WithValue(ctx, "request_id", "req-789")

	logger := &MockLogger{}

	// Example of how logging would work with contextual information
	user, err := NewUser(ctx, "user-123", "https://example.com/user-123#me", "user@example.com", UserProfile{
		Name: "Test User",
	})
	require.NoError(t, err)

	// Logger is available for testing but not used in current implementation
	_ = logger

	// The logging inside the NewUser method would have access to:
	// - user_id from context (for authorization/auditing)
	// - account_id from context (for multi-tenant scenarios)
	// - request_id from context (for request tracing)
	// - The actual operation details from the method parameters

	t.Logf("User created successfully with contextual logging: %s", user.ID())
}

package middleware

import (
	"context"
	"testing"
	"time"

	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTimeout(t *testing.T) {
	tests := []struct {
		name          string
		timeout       time.Duration
		handlerDelay  time.Duration
		expectTimeout bool
		expectedError string
	}{
		{
			name:          "Request completes within timeout",
			timeout:       100 * time.Millisecond,
			handlerDelay:  50 * time.Millisecond,
			expectTimeout: false,
		},
		{
			name:          "Request times out",
			timeout:       50 * time.Millisecond,
			handlerDelay:  100 * time.Millisecond,
			expectTimeout: true,
			expectedError: "context deadline exceeded",
		},
		{
			name:          "Zero timeout should not timeout",
			timeout:       0,
			handlerDelay:  100 * time.Millisecond,
			expectTimeout: false,
		},
		{
			name:          "Negative timeout should not timeout",
			timeout:       -1 * time.Second,
			handlerDelay:  100 * time.Millisecond,
			expectTimeout: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create timeout middleware
			timeoutMiddleware := Timeout(tt.timeout)

			// Create a handler that takes some time to complete
			handler := func(ctx context.Context, req interface{}) (interface{}, error) {
				select {
				case <-time.After(tt.handlerDelay):
					return "success", nil
				case <-ctx.Done():
					return nil, ctx.Err()
				}
			}

			// Wrap handler with timeout middleware
			wrappedHandler := timeoutMiddleware(handler)

			// Execute the handler
			ctx := context.Background()
			result, err := wrappedHandler(ctx, "test-request")

			if tt.expectTimeout {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				assert.Equal(t, "success", result)
			}
		})
	}
}

func TestTimeoutWithCancellation(t *testing.T) {
	timeout := 200 * time.Millisecond
	timeoutMiddleware := Timeout(timeout)

	// Create a handler that checks for context cancellation
	handlerCalled := false
	contextCancelled := false

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		handlerCalled = true

		// Simulate some work and check for cancellation
		select {
		case <-time.After(300 * time.Millisecond): // This should be interrupted
			return "should not reach here", nil
		case <-ctx.Done():
			contextCancelled = true
			return nil, ctx.Err()
		}
	}

	wrappedHandler := timeoutMiddleware(handler)

	// Execute the handler
	ctx := context.Background()
	result, err := wrappedHandler(ctx, "test-request")

	// Assertions
	assert.True(t, handlerCalled, "Handler should have been called")
	assert.True(t, contextCancelled, "Context should have been cancelled")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "context deadline exceeded")
	assert.Nil(t, result)
}

func TestTimeoutChaining(t *testing.T) {
	// Test that timeout middleware can be chained with other middleware
	timeout := 100 * time.Millisecond
	timeoutMiddleware := Timeout(timeout)

	// Create a simple logging middleware for testing chaining
	loggingMiddleware := func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			// Add a marker to context to verify middleware was called
			ctx = context.WithValue(ctx, "logged", true)
			return handler(ctx, req)
		}
	}

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		// Verify that both middleware were applied
		logged := ctx.Value("logged")
		if logged != true {
			t.Error("Logging middleware was not applied")
		}

		// Check that timeout context is set
		deadline, ok := ctx.Deadline()
		if !ok {
			t.Error("Timeout context was not set")
		}

		if time.Until(deadline) > timeout {
			t.Error("Timeout deadline is incorrect")
		}

		return "chained success", nil
	}

	// Chain the middleware
	wrappedHandler := timeoutMiddleware(loggingMiddleware(handler))

	// Execute
	ctx := context.Background()
	result, err := wrappedHandler(ctx, "test-request")

	require.NoError(t, err)
	assert.Equal(t, "chained success", result)
}

func TestTimeoutWithExistingDeadline(t *testing.T) {
	// Test behavior when context already has a deadline
	timeout := 200 * time.Millisecond
	timeoutMiddleware := Timeout(timeout)

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		deadline, ok := ctx.Deadline()
		require.True(t, ok, "Context should have a deadline")

		// The deadline should be the shorter of the two (100ms vs 200ms)
		expectedDeadline := time.Now().Add(100 * time.Millisecond)
		timeDiff := deadline.Sub(expectedDeadline).Abs()
		assert.Less(t, timeDiff, 10*time.Millisecond, "Deadline should be close to the shorter timeout")

		return "existing deadline", nil
	}

	wrappedHandler := timeoutMiddleware(handler)

	// Create context with existing shorter deadline
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	result, err := wrappedHandler(ctx, "test-request")

	require.NoError(t, err)
	assert.Equal(t, "existing deadline", result)
}

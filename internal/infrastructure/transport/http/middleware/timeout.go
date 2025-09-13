package middleware

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/middleware"
)

// Timeout returns a middleware that enforces a timeout on request processing
func Timeout(timeout time.Duration) middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			// If timeout is zero or negative, don't apply timeout
			if timeout <= 0 {
				return handler(ctx, req)
			}

			// Check if context already has a deadline
			if existingDeadline, ok := ctx.Deadline(); ok {
				// If existing deadline is sooner, use it
				if time.Until(existingDeadline) < timeout {
					return handler(ctx, req)
				}
			}

			// Create context with timeout
			timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
			defer cancel()

			// Execute handler with timeout context directly
			return handler(timeoutCtx, req)
		}
	}
}

package middleware

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/google/uuid"
)

// Context keys for request information
type contextKey string

const (
	correlationIDKey contextKey = "correlation_id"
	requestInfoKey   contextKey = "request_info"
)

// RequestInfo holds information about the HTTP request
type RequestInfo struct {
	Method string
	Path   string
	Status int
}

// WithCorrelationID adds a correlation ID to the context
func WithCorrelationID(ctx context.Context, correlationID string) context.Context {
	return context.WithValue(ctx, correlationIDKey, correlationID)
}

// GetCorrelationID retrieves the correlation ID from the context
func GetCorrelationID(ctx context.Context) string {
	if correlationID, ok := ctx.Value(correlationIDKey).(string); ok {
		return correlationID
	}
	return ""
}

// WithRequestInfo adds request information to the context
func WithRequestInfo(ctx context.Context, info RequestInfo) context.Context {
	return context.WithValue(ctx, requestInfoKey, info)
}

// GetRequestInfo retrieves request information from the context
func GetRequestInfo(ctx context.Context) RequestInfo {
	if info, ok := ctx.Value(requestInfoKey).(RequestInfo); ok {
		return info
	}
	return RequestInfo{}
}

// StructuredLogging returns a middleware that provides structured logging with correlation IDs
func StructuredLogging(logger log.Logger) middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			// Generate or use existing correlation ID
			correlationID := GetCorrelationID(ctx)
			if correlationID == "" {
				correlationID = generateCorrelationID()
				ctx = WithCorrelationID(ctx, correlationID)
			}

			// Get request info
			requestInfo := GetRequestInfo(ctx)

			// Record start time
			start := time.Now()

			// Execute handler
			result, err := handler(ctx, req)

			// Calculate duration
			duration := time.Since(start)

			// Prepare log fields
			logFields := []interface{}{
				"correlation_id", correlationID,
				"method", requestInfo.Method,
				"path", requestInfo.Path,
				"duration", duration.String(),
				"timestamp", start.Unix(),
			}

			// Log based on result
			if err != nil {
				logFields = append(logFields, "error", err.Error())
				if requestInfo.Status == 0 {
					logFields = append(logFields, "status", 500)
				} else {
					logFields = append(logFields, "status", requestInfo.Status)
				}
				logger.Log(log.LevelError, logFields...)
			} else {
				if requestInfo.Status == 0 {
					logFields = append(logFields, "status", 200)
				} else {
					logFields = append(logFields, "status", requestInfo.Status)
				}
				logger.Log(log.LevelInfo, logFields...)
			}

			return result, err
		}
	}
}

// generateCorrelationID generates a new UUID-based correlation ID
func generateCorrelationID() string {
	return uuid.New().String()
}

// LoggingWithConfig returns a structured logging middleware with custom configuration
type LoggingConfig struct {
	Logger             log.Logger
	SkipPaths          []string
	EnableRequestBody  bool
	EnableResponseBody bool
}

// StructuredLoggingWithConfig returns a middleware with custom logging configuration
func StructuredLoggingWithConfig(config LoggingConfig) middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			requestInfo := GetRequestInfo(ctx)

			// Skip logging for specified paths
			for _, skipPath := range config.SkipPaths {
				if requestInfo.Path == skipPath {
					return handler(ctx, req)
				}
			}

			// Generate or use existing correlation ID
			correlationID := GetCorrelationID(ctx)
			if correlationID == "" {
				correlationID = generateCorrelationID()
				ctx = WithCorrelationID(ctx, correlationID)
			}

			// Record start time
			start := time.Now()

			// Execute handler
			result, err := handler(ctx, req)

			// Calculate duration
			duration := time.Since(start)

			// Prepare log fields
			logFields := []interface{}{
				"correlation_id", correlationID,
				"method", requestInfo.Method,
				"path", requestInfo.Path,
				"duration", duration.String(),
				"timestamp", start.Unix(),
			}

			// Add request body if enabled
			if config.EnableRequestBody && req != nil {
				logFields = append(logFields, "request_body", req)
			}

			// Add response body if enabled
			if config.EnableResponseBody && result != nil {
				logFields = append(logFields, "response_body", result)
			}

			// Log based on result
			logger := config.Logger
			if logger == nil {
				logger = log.DefaultLogger
			}

			if err != nil {
				logFields = append(logFields, "error", err.Error())
				if requestInfo.Status == 0 {
					logFields = append(logFields, "status", 500)
				} else {
					logFields = append(logFields, "status", requestInfo.Status)
				}
				logger.Log(log.LevelError, logFields...)
			} else {
				if requestInfo.Status == 0 {
					logFields = append(logFields, "status", 200)
				} else {
					logFields = append(logFields, "status", requestInfo.Status)
				}
				logger.Log(log.LevelInfo, logFields...)
			}

			return result, err
		}
	}
}

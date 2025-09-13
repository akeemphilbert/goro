package middleware

import (
	"net/http"
	"strconv"
	"strings"

	khttp "github.com/go-kratos/kratos/v2/transport/http"
)

// CORSConfig holds configuration for CORS middleware
type CORSConfig struct {
	AllowedOrigins []string
	AllowedMethods []string
	AllowedHeaders []string
	MaxAge         int
}

// DefaultCORSConfig returns a default CORS configuration
func DefaultCORSConfig() CORSConfig {
	return CORSConfig{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS"},
		AllowedHeaders: []string{"Content-Type", "Authorization", "X-Requested-With"},
		MaxAge:         86400, // 24 hours
	}
}

// CORS returns a CORS filter with default configuration
func CORS() khttp.FilterFunc {
	return CORSWithConfig(DefaultCORSConfig())
}

// CORSWithConfig returns a CORS filter with custom configuration
func CORSWithConfig(config CORSConfig) khttp.FilterFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			// Set allowed origin
			if len(config.AllowedOrigins) > 0 {
				if contains(config.AllowedOrigins, "*") {
					w.Header().Set("Access-Control-Allow-Origin", "*")
				} else if origin != "" && contains(config.AllowedOrigins, origin) {
					w.Header().Set("Access-Control-Allow-Origin", origin)
				}
			}

			// Set allowed methods
			if len(config.AllowedMethods) > 0 {
				w.Header().Set("Access-Control-Allow-Methods", strings.Join(config.AllowedMethods, ","))
			}

			// Set allowed headers
			if len(config.AllowedHeaders) > 0 {
				w.Header().Set("Access-Control-Allow-Headers", strings.Join(config.AllowedHeaders, ","))
			}

			// Set max age for preflight requests
			if config.MaxAge > 0 {
				w.Header().Set("Access-Control-Max-Age", strconv.Itoa(config.MaxAge))
			}

			// Handle preflight OPTIONS request
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			// Continue to next handler
			next.ServeHTTP(w, r)
		})
	}
}

// contains checks if a slice contains a specific string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

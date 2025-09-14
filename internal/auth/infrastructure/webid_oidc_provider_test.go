package infrastructure_test

import (
	"testing"

	"github.com/akeemphilbert/goro/internal/auth/infrastructure"
)

func TestWebIDOIDCProvider_Basic(t *testing.T) {
	config := &infrastructure.WebIDOIDCConfig{
		Timeout:  "30s",
		CacheTTL: "1h",
	}

	provider := infrastructure.NewWebIDOIDCProvider(config)
	if provider == nil {
		t.Error("Expected non-nil provider")
	}
}

func TestJWTValidator_Basic(t *testing.T) {
	validator := infrastructure.NewJWTValidator()
	if validator == nil {
		t.Error("Expected non-nil validator")
	}

	// Test with invalid token format
	_, err := validator.ParseUnverified("invalid-token")
	if err == nil {
		t.Error("Expected error for invalid token format")
	}
}

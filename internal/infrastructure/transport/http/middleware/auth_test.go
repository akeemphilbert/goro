package middleware

import (
	"context"
	"net/http/httptest"
	"testing"
)

func TestExtractToken(t *testing.T) {
	tests := []struct {
		name          string
		authHeader    string
		queryParams   map[string]string
		expectedToken string
	}{
		{
			name:          "Bearer token in header",
			authHeader:    "Bearer abc123",
			expectedToken: "abc123",
		},
		{
			name:          "token in query param",
			queryParams:   map[string]string{"token": "xyz789"},
			expectedToken: "xyz789",
		},
		{
			name:          "no token",
			expectedToken: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)

			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			if len(tt.queryParams) > 0 {
				q := req.URL.Query()
				for key, value := range tt.queryParams {
					q.Set(key, value)
				}
				req.URL.RawQuery = q.Encode()
			}

			token := extractToken(req)

			if token != tt.expectedToken {
				t.Errorf("Expected token %s, got %s", tt.expectedToken, token)
			}
		})
	}
}

func TestAuthInfoContextMethods(t *testing.T) {
	authInfo := AuthInfo{
		UserID:          "user123",
		SessionID:       "session456",
		WebID:           "https://example.com/user123#me",
		AccountID:       "account789",
		RoleID:          "admin",
		IsAuthenticated: true,
	}

	ctx := context.Background()
	ctx = WithAuthInfo(ctx, authInfo)

	if GetUserID(ctx) != authInfo.UserID {
		t.Errorf("Expected user ID %s, got %s", authInfo.UserID, GetUserID(ctx))
	}

	if !IsAuthenticated(ctx) {
		t.Error("Expected authentication to be true")
	}

	retrievedInfo := GetAuthInfo(ctx)
	if retrievedInfo != authInfo {
		t.Errorf("Expected auth info %+v, got %+v", authInfo, retrievedInfo)
	}

	emptyCtx := context.Background()
	if GetUserID(emptyCtx) != "" {
		t.Error("Expected empty user ID from empty context")
	}
	if IsAuthenticated(emptyCtx) {
		t.Error("Expected authentication to be false for empty context")
	}
}

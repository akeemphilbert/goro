package domain_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/akeemphilbert/goro/internal/auth/domain"
	"golang.org/x/crypto/bcrypt"
)

func TestBCryptPasswordHasher_Hash(t *testing.T) {
	hasher := domain.NewBCryptPasswordHasher(bcrypt.DefaultCost)

	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{
			name:     "valid password",
			password: "validPassword123!",
			wantErr:  false,
		},
		{
			name:     "empty password",
			password: "",
			wantErr:  true,
		},
		{
			name:     "long password",
			password: strings.Repeat("a", 100),
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, salt, err := hasher.Hash(tt.password)

			if tt.wantErr {
				if err == nil {
					t.Errorf("BCryptPasswordHasher.Hash() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("BCryptPasswordHasher.Hash() unexpected error: %v", err)
				return
			}

			if hash == "" {
				t.Error("BCryptPasswordHasher.Hash() returned empty hash")
			}

			if salt == "" {
				t.Error("BCryptPasswordHasher.Hash() returned empty salt")
			}

			// Verify the hash can be used for verification
			if !hasher.Verify(tt.password, hash, salt) {
				t.Error("BCryptPasswordHasher.Hash() generated hash that fails verification")
			}
		})
	}
}

func TestBCryptPasswordHasher_Verify(t *testing.T) {
	hasher := domain.NewBCryptPasswordHasher(bcrypt.DefaultCost)
	password := "testPassword123!"
	hash, salt, err := hasher.Hash(password)
	if err != nil {
		t.Fatalf("Failed to hash password for test: %v", err)
	}

	tests := []struct {
		name     string
		password string
		hash     string
		salt     string
		want     bool
	}{
		{
			name:     "correct password",
			password: password,
			hash:     hash,
			salt:     salt,
			want:     true,
		},
		{
			name:     "incorrect password",
			password: "wrongPassword",
			hash:     hash,
			salt:     salt,
			want:     false,
		},
		{
			name:     "empty password",
			password: "",
			hash:     hash,
			salt:     salt,
			want:     false,
		},
		{
			name:     "empty hash",
			password: password,
			hash:     "",
			salt:     salt,
			want:     false,
		},
		{
			name:     "empty salt",
			password: password,
			hash:     hash,
			salt:     "",
			want:     false,
		},
		{
			name:     "wrong salt",
			password: password,
			hash:     hash,
			salt:     "wrongSalt",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hasher.Verify(tt.password, tt.hash, tt.salt)
			if got != tt.want {
				t.Errorf("BCryptPasswordHasher.Verify() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBCryptPasswordHasher_DifferentCosts(t *testing.T) {
	// Use reasonable costs for testing (MaxCost is too expensive for tests)
	costs := []int{bcrypt.MinCost, bcrypt.DefaultCost, 12}
	password := "testPassword123!"

	for _, cost := range costs {
		t.Run(fmt.Sprintf("cost_%d", cost), func(t *testing.T) {
			hasher := domain.NewBCryptPasswordHasher(cost)
			hash, salt, err := hasher.Hash(password)
			if err != nil {
				t.Errorf("BCryptPasswordHasher.Hash() with cost %d failed: %v", cost, err)
				return
			}

			if !hasher.Verify(password, hash, salt) {
				t.Errorf("BCryptPasswordHasher.Verify() failed with cost %d", cost)
			}
		})
	}
}

func TestBCryptPasswordHasher_InvalidCost(t *testing.T) {
	// Test with invalid costs - should still work (implementation handles it internally)
	invalidCosts := []int{-1, 0, bcrypt.MaxCost + 1, 100}
	password := "testPassword123!"

	for _, cost := range invalidCosts {
		hasher := domain.NewBCryptPasswordHasher(cost)
		// Test that it still works regardless of invalid cost
		hash, salt, err := hasher.Hash(password)
		if err != nil {
			t.Errorf("NewBCryptPasswordHasher(%d) should handle invalid cost gracefully, got error: %v", cost, err)
		}
		if !hasher.Verify(password, hash, salt) {
			t.Errorf("NewBCryptPasswordHasher(%d) should still work with invalid cost", cost)
		}
	}
}

func TestCryptoSecureTokenGenerator_GenerateToken(t *testing.T) {
	generator := domain.NewCryptoSecureTokenGenerator(32)

	// Test multiple token generations
	tokens := make(map[string]bool)
	for i := 0; i < 100; i++ {
		token, err := generator.GenerateToken()
		if err != nil {
			t.Errorf("CryptoSecureTokenGenerator.GenerateToken() error: %v", err)
			continue
		}

		if token == "" {
			t.Error("CryptoSecureTokenGenerator.GenerateToken() returned empty token")
			continue
		}

		// Check for uniqueness
		if tokens[token] {
			t.Error("CryptoSecureTokenGenerator.GenerateToken() generated duplicate token")
		}
		tokens[token] = true
	}
}

func TestCryptoSecureTokenGenerator_DifferentLengths(t *testing.T) {
	lengths := []int{16, 32, 64}

	for _, length := range lengths {
		t.Run(string(rune(length)), func(t *testing.T) {
			generator := domain.NewCryptoSecureTokenGenerator(length)
			token, err := generator.GenerateToken()
			if err != nil {
				t.Errorf("CryptoSecureTokenGenerator.GenerateToken() with length %d failed: %v", length, err)
				return
			}

			if token == "" {
				t.Errorf("CryptoSecureTokenGenerator.GenerateToken() with length %d returned empty token", length)
			}
		})
	}
}

func TestCryptoSecureTokenGenerator_InvalidLength(t *testing.T) {
	// Test with invalid lengths - should still work (implementation handles it internally)
	invalidLengths := []int{-1, 0}

	for _, length := range invalidLengths {
		generator := domain.NewCryptoSecureTokenGenerator(length)
		// Test that it still works regardless of invalid length
		token, err := generator.GenerateToken()
		if err != nil {
			t.Errorf("NewCryptoSecureTokenGenerator(%d) should handle invalid length gracefully, got error: %v", length, err)
		}
		if token == "" {
			t.Errorf("NewCryptoSecureTokenGenerator(%d) should still generate tokens with invalid length", length)
		}
	}
}

func TestDefaultPasswordValidator_Validate(t *testing.T) {
	validator := domain.NewDefaultPasswordValidator()

	tests := []struct {
		name     string
		password string
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "valid strong password",
			password: "MyStr0ng!P4ss",
			wantErr:  false,
		},
		{
			name:     "too short",
			password: "Sh0rt!",
			wantErr:  true,
			errMsg:   "at least 8 characters long",
		},
		{
			name:     "no uppercase",
			password: "lowercase987!",
			wantErr:  true,
			errMsg:   "uppercase letter",
		},
		{
			name:     "no lowercase",
			password: "UPPERCASE987!",
			wantErr:  true,
			errMsg:   "lowercase letter",
		},
		{
			name:     "no numbers",
			password: "NoNumbers!",
			wantErr:  true,
			errMsg:   "number",
		},
		{
			name:     "no special characters",
			password: "NoSpecial987",
			wantErr:  true,
			errMsg:   "special character",
		},
		{
			name:     "common weak password",
			password: "Password987!",
			wantErr:  true,
			errMsg:   "weak pattern",
		},
		{
			name:     "sequential characters",
			password: "Abcd9876!",
			wantErr:  true,
			errMsg:   "sequential characters",
		},
		{
			name:     "repeated characters",
			password: "Aaaa9876!",
			wantErr:  true,
			errMsg:   "repeated characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(tt.password)

			if tt.wantErr {
				if err == nil {
					t.Errorf("DefaultPasswordValidator.Validate() expected error for %q, got nil", tt.password)
					return
				}
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("DefaultPasswordValidator.Validate() error = %q, want to contain %q", err.Error(), tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("DefaultPasswordValidator.Validate() unexpected error for %q: %v", tt.password, err)
				}
			}
		})
	}
}

func TestPasswordValidator_CustomRules(t *testing.T) {
	// Create validator with relaxed rules
	validator := domain.NewPasswordValidator(6, false, true, true, false)

	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{
			name:     "meets custom requirements",
			password: "lower987",
			wantErr:  false,
		},
		{
			name:     "too short for custom rules",
			password: "low98",
			wantErr:  true,
		},
		{
			name:     "no lowercase (required)",
			password: "UPPER987",
			wantErr:  true,
		},
		{
			name:     "no numbers (required)",
			password: "lowercase",
			wantErr:  true,
		},
		{
			name:     "no uppercase (not required)",
			password: "lower987",
			wantErr:  false,
		},
		{
			name:     "no special (not required)",
			password: "lower987",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(tt.password)

			if tt.wantErr && err == nil {
				t.Errorf("PasswordValidator.Validate() expected error for %q, got nil", tt.password)
			} else if !tt.wantErr && err != nil {
				t.Errorf("PasswordValidator.Validate() unexpected error for %q: %v", tt.password, err)
			}
		})
	}
}

func TestPasswordValidator_WeakPatterns(t *testing.T) {
	validator := domain.NewDefaultPasswordValidator()

	weakPasswords := []string{
		"Password123!",   // contains "password"
		"MyQwerty123!",   // contains "qwerty"
		"Admin123!",      // contains "admin"
		"Abc123456!",     // sequential abc and 123456
		"MyPassword111!", // repeated 1s
	}

	for _, password := range weakPasswords {
		t.Run(password, func(t *testing.T) {
			err := validator.Validate(password)
			if err == nil {
				t.Errorf("PasswordValidator.Validate() should reject weak password %q", password)
			}
		})
	}
}

func TestPasswordSecurity_Integration(t *testing.T) {
	// Test integration between hasher and validator
	hasher := domain.NewBCryptPasswordHasher(bcrypt.DefaultCost)
	validator := domain.NewDefaultPasswordValidator()
	generator := domain.NewCryptoSecureTokenGenerator(32)

	// Generate a strong password
	strongPassword := "MyStr0ng!P4ss"

	// Validate the password
	if err := validator.Validate(strongPassword); err != nil {
		t.Fatalf("Strong password failed validation: %v", err)
	}

	// Hash the password
	hash, salt, err := hasher.Hash(strongPassword)
	if err != nil {
		t.Fatalf("Failed to hash strong password: %v", err)
	}

	// Verify the password
	if !hasher.Verify(strongPassword, hash, salt) {
		t.Error("Failed to verify hashed strong password")
	}

	// Generate a secure token
	token, err := generator.GenerateToken()
	if err != nil {
		t.Fatalf("Failed to generate secure token: %v", err)
	}

	if len(token) == 0 {
		t.Error("Generated token is empty")
	}
}

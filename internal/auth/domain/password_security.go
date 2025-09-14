package domain

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"regexp"
	"strings"
	"unicode"

	"golang.org/x/crypto/bcrypt"
)

// PasswordHasher defines the interface for password hashing operations
type PasswordHasher interface {
	Hash(password string) (hash, salt string, err error)
	Verify(password, hash, salt string) bool
}

// SecureTokenGenerator defines the interface for generating secure tokens
type SecureTokenGenerator interface {
	GenerateToken() (string, error)
}

// PasswordValidator defines the interface for password strength validation
type PasswordValidator interface {
	Validate(password string) error
}

// BCryptPasswordHasher implements PasswordHasher using bcrypt
type BCryptPasswordHasher struct {
	cost int
}

// NewBCryptPasswordHasher creates a new BCrypt password hasher with the specified cost
func NewBCryptPasswordHasher(cost int) *BCryptPasswordHasher {
	if cost < bcrypt.MinCost || cost > bcrypt.MaxCost {
		cost = bcrypt.DefaultCost
	}
	return &BCryptPasswordHasher{cost: cost}
}

// Hash generates a bcrypt hash and salt for the given password
func (h *BCryptPasswordHasher) Hash(password string) (hash, salt string, err error) {
	if password == "" {
		return "", "", fmt.Errorf("password cannot be empty")
	}

	// Generate salt
	saltBytes := make([]byte, 16)
	if _, err := rand.Read(saltBytes); err != nil {
		return "", "", fmt.Errorf("failed to generate salt: %w", err)
	}
	salt = base64.StdEncoding.EncodeToString(saltBytes)

	// Hash password with salt
	saltedPassword := password + salt

	// BCrypt has a 72-byte limit, so truncate if necessary
	if len(saltedPassword) > 72 {
		saltedPassword = saltedPassword[:72]
	}

	hashBytes, err := bcrypt.GenerateFromPassword([]byte(saltedPassword), h.cost)
	if err != nil {
		return "", "", fmt.Errorf("failed to hash password: %w", err)
	}

	hash = string(hashBytes)
	return hash, salt, nil
}

// Verify checks if the provided password matches the stored hash and salt
func (h *BCryptPasswordHasher) Verify(password, hash, salt string) bool {
	if password == "" || hash == "" || salt == "" {
		return false
	}

	saltedPassword := password + salt

	// BCrypt has a 72-byte limit, so truncate if necessary (same as during hashing)
	if len(saltedPassword) > 72 {
		saltedPassword = saltedPassword[:72]
	}

	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(saltedPassword))
	return err == nil
}

// CryptoSecureTokenGenerator implements SecureTokenGenerator using crypto/rand
type CryptoSecureTokenGenerator struct {
	tokenLength int
}

// NewCryptoSecureTokenGenerator creates a new secure token generator
func NewCryptoSecureTokenGenerator(tokenLength int) *CryptoSecureTokenGenerator {
	if tokenLength <= 0 {
		tokenLength = 32 // Default to 32 bytes (256 bits)
	}
	return &CryptoSecureTokenGenerator{tokenLength: tokenLength}
}

// GenerateToken generates a cryptographically secure random token
func (g *CryptoSecureTokenGenerator) GenerateToken() (string, error) {
	bytes := make([]byte, g.tokenLength)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate secure token: %w", err)
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// DefaultPasswordValidator implements PasswordValidator with common security rules
type DefaultPasswordValidator struct {
	minLength        int
	requireUppercase bool
	requireLowercase bool
	requireNumbers   bool
	requireSpecial   bool
}

// NewDefaultPasswordValidator creates a new password validator with default rules
func NewDefaultPasswordValidator() *DefaultPasswordValidator {
	return &DefaultPasswordValidator{
		minLength:        8,
		requireUppercase: true,
		requireLowercase: true,
		requireNumbers:   true,
		requireSpecial:   true,
	}
}

// NewPasswordValidator creates a new password validator with custom rules
func NewPasswordValidator(minLength int, requireUppercase, requireLowercase, requireNumbers, requireSpecial bool) *DefaultPasswordValidator {
	return &DefaultPasswordValidator{
		minLength:        minLength,
		requireUppercase: requireUppercase,
		requireLowercase: requireLowercase,
		requireNumbers:   requireNumbers,
		requireSpecial:   requireSpecial,
	}
}

// Validate checks if the password meets the security requirements
func (v *DefaultPasswordValidator) Validate(password string) error {
	if len(password) < v.minLength {
		return fmt.Errorf("password must be at least %d characters long", v.minLength)
	}

	var hasUpper, hasLower, hasNumber, hasSpecial bool

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsNumber(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	var errors []string

	if v.requireUppercase && !hasUpper {
		errors = append(errors, "at least one uppercase letter")
	}
	if v.requireLowercase && !hasLower {
		errors = append(errors, "at least one lowercase letter")
	}
	if v.requireNumbers && !hasNumber {
		errors = append(errors, "at least one number")
	}
	if v.requireSpecial && !hasSpecial {
		errors = append(errors, "at least one special character")
	}

	if len(errors) > 0 {
		return fmt.Errorf("password must contain %s", strings.Join(errors, ", "))
	}

	// Check for common weak patterns
	if err := v.checkWeakPatterns(password); err != nil {
		return err
	}

	return nil
}

// checkWeakPatterns checks for common weak password patterns
func (v *DefaultPasswordValidator) checkWeakPatterns(password string) error {
	lower := strings.ToLower(password)

	// Check for common weak passwords
	weakPasswords := []string{
		"password", "123456", "qwerty", "abc123", "letmein",
		"welcome", "monkey", "dragon", "master", "admin",
	}

	for _, weak := range weakPasswords {
		if strings.Contains(lower, weak) {
			return fmt.Errorf("password contains common weak pattern")
		}
	}

	// Check for sequential characters (4+ consecutive)
	sequential := regexp.MustCompile(`(0123|1234|2345|3456|4567|5678|6789|abcd|bcde|cdef|defg|efgh|fghi|ghij|hijk|ijkl|jklm|klmn|lmno|mnop|nopq|opqr|pqrs|qrst|rstu|stuv|tuvw|uvwx|vwxy|wxyz)`)
	if sequential.MatchString(lower) {
		return fmt.Errorf("password contains sequential characters")
	}

	// Check for repeated characters (3+ of the same character)
	for i := 0; i < len(password)-2; i++ {
		if password[i] == password[i+1] && password[i+1] == password[i+2] {
			return fmt.Errorf("password contains too many repeated characters")
		}
	}

	return nil
}

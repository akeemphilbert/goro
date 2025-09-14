package infrastructure

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/akeemphilbert/goro/internal/user/domain"
)

var (
	// ErrUserNotFound is returned when a user's data is not found
	ErrUserNotFound = fmt.Errorf("user not found")

	// ErrInvalidUserID is returned when a user ID is invalid
	ErrInvalidUserID = fmt.Errorf("invalid user ID")
)

// UserFileStorage defines the interface for user file storage operations
type UserFileStorage interface {
	// Profile operations
	StoreUserProfile(ctx context.Context, userID string, profile domain.UserProfile) error
	LoadUserProfile(ctx context.Context, userID string) (domain.UserProfile, error)
	UserProfileExists(ctx context.Context, userID string) (bool, error)

	// WebID document operations
	StoreWebIDDocument(ctx context.Context, userID, webID, document string) error
	LoadWebIDDocument(ctx context.Context, userID string) (string, error)
	WebIDDocumentExists(ctx context.Context, userID string) (bool, error)

	// User data management
	DeleteUserData(ctx context.Context, userID string) error
	GetUserDirectory(userID string) string
}

// userFileStorage implements the UserFileStorage interface
type userFileStorage struct {
	baseDir string
}

// NewUserFileStorage creates a new user file storage with the given base directory
func NewUserFileStorage(baseDir string) UserFileStorage {
	return &userFileStorage{
		baseDir: baseDir,
	}
}

// NewUserFileStorageWithValidation creates a new user file storage with validation
func NewUserFileStorageWithValidation(baseDir string) (UserFileStorage, error) {
	if strings.TrimSpace(baseDir) == "" {
		return nil, fmt.Errorf("base directory is required")
	}

	return &userFileStorage{
		baseDir: baseDir,
	}, nil
}

// StoreUserProfile stores a user profile to the file system
func (s *userFileStorage) StoreUserProfile(ctx context.Context, userID string, profile domain.UserProfile) error {
	if strings.TrimSpace(userID) == "" {
		return ErrInvalidUserID
	}

	// Validate profile
	if err := profile.Validate(); err != nil {
		return fmt.Errorf("invalid profile: %w", err)
	}

	// Ensure user directory exists
	userDir := s.GetUserDirectory(userID)
	if err := os.MkdirAll(userDir, 0755); err != nil {
		return fmt.Errorf("failed to create user directory: %w", err)
	}

	// Marshal profile to JSON
	profileData, err := json.MarshalIndent(profile, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal profile: %w", err)
	}

	// Write profile to file atomically
	profilePath := filepath.Join(userDir, "profile.json")
	tempPath := profilePath + ".tmp"

	if err := os.WriteFile(tempPath, profileData, 0644); err != nil {
		return fmt.Errorf("failed to write profile file: %w", err)
	}

	if err := os.Rename(tempPath, profilePath); err != nil {
		os.Remove(tempPath) // Clean up temp file on failure
		return fmt.Errorf("failed to commit profile file: %w", err)
	}

	return nil
}

// LoadUserProfile loads a user profile from the file system
func (s *userFileStorage) LoadUserProfile(ctx context.Context, userID string) (domain.UserProfile, error) {
	var profile domain.UserProfile

	if strings.TrimSpace(userID) == "" {
		return profile, ErrInvalidUserID
	}

	profilePath := filepath.Join(s.GetUserDirectory(userID), "profile.json")

	// Check if file exists
	if _, err := os.Stat(profilePath); os.IsNotExist(err) {
		return profile, fmt.Errorf("%w: profile not found for user %s", ErrUserNotFound, userID)
	}

	// Read profile file
	profileData, err := os.ReadFile(profilePath)
	if err != nil {
		return profile, fmt.Errorf("failed to read profile file: %w", err)
	}

	// Unmarshal JSON
	if err := json.Unmarshal(profileData, &profile); err != nil {
		return profile, fmt.Errorf("failed to parse profile: %w", err)
	}

	return profile, nil
}

// UserProfileExists checks if a user profile exists
func (s *userFileStorage) UserProfileExists(ctx context.Context, userID string) (bool, error) {
	if strings.TrimSpace(userID) == "" {
		return false, ErrInvalidUserID
	}

	profilePath := filepath.Join(s.GetUserDirectory(userID), "profile.json")
	_, err := os.Stat(profilePath)

	if os.IsNotExist(err) {
		return false, nil
	}

	if err != nil {
		return false, fmt.Errorf("failed to check profile existence: %w", err)
	}

	return true, nil
}

// StoreWebIDDocument stores a WebID document to the file system
func (s *userFileStorage) StoreWebIDDocument(ctx context.Context, userID, webID, document string) error {
	if strings.TrimSpace(userID) == "" {
		return ErrInvalidUserID
	}

	if strings.TrimSpace(webID) == "" {
		return fmt.Errorf("WebID is required")
	}

	if strings.TrimSpace(document) == "" {
		return fmt.Errorf("document content is required")
	}

	// Ensure user directory exists
	userDir := s.GetUserDirectory(userID)
	if err := os.MkdirAll(userDir, 0755); err != nil {
		return fmt.Errorf("failed to create user directory: %w", err)
	}

	// Write WebID document to file atomically
	webidPath := filepath.Join(userDir, "webid.ttl")
	tempPath := webidPath + ".tmp"

	if err := os.WriteFile(tempPath, []byte(document), 0644); err != nil {
		return fmt.Errorf("failed to write WebID document: %w", err)
	}

	if err := os.Rename(tempPath, webidPath); err != nil {
		os.Remove(tempPath) // Clean up temp file on failure
		return fmt.Errorf("failed to commit WebID document: %w", err)
	}

	return nil
}

// LoadWebIDDocument loads a WebID document from the file system
func (s *userFileStorage) LoadWebIDDocument(ctx context.Context, userID string) (string, error) {
	if strings.TrimSpace(userID) == "" {
		return "", ErrInvalidUserID
	}

	webidPath := filepath.Join(s.GetUserDirectory(userID), "webid.ttl")

	// Check if file exists
	if _, err := os.Stat(webidPath); os.IsNotExist(err) {
		return "", fmt.Errorf("%w: WebID document not found for user %s", ErrUserNotFound, userID)
	}

	// Read WebID document
	documentData, err := os.ReadFile(webidPath)
	if err != nil {
		return "", fmt.Errorf("failed to read WebID document: %w", err)
	}

	return string(documentData), nil
}

// WebIDDocumentExists checks if a WebID document exists
func (s *userFileStorage) WebIDDocumentExists(ctx context.Context, userID string) (bool, error) {
	if strings.TrimSpace(userID) == "" {
		return false, ErrInvalidUserID
	}

	webidPath := filepath.Join(s.GetUserDirectory(userID), "webid.ttl")
	_, err := os.Stat(webidPath)

	if os.IsNotExist(err) {
		return false, nil
	}

	if err != nil {
		return false, fmt.Errorf("failed to check WebID document existence: %w", err)
	}

	return true, nil
}

// DeleteUserData deletes all user data from the file system
func (s *userFileStorage) DeleteUserData(ctx context.Context, userID string) error {
	if strings.TrimSpace(userID) == "" {
		return ErrInvalidUserID
	}

	userDir := s.GetUserDirectory(userID)

	// Check if directory exists
	if _, err := os.Stat(userDir); os.IsNotExist(err) {
		// Directory doesn't exist, operation is idempotent
		return nil
	}

	// Remove entire user directory
	if err := os.RemoveAll(userDir); err != nil {
		return fmt.Errorf("failed to delete user data: %w", err)
	}

	return nil
}

// GetUserDirectory returns the directory path for a user
func (s *userFileStorage) GetUserDirectory(userID string) string {
	return filepath.Join(s.baseDir, "users", userID)
}

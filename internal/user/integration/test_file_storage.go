package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/akeemphilbert/goro/internal/user/domain"
)

// testFileStorage implements application.FileStorage for testing
type testFileStorage struct {
	baseDir string
}

// NewTestFileStorage creates a new test file storage
func NewTestFileStorage(baseDir string) *testFileStorage {
	return &testFileStorage{
		baseDir: baseDir,
	}
}

// WriteUserProfile writes a user profile to file storage
func (fs *testFileStorage) WriteUserProfile(ctx context.Context, userID string, profile domain.UserProfile) error {
	userDir := filepath.Join(fs.baseDir, "users", userID)
	if err := os.MkdirAll(userDir, 0755); err != nil {
		return fmt.Errorf("failed to create user directory: %w", err)
	}

	profileFile := filepath.Join(userDir, "profile.json")
	data, err := json.Marshal(profile)
	if err != nil {
		return fmt.Errorf("failed to marshal profile: %w", err)
	}

	return os.WriteFile(profileFile, data, 0644)
}

// WriteWebIDDocument writes a WebID document to file storage
func (fs *testFileStorage) WriteWebIDDocument(ctx context.Context, userID, webID, document string) error {
	userDir := filepath.Join(fs.baseDir, "users", userID)
	if err := os.MkdirAll(userDir, 0755); err != nil {
		return fmt.Errorf("failed to create user directory: %w", err)
	}

	webidFile := filepath.Join(userDir, "webid.ttl")
	return os.WriteFile(webidFile, []byte(document), 0644)
}

// DeleteUserFiles deletes all user files
func (fs *testFileStorage) DeleteUserFiles(ctx context.Context, userID string) error {
	userDir := filepath.Join(fs.baseDir, "users", userID)
	return os.RemoveAll(userDir)
}

// ReadUserProfile reads a user profile from file storage
func (fs *testFileStorage) ReadUserProfile(ctx context.Context, userID string) (domain.UserProfile, error) {
	profileFile := filepath.Join(fs.baseDir, "users", userID, "profile.json")
	data, err := os.ReadFile(profileFile)
	if err != nil {
		return domain.UserProfile{}, fmt.Errorf("failed to read profile file: %w", err)
	}

	var profile domain.UserProfile
	if err := json.Unmarshal(data, &profile); err != nil {
		return domain.UserProfile{}, fmt.Errorf("failed to unmarshal profile: %w", err)
	}

	return profile, nil
}

// ReadWebIDDocument reads a WebID document from file storage
func (fs *testFileStorage) ReadWebIDDocument(ctx context.Context, userID string) (string, error) {
	webidFile := filepath.Join(fs.baseDir, "users", userID, "webid.ttl")
	data, err := os.ReadFile(webidFile)
	if err != nil {
		return "", fmt.Errorf("failed to read WebID file: %w", err)
	}

	return string(data), nil
}

// UserExists checks if a user exists in file storage
func (fs *testFileStorage) UserExists(ctx context.Context, userID string) (bool, error) {
	userDir := filepath.Join(fs.baseDir, "users", userID)
	_, err := os.Stat(userDir)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

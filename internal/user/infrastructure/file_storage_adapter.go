package infrastructure

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/akeemphilbert/goro/internal/user/domain"
)

// fileStorageAdapter adapts UserFileStorage to implement domain.FileStorage
type fileStorageAdapter struct {
	userStorage UserFileStorage
}

// NewFileStorageAdapter creates a new file storage adapter
func NewFileStorageAdapter(baseDir string) (domain.FileStorage, error) {
	userStorage := NewUserFileStorage(baseDir)
	return &fileStorageAdapter{
		userStorage: userStorage,
	}, nil
}

// StoreUserProfile stores a user profile
func (f *fileStorageAdapter) StoreUserProfile(ctx context.Context, userID string, profile *domain.User) error {
	if profile == nil {
		return fmt.Errorf("profile cannot be nil")
	}
	return f.userStorage.StoreUserProfile(ctx, userID, profile.Profile)
}

// LoadUserProfile loads a user profile
func (f *fileStorageAdapter) LoadUserProfile(ctx context.Context, userID string) (*domain.User, error) {
	profile, err := f.userStorage.LoadUserProfile(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Create a minimal user with the profile
	// Note: This is a simplified implementation for the adapter
	// In a real scenario, we would need to load the full user data
	user, err := domain.NewUser(ctx, userID, "", "", profile)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}
	return user, nil
}

// DeleteUserProfile deletes a user profile
func (f *fileStorageAdapter) DeleteUserProfile(ctx context.Context, userID string) error {
	return f.userStorage.DeleteUserData(ctx, userID)
}

// StoreWebIDDocument stores a WebID document
func (f *fileStorageAdapter) StoreWebIDDocument(ctx context.Context, userID, webIDDoc string) error {
	// We need to extract the WebID from the document or pass it separately
	// For now, we'll use a placeholder WebID
	webID := fmt.Sprintf("https://example.com/users/%s#me", userID)
	return f.userStorage.StoreWebIDDocument(ctx, userID, webID, webIDDoc)
}

// LoadWebIDDocument loads a WebID document
func (f *fileStorageAdapter) LoadWebIDDocument(ctx context.Context, userID string) (string, error) {
	return f.userStorage.LoadWebIDDocument(ctx, userID)
}

// DeleteWebIDDocument deletes a WebID document
func (f *fileStorageAdapter) DeleteWebIDDocument(ctx context.Context, userID string) error {
	return f.userStorage.DeleteUserData(ctx, userID)
}

// StoreAccountData stores account data
func (f *fileStorageAdapter) StoreAccountData(ctx context.Context, accountID string, account *domain.Account) error {
	if account == nil {
		return fmt.Errorf("account cannot be nil")
	}

	// Store account data as JSON in the accounts directory
	accountDir := filepath.Join(f.userStorage.GetUserDirectory(""), "accounts", accountID)
	accountFile := filepath.Join(accountDir, "account.json")

	data, err := json.Marshal(account)
	if err != nil {
		return fmt.Errorf("failed to marshal account: %w", err)
	}

	return writeFileAtomic(accountFile, data)
}

// LoadAccountData loads account data
func (f *fileStorageAdapter) LoadAccountData(ctx context.Context, accountID string) (*domain.Account, error) {
	// Load account data from JSON in the accounts directory
	accountDir := filepath.Join(f.userStorage.GetUserDirectory(""), "accounts", accountID)
	accountFile := filepath.Join(accountDir, "account.json")

	data, err := readFile(accountFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read account file: %w", err)
	}

	var account domain.Account
	if err := json.Unmarshal(data, &account); err != nil {
		return nil, fmt.Errorf("failed to unmarshal account: %w", err)
	}

	return &account, nil
}

// DeleteAccountData deletes account data
func (f *fileStorageAdapter) DeleteAccountData(ctx context.Context, accountID string) error {
	// Delete account directory
	accountDir := filepath.Join(f.userStorage.GetUserDirectory(""), "accounts", accountID)
	return removeAll(accountDir)
}

// Helper functions

// writeFileAtomic writes data to a file atomically
func writeFileAtomic(filename string, data []byte) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Write to temporary file first
	tempFile := filename + ".tmp"
	if err := ioutil.WriteFile(tempFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write temp file: %w", err)
	}

	// Atomically rename temp file to final file
	if err := os.Rename(tempFile, filename); err != nil {
		os.Remove(tempFile) // Clean up temp file on error
		return fmt.Errorf("failed to rename temp file: %w", err)
	}

	return nil
}

// readFile reads data from a file
func readFile(filename string) ([]byte, error) {
	return ioutil.ReadFile(filename)
}

// removeAll removes a directory and all its contents
func removeAll(path string) error {
	return os.RemoveAll(path)
}

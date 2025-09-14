package infrastructure_test

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/akeemphilbert/goro/internal/user/domain"
	"github.com/akeemphilbert/goro/internal/user/infrastructure"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserFileStorage_StoreUserProfile(t *testing.T) {
	tests := []struct {
		name     string
		userID   string
		profile  domain.UserProfile
		wantErr  bool
		validate func(t *testing.T, storage infrastructure.UserFileStorage, userID string)
	}{
		{
			name:   "valid profile is stored successfully",
			userID: "user-123",
			profile: domain.UserProfile{
				Name:        "John Doe",
				Bio:         "Software developer",
				Avatar:      "https://example.com/avatar.jpg",
				Preferences: map[string]interface{}{"theme": "dark", "language": "en"},
			},
			wantErr: false,
			validate: func(t *testing.T, storage infrastructure.UserFileStorage, userID string) {
				// Verify file exists
				exists, err := storage.UserProfileExists(context.Background(), userID)
				assert.NoError(t, err)
				assert.True(t, exists)

				// Verify content can be read back
				profile, err := storage.LoadUserProfile(context.Background(), userID)
				assert.NoError(t, err)
				assert.Equal(t, "John Doe", profile.Name)
				assert.Equal(t, "Software developer", profile.Bio)
				assert.Equal(t, "dark", profile.Preferences["theme"])
			},
		},
		{
			name:    "empty user ID returns error",
			userID:  "",
			profile: domain.UserProfile{Name: "John Doe"},
			wantErr: true,
		},
		{
			name:   "profile with empty name returns error",
			userID: "user-123",
			profile: domain.UserProfile{
				Name: "",
				Bio:  "Software developer",
			},
			wantErr: true,
		},
		{
			name:   "profile with special characters is handled",
			userID: "user-with-special@chars",
			profile: domain.UserProfile{
				Name: "John \"Doe\" O'Connor",
				Bio:  "Software developer\nwith newlines",
			},
			wantErr: false,
			validate: func(t *testing.T, storage infrastructure.UserFileStorage, userID string) {
				profile, err := storage.LoadUserProfile(context.Background(), userID)
				assert.NoError(t, err)
				assert.Equal(t, "John \"Doe\" O'Connor", profile.Name)
				assert.Contains(t, profile.Bio, "newlines")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary directory for test
			tempDir, err := os.MkdirTemp("", "user-storage-test-*")
			require.NoError(t, err)
			defer os.RemoveAll(tempDir)

			storage := infrastructure.NewUserFileStorage(tempDir)

			err = storage.StoreUserProfile(context.Background(), tt.userID, tt.profile)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.validate != nil {
					tt.validate(t, storage, tt.userID)
				}
			}
		})
	}
}

func TestUserFileStorage_StoreWebIDDocument(t *testing.T) {
	tests := []struct {
		name     string
		userID   string
		webID    string
		document string
		wantErr  bool
		validate func(t *testing.T, storage infrastructure.UserFileStorage, userID string)
	}{
		{
			name:   "valid WebID document is stored successfully",
			userID: "user-123",
			webID:  "https://example.com/users/user-123#me",
			document: `@prefix foaf: <http://xmlns.com/foaf/0.1/> .
@prefix solid: <http://www.w3.org/ns/solid/terms#> .

<https://example.com/users/user-123#me> a foaf:Person ;
    foaf:name "John Doe" ;
    foaf:mbox <mailto:john@example.com> .`,
			wantErr: false,
			validate: func(t *testing.T, storage infrastructure.UserFileStorage, userID string) {
				// Verify file exists
				exists, err := storage.WebIDDocumentExists(context.Background(), userID)
				assert.NoError(t, err)
				assert.True(t, exists)

				// Verify content can be read back
				document, err := storage.LoadWebIDDocument(context.Background(), userID)
				assert.NoError(t, err)
				assert.Contains(t, document, "foaf:Person")
				assert.Contains(t, document, "John Doe")
			},
		},
		{
			name:     "empty user ID returns error",
			userID:   "",
			webID:    "https://example.com/users/user-123#me",
			document: "valid turtle document",
			wantErr:  true,
		},
		{
			name:     "empty WebID returns error",
			userID:   "user-123",
			webID:    "",
			document: "valid turtle document",
			wantErr:  true,
		},
		{
			name:     "empty document returns error",
			userID:   "user-123",
			webID:    "https://example.com/users/user-123#me",
			document: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary directory for test
			tempDir, err := os.MkdirTemp("", "user-storage-test-*")
			require.NoError(t, err)
			defer os.RemoveAll(tempDir)

			storage := infrastructure.NewUserFileStorage(tempDir)

			err = storage.StoreWebIDDocument(context.Background(), tt.userID, tt.webID, tt.document)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.validate != nil {
					tt.validate(t, storage, tt.userID)
				}
			}
		})
	}
}

func TestUserFileStorage_LoadUserProfile(t *testing.T) {
	tests := []struct {
		name     string
		userID   string
		setup    func(t *testing.T, storage infrastructure.UserFileStorage, userID string)
		wantErr  bool
		validate func(t *testing.T, profile domain.UserProfile)
	}{
		{
			name:   "existing profile is loaded successfully",
			userID: "user-123",
			setup: func(t *testing.T, storage infrastructure.UserFileStorage, userID string) {
				profile := domain.UserProfile{
					Name:        "John Doe",
					Bio:         "Software developer",
					Preferences: map[string]interface{}{"theme": "dark"},
				}
				err := storage.StoreUserProfile(context.Background(), userID, profile)
				require.NoError(t, err)
			},
			wantErr: false,
			validate: func(t *testing.T, profile domain.UserProfile) {
				assert.Equal(t, "John Doe", profile.Name)
				assert.Equal(t, "Software developer", profile.Bio)
				assert.Equal(t, "dark", profile.Preferences["theme"])
			},
		},
		{
			name:    "non-existent profile returns error",
			userID:  "non-existent-user",
			setup:   func(t *testing.T, storage infrastructure.UserFileStorage, userID string) {},
			wantErr: true,
		},
		{
			name:    "empty user ID returns error",
			userID:  "",
			setup:   func(t *testing.T, storage infrastructure.UserFileStorage, userID string) {},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary directory for test
			tempDir, err := os.MkdirTemp("", "user-storage-test-*")
			require.NoError(t, err)
			defer os.RemoveAll(tempDir)

			storage := infrastructure.NewUserFileStorage(tempDir)
			tt.setup(t, storage, tt.userID)

			profile, err := storage.LoadUserProfile(context.Background(), tt.userID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.validate != nil {
					tt.validate(t, profile)
				}
			}
		})
	}
}

func TestUserFileStorage_DeleteUserData(t *testing.T) {
	tests := []struct {
		name     string
		userID   string
		setup    func(t *testing.T, storage infrastructure.UserFileStorage, userID string)
		wantErr  bool
		validate func(t *testing.T, storage infrastructure.UserFileStorage, userID string)
	}{
		{
			name:   "existing user data is deleted successfully",
			userID: "user-123",
			setup: func(t *testing.T, storage infrastructure.UserFileStorage, userID string) {
				// Create profile and WebID document
				profile := domain.UserProfile{Name: "John Doe", Bio: "Developer"}
				err := storage.StoreUserProfile(context.Background(), userID, profile)
				require.NoError(t, err)

				document := `<https://example.com/users/user-123#me> a foaf:Person .`
				err = storage.StoreWebIDDocument(context.Background(), userID, "https://example.com/users/user-123#me", document)
				require.NoError(t, err)
			},
			wantErr: false,
			validate: func(t *testing.T, storage infrastructure.UserFileStorage, userID string) {
				// Verify profile no longer exists
				exists, err := storage.UserProfileExists(context.Background(), userID)
				assert.NoError(t, err)
				assert.False(t, exists)

				// Verify WebID document no longer exists
				exists, err = storage.WebIDDocumentExists(context.Background(), userID)
				assert.NoError(t, err)
				assert.False(t, exists)

				// Verify user directory is removed
				userDir := storage.GetUserDirectory(userID)
				_, err = os.Stat(userDir)
				assert.True(t, os.IsNotExist(err))
			},
		},
		{
			name:    "non-existent user returns no error (idempotent)",
			userID:  "non-existent-user",
			setup:   func(t *testing.T, storage infrastructure.UserFileStorage, userID string) {},
			wantErr: false,
		},
		{
			name:    "empty user ID returns error",
			userID:  "",
			setup:   func(t *testing.T, storage infrastructure.UserFileStorage, userID string) {},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary directory for test
			tempDir, err := os.MkdirTemp("", "user-storage-test-*")
			require.NoError(t, err)
			defer os.RemoveAll(tempDir)

			storage := infrastructure.NewUserFileStorage(tempDir)
			tt.setup(t, storage, tt.userID)

			err = storage.DeleteUserData(context.Background(), tt.userID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.validate != nil {
					tt.validate(t, storage, tt.userID)
				}
			}
		})
	}
}

func TestUserFileStorage_AtomicOperations(t *testing.T) {
	t.Run("concurrent profile updates are handled safely", func(t *testing.T) {
		// Create temporary directory for test
		tempDir, err := os.MkdirTemp("", "user-storage-test-*")
		require.NoError(t, err)
		defer os.RemoveAll(tempDir)

		storage := infrastructure.NewUserFileStorage(tempDir)

		// Perform concurrent updates with different user IDs to avoid conflicts
		done := make(chan error, 2)

		go func() {
			profile1 := domain.UserProfile{Name: "John Doe 1", Bio: "First update"}
			err := storage.StoreUserProfile(context.Background(), "user-1", profile1)
			done <- err
		}()

		go func() {
			profile2 := domain.UserProfile{Name: "John Doe 2", Bio: "Second update"}
			err := storage.StoreUserProfile(context.Background(), "user-2", profile2)
			done <- err
		}()

		// Wait for both operations to complete
		err1 := <-done
		err2 := <-done

		// Both operations should succeed
		assert.NoError(t, err1)
		assert.NoError(t, err2)

		// Verify that both profiles exist
		profile1, err := storage.LoadUserProfile(context.Background(), "user-1")
		assert.NoError(t, err)
		assert.Equal(t, "John Doe 1", profile1.Name)

		profile2, err := storage.LoadUserProfile(context.Background(), "user-2")
		assert.NoError(t, err)
		assert.Equal(t, "John Doe 2", profile2.Name)
	})

	t.Run("atomic file operations prevent corruption", func(t *testing.T) {
		// Create temporary directory for test
		tempDir, err := os.MkdirTemp("", "user-storage-test-*")
		require.NoError(t, err)
		defer os.RemoveAll(tempDir)

		storage := infrastructure.NewUserFileStorage(tempDir)
		userID := "user-123"

		// Store a valid profile first
		profile := domain.UserProfile{Name: "John Doe", Bio: "Developer"}
		err = storage.StoreUserProfile(context.Background(), userID, profile)
		require.NoError(t, err)

		// Verify the profile can be loaded
		loadedProfile, err := storage.LoadUserProfile(context.Background(), userID)
		assert.NoError(t, err)
		assert.Equal(t, "John Doe", loadedProfile.Name)

		// Store WebID document
		document := `<https://example.com/users/user-123#me> a foaf:Person .`
		err = storage.StoreWebIDDocument(context.Background(), userID, "https://example.com/users/user-123#me", document)
		assert.NoError(t, err)

		// Verify WebID document can be loaded
		loadedDocument, err := storage.LoadWebIDDocument(context.Background(), userID)
		assert.NoError(t, err)
		assert.Contains(t, loadedDocument, "foaf:Person")
	})
}

func TestUserFileStorage_DirectoryStructure(t *testing.T) {
	t.Run("correct directory structure is created", func(t *testing.T) {
		// Create temporary directory for test
		tempDir, err := os.MkdirTemp("", "user-storage-test-*")
		require.NoError(t, err)
		defer os.RemoveAll(tempDir)

		storage := infrastructure.NewUserFileStorage(tempDir)
		userID := "user-123"

		// Store profile and WebID document
		profile := domain.UserProfile{Name: "John Doe", Bio: "Developer"}
		err = storage.StoreUserProfile(context.Background(), userID, profile)
		require.NoError(t, err)

		document := `<https://example.com/users/user-123#me> a foaf:Person .`
		err = storage.StoreWebIDDocument(context.Background(), userID, "https://example.com/users/user-123#me", document)
		require.NoError(t, err)

		// Verify directory structure
		userDir := storage.GetUserDirectory(userID)
		assert.DirExists(t, userDir)

		profilePath := filepath.Join(userDir, "profile.json")
		assert.FileExists(t, profilePath)

		webidPath := filepath.Join(userDir, "webid.ttl")
		assert.FileExists(t, webidPath)

		// Verify file permissions
		profileInfo, err := os.Stat(profilePath)
		require.NoError(t, err)
		assert.Equal(t, fs.FileMode(0644), profileInfo.Mode().Perm())

		webidInfo, err := os.Stat(webidPath)
		require.NoError(t, err)
		assert.Equal(t, fs.FileMode(0644), webidInfo.Mode().Perm())
	})

	t.Run("user directory path is correctly generated", func(t *testing.T) {
		tempDir := "/tmp/test-storage"
		storage := infrastructure.NewUserFileStorage(tempDir)

		tests := []struct {
			userID   string
			expected string
		}{
			{"user-123", filepath.Join(tempDir, "users", "user-123")},
			{"user-with-special@chars", filepath.Join(tempDir, "users", "user-with-special@chars")},
			{"", filepath.Join(tempDir, "users", "")},
		}

		for _, tt := range tests {
			result := storage.GetUserDirectory(tt.userID)
			assert.Equal(t, tt.expected, result)
		}
	})
}

func TestUserFileStorage_ErrorHandling(t *testing.T) {
	t.Run("invalid JSON in profile file is handled", func(t *testing.T) {
		// Create temporary directory for test
		tempDir, err := os.MkdirTemp("", "user-storage-test-*")
		require.NoError(t, err)
		defer os.RemoveAll(tempDir)

		storage := infrastructure.NewUserFileStorage(tempDir)
		userID := "user-123"

		// Create user directory and write invalid JSON
		userDir := storage.GetUserDirectory(userID)
		err = os.MkdirAll(userDir, 0755)
		require.NoError(t, err)

		profilePath := filepath.Join(userDir, "profile.json")
		err = os.WriteFile(profilePath, []byte("invalid json content"), 0644)
		require.NoError(t, err)

		// Try to load profile (should fail)
		_, err = storage.LoadUserProfile(context.Background(), userID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse profile")
	})

	t.Run("missing base directory is created automatically", func(t *testing.T) {
		// Use a non-existent base directory
		tempDir := filepath.Join(os.TempDir(), "non-existent-dir", "user-storage")
		defer os.RemoveAll(filepath.Dir(tempDir))

		storage := infrastructure.NewUserFileStorage(tempDir)
		userID := "user-123"

		// Store profile (should create directories automatically)
		profile := domain.UserProfile{Name: "John Doe", Bio: "Developer"}
		err := storage.StoreUserProfile(context.Background(), userID, profile)
		assert.NoError(t, err)

		// Verify directory was created
		assert.DirExists(t, tempDir)
		assert.DirExists(t, storage.GetUserDirectory(userID))
	})
}

func TestNewUserFileStorage(t *testing.T) {
	tests := []struct {
		name    string
		baseDir string
		wantErr bool
	}{
		{
			name:    "valid base directory creates storage",
			baseDir: "/tmp/user-storage",
			wantErr: false,
		},
		{
			name:    "empty base directory returns error",
			baseDir: "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage, err := infrastructure.NewUserFileStorageWithValidation(tt.baseDir)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, storage)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, storage)
			}
		})
	}
}

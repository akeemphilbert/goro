package infrastructure

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/akeemphilbert/goro/internal/user/domain"
)

// setupTestDB creates a temporary in-memory SQLite database for testing
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	return db
}

func TestUserModel_Validation(t *testing.T) {
	db := setupTestDB(t)

	// Test that UserModel can be created and migrated
	err := db.AutoMigrate(&UserModel{})
	require.NoError(t, err)

	// Test valid user model
	user := &UserModel{
		ID:        "user-123",
		WebID:     "https://example.com/user/123",
		Email:     "user@example.com",
		Name:      "John Doe",
		Status:    string(domain.UserStatusActive),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = db.Create(user).Error
	require.NoError(t, err)

	// Test unique constraints
	duplicateUser := &UserModel{
		ID:        "user-456",
		WebID:     "https://example.com/user/123", // Same WebID
		Email:     "different@example.com",
		Name:      "Jane Doe",
		Status:    string(domain.UserStatusActive),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = db.Create(duplicateUser).Error
	assert.Error(t, err, "Should fail due to unique WebID constraint")

	// Test unique email constraint
	duplicateEmailUser := &UserModel{
		ID:        "user-789",
		WebID:     "https://example.com/user/789",
		Email:     "user@example.com", // Same email
		Name:      "Bob Smith",
		Status:    string(domain.UserStatusActive),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = db.Create(duplicateEmailUser).Error
	assert.Error(t, err, "Should fail due to unique email constraint")
}

func TestRoleModel_Validation(t *testing.T) {
	db := setupTestDB(t)

	// Test that RoleModel can be created and migrated
	err := db.AutoMigrate(&RoleModel{})
	require.NoError(t, err)

	// Test valid role model with JSON permissions
	permissions := []domain.Permission{
		{Resource: "user", Action: "read", Scope: "account"},
		{Resource: "resource", Action: "*", Scope: "own"},
	}
	permissionsJSON, err := json.Marshal(permissions)
	require.NoError(t, err)

	role := &RoleModel{
		ID:          "role-123",
		Name:        "Test Role",
		Description: "A test role",
		Permissions: string(permissionsJSON),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err = db.Create(role).Error
	require.NoError(t, err)

	// Test retrieval and JSON deserialization
	var retrievedRole RoleModel
	err = db.First(&retrievedRole, "id = ?", "role-123").Error
	require.NoError(t, err)

	var retrievedPermissions []domain.Permission
	err = json.Unmarshal([]byte(retrievedRole.Permissions), &retrievedPermissions)
	require.NoError(t, err)
	assert.Equal(t, permissions, retrievedPermissions)
}

func TestAccountModel_Validation(t *testing.T) {
	db := setupTestDB(t)

	// Test that AccountModel can be created and migrated
	err := db.AutoMigrate(&AccountModel{})
	require.NoError(t, err)

	// Test valid account model with JSON settings
	settings := domain.AccountSettings{
		AllowInvitations: true,
		DefaultRoleID:    "member",
		MaxMembers:       100,
	}
	settingsJSON, err := json.Marshal(settings)
	require.NoError(t, err)

	account := &AccountModel{
		ID:          "account-123",
		OwnerID:     "user-123",
		Name:        "Test Account",
		Description: "A test account",
		Settings:    string(settingsJSON),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err = db.Create(account).Error
	require.NoError(t, err)

	// Test retrieval and JSON deserialization
	var retrievedAccount AccountModel
	err = db.First(&retrievedAccount, "id = ?", "account-123").Error
	require.NoError(t, err)

	var retrievedSettings domain.AccountSettings
	err = json.Unmarshal([]byte(retrievedAccount.Settings), &retrievedSettings)
	require.NoError(t, err)
	assert.Equal(t, settings, retrievedSettings)

	// Test owner index exists (should be able to query by owner)
	var accountsByOwner []AccountModel
	err = db.Where("owner_id = ?", "user-123").Find(&accountsByOwner).Error
	require.NoError(t, err)
	assert.Len(t, accountsByOwner, 1)
}

func TestAccountMemberModel_Validation(t *testing.T) {
	db := setupTestDB(t)

	// Test that AccountMemberModel can be created and migrated
	err := db.AutoMigrate(&AccountMemberModel{})
	require.NoError(t, err)

	// Test valid account member model
	member := &AccountMemberModel{
		ID:        "member-123",
		AccountID: "account-123",
		UserID:    "user-123",
		RoleID:    "role-123",
		InvitedBy: "user-456",
		JoinedAt:  time.Now(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = db.Create(member).Error
	require.NoError(t, err)

	// Test unique constraint on account+user combination
	duplicateMember := &AccountMemberModel{
		ID:        "member-456",
		AccountID: "account-123", // Same account
		UserID:    "user-123",    // Same user
		RoleID:    "role-456",
		InvitedBy: "user-789",
		JoinedAt:  time.Now(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = db.Create(duplicateMember).Error
	assert.Error(t, err, "Should fail due to unique account+user constraint")

	// Test indexes exist (should be able to query by account and user)
	var membersByAccount []AccountMemberModel
	err = db.Where("account_id = ?", "account-123").Find(&membersByAccount).Error
	require.NoError(t, err)
	assert.Len(t, membersByAccount, 1)

	var membersByUser []AccountMemberModel
	err = db.Where("user_id = ?", "user-123").Find(&membersByUser).Error
	require.NoError(t, err)
	assert.Len(t, membersByUser, 1)
}

func TestInvitationModel_Validation(t *testing.T) {
	db := setupTestDB(t)

	// Test that InvitationModel can be created and migrated
	err := db.AutoMigrate(&InvitationModel{})
	require.NoError(t, err)

	// Test valid invitation model
	invitation := &InvitationModel{
		ID:        "invitation-123",
		AccountID: "account-123",
		Email:     "invited@example.com",
		RoleID:    "role-123",
		Token:     "unique-token-123",
		InvitedBy: "user-123",
		Status:    string(domain.InvitationStatusPending),
		ExpiresAt: time.Now().Add(24 * time.Hour),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = db.Create(invitation).Error
	require.NoError(t, err)

	// Test unique token constraint
	duplicateTokenInvitation := &InvitationModel{
		ID:        "invitation-456",
		AccountID: "account-456",
		Email:     "other@example.com",
		RoleID:    "role-456",
		Token:     "unique-token-123", // Same token
		InvitedBy: "user-456",
		Status:    string(domain.InvitationStatusPending),
		ExpiresAt: time.Now().Add(24 * time.Hour),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = db.Create(duplicateTokenInvitation).Error
	assert.Error(t, err, "Should fail due to unique token constraint")

	// Test indexes exist (should be able to query by account)
	var invitationsByAccount []InvitationModel
	err = db.Where("account_id = ?", "account-123").Find(&invitationsByAccount).Error
	require.NoError(t, err)
	assert.Len(t, invitationsByAccount, 1)
}

func TestModels_Relationships(t *testing.T) {
	db := setupTestDB(t)

	// Migrate all models
	err := db.AutoMigrate(&UserModel{}, &AccountModel{}, &RoleModel{}, &AccountMemberModel{}, &InvitationModel{})
	require.NoError(t, err)

	// Create test data
	user := &UserModel{
		ID:        "user-123",
		WebID:     "https://example.com/user/123",
		Email:     "user@example.com",
		Name:      "John Doe",
		Status:    string(domain.UserStatusActive),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	err = db.Create(user).Error
	require.NoError(t, err)

	role := &RoleModel{
		ID:          "role-123",
		Name:        "Member",
		Description: "Standard member role",
		Permissions: `[{"resource":"user","action":"read","scope":"account"}]`,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	err = db.Create(role).Error
	require.NoError(t, err)

	account := &AccountModel{
		ID:          "account-123",
		OwnerID:     "user-123",
		Name:        "Test Account",
		Description: "A test account",
		Settings:    `{"allow_invitations":true,"default_role_id":"member","max_members":100}`,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	err = db.Create(account).Error
	require.NoError(t, err)

	member := &AccountMemberModel{
		ID:        "member-123",
		AccountID: "account-123",
		UserID:    "user-123",
		RoleID:    "role-123",
		InvitedBy: "user-123",
		JoinedAt:  time.Now(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	err = db.Create(member).Error
	require.NoError(t, err)

	invitation := &InvitationModel{
		ID:        "invitation-123",
		AccountID: "account-123",
		Email:     "invited@example.com",
		RoleID:    "role-123",
		Token:     "unique-token-123",
		InvitedBy: "user-123",
		Status:    string(domain.InvitationStatusPending),
		ExpiresAt: time.Now().Add(24 * time.Hour),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	err = db.Create(invitation).Error
	require.NoError(t, err)

	// Test that all relationships work correctly
	// Query members by account
	var members []AccountMemberModel
	err = db.Where("account_id = ?", "account-123").Find(&members).Error
	require.NoError(t, err)
	assert.Len(t, members, 1)

	// Query invitations by account
	var invitations []InvitationModel
	err = db.Where("account_id = ?", "account-123").Find(&invitations).Error
	require.NoError(t, err)
	assert.Len(t, invitations, 1)

	// Query accounts by owner
	var accounts []AccountModel
	err = db.Where("owner_id = ?", "user-123").Find(&accounts).Error
	require.NoError(t, err)
	assert.Len(t, accounts, 1)
}

func TestJSON_Serialization(t *testing.T) {
	// Test UserProfile JSON serialization
	profile := domain.UserProfile{
		Name:        "John Doe",
		Bio:         "Software developer",
		Avatar:      "https://example.com/avatar.jpg",
		Preferences: map[string]interface{}{"theme": "dark", "notifications": true},
	}

	profileJSON, err := json.Marshal(profile)
	require.NoError(t, err)

	var deserializedProfile domain.UserProfile
	err = json.Unmarshal(profileJSON, &deserializedProfile)
	require.NoError(t, err)
	assert.Equal(t, profile, deserializedProfile)

	// Test AccountSettings JSON serialization
	settings := domain.AccountSettings{
		AllowInvitations: true,
		DefaultRoleID:    "member",
		MaxMembers:       100,
	}

	settingsJSON, err := json.Marshal(settings)
	require.NoError(t, err)

	var deserializedSettings domain.AccountSettings
	err = json.Unmarshal(settingsJSON, &deserializedSettings)
	require.NoError(t, err)
	assert.Equal(t, settings, deserializedSettings)

	// Test Permissions JSON serialization
	permissions := []domain.Permission{
		{Resource: "user", Action: "read", Scope: "account"},
		{Resource: "resource", Action: "*", Scope: "own"},
		{Resource: "*", Action: "*", Scope: "global"},
	}

	permissionsJSON, err := json.Marshal(permissions)
	require.NoError(t, err)

	var deserializedPermissions []domain.Permission
	err = json.Unmarshal(permissionsJSON, &deserializedPermissions)
	require.NoError(t, err)
	assert.Equal(t, permissions, deserializedPermissions)
}

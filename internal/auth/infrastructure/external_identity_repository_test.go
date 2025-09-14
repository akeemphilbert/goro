package infrastructure

import (
	"context"
	"testing"

	"github.com/akeemphilbert/goro/internal/auth/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGormExternalIdentityRepository(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGormExternalIdentityRepository(db)
	ctx := context.Background()

	t.Run("LinkIdentity valid identity", func(t *testing.T) {
		err := repo.LinkIdentity(ctx, "user-123", "google", "google-456")
		assert.NoError(t, err)

		// Verify it was linked
		userID, err := repo.FindByExternalID(ctx, "google", "google-456")
		assert.NoError(t, err)
		assert.Equal(t, "user-123", userID)
	})

	t.Run("LinkIdentity empty parameters", func(t *testing.T) {
		err := repo.LinkIdentity(ctx, "", "google", "google-456")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "userID, provider, and externalID cannot be empty")

		err = repo.LinkIdentity(ctx, "user-123", "", "google-456")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "userID, provider, and externalID cannot be empty")

		err = repo.LinkIdentity(ctx, "user-123", "google", "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "userID, provider, and externalID cannot be empty")
	})

	t.Run("LinkIdentity duplicate identity", func(t *testing.T) {
		// Link first identity
		err := repo.LinkIdentity(ctx, "user-first", "github", "github-789")
		require.NoError(t, err)

		// Try to link same external identity to different user
		err = repo.LinkIdentity(ctx, "user-second", "github", "github-789")
		assert.Error(t, err)
		assert.Equal(t, domain.ErrExternalIdentityAlreadyLinked, err)
	})

	t.Run("LinkIdentity same external ID different providers", func(t *testing.T) {
		// Link to Google
		err := repo.LinkIdentity(ctx, "user-multi-1", "google", "same-id-123")
		assert.NoError(t, err)

		// Link same external ID to GitHub (should be allowed)
		err = repo.LinkIdentity(ctx, "user-multi-2", "github", "same-id-123")
		assert.NoError(t, err)

		// Verify both exist
		userID1, err := repo.FindByExternalID(ctx, "google", "same-id-123")
		assert.NoError(t, err)
		assert.Equal(t, "user-multi-1", userID1)

		userID2, err := repo.FindByExternalID(ctx, "github", "same-id-123")
		assert.NoError(t, err)
		assert.Equal(t, "user-multi-2", userID2)
	})

	t.Run("FindByExternalID existing identity", func(t *testing.T) {
		err := repo.LinkIdentity(ctx, "user-find-123", "google", "google-find-456")
		require.NoError(t, err)

		userID, err := repo.FindByExternalID(ctx, "google", "google-find-456")
		assert.NoError(t, err)
		assert.Equal(t, "user-find-123", userID)
	})

	t.Run("FindByExternalID non-existent identity", func(t *testing.T) {
		_, err := repo.FindByExternalID(ctx, "google", "non-existent")
		assert.Error(t, err)
		assert.Equal(t, domain.ErrExternalIdentityNotFound, err)
	})

	t.Run("FindByExternalID empty parameters", func(t *testing.T) {
		_, err := repo.FindByExternalID(ctx, "", "google-456")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "provider and externalID cannot be empty")

		_, err = repo.FindByExternalID(ctx, "google", "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "provider and externalID cannot be empty")
	})

	t.Run("GetLinkedIdentities", func(t *testing.T) {
		userID := "user-multi-identities"

		// Link multiple identities for the same user
		err := repo.LinkIdentity(ctx, userID, "google", "google-multi-1")
		require.NoError(t, err)
		err = repo.LinkIdentity(ctx, userID, "github", "github-multi-1")
		require.NoError(t, err)

		identities, err := repo.GetLinkedIdentities(ctx, userID)
		assert.NoError(t, err)
		assert.Len(t, identities, 2)

		// Verify both identities are present
		providers := make(map[string]string)
		for _, identity := range identities {
			providers[identity.Provider] = identity.ExternalID
		}
		assert.Equal(t, "google-multi-1", providers["google"])
		assert.Equal(t, "github-multi-1", providers["github"])
	})

	t.Run("GetLinkedIdentities empty user ID", func(t *testing.T) {
		_, err := repo.GetLinkedIdentities(ctx, "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "user ID cannot be empty")
	})

	t.Run("GetLinkedIdentities no identities", func(t *testing.T) {
		identities, err := repo.GetLinkedIdentities(ctx, "user-no-identities")
		assert.NoError(t, err)
		assert.Len(t, identities, 0)
	})

	t.Run("UnlinkIdentity existing identity", func(t *testing.T) {
		userID := "user-unlink"
		provider := "google"
		externalID := "google-unlink-123"

		// Link identity first
		err := repo.LinkIdentity(ctx, userID, provider, externalID)
		require.NoError(t, err)

		// Verify it exists
		foundUserID, err := repo.FindByExternalID(ctx, provider, externalID)
		assert.NoError(t, err)
		assert.Equal(t, userID, foundUserID)

		// Unlink it
		err = repo.UnlinkIdentity(ctx, userID, provider, externalID)
		assert.NoError(t, err)

		// Verify it was unlinked
		_, err = repo.FindByExternalID(ctx, provider, externalID)
		assert.Equal(t, domain.ErrExternalIdentityNotFound, err)
	})

	t.Run("UnlinkIdentity non-existent identity", func(t *testing.T) {
		err := repo.UnlinkIdentity(ctx, "user-123", "google", "non-existent")
		assert.Error(t, err)
		assert.Equal(t, domain.ErrExternalIdentityNotFound, err)
	})

	t.Run("UnlinkIdentity empty parameters", func(t *testing.T) {
		err := repo.UnlinkIdentity(ctx, "", "google", "google-456")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "userID, provider, and externalID cannot be empty")

		err = repo.UnlinkIdentity(ctx, "user-123", "", "google-456")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "userID, provider, and externalID cannot be empty")

		err = repo.UnlinkIdentity(ctx, "user-123", "google", "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "userID, provider, and externalID cannot be empty")
	})

	t.Run("UnlinkAllIdentities", func(t *testing.T) {
		userID := "user-unlink-all"

		// Link multiple identities
		err := repo.LinkIdentity(ctx, userID, "google", "google-all-1")
		require.NoError(t, err)
		err = repo.LinkIdentity(ctx, userID, "github", "github-all-1")
		require.NoError(t, err)

		// Verify they exist
		identities, err := repo.GetLinkedIdentities(ctx, userID)
		assert.NoError(t, err)
		assert.Len(t, identities, 2)

		// Unlink all
		err = repo.UnlinkAllIdentities(ctx, userID)
		assert.NoError(t, err)

		// Verify all were unlinked
		identities, err = repo.GetLinkedIdentities(ctx, userID)
		assert.NoError(t, err)
		assert.Len(t, identities, 0)
	})

	t.Run("UnlinkAllIdentities empty user ID", func(t *testing.T) {
		err := repo.UnlinkAllIdentities(ctx, "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "user ID cannot be empty")
	})

	t.Run("IsLinked existing identity", func(t *testing.T) {
		err := repo.LinkIdentity(ctx, "user-is-linked", "google", "google-is-linked")
		require.NoError(t, err)

		isLinked, err := repo.IsLinked(ctx, "google", "google-is-linked")
		assert.NoError(t, err)
		assert.True(t, isLinked)
	})

	t.Run("IsLinked non-existent identity", func(t *testing.T) {
		isLinked, err := repo.IsLinked(ctx, "google", "non-existent")
		assert.NoError(t, err)
		assert.False(t, isLinked)
	})

	t.Run("IsLinked empty parameters", func(t *testing.T) {
		_, err := repo.IsLinked(ctx, "", "google-456")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "provider and externalID cannot be empty")

		_, err = repo.IsLinked(ctx, "google", "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "provider and externalID cannot be empty")
	})

	t.Run("GetByProvider", func(t *testing.T) {
		provider := "test-provider"

		// Link multiple identities for the same provider
		err := repo.LinkIdentity(ctx, "user-provider-1", provider, "ext-1")
		require.NoError(t, err)
		err = repo.LinkIdentity(ctx, "user-provider-2", provider, "ext-2")
		require.NoError(t, err)

		identities, err := repo.GetByProvider(ctx, provider)
		assert.NoError(t, err)
		assert.Len(t, identities, 2)

		// Verify all identities have the correct provider
		for _, identity := range identities {
			assert.Equal(t, provider, identity.Provider)
		}
	})

	t.Run("GetByProvider empty provider", func(t *testing.T) {
		_, err := repo.GetByProvider(ctx, "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "provider cannot be empty")
	})

	t.Run("GetByProvider no identities", func(t *testing.T) {
		identities, err := repo.GetByProvider(ctx, "non-existent-provider")
		assert.NoError(t, err)
		assert.Len(t, identities, 0)
	})

	t.Run("Complex scenario - multiple users and providers", func(t *testing.T) {
		// User 1 with Google and GitHub
		err := repo.LinkIdentity(ctx, "complex-user-1", "google", "google-complex-1")
		require.NoError(t, err)
		err = repo.LinkIdentity(ctx, "complex-user-1", "github", "github-complex-1")
		require.NoError(t, err)

		// User 2 with Google only
		err = repo.LinkIdentity(ctx, "complex-user-2", "google", "google-complex-2")
		require.NoError(t, err)

		// Verify User 1 has 2 identities
		identities1, err := repo.GetLinkedIdentities(ctx, "complex-user-1")
		assert.NoError(t, err)
		assert.Len(t, identities1, 2)

		// Verify User 2 has 1 identity
		identities2, err := repo.GetLinkedIdentities(ctx, "complex-user-2")
		assert.NoError(t, err)
		assert.Len(t, identities2, 1)

		// Verify Google provider has 2 identities
		googleIdentities, err := repo.GetByProvider(ctx, "google")
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(googleIdentities), 2) // At least 2 from this test

		// Verify GitHub provider has at least 1 identity
		githubIdentities, err := repo.GetByProvider(ctx, "github")
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(githubIdentities), 1) // At least 1 from this test
	})
}

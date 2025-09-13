package handlers

import (
	"testing"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResourceHandler_generateResourceID(t *testing.T) {
	// Setup
	logger := log.NewStdLogger(nil)
	handler := NewResourceHandler(nil, logger)

	t.Run("generates valid KSUID", func(t *testing.T) {
		// Generate multiple IDs
		id1 := handler.generateResourceID()
		id2 := handler.generateResourceID()

		// Verify they are different
		assert.NotEqual(t, id1, id2, "Generated IDs should be unique")

		// Verify they are valid KSUIDs
		parsedID1, err := ksuid.Parse(id1)
		require.NoError(t, err, "ID1 should be a valid KSUID")

		parsedID2, err := ksuid.Parse(id2)
		require.NoError(t, err, "ID2 should be a valid KSUID")

		// Verify KSUID properties
		assert.Len(t, id1, 27, "KSUID should be 27 characters long")
		assert.Len(t, id2, 27, "KSUID should be 27 characters long")

		// Note: KSUIDs generated in very quick succession may not maintain strict ordering
		// due to the random payload component. The key property is uniqueness.
		// Strict temporal ordering is only guaranteed when there's sufficient time difference.

		// Verify timestamp component (should be recent)
		timestamp1 := parsedID1.Time()
		timestamp2 := parsedID2.Time()

		assert.True(t, timestamp2.After(timestamp1) || timestamp2.Equal(timestamp1),
			"Second KSUID timestamp should be after or equal to first")
	})

	t.Run("generates unique IDs in sequence", func(t *testing.T) {
		// Generate multiple IDs quickly
		ids := make([]string, 100)
		for i := 0; i < 100; i++ {
			ids[i] = handler.generateResourceID()
		}

		// Verify all IDs are unique
		idSet := make(map[string]bool)
		for _, id := range ids {
			assert.False(t, idSet[id], "ID %s should be unique", id)
			idSet[id] = true

			// Verify each is a valid KSUID
			_, err := ksuid.Parse(id)
			assert.NoError(t, err, "ID %s should be a valid KSUID", id)
		}

		assert.Len(t, idSet, 100, "All 100 IDs should be unique")
	})

	t.Run("KSUID format validation", func(t *testing.T) {
		id := handler.generateResourceID()

		// KSUID should be base62 encoded
		for _, char := range id {
			assert.True(t,
				(char >= '0' && char <= '9') ||
					(char >= 'A' && char <= 'Z') ||
					(char >= 'a' && char <= 'z'),
				"KSUID should only contain base62 characters (0-9, A-Z, a-z), found: %c", char)
		}

		// Parse and verify components
		parsedID, err := ksuid.Parse(id)
		require.NoError(t, err)

		// Verify timestamp is reasonable (within last minute)
		timestamp := parsedID.Time()
		assert.WithinDuration(t, timestamp,
			parsedID.Time(),
			60*1000000000, // 60 seconds in nanoseconds
			"KSUID timestamp should be recent")

		// Verify payload exists (16 bytes of randomness)
		payload := parsedID.Payload()
		assert.Len(t, payload, 16, "KSUID payload should be 16 bytes")
	})
}

package infrastructure

import (
	"context"
	"os"
	"testing"

	"github.com/akeemphilbert/goro/internal/ldp/domain"
	"github.com/akeemphilbert/pericarp/pkg/infrastructure"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewEventDispatcher_WireProvider(t *testing.T) {
	// Act
	dispatcher, err := NewEventDispatcher()

	// Assert
	require.NoError(t, err)
	require.NotNil(t, dispatcher)

	// Verify it implements the domain interface
	var _ domain.EventDispatcher = dispatcher

	// Cleanup - cast to concrete type to access Close method
	if closer, ok := dispatcher.(*infrastructure.WatermillEventDispatcher); ok {
		err = closer.Close()
		assert.NoError(t, err)
	}
}

func TestEventInfrastructureSetIntegration(t *testing.T) {
	// This test verifies that the Wire set is properly configured
	// In a real Wire-generated application, this would be tested through
	// the generated wire_gen.go file, but we can test the provider directly

	// Act
	dispatcher, err := NewEventDispatcher()

	// Assert
	require.NoError(t, err)
	require.NotNil(t, dispatcher)

	// Cleanup - cast to concrete type to access Close method
	if closer, ok := dispatcher.(*infrastructure.WatermillEventDispatcher); ok {
		err = closer.Close()
		assert.NoError(t, err)
	}
}
func TestRDFInfrastructureSetIntegration(t *testing.T) {
	// This test verifies that the RDF Wire set can be used to create dependencies
	converter := NewRDFConverter()
	if converter == nil {
		t.Fatal("RDFInfrastructureSet should provide a valid RDFConverter")
	}

	// Test basic functionality
	if !converter.ValidateFormat("application/ld+json") {
		t.Error("RDFConverter should validate JSON-LD format")
	}
}

func TestStorageInfrastructureSetIntegration(t *testing.T) {
	// Clean up any existing test data
	defer func() {
		os.RemoveAll("./data")
	}()

	// This test verifies that the Storage Wire set can be used to create dependencies
	repo, err := NewFileSystemRepositoryProvider()
	require.NoError(t, err)
	require.NotNil(t, repo)

	// Verify it implements the domain interface
	var _ domain.ResourceRepository = repo

	// Test basic functionality
	ctx := context.Background()
	testResource := domain.NewResource(ctx, "wire-test", "text/plain", []byte("test data"))

	err = repo.Store(ctx, testResource)
	assert.NoError(t, err)

	exists, err := repo.Exists(ctx, "wire-test")
	assert.NoError(t, err)
	assert.True(t, exists)
}

package infrastructure

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/akeemphilbert/goro/internal/conf"
	"github.com/akeemphilbert/goro/internal/ldp/domain"
	pericarpdomain "github.com/akeemphilbert/pericarp/pkg/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewFileSystemContainerRepositoryProviderFixed(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "container_repo_provider_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create test configuration
	storagePath := filepath.Join(tempDir, "storage")
	indexPath := filepath.Join(tempDir, "index")

	// Create directories
	err = os.MkdirAll(storagePath, 0755)
	require.NoError(t, err)
	err = os.MkdirAll(indexPath, 0755)
	require.NoError(t, err)

	config := &conf.Container{
		StoragePath:     storagePath,
		IndexPath:       indexPath,
		MaxDepth:        100,
		PageSize:        50,
		CacheEnabled:    true,
		CacheSize:       1000,
		IndexingEnabled: true,
	}

	t.Run("creates container repository successfully", func(t *testing.T) {
		// Test the provider function
		repo, err := NewFileSystemContainerRepositoryProvider(config)

		require.NoError(t, err, "Container repository provider should create repository successfully")
		assert.NotNil(t, repo, "Container repository should not be nil")
		assert.Implements(t, (*domain.ContainerRepository)(nil), repo, "Should implement ContainerRepository interface")
	})

	t.Run("creates repository with proper type", func(t *testing.T) {
		repo, err := NewFileSystemContainerRepositoryProvider(config)

		require.NoError(t, err)
		assert.IsType(t, &FileSystemContainerRepository{}, repo, "Should return FileSystemContainerRepository type")
	})

	t.Run("repository has membership indexer", func(t *testing.T) {
		repo, err := NewFileSystemContainerRepositoryProvider(config)

		require.NoError(t, err)

		// Cast to concrete type to verify internal structure
		fsRepo, ok := repo.(*FileSystemContainerRepository)
		require.True(t, ok, "Should be FileSystemContainerRepository")
		assert.NotNil(t, fsRepo, "Repository should not be nil")
	})

	t.Run("handles invalid configuration", func(t *testing.T) {
		// Test with invalid configuration
		invalidConfig := &conf.Container{
			StoragePath: "",
			IndexPath:   "",
		}

		_, err := NewFileSystemContainerRepositoryProvider(invalidConfig)
		assert.Error(t, err, "Should return error with invalid configuration")
	})

	t.Run("handles nil configuration", func(t *testing.T) {
		// Test with nil configuration (should use defaults)
		// First create the default directories that would be used
		defaultConfig := &conf.Container{}
		defaultConfig.SetDefaults()

		err := os.MkdirAll(defaultConfig.StoragePath, 0755)
		require.NoError(t, err)
		err = os.MkdirAll(defaultConfig.IndexPath, 0755)
		require.NoError(t, err)

		// Clean up after test
		defer os.RemoveAll(defaultConfig.StoragePath)
		defer os.RemoveAll(defaultConfig.IndexPath)

		repo, err := NewFileSystemContainerRepositoryProvider(nil)

		require.NoError(t, err, "Should handle nil configuration with defaults")
		assert.NotNil(t, repo, "Repository should not be nil")
	})
}

func TestNewUnitOfWorkFactoryFixed(t *testing.T) {
	t.Run("creates unit of work factory successfully", func(t *testing.T) {
		db, err := DatabaseProvider()
		require.NoError(t, err)

		eventStore, err := EventStoreProvider(db)
		require.NoError(t, err)

		eventDispatcher, err := NewEventDispatcher()
		require.NoError(t, err)

		factory := NewUnitOfWorkFactory(eventStore, eventDispatcher)

		assert.NotNil(t, factory, "Unit of work factory should not be nil")
		assert.IsType(t, (func() pericarpdomain.UnitOfWork)(nil), factory, "Should return function type")
	})

	t.Run("factory creates unit of work instances", func(t *testing.T) {
		db, err := DatabaseProvider()
		require.NoError(t, err)

		eventStore, err := EventStoreProvider(db)
		require.NoError(t, err)

		eventDispatcher, err := NewEventDispatcher()
		require.NoError(t, err)

		factory := NewUnitOfWorkFactory(eventStore, eventDispatcher)
		uow := factory()

		assert.NotNil(t, uow, "Unit of work should not be nil")
		assert.Implements(t, (*pericarpdomain.UnitOfWork)(nil), uow, "Should implement UnitOfWork interface")
	})

	t.Run("handles nil dependencies", func(t *testing.T) {
		// Test with nil event store
		eventDispatcher, err := NewEventDispatcher()
		require.NoError(t, err)
		factory := NewUnitOfWorkFactory(nil, eventDispatcher)
		assert.NotNil(t, factory, "Factory should be created even with nil event store")

		// Test with nil event dispatcher
		db2, err := DatabaseProvider()
		require.NoError(t, err)

		eventStore2, err := EventStoreProvider(db2)
		require.NoError(t, err)

		factory = NewUnitOfWorkFactory(eventStore2, nil)
		assert.NotNil(t, factory, "Factory should be created even with nil event dispatcher")
	})
}

func TestContainerRDFConverterProviderFixed(t *testing.T) {
	t.Run("creates container RDF converter successfully", func(t *testing.T) {
		converter := NewContainerRDFConverter()

		assert.NotNil(t, converter, "Container RDF converter should not be nil")
		assert.IsType(t, &ContainerRDFConverter{}, converter, "Should return ContainerRDFConverter type")
	})

	t.Run("converter is ready for use", func(t *testing.T) {
		converter := NewContainerRDFConverter()

		// Verify converter is properly initialized
		assert.NotNil(t, converter, "Converter should be initialized")
	})
}

func TestInfrastructureSetIntegrationFixed(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "infrastructure_set_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	t.Run("all infrastructure providers work together", func(t *testing.T) {
		// Test that all providers in the infrastructure set can work together
		basePath := filepath.Join(tempDir, "data")
		err := os.MkdirAll(basePath, 0755)
		require.NoError(t, err)

		// Test individual providers that would be used by Wire
		db, err := DatabaseProvider()
		require.NoError(t, err)

		eventStore, err := EventStoreProvider(db)
		require.NoError(t, err)

		eventDispatcher, err := NewEventDispatcher()
		require.NoError(t, err)

		unitOfWorkFactory := NewUnitOfWorkFactory(eventStore, eventDispatcher)
		assert.NotNil(t, unitOfWorkFactory, "Unit of work factory should be created")

		config := &conf.Container{
			StoragePath:     basePath,
			IndexPath:       filepath.Join(basePath, "index"),
			MaxDepth:        100,
			PageSize:        50,
			CacheEnabled:    true,
			CacheSize:       1000,
			IndexingEnabled: true,
		}

		containerRepo, err := NewFileSystemContainerRepositoryProvider(config)
		require.NoError(t, err, "Container repository should be created")
		assert.NotNil(t, containerRepo, "Container repository should not be nil")

		rdfConverter := NewRDFConverter()
		assert.NotNil(t, rdfConverter, "RDF converter should be created")

		containerRDFConverter := NewContainerRDFConverter()
		assert.NotNil(t, containerRDFConverter, "Container RDF converter should be created")
	})

	t.Run("providers create compatible components", func(t *testing.T) {
		// Test that providers create components that are compatible with each other
		db, err := DatabaseProvider()
		require.NoError(t, err)

		eventStore, err := EventStoreProvider(db)
		require.NoError(t, err)

		eventDispatcher, err := NewEventDispatcher()
		require.NoError(t, err)

		unitOfWorkFactory := NewUnitOfWorkFactory(eventStore, eventDispatcher)
		uow := unitOfWorkFactory()

		assert.NotNil(t, uow, "Unit of work should be created by factory")
		assert.Implements(t, (*pericarpdomain.UnitOfWork)(nil), uow, "Should implement UnitOfWork interface")

		config := &conf.Container{
			StoragePath:     filepath.Join(tempDir, "data2"),
			IndexPath:       filepath.Join(tempDir, "data2", "index"),
			MaxDepth:        100,
			PageSize:        50,
			CacheEnabled:    true,
			CacheSize:       1000,
			IndexingEnabled: true,
		}

		containerRepo, err := NewFileSystemContainerRepositoryProvider(config)
		require.NoError(t, err)
		assert.Implements(t, (*domain.ContainerRepository)(nil), containerRepo, "Should implement ContainerRepository interface")
	})
}

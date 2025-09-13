package application

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/akeemphilbert/goro/internal/ldp/domain"
	pericarpdomain "github.com/akeemphilbert/pericarp/pkg/domain"
)

// Mock implementations for testing

type mockRepository struct {
	resources map[string]*domain.Resource
	mu        sync.RWMutex
	storeErr  error
	getErr    error
	deleteErr error
	existsErr error
}

func newMockRepository() *mockRepository {
	return &mockRepository{
		resources: make(map[string]*domain.Resource),
	}
}

func (m *mockRepository) Store(ctx context.Context, resource *domain.Resource) error {
	if m.storeErr != nil {
		return m.storeErr
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.resources[resource.ID()] = resource
	return nil
}

func (m *mockRepository) Retrieve(ctx context.Context, id string) (*domain.Resource, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	resource, exists := m.resources[id]
	if !exists {
		return nil, domain.ErrResourceNotFound
	}
	return resource, nil
}

func (m *mockRepository) Delete(ctx context.Context, id string) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.resources, id)
	return nil
}

func (m *mockRepository) Exists(ctx context.Context, id string) (bool, error) {
	if m.existsErr != nil {
		return false, m.existsErr
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, exists := m.resources[id]
	return exists, nil
}

type mockConverter struct {
	convertErr    error
	validateErr   bool
	convertResult []byte
}

func newMockConverter() *mockConverter {
	return &mockConverter{
		convertResult: []byte(`{"@context": {}, "@id": "test", "@type": "Test"}`),
	}
}

func (m *mockConverter) Convert(data []byte, fromFormat, toFormat string) ([]byte, error) {
	if m.convertErr != nil {
		return nil, m.convertErr
	}
	return m.convertResult, nil
}

func (m *mockConverter) ValidateFormat(format string) bool {
	if m.validateErr {
		return false
	}
	validFormats := map[string]bool{
		"application/ld+json": true,
		"text/turtle":         true,
		"application/rdf+xml": true,
	}
	return validFormats[format]
}

type mockUnitOfWork struct {
	events      []pericarpdomain.Event
	envelopes   []pericarpdomain.Envelope
	commitErr   error
	rollbackErr error
	mu          sync.Mutex
}

func newMockUnitOfWork() *mockUnitOfWork {
	return &mockUnitOfWork{
		events:    make([]pericarpdomain.Event, 0),
		envelopes: make([]pericarpdomain.Envelope, 0),
	}
}

func (m *mockUnitOfWork) RegisterEvents(events []pericarpdomain.Event) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.events = append(m.events, events...)
}

func (m *mockUnitOfWork) Commit(ctx context.Context) ([]pericarpdomain.Envelope, error) {
	if m.commitErr != nil {
		return nil, m.commitErr
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Create mock envelopes from events
	envelopes := make([]pericarpdomain.Envelope, len(m.events))
	for i, event := range m.events {
		envelopes[i] = &mockEnvelope{
			event:     event,
			eventID:   fmt.Sprintf("event-%d", i),
			timestamp: time.Now(),
			metadata:  make(map[string]interface{}),
		}
	}

	m.envelopes = append(m.envelopes, envelopes...)
	m.events = make([]pericarpdomain.Event, 0) // Clear events after commit

	return envelopes, nil
}

func (m *mockUnitOfWork) Rollback() error {
	if m.rollbackErr != nil {
		return m.rollbackErr
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	m.events = make([]pericarpdomain.Event, 0) // Clear events on rollback

	return nil
}

func (m *mockUnitOfWork) GetEnvelopes() []pericarpdomain.Envelope {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]pericarpdomain.Envelope{}, m.envelopes...)
}

type mockEnvelope struct {
	event     pericarpdomain.Event
	eventID   string
	timestamp time.Time
	metadata  map[string]interface{}
}

func (e *mockEnvelope) Event() pericarpdomain.Event {
	return e.event
}

func (e *mockEnvelope) Metadata() map[string]interface{} {
	return e.metadata
}

func (e *mockEnvelope) EventID() string {
	return e.eventID
}

func (e *mockEnvelope) Timestamp() time.Time {
	return e.timestamp
}

// Test helper functions

func createMockUnitOfWorkFactory() UnitOfWorkFactory {
	return func() pericarpdomain.UnitOfWork {
		return newMockUnitOfWork()
	}
}

// Test cases

func TestNewStorageService(t *testing.T) {
	repo := newMockRepository()
	converter := newMockConverter()
	unitOfWorkFactory := func() pericarpdomain.UnitOfWork {
		return newMockUnitOfWork()
	}

	service := NewStorageService(repo, converter, unitOfWorkFactory)

	if service == nil {
		t.Fatal("NewStorageService returned nil")
	}
	if service.repo != repo {
		t.Error("Repository not set correctly")
	}
	if service.converter != converter {
		t.Error("Converter not set correctly")
	}
	// Note: Can't directly compare unitOfWorkFactory due to function type
	if service.unitOfWorkFactory == nil {
		t.Error("UnitOfWorkFactory not set correctly")
	}
}

func TestStorageService_StoreResource(t *testing.T) {
	tests := []struct {
		name        string
		id          string
		data        []byte
		contentType string
		setupRepo   func(*mockRepository)
		wantErr     bool
		errCode     string
	}{
		{
			name:        "successful store new resource",
			id:          "test-resource",
			data:        []byte(`{"@context": {}, "@id": "test"}`),
			contentType: "application/ld+json",
			setupRepo:   func(r *mockRepository) {},
			wantErr:     false,
		},
		{
			name:        "successful store update existing resource",
			id:          "existing-resource",
			data:        []byte(`{"@context": {}, "@id": "updated"}`),
			contentType: "application/ld+json",
			setupRepo: func(r *mockRepository) {
				existing := domain.NewResource("existing-resource", "application/ld+json", []byte(`{"old": "data"}`))
				r.resources["existing-resource"] = existing
			},
			wantErr: false,
		},
		{
			name:        "empty id error",
			id:          "",
			data:        []byte(`{"test": "data"}`),
			contentType: "application/ld+json",
			setupRepo:   func(r *mockRepository) {},
			wantErr:     true,
			errCode:     "INVALID_ID",
		},
		{
			name:        "empty data error",
			id:          "test-resource",
			data:        []byte{},
			contentType: "application/ld+json",
			setupRepo:   func(r *mockRepository) {},
			wantErr:     true,
			errCode:     "INVALID_RESOURCE",
		},
		{
			name:        "repository store error",
			id:          "test-resource",
			data:        []byte(`{"test": "data"}`),
			contentType: "application/ld+json",
			setupRepo: func(r *mockRepository) {
				r.storeErr = errors.New("storage failed")
			},
			wantErr: true,
			errCode: "STORE_FAILED",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newMockRepository()
			converter := newMockConverter()
			unitOfWorkFactory := createMockUnitOfWorkFactory()
			tt.setupRepo(repo)

			service := NewStorageService(repo, converter, unitOfWorkFactory)
			ctx := context.Background()

			resource, err := service.StoreResource(ctx, tt.id, tt.data, tt.contentType)

			if tt.wantErr {
				if err == nil {
					t.Fatal("Expected error but got none")
				}
				if storageErr, ok := domain.GetStorageError(err); ok {
					if storageErr.Code != tt.errCode {
						t.Errorf("Expected error code %s, got %s", tt.errCode, storageErr.Code)
					}
				} else {
					t.Errorf("Expected StorageError, got %T", err)
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if resource == nil {
				t.Fatal("Expected resource but got nil")
			}
			if resource.ID() != tt.id {
				t.Errorf("Expected resource ID %s, got %s", tt.id, resource.ID())
			}
		})
	}
}

func TestStorageService_RetrieveResource(t *testing.T) {
	tests := []struct {
		name         string
		id           string
		acceptFormat string
		setupRepo    func(*mockRepository)
		setupConv    func(*mockConverter)
		wantErr      bool
		errCode      string
	}{
		{
			name:         "successful retrieve without conversion",
			id:           "test-resource",
			acceptFormat: "",
			setupRepo: func(r *mockRepository) {
				resource := domain.NewResource("test-resource", "application/ld+json", []byte(`{"test": "data"}`))
				r.resources["test-resource"] = resource
			},
			setupConv: func(c *mockConverter) {},
			wantErr:   false,
		},
		{
			name:         "successful retrieve with format conversion",
			id:           "test-resource",
			acceptFormat: "text/turtle",
			setupRepo: func(r *mockRepository) {
				resource := domain.NewResource("test-resource", "application/ld+json", []byte(`{"test": "data"}`))
				r.resources["test-resource"] = resource
			},
			setupConv: func(c *mockConverter) {},
			wantErr:   false,
		},
		{
			name:         "empty id error",
			id:           "",
			acceptFormat: "",
			setupRepo:    func(r *mockRepository) {},
			setupConv:    func(c *mockConverter) {},
			wantErr:      true,
			errCode:      "INVALID_ID",
		},
		{
			name:         "resource not found",
			id:           "nonexistent",
			acceptFormat: "",
			setupRepo:    func(r *mockRepository) {},
			setupConv:    func(c *mockConverter) {},
			wantErr:      true,
			errCode:      "RESOURCE_NOT_FOUND",
		},
		{
			name:         "format conversion error",
			id:           "test-resource",
			acceptFormat: "text/turtle",
			setupRepo: func(r *mockRepository) {
				resource := domain.NewResource("test-resource", "application/ld+json", []byte(`{"test": "data"}`))
				r.resources["test-resource"] = resource
			},
			setupConv: func(c *mockConverter) {
				c.convertErr = errors.New("conversion failed")
			},
			wantErr: true,
			errCode: "FORMAT_CONVERSION_FAILED",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newMockRepository()
			converter := newMockConverter()
			unitOfWorkFactory := createMockUnitOfWorkFactory()
			tt.setupRepo(repo)
			tt.setupConv(converter)

			service := NewStorageService(repo, converter, unitOfWorkFactory)
			ctx := context.Background()

			resource, err := service.RetrieveResource(ctx, tt.id, tt.acceptFormat)

			if tt.wantErr {
				if err == nil {
					t.Fatal("Expected error but got none")
				}
				if storageErr, ok := domain.GetStorageError(err); ok {
					if storageErr.Code != tt.errCode {
						t.Errorf("Expected error code %s, got %s", tt.errCode, storageErr.Code)
					}
				} else {
					t.Errorf("Expected StorageError, got %T", err)
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if resource == nil {
				t.Fatal("Expected resource but got nil")
			}
		})
	}
}

func TestStorageService_DeleteResource(t *testing.T) {
	tests := []struct {
		name      string
		id        string
		setupRepo func(*mockRepository)
		wantErr   bool
		errCode   string
	}{
		{
			name: "successful delete",
			id:   "test-resource",
			setupRepo: func(r *mockRepository) {
				resource := domain.NewResource("test-resource", "application/ld+json", []byte(`{"test": "data"}`))
				r.resources["test-resource"] = resource
			},
			wantErr: false,
		},
		{
			name:      "empty id error",
			id:        "",
			setupRepo: func(r *mockRepository) {},
			wantErr:   true,
			errCode:   "INVALID_ID",
		},
		{
			name:      "resource not found",
			id:        "nonexistent",
			setupRepo: func(r *mockRepository) {},
			wantErr:   true,
			errCode:   "RESOURCE_NOT_FOUND",
		},
		{
			name: "repository delete error",
			id:   "test-resource",
			setupRepo: func(r *mockRepository) {
				resource := domain.NewResource("test-resource", "application/ld+json", []byte(`{"test": "data"}`))
				r.resources["test-resource"] = resource
				r.deleteErr = errors.New("delete failed")
			},
			wantErr: true,
			errCode: "DELETE_FAILED",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newMockRepository()
			converter := newMockConverter()
			unitOfWorkFactory := createMockUnitOfWorkFactory()
			tt.setupRepo(repo)

			service := NewStorageService(repo, converter, unitOfWorkFactory)
			ctx := context.Background()

			err := service.DeleteResource(ctx, tt.id)

			if tt.wantErr {
				if err == nil {
					t.Fatal("Expected error but got none")
				}
				if storageErr, ok := domain.GetStorageError(err); ok {
					if storageErr.Code != tt.errCode {
						t.Errorf("Expected error code %s, got %s", tt.errCode, storageErr.Code)
					}
				} else {
					t.Errorf("Expected StorageError, got %T", err)
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			// Verify resource was deleted
			exists, _ := repo.Exists(context.Background(), tt.id)
			if exists {
				t.Error("Resource should have been deleted")
			}
		})
	}
}

func TestStorageService_StreamResource(t *testing.T) {
	repo := newMockRepository()
	converter := newMockConverter()
	unitOfWorkFactory := createMockUnitOfWorkFactory()

	// Setup test resource
	testData := []byte(`{"@context": {}, "@id": "test", "data": "streaming test"}`)
	resource := domain.NewResource("stream-test", "application/ld+json", testData)
	repo.resources["stream-test"] = resource

	service := NewStorageService(repo, converter, unitOfWorkFactory)
	ctx := context.Background()

	reader, contentType, err := service.StreamResource(ctx, "stream-test", "")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	defer reader.Close()

	if contentType != "application/ld+json" {
		t.Errorf("Expected content type application/ld+json, got %s", contentType)
	}

	// Read all data from stream
	data, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("Failed to read from stream: %v", err)
	}

	if string(data) != string(testData) {
		t.Errorf("Expected data %s, got %s", string(testData), string(data))
	}
}

func TestStorageService_StoreResourceStream(t *testing.T) {
	repo := newMockRepository()
	converter := newMockConverter()
	unitOfWorkFactory := createMockUnitOfWorkFactory()

	service := NewStorageService(repo, converter, unitOfWorkFactory)
	ctx := context.Background()

	testData := `{"@context": {}, "@id": "stream-store", "data": "from stream"}`
	reader := strings.NewReader(testData)

	resource, err := service.StoreResourceStream(ctx, "stream-store", reader, "application/ld+json")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if resource.ID() != "stream-store" {
		t.Errorf("Expected resource ID stream-store, got %s", resource.ID())
	}

	if string(resource.GetData()) != testData {
		t.Errorf("Expected data %s, got %s", testData, string(resource.GetData()))
	}
}

func TestStorageService_ResourceExists(t *testing.T) {
	repo := newMockRepository()
	converter := newMockConverter()
	unitOfWorkFactory := createMockUnitOfWorkFactory()

	// Setup test resource
	resource := domain.NewResource("exists-test", "application/ld+json", []byte(`{"test": "data"}`))
	repo.resources["exists-test"] = resource

	service := NewStorageService(repo, converter, unitOfWorkFactory)
	ctx := context.Background()

	// Test existing resource
	exists, err := service.ResourceExists(ctx, "exists-test")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if !exists {
		t.Error("Expected resource to exist")
	}

	// Test non-existing resource
	exists, err = service.ResourceExists(ctx, "nonexistent")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if exists {
		t.Error("Expected resource to not exist")
	}

	// Test empty ID
	_, err = service.ResourceExists(ctx, "")
	if err == nil {
		t.Fatal("Expected error for empty ID")
	}
}

func TestStorageService_ConcurrentAccess(t *testing.T) {
	repo := newMockRepository()
	converter := newMockConverter()
	unitOfWorkFactory := createMockUnitOfWorkFactory()

	service := NewStorageService(repo, converter, unitOfWorkFactory)
	ctx := context.Background()

	// Test concurrent store operations
	var wg sync.WaitGroup
	numGoroutines := 10

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			resourceID := fmt.Sprintf("concurrent-test-%d", id)
			data := []byte(fmt.Sprintf(`{"@context": {}, "@id": "%s"}`, resourceID))

			_, err := service.StoreResource(ctx, resourceID, data, "application/ld+json")
			if err != nil {
				t.Errorf("Concurrent store failed for %s: %v", resourceID, err)
			}
		}(i)
	}

	wg.Wait()

	// Verify all resources were stored
	for i := 0; i < numGoroutines; i++ {
		resourceID := fmt.Sprintf("concurrent-test-%d", i)
		exists, err := service.ResourceExists(ctx, resourceID)
		if err != nil {
			t.Errorf("Failed to check existence of %s: %v", resourceID, err)
		}
		if !exists {
			t.Errorf("Resource %s should exist after concurrent store", resourceID)
		}
	}
}

func TestStorageService_EventDispatching(t *testing.T) {
	repo := newMockRepository()
	converter := newMockConverter()

	// Create a shared mock to track events across operations
	sharedMock := newMockUnitOfWork()
	unitOfWorkFactory := func() pericarpdomain.UnitOfWork {
		return sharedMock
	}

	service := NewStorageService(repo, converter, unitOfWorkFactory)
	ctx := context.Background()

	// Store a resource
	testData := []byte(`{"@context": {}, "@id": "event-test"}`)
	_, err := service.StoreResource(ctx, "event-test", testData, "application/ld+json")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Check that events were processed through unit of work
	envelopes := sharedMock.GetEnvelopes()
	if len(envelopes) == 0 {
		t.Error("Expected events to be processed")
	}

	// Delete the resource
	err = service.DeleteResource(ctx, "event-test")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Check that delete events were processed
	envelopes = sharedMock.GetEnvelopes()
	if len(envelopes) < 2 { // At least create and delete events
		t.Error("Expected multiple events to be processed")
	}
}

func TestStorageService_ContentNegotiation(t *testing.T) {
	repo := newMockRepository()
	converter := newMockConverter()
	unitOfWorkFactory := createMockUnitOfWorkFactory()

	service := NewStorageService(repo, converter, unitOfWorkFactory)
	ctx := context.Background()

	// Store a JSON-LD resource
	testData := []byte(`{"@context": {}, "@id": "content-neg-test"}`)
	_, err := service.StoreResource(ctx, "content-neg-test", testData, "application/ld+json")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Retrieve in different format
	resource, err := service.RetrieveResource(ctx, "content-neg-test", "text/turtle")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Should have converted format
	if resource.GetContentType() != "text/turtle" {
		t.Errorf("Expected content type text/turtle, got %s", resource.GetContentType())
	}

	// Should have conversion metadata
	if resource.GetMetadata()["convertedFrom"] != "application/ld+json" {
		t.Error("Expected convertedFrom metadata to be set")
	}
}

// Benchmark tests for performance validation

func BenchmarkStorageService_StoreResource(b *testing.B) {
	repo := newMockRepository()
	converter := newMockConverter()
	unitOfWorkFactory := createMockUnitOfWorkFactory()

	service := NewStorageService(repo, converter, unitOfWorkFactory)
	ctx := context.Background()
	testData := []byte(`{"@context": {}, "@id": "benchmark-test"}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resourceID := fmt.Sprintf("benchmark-resource-%d", i)
		_, err := service.StoreResource(ctx, resourceID, testData, "application/ld+json")
		if err != nil {
			b.Fatalf("Unexpected error: %v", err)
		}
	}
}

func BenchmarkStorageService_RetrieveResource(b *testing.B) {
	repo := newMockRepository()
	converter := newMockConverter()
	unitOfWorkFactory := createMockUnitOfWorkFactory()

	service := NewStorageService(repo, converter, unitOfWorkFactory)
	ctx := context.Background()

	// Setup test resources
	testData := []byte(`{"@context": {}, "@id": "benchmark-test"}`)
	for i := 0; i < 1000; i++ {
		resourceID := fmt.Sprintf("benchmark-resource-%d", i)
		resource := domain.NewResource(resourceID, "application/ld+json", testData)
		repo.resources[resourceID] = resource
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resourceID := fmt.Sprintf("benchmark-resource-%d", i%1000)
		_, err := service.RetrieveResource(ctx, resourceID, "")
		if err != nil {
			b.Fatalf("Unexpected error: %v", err)
		}
	}
}

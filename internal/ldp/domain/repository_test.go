package domain

import (
	"context"
	"testing"
)

// MockResourceRepository is a test implementation of ResourceRepository
type MockResourceRepository struct {
	resources map[string]*Resource
	errors    map[string]error // Map operation to error for testing error scenarios
}

func NewMockResourceRepository() *MockResourceRepository {
	return &MockResourceRepository{
		resources: make(map[string]*Resource),
		errors:    make(map[string]error),
	}
}

func (m *MockResourceRepository) Store(ctx context.Context, resource *Resource) error {
	if err, exists := m.errors["Store"]; exists {
		return err
	}
	m.resources[resource.ID()] = resource
	return nil
}

func (m *MockResourceRepository) Retrieve(ctx context.Context, id string) (*Resource, error) {
	if err, exists := m.errors["Retrieve"]; exists {
		return nil, err
	}
	resource, exists := m.resources[id]
	if !exists {
		return nil, &StorageError{
			Code:    ErrResourceNotFound.Code,
			Message: ErrResourceNotFound.Message,
		}
	}
	return resource, nil
}

func (m *MockResourceRepository) Delete(ctx context.Context, id string) error {
	if err, exists := m.errors["Delete"]; exists {
		return err
	}
	if _, exists := m.resources[id]; !exists {
		return &StorageError{
			Code:    ErrResourceNotFound.Code,
			Message: ErrResourceNotFound.Message,
		}
	}
	delete(m.resources, id)
	return nil
}

func (m *MockResourceRepository) Exists(ctx context.Context, id string) (bool, error) {
	if err, exists := m.errors["Exists"]; exists {
		return false, err
	}
	_, exists := m.resources[id]
	return exists, nil
}

// SetError configures the mock to return an error for a specific operation
func (m *MockResourceRepository) SetError(operation string, err error) {
	m.errors[operation] = err
}

// ClearError removes the error configuration for a specific operation
func (m *MockResourceRepository) ClearError(operation string) {
	delete(m.errors, operation)
}

// Reset clears all resources and errors
func (m *MockResourceRepository) Reset() {
	m.resources = make(map[string]*Resource)
	m.errors = make(map[string]error)
}

func TestResourceRepository_Interface(t *testing.T) {
	// Test that MockResourceRepository implements ResourceRepository interface
	var _ ResourceRepository = (*MockResourceRepository)(nil)
}

func TestMockResourceRepository_Store(t *testing.T) {
	repo := NewMockResourceRepository()
	ctx := context.Background()

	resource := NewResource("test-123", "text/turtle", []byte("test data"))

	err := repo.Store(ctx, resource)
	if err != nil {
		t.Errorf("Store() error = %v, want nil", err)
	}

	// Verify resource was stored
	stored, exists := repo.resources[resource.ID()]
	if !exists {
		t.Error("Resource was not stored")
	}
	if stored != resource {
		t.Error("Stored resource does not match original")
	}
}

func TestMockResourceRepository_Store_Error(t *testing.T) {
	repo := NewMockResourceRepository()
	ctx := context.Background()

	expectedErr := &StorageError{Code: "STORE_ERROR", Message: "store failed"}
	repo.SetError("Store", expectedErr)

	resource := NewResource("test-123", "text/turtle", []byte("test data"))

	err := repo.Store(ctx, resource)
	if err != expectedErr {
		t.Errorf("Store() error = %v, want %v", err, expectedErr)
	}
}

func TestMockResourceRepository_Retrieve(t *testing.T) {
	repo := NewMockResourceRepository()
	ctx := context.Background()

	resource := NewResource("test-123", "text/turtle", []byte("test data"))
	repo.resources[resource.ID()] = resource

	retrieved, err := repo.Retrieve(ctx, resource.ID())
	if err != nil {
		t.Errorf("Retrieve() error = %v, want nil", err)
	}
	if retrieved != resource {
		t.Error("Retrieved resource does not match stored resource")
	}
}

func TestMockResourceRepository_Retrieve_NotFound(t *testing.T) {
	repo := NewMockResourceRepository()
	ctx := context.Background()

	retrieved, err := repo.Retrieve(ctx, "nonexistent")
	if retrieved != nil {
		t.Error("Retrieve() should return nil for nonexistent resource")
	}
	if !IsResourceNotFound(err) {
		t.Errorf("Retrieve() should return ResourceNotFound error, got %v", err)
	}
}

func TestMockResourceRepository_Retrieve_Error(t *testing.T) {
	repo := NewMockResourceRepository()
	ctx := context.Background()

	expectedErr := &StorageError{Code: "RETRIEVE_ERROR", Message: "retrieve failed"}
	repo.SetError("Retrieve", expectedErr)

	retrieved, err := repo.Retrieve(ctx, "test-123")
	if retrieved != nil {
		t.Error("Retrieve() should return nil when error occurs")
	}
	if err != expectedErr {
		t.Errorf("Retrieve() error = %v, want %v", err, expectedErr)
	}
}

func TestMockResourceRepository_Delete(t *testing.T) {
	repo := NewMockResourceRepository()
	ctx := context.Background()

	resource := NewResource("test-123", "text/turtle", []byte("test data"))
	repo.resources[resource.ID()] = resource

	err := repo.Delete(ctx, resource.ID())
	if err != nil {
		t.Errorf("Delete() error = %v, want nil", err)
	}

	// Verify resource was deleted
	if _, exists := repo.resources[resource.ID()]; exists {
		t.Error("Resource was not deleted")
	}
}

func TestMockResourceRepository_Delete_NotFound(t *testing.T) {
	repo := NewMockResourceRepository()
	ctx := context.Background()

	err := repo.Delete(ctx, "nonexistent")
	if !IsResourceNotFound(err) {
		t.Errorf("Delete() should return ResourceNotFound error, got %v", err)
	}
}

func TestMockResourceRepository_Delete_Error(t *testing.T) {
	repo := NewMockResourceRepository()
	ctx := context.Background()

	expectedErr := &StorageError{Code: "DELETE_ERROR", Message: "delete failed"}
	repo.SetError("Delete", expectedErr)

	err := repo.Delete(ctx, "test-123")
	if err != expectedErr {
		t.Errorf("Delete() error = %v, want %v", err, expectedErr)
	}
}

func TestMockResourceRepository_Exists(t *testing.T) {
	repo := NewMockResourceRepository()
	ctx := context.Background()

	resource := NewResource("test-123", "text/turtle", []byte("test data"))
	repo.resources[resource.ID()] = resource

	exists, err := repo.Exists(ctx, resource.ID())
	if err != nil {
		t.Errorf("Exists() error = %v, want nil", err)
	}
	if !exists {
		t.Error("Exists() should return true for existing resource")
	}

	exists, err = repo.Exists(ctx, "nonexistent")
	if err != nil {
		t.Errorf("Exists() error = %v, want nil", err)
	}
	if exists {
		t.Error("Exists() should return false for nonexistent resource")
	}
}

func TestMockResourceRepository_Exists_Error(t *testing.T) {
	repo := NewMockResourceRepository()
	ctx := context.Background()

	expectedErr := &StorageError{Code: "EXISTS_ERROR", Message: "exists check failed"}
	repo.SetError("Exists", expectedErr)

	exists, err := repo.Exists(ctx, "test-123")
	if exists {
		t.Error("Exists() should return false when error occurs")
	}
	if err != expectedErr {
		t.Errorf("Exists() error = %v, want %v", err, expectedErr)
	}
}

func TestMockResourceRepository_ErrorManagement(t *testing.T) {
	repo := NewMockResourceRepository()

	// Test setting and clearing errors
	testErr := &StorageError{Code: "TEST_ERROR", Message: "test"}
	repo.SetError("Store", testErr)

	if repo.errors["Store"] != testErr {
		t.Error("SetError() did not set error correctly")
	}

	repo.ClearError("Store")
	if _, exists := repo.errors["Store"]; exists {
		t.Error("ClearError() did not clear error")
	}
}

func TestMockResourceRepository_Reset(t *testing.T) {
	repo := NewMockResourceRepository()
	ctx := context.Background()

	// Add some data and errors
	resource := NewResource("test-123", "text/turtle", []byte("test data"))
	repo.Store(ctx, resource)
	repo.SetError("Store", &StorageError{Code: "TEST", Message: "test"})

	// Reset
	repo.Reset()

	// Verify everything is cleared
	if len(repo.resources) != 0 {
		t.Error("Reset() did not clear resources")
	}
	if len(repo.errors) != 0 {
		t.Error("Reset() did not clear errors")
	}
}

func TestResourceRepository_ErrorScenarios(t *testing.T) {
	repo := NewMockResourceRepository()
	ctx := context.Background()

	tests := []struct {
		name      string
		operation string
		setupErr  error
		expectErr bool
	}{
		{
			name:      "insufficient storage on store",
			operation: "Store",
			setupErr:  &StorageError{Code: ErrInsufficientStorage.Code, Message: "no space"},
			expectErr: true,
		},
		{
			name:      "data corruption on retrieve",
			operation: "Retrieve",
			setupErr:  &StorageError{Code: ErrDataCorruption.Code, Message: "corrupted data"},
			expectErr: true,
		},
		{
			name:      "storage operation failed on delete",
			operation: "Delete",
			setupErr:  &StorageError{Code: ErrStorageOperation.Code, Message: "operation failed"},
			expectErr: true,
		},
		{
			name:      "storage operation failed on exists",
			operation: "Exists",
			setupErr:  &StorageError{Code: ErrStorageOperation.Code, Message: "operation failed"},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo.Reset()
			repo.SetError(tt.operation, tt.setupErr)

			resource := NewResource("test-123", "text/turtle", []byte("test data"))

			var err error
			switch tt.operation {
			case "Store":
				err = repo.Store(ctx, resource)
			case "Retrieve":
				_, err = repo.Retrieve(ctx, "test-123")
			case "Delete":
				err = repo.Delete(ctx, "test-123")
			case "Exists":
				_, err = repo.Exists(ctx, "test-123")
			}

			if tt.expectErr && err == nil {
				t.Error("Expected error but got nil")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("Expected no error but got %v", err)
			}
			if tt.expectErr && err != tt.setupErr {
				t.Errorf("Expected error %v but got %v", tt.setupErr, err)
			}
		})
	}
}

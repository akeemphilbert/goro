package application

import (
	"context"
	"fmt"
	"testing"

	"github.com/akeemphilbert/goro/internal/ldp/domain"
	"github.com/akeemphilbert/goro/internal/ldp/infrastructure"
	pericarpdomain "github.com/akeemphilbert/pericarp/pkg/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// TestPaginationValidation tests pagination parameter validation
func TestPaginationValidation(t *testing.T) {
	tests := []struct {
		name          string
		pagination    domain.PaginationOptions
		expectValid   bool
		expectDefault bool
	}{
		{
			name:          "Valid pagination",
			pagination:    domain.PaginationOptions{Limit: 50, Offset: 0},
			expectValid:   true,
			expectDefault: false,
		},
		{
			name:          "Valid pagination with offset",
			pagination:    domain.PaginationOptions{Limit: 25, Offset: 100},
			expectValid:   true,
			expectDefault: false,
		},
		{
			name:          "Invalid - zero limit",
			pagination:    domain.PaginationOptions{Limit: 0, Offset: 0},
			expectValid:   false,
			expectDefault: true,
		},
		{
			name:          "Invalid - negative limit",
			pagination:    domain.PaginationOptions{Limit: -1, Offset: 0},
			expectValid:   false,
			expectDefault: true,
		},
		{
			name:          "Invalid - negative offset",
			pagination:    domain.PaginationOptions{Limit: 50, Offset: -1},
			expectValid:   false,
			expectDefault: true,
		},
		{
			name:          "Invalid - limit too large",
			pagination:    domain.PaginationOptions{Limit: 1001, Offset: 0},
			expectValid:   false,
			expectDefault: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := tt.pagination.IsValid()
			assert.Equal(t, tt.expectValid, isValid)

			if tt.expectDefault {
				defaultPagination := domain.GetDefaultPagination()
				assert.Equal(t, 50, defaultPagination.Limit)
				assert.Equal(t, 0, defaultPagination.Offset)
			}
		})
	}
}

// TestContainerPaginationBasic tests basic pagination functionality
func TestContainerPaginationBasic(t *testing.T) {
	ctx := context.Background()
	containerRepo := &MockContainerRepository{}
	unitOfWorkFactory := func() pericarpdomain.UnitOfWork {
		return &MockUnitOfWork{}
	}
	rdfConverter := &infrastructure.ContainerRDFConverter{}
	service := NewContainerService(containerRepo, unitOfWorkFactory, rdfConverter)

	// Setup mocks for container creation
	containerID := "pagination-test-container"
	containerRepo.On("ContainerExists", ctx, containerID).Return(false, nil)
	containerRepo.On("ContainerExists", ctx, "").Return(false, nil)

	mockUoW := &MockUnitOfWork{}
	mockUoW.On("RegisterEvents", mock.Anything).Return()
	mockUoW.On("Commit", ctx).Return([]pericarpdomain.Envelope{}, nil)
	unitOfWorkFactory = func() pericarpdomain.UnitOfWork {
		return mockUoW
	}
	service = NewContainerService(containerRepo, unitOfWorkFactory, rdfConverter)

	// Create container
	_, err := service.CreateContainer(ctx, containerID, "", domain.BasicContainer)
	require.NoError(t, err)

	// Add 100 members
	memberCount := 100
	for i := 0; i < memberCount; i++ {
		memberID := fmt.Sprintf("member-%03d", i) // Zero-padded for consistent ordering
		err := service.AddResource(ctx, containerID, memberID, domain.NewResource(ctx, memberID, "text/plain", []byte("Hello, World!")))
		require.NoError(t, err)
	}

	// Test first page
	pagination := domain.PaginationOptions{Limit: 20, Offset: 0}
	listing, err := service.ListContainerMembers(ctx, containerID, pagination)
	require.NoError(t, err)
	assert.Equal(t, containerID, listing.ContainerID)
	assert.LessOrEqual(t, len(listing.Members), 20)
	assert.Equal(t, pagination, listing.Pagination)

	// Test second page
	pagination = domain.PaginationOptions{Limit: 20, Offset: 20}
	listing, err = service.ListContainerMembers(ctx, containerID, pagination)
	require.NoError(t, err)
	assert.Equal(t, containerID, listing.ContainerID)
	assert.LessOrEqual(t, len(listing.Members), 20)
	assert.Equal(t, pagination, listing.Pagination)

	// Test last page (partial)
	pagination = domain.PaginationOptions{Limit: 20, Offset: 80}
	listing, err = service.ListContainerMembers(ctx, containerID, pagination)
	require.NoError(t, err)
	assert.Equal(t, containerID, listing.ContainerID)
	assert.LessOrEqual(t, len(listing.Members), 20)
	assert.Equal(t, pagination, listing.Pagination)
}

// TestContainerPaginationEdgeCases tests pagination edge cases
func TestContainerPaginationEdgeCases(t *testing.T) {
	ctx := context.Background()
	containerRepo := &MockContainerRepository{}
	unitOfWorkFactory := func() pericarpdomain.UnitOfWork {
		return &MockUnitOfWork{}
	}
	rdfConverter := &infrastructure.ContainerRDFConverter{}
	service := NewContainerService(containerRepo, unitOfWorkFactory, rdfConverter)

	// Create empty container
	containerID := "empty-container"
	_, err := service.CreateContainer(ctx, containerID, "", domain.BasicContainer)
	require.NoError(t, err)

	// Test pagination on empty container
	pagination := domain.PaginationOptions{Limit: 20, Offset: 0}
	listing, err := service.ListContainerMembers(ctx, containerID, pagination)
	require.NoError(t, err)
	assert.Equal(t, containerID, listing.ContainerID)
	assert.Empty(t, listing.Members)

	// Add single member
	err = service.AddResource(ctx, containerID, "single-member", domain.NewResource(ctx, "single-member", "text/plain", []byte("Hello, World!")))
	require.NoError(t, err)

	// Test pagination with single member
	listing, err = service.ListContainerMembers(ctx, containerID, pagination)
	require.NoError(t, err)
	assert.Equal(t, containerID, listing.ContainerID)
	assert.Len(t, listing.Members, 1)
	assert.Equal(t, "single-member", listing.Members[0])

	// Test offset beyond available members
	pagination = domain.PaginationOptions{Limit: 20, Offset: 100}
	listing, err = service.ListContainerMembers(ctx, containerID, pagination)
	require.NoError(t, err)
	assert.Equal(t, containerID, listing.ContainerID)
	assert.Empty(t, listing.Members) // Should return empty list, not error
}

// TestContainerPaginationConsistency tests pagination consistency across multiple calls
func TestContainerPaginationConsistency(t *testing.T) {
	ctx := context.Background()
	containerRepo := &MockContainerRepository{}
	unitOfWorkFactory := func() pericarpdomain.UnitOfWork {
		return &MockUnitOfWork{}
	}
	rdfConverter := &infrastructure.ContainerRDFConverter{}
	service := NewContainerService(containerRepo, unitOfWorkFactory, rdfConverter)

	// Create container with members
	containerID := "consistency-test-container"
	_, err := service.CreateContainer(ctx, containerID, "", domain.BasicContainer)
	require.NoError(t, err)

	memberCount := 50
	expectedMembers := make([]string, memberCount)
	for i := 0; i < memberCount; i++ {
		memberID := fmt.Sprintf("member-%03d", i)
		expectedMembers[i] = memberID
		err := service.AddResource(ctx, containerID, memberID, domain.NewResource(ctx, memberID, "text/plain", []byte("Hello, World!")))
		require.NoError(t, err)
	}

	// Retrieve all members through pagination
	pageSize := 10
	var allRetrievedMembers []string

	for offset := 0; offset < memberCount; offset += pageSize {
		pagination := domain.PaginationOptions{Limit: pageSize, Offset: offset}
		listing, err := service.ListContainerMembers(ctx, containerID, pagination)
		require.NoError(t, err)

		allRetrievedMembers = append(allRetrievedMembers, listing.Members...)
	}

	// Verify we got all members
	assert.Equal(t, memberCount, len(allRetrievedMembers))

	// Verify no duplicates
	memberSet := make(map[string]bool)
	for _, member := range allRetrievedMembers {
		assert.False(t, memberSet[member], "Duplicate member found: %s", member)
		memberSet[member] = true
	}
}

// TestContainerPaginationWithInvalidOptions tests service behavior with invalid pagination
func TestContainerPaginationWithInvalidOptions(t *testing.T) {
	ctx := context.Background()
	containerRepo := &MockContainerRepository{}
	unitOfWorkFactory := func() pericarpdomain.UnitOfWork {
		return &MockUnitOfWork{}
	}
	rdfConverter := &infrastructure.ContainerRDFConverter{}
	service := NewContainerService(containerRepo, unitOfWorkFactory, rdfConverter)

	// Create container with members
	containerID := "invalid-pagination-test"
	_, err := service.CreateContainer(ctx, containerID, "", domain.BasicContainer)
	require.NoError(t, err)

	// Add some members
	for i := 0; i < 10; i++ {
		memberID := fmt.Sprintf("member-%d", i)
		err := service.AddResource(ctx, containerID, memberID, domain.NewResource(ctx, memberID, "text/plain", []byte("Hello, World!")))
		require.NoError(t, err)
	}

	// Test with invalid pagination - should use defaults
	invalidPagination := domain.PaginationOptions{Limit: 0, Offset: -1}
	listing, err := service.ListContainerMembers(ctx, containerID, invalidPagination)
	require.NoError(t, err)

	// Should use default pagination
	defaultPagination := domain.GetDefaultPagination()
	assert.Equal(t, defaultPagination, listing.Pagination)
	assert.LessOrEqual(t, len(listing.Members), defaultPagination.Limit)
}

// TestContainerPaginationBoundaryValues tests pagination with boundary values
func TestContainerPaginationBoundaryValues(t *testing.T) {
	ctx := context.Background()
	containerRepo := &MockContainerRepository{}
	unitOfWorkFactory := func() pericarpdomain.UnitOfWork {
		return &MockUnitOfWork{}
	}
	rdfConverter := &infrastructure.ContainerRDFConverter{}
	service := NewContainerService(containerRepo, unitOfWorkFactory, rdfConverter)

	// Create container with members
	containerID := "boundary-test-container"
	_, err := service.CreateContainer(ctx, containerID, "", domain.BasicContainer)
	require.NoError(t, err)

	memberCount := 100
	for i := 0; i < memberCount; i++ {
		memberID := fmt.Sprintf("member-%d", i)
		err := service.AddResource(ctx, containerID, memberID, domain.NewResource(ctx, memberID, "text/plain", []byte("Hello, World!")))
		require.NoError(t, err)
	}

	tests := []struct {
		name        string
		pagination  domain.PaginationOptions
		expectError bool
	}{
		{
			name:        "Minimum valid limit",
			pagination:  domain.PaginationOptions{Limit: 1, Offset: 0},
			expectError: false,
		},
		{
			name:        "Maximum valid limit",
			pagination:  domain.PaginationOptions{Limit: 1000, Offset: 0},
			expectError: false,
		},
		{
			name:        "Large offset",
			pagination:  domain.PaginationOptions{Limit: 10, Offset: 1000000},
			expectError: false, // Should return empty results, not error
		},
		{
			name:        "Exact member count limit",
			pagination:  domain.PaginationOptions{Limit: memberCount, Offset: 0},
			expectError: false,
		},
		{
			name:        "Offset at last member",
			pagination:  domain.PaginationOptions{Limit: 10, Offset: memberCount - 1},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			listing, err := service.ListContainerMembers(ctx, containerID, tt.pagination)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, listing)
				assert.Equal(t, containerID, listing.ContainerID)
			}
		})
	}
}

// TestContainerPaginationWithTotalCount tests pagination with total count information
func TestContainerPaginationWithTotalCount(t *testing.T) {
	ctx := context.Background()
	containerRepo := &MockContainerRepository{}
	unitOfWorkFactory := func() pericarpdomain.UnitOfWork {
		return &MockUnitOfWork{}
	}
	rdfConverter := &infrastructure.ContainerRDFConverter{}
	service := NewContainerService(containerRepo, unitOfWorkFactory, rdfConverter)

	// Create container with known number of members
	containerID := "total-count-test"
	_, err := service.CreateContainer(ctx, containerID, "", domain.BasicContainer)
	require.NoError(t, err)

	memberCount := 75
	for i := 0; i < memberCount; i++ {
		memberID := fmt.Sprintf("member-%d", i)
		err := service.AddResource(ctx, containerID, memberID, domain.NewResource(ctx, memberID, "text/plain", []byte("Hello, World!")))
		require.NoError(t, err)
	}

	// Test pagination with total count
	pagination := domain.PaginationOptions{Limit: 20, Offset: 0}
	listing, err := service.ListContainerMembers(ctx, containerID, pagination)
	require.NoError(t, err)

	// Verify pagination info
	assert.Equal(t, containerID, listing.ContainerID)
	assert.LessOrEqual(t, len(listing.Members), 20)
	assert.Equal(t, pagination, listing.Pagination)

	// Test different pages
	pages := []struct {
		offset             int
		expectedMaxMembers int
	}{
		{0, 20},  // First page
		{20, 20}, // Second page
		{40, 20}, // Third page
		{60, 15}, // Last page (partial)
		{80, 0},  // Beyond available members
	}

	for _, page := range pages {
		pagination := domain.PaginationOptions{Limit: 20, Offset: page.offset}
		listing, err := service.ListContainerMembers(ctx, containerID, pagination)
		require.NoError(t, err)
		assert.LessOrEqual(t, len(listing.Members), page.expectedMaxMembers)
	}
}

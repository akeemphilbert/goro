package domain

import (
	"context"
	"testing"
)

// TestContainerValidation_CircularReferenceDetection tests circular reference detection
func TestContainerValidation_CircularReferenceDetection(t *testing.T) {
	tests := []struct {
		name         string
		containerID  string
		parentID     string
		ancestorPath []string
		expectError  bool
		errorType    string
	}{
		{
			name:         "valid hierarchy - no circular reference",
			containerID:  "container-c",
			parentID:     "container-b",
			ancestorPath: []string{"container-a", "container-b"},
			expectError:  false,
		},
		{
			name:         "self reference - container as its own parent",
			containerID:  "container-a",
			parentID:     "container-a",
			ancestorPath: []string{},
			expectError:  true,
			errorType:    "self-reference",
		},
		{
			name:         "circular reference - container in ancestor path",
			containerID:  "container-c",
			parentID:     "container-a",
			ancestorPath: []string{"container-a", "container-b", "container-c"},
			expectError:  true,
			errorType:    "circular-reference",
		},
		{
			name:         "deep hierarchy - valid",
			containerID:  "container-e",
			parentID:     "container-d",
			ancestorPath: []string{"container-a", "container-b", "container-c", "container-d"},
			expectError:  false,
		},
		{
			name:         "circular reference at depth",
			containerID:  "container-b",
			parentID:     "container-e",
			ancestorPath: []string{"container-a", "container-b", "container-c", "container-d", "container-e"},
			expectError:  true,
			errorType:    "circular-reference",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			container := NewContainer(ctx, tt.containerID, tt.parentID, BasicContainer)

			err := container.ValidateHierarchy(tt.ancestorPath)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
					return
				}

				// Verify error type
				switch tt.errorType {
				case "self-reference":
					if !containsString(err.Error(), "cannot be its own parent") {
						t.Errorf("expected self-reference error, got: %v", err)
					}
				case "circular-reference":
					if !containsString(err.Error(), "circular reference detected") {
						t.Errorf("expected circular reference error, got: %v", err)
					}
				}
			} else {
				if err != nil {
					t.Errorf("expected no error but got: %v", err)
				}
			}
		})
	}
}

// TestContainerValidation_EmptinessValidation tests container emptiness validation
func TestContainerValidation_EmptinessValidation(t *testing.T) {
	tests := []struct {
		name        string
		members     []string
		expectEmpty bool
	}{
		{
			name:        "empty container",
			members:     []string{},
			expectEmpty: true,
		},
		{
			name:        "container with one member",
			members:     []string{"resource-1"},
			expectEmpty: false,
		},
		{
			name:        "container with multiple members",
			members:     []string{"resource-1", "resource-2", "container-child"},
			expectEmpty: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			container := NewContainer(ctx, "test-container", "", BasicContainer)

			// Add members
			for _, memberID := range tt.members {
				// Create a resource for each member
				resource := NewResource(ctx, memberID, "text/plain", []byte("test data"))
				err := container.AddMember(ctx, resource)
				if err != nil {
					t.Fatalf("failed to add member %s: %v", memberID, err)
				}
			}

			isEmpty := container.IsEmpty()
			if isEmpty != tt.expectEmpty {
				t.Errorf("expected IsEmpty() = %v, got %v", tt.expectEmpty, isEmpty)
			}

			// Test deletion - validation is now done at service layer
			container.Delete(ctx)
			// Note: Deletion validation is now handled at the service layer, not in the domain entity
		})
	}
}

// TestContainerValidation_MembershipConsistency tests membership consistency validation
func TestContainerValidation_MembershipConsistency(t *testing.T) {
	tests := []struct {
		name           string
		initialMembers []string
		addMember      string
		removeMember   string
		expectAddError bool
		expectRemError bool
	}{
		{
			name:           "add new member - success",
			initialMembers: []string{"resource-1"},
			addMember:      "resource-2",
			expectAddError: false,
		},
		{
			name:           "add duplicate member - error",
			initialMembers: []string{"resource-1", "resource-2"},
			addMember:      "resource-1",
			expectAddError: true,
		},
		{
			name:           "remove existing member - success",
			initialMembers: []string{"resource-1", "resource-2"},
			removeMember:   "resource-1",
			expectRemError: false,
		},
		{
			name:           "remove non-existent member - error",
			initialMembers: []string{"resource-1"},
			removeMember:   "resource-2",
			expectRemError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			container := NewContainer(ctx, "test-container", "", BasicContainer)

			// Add initial members
			for _, memberID := range tt.initialMembers {
				// Create a resource for each initial member
				resource := NewResource(ctx, memberID, "text/plain", []byte("test data"))
				err := container.AddMember(ctx, resource)
				if err != nil {
					t.Fatalf("failed to add initial member %s: %v", memberID, err)
				}
			}

			// Test adding member
			if tt.addMember != "" {
				// Create a resource for the new member
				resource := NewResource(ctx, tt.addMember, "text/plain", []byte("test data"))
				err := container.AddMember(ctx, resource)
				if tt.expectAddError {
					if err == nil {
						t.Errorf("expected error when adding duplicate member")
					}
					if !containsString(err.Error(), "member already exists") {
						t.Errorf("expected 'member already exists' error, got: %v", err)
					}
				} else {
					if err != nil {
						t.Errorf("expected no error when adding new member, got: %v", err)
					}
				}
			}

			// Test removing member
			if tt.removeMember != "" {
				err := container.RemoveMember(ctx, tt.removeMember)
				if tt.expectRemError {
					if err == nil {
						t.Errorf("expected error when removing non-existent member")
					}
					if !containsString(err.Error(), "member not found") {
						t.Errorf("expected 'member not found' error, got: %v", err)
					}
				} else {
					if err != nil {
						t.Errorf("expected no error when removing existing member, got: %v", err)
					}
				}
			}
		})
	}
}

// TestContainerValidation_ContainerTypeValidation tests container type validation
func TestContainerValidation_ContainerTypeValidation(t *testing.T) {
	tests := []struct {
		name          string
		containerType ContainerType
		expectValid   bool
	}{
		{
			name:          "valid BasicContainer type",
			containerType: BasicContainer,
			expectValid:   true,
		},
		{
			name:          "valid DirectContainer type",
			containerType: DirectContainer,
			expectValid:   true,
		},
		{
			name:          "invalid container type",
			containerType: ContainerType("InvalidContainer"),
			expectValid:   false,
		},
		{
			name:          "empty container type",
			containerType: ContainerType(""),
			expectValid:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			isValid := tt.containerType.IsValid()
			if isValid != tt.expectValid {
				t.Errorf("expected IsValid() = %v for type %s, got %v", tt.expectValid, tt.containerType, isValid)
			}

			// Test container creation with type validation
			container := NewContainer(ctx, "test-container", "", tt.containerType)
			if tt.expectValid {
				if container.ContainerType != tt.containerType {
					t.Errorf("expected container type %s, got %s", tt.containerType, container.ContainerType)
				}
			}
		})
	}
}

// TestContainerValidation_PaginationOptionsValidation tests pagination options validation
func TestContainerValidation_PaginationOptionsValidation(t *testing.T) {
	tests := []struct {
		name        string
		pagination  PaginationOptions
		expectValid bool
	}{
		{
			name:        "valid pagination - default",
			pagination:  GetDefaultPagination(),
			expectValid: true,
		},
		{
			name:        "valid pagination - custom",
			pagination:  PaginationOptions{Limit: 100, Offset: 50},
			expectValid: true,
		},
		{
			name:        "invalid pagination - zero limit",
			pagination:  PaginationOptions{Limit: 0, Offset: 0},
			expectValid: false,
		},
		{
			name:        "invalid pagination - negative limit",
			pagination:  PaginationOptions{Limit: -1, Offset: 0},
			expectValid: false,
		},
		{
			name:        "invalid pagination - negative offset",
			pagination:  PaginationOptions{Limit: 50, Offset: -1},
			expectValid: false,
		},
		{
			name:        "invalid pagination - limit too large",
			pagination:  PaginationOptions{Limit: 1001, Offset: 0},
			expectValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := tt.pagination.IsValid()
			if isValid != tt.expectValid {
				t.Errorf("expected IsValid() = %v for pagination %+v, got %v", tt.expectValid, tt.pagination, isValid)
			}
		})
	}
}

// TestContainerValidation_SortOptionsValidation tests sort options validation
func TestContainerValidation_SortOptionsValidation(t *testing.T) {
	tests := []struct {
		name        string
		sort        SortOptions
		expectValid bool
	}{
		{
			name:        "valid sort - default",
			sort:        GetDefaultSort(),
			expectValid: true,
		},
		{
			name:        "valid sort - name asc",
			sort:        SortOptions{Field: "name", Direction: "asc"},
			expectValid: true,
		},
		{
			name:        "valid sort - size desc",
			sort:        SortOptions{Field: "size", Direction: "desc"},
			expectValid: true,
		},
		{
			name:        "invalid sort - invalid field",
			sort:        SortOptions{Field: "invalid", Direction: "asc"},
			expectValid: false,
		},
		{
			name:        "invalid sort - invalid direction",
			sort:        SortOptions{Field: "name", Direction: "invalid"},
			expectValid: false,
		},
		{
			name:        "invalid sort - empty field",
			sort:        SortOptions{Field: "", Direction: "asc"},
			expectValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := tt.sort.IsValid()
			if isValid != tt.expectValid {
				t.Errorf("expected IsValid() = %v for sort %+v, got %v", tt.expectValid, tt.sort, isValid)
			}
		})
	}
}

// TestContainerValidation_ListingOptionsValidation tests listing options validation
func TestContainerValidation_ListingOptionsValidation(t *testing.T) {
	tests := []struct {
		name        string
		listing     ListingOptions
		expectValid bool
	}{
		{
			name:        "valid listing - default",
			listing:     GetDefaultListingOptions(),
			expectValid: true,
		},
		{
			name: "valid listing - custom",
			listing: ListingOptions{
				Pagination: PaginationOptions{Limit: 25, Offset: 10},
				Sort:       SortOptions{Field: "updatedAt", Direction: "desc"},
			},
			expectValid: true,
		},
		{
			name: "invalid listing - invalid pagination",
			listing: ListingOptions{
				Pagination: PaginationOptions{Limit: 0, Offset: 0},
				Sort:       GetDefaultSort(),
			},
			expectValid: false,
		},
		{
			name: "invalid listing - invalid sort",
			listing: ListingOptions{
				Pagination: GetDefaultPagination(),
				Sort:       SortOptions{Field: "invalid", Direction: "asc"},
			},
			expectValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := tt.listing.IsValid()
			if isValid != tt.expectValid {
				t.Errorf("expected IsValid() = %v for listing %+v, got %v", tt.expectValid, tt.listing, isValid)
			}
		})
	}
}

// Helper function to check if a string contains a substring
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || (len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			findSubstring(s, substr))))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

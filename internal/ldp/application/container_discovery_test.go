package application

import (
	"context"
	"testing"

	"github.com/akeemphilbert/goro/internal/ldp/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// TestContainerService_GenerateBreadcrumbs tests breadcrumb generation for container hierarchies
func TestContainerService_GenerateBreadcrumbs(t *testing.T) {
	tests := []struct {
		name           string
		containerID    string
		setupMocks     func(*MockContainerRepository)
		expectedCrumbs []BreadcrumbItem
		expectError    bool
	}{
		{
			name:        "root container",
			containerID: "root",
			setupMocks: func(mockRepo *MockContainerRepository) {
				container := domain.NewContainer("root", "", domain.BasicContainer)
				container.SetTitle("Root Container")
				mockRepo.On("GetContainer", mock.Anything, "root").Return(container, nil)
				mockRepo.On("GetPath", mock.Anything, "root").Return([]string{"root"}, nil)
			},
			expectedCrumbs: []BreadcrumbItem{
				{ID: "root", Title: "Root Container", Path: "/root"},
			},
			expectError: false,
		},
		{
			name:        "nested container",
			containerID: "child",
			setupMocks: func(mockRepo *MockContainerRepository) {
				rootContainer := domain.NewContainer("root", "", domain.BasicContainer)
				rootContainer.SetTitle("Root")
				childContainer := domain.NewContainer("child", "root", domain.BasicContainer)
				childContainer.SetTitle("Child")

				mockRepo.On("GetContainer", mock.Anything, "child").Return(childContainer, nil)
				mockRepo.On("GetPath", mock.Anything, "child").Return([]string{"root", "child"}, nil)
				mockRepo.On("GetContainer", mock.Anything, "root").Return(rootContainer, nil)
			},
			expectedCrumbs: []BreadcrumbItem{
				{ID: "root", Title: "Root", Path: "/root"},
				{ID: "child", Title: "Child", Path: "/root/child"},
			},
			expectError: false,
		},
		{
			name:        "deep hierarchy",
			containerID: "grandchild",
			setupMocks: func(mockRepo *MockContainerRepository) {
				rootContainer := domain.NewContainer("root", "", domain.BasicContainer)
				rootContainer.SetTitle("Root")
				childContainer := domain.NewContainer("child", "root", domain.BasicContainer)
				childContainer.SetTitle("Child")
				grandchildContainer := domain.NewContainer("grandchild", "child", domain.BasicContainer)
				grandchildContainer.SetTitle("Grandchild")

				mockRepo.On("GetContainer", mock.Anything, "grandchild").Return(grandchildContainer, nil)
				mockRepo.On("GetPath", mock.Anything, "grandchild").Return([]string{"root", "child", "grandchild"}, nil)
				mockRepo.On("GetContainer", mock.Anything, "root").Return(rootContainer, nil)
				mockRepo.On("GetContainer", mock.Anything, "child").Return(childContainer, nil)
			},
			expectedCrumbs: []BreadcrumbItem{
				{ID: "root", Title: "Root", Path: "/root"},
				{ID: "child", Title: "Child", Path: "/root/child"},
				{ID: "grandchild", Title: "Grandchild", Path: "/root/child/grandchild"},
			},
			expectError: false,
		},
		{
			name:        "container not found",
			containerID: "nonexistent",
			setupMocks: func(mockRepo *MockContainerRepository) {
				mockRepo.On("GetContainer", mock.Anything, "nonexistent").Return(nil, domain.ErrResourceNotFound)
			},
			expectedCrumbs: nil,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			mockRepo := new(MockContainerRepository)
			tt.setupMocks(mockRepo)

			service := NewContainerService(mockRepo, nil, nil)
			ctx := context.Background()

			// Act
			breadcrumbs, err := service.GenerateBreadcrumbs(ctx, tt.containerID)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, breadcrumbs)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedCrumbs, breadcrumbs)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

// TestContainerService_ResolveContainerPath tests path-based container resolution
func TestContainerService_ResolveContainerPath(t *testing.T) {
	tests := []struct {
		name           string
		path           string
		setupMocks     func(*MockContainerRepository)
		expectedResult *ContainerPathResolution
		expectError    bool
	}{
		{
			name: "root path",
			path: "/root",
			setupMocks: func(mockRepo *MockContainerRepository) {
				container := domain.NewContainer("root", "", domain.BasicContainer)
				container.SetTitle("Root Container")
				mockRepo.On("FindByPath", mock.Anything, "/root").Return(container, nil)
				// Mocks for breadcrumb generation
				mockRepo.On("GetContainer", mock.Anything, "root").Return(container, nil)
				mockRepo.On("GetPath", mock.Anything, "root").Return([]string{"root"}, nil)
			},
			expectedResult: &ContainerPathResolution{
				Container:   &ContainerInfo{ID: "root", Title: "Root Container", Type: "BasicContainer", ParentID: ""},
				Path:        "/root",
				Exists:      true,
				IsContainer: true,
				Breadcrumbs: []BreadcrumbItem{{ID: "root", Title: "Root Container", Path: "/root"}},
			},
			expectError: false,
		},
		{
			name: "nested path",
			path: "/root/documents",
			setupMocks: func(mockRepo *MockContainerRepository) {
				container := domain.NewContainer("documents", "root", domain.BasicContainer)
				container.SetTitle("Documents")
				mockRepo.On("FindByPath", mock.Anything, "/root/documents").Return(container, nil)
				mockRepo.On("GetPath", mock.Anything, "documents").Return([]string{"root", "documents"}, nil)

				rootContainer := domain.NewContainer("root", "", domain.BasicContainer)
				rootContainer.SetTitle("Root")
				mockRepo.On("GetContainer", mock.Anything, "root").Return(rootContainer, nil)
				mockRepo.On("GetContainer", mock.Anything, "documents").Return(container, nil)
			},
			expectedResult: &ContainerPathResolution{
				Container:   &ContainerInfo{ID: "documents", Title: "Documents", Type: "BasicContainer", ParentID: "root"},
				Path:        "/root/documents",
				Exists:      true,
				IsContainer: true,
				Breadcrumbs: []BreadcrumbItem{
					{ID: "root", Title: "Root", Path: "/root"},
					{ID: "documents", Title: "Documents", Path: "/root/documents"},
				},
			},
			expectError: false,
		},
		{
			name: "nonexistent path",
			path: "/nonexistent",
			setupMocks: func(mockRepo *MockContainerRepository) {
				mockRepo.On("FindByPath", mock.Anything, "/nonexistent").Return(nil, domain.ErrResourceNotFound)
			},
			expectedResult: &ContainerPathResolution{
				Container:   nil,
				Path:        "/nonexistent",
				Exists:      false,
				IsContainer: false,
				Breadcrumbs: nil,
			},
			expectError: false,
		},
		{
			name: "invalid path",
			path: "",
			setupMocks: func(mockRepo *MockContainerRepository) {
				// No mocks needed for validation error
			},
			expectedResult: nil,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			mockRepo := new(MockContainerRepository)
			tt.setupMocks(mockRepo)

			service := NewContainerService(mockRepo, nil, nil)
			ctx := context.Background()

			// Act
			result, err := service.ResolveContainerPath(ctx, tt.path)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult, result)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

// TestContainerService_GetContainerTypeInfo tests container type information exposure
func TestContainerService_GetContainerTypeInfo(t *testing.T) {
	tests := []struct {
		name         string
		containerID  string
		setupMocks   func(*MockContainerRepository)
		expectedInfo *ContainerTypeInfo
		expectError  bool
	}{
		{
			name:        "basic container with members",
			containerID: "container1",
			setupMocks: func(mockRepo *MockContainerRepository) {
				container := domain.NewContainer("container1", "", domain.BasicContainer)
				container.SetTitle("Test Container")
				container.SetDescription("A test container")
				container.AddMember("resource1")
				container.AddMember("resource2")

				mockRepo.On("GetContainer", mock.Anything, "container1").Return(container, nil)
				mockRepo.On("ListMembers", mock.Anything, "container1", mock.AnythingOfType("domain.PaginationOptions")).Return([]string{"resource1", "resource2"}, nil)
				mockRepo.On("GetChildren", mock.Anything, "container1").Return([]*domain.Container{}, nil)
			},
			expectedInfo: &ContainerTypeInfo{
				ID:            "container1",
				Type:          "BasicContainer",
				Title:         "Test Container",
				Description:   "A test container",
				MemberCount:   2,
				ChildCount:    0,
				IsEmpty:       false,
				AcceptedTypes: []string{"*/*"},
				Capabilities:  []string{"create", "read", "update", "delete", "list"},
			},
			expectError: false,
		},
		{
			name:        "empty container",
			containerID: "empty",
			setupMocks: func(mockRepo *MockContainerRepository) {
				container := domain.NewContainer("empty", "", domain.BasicContainer)
				container.SetTitle("Empty Container")

				mockRepo.On("GetContainer", mock.Anything, "empty").Return(container, nil)
				mockRepo.On("ListMembers", mock.Anything, "empty", mock.AnythingOfType("domain.PaginationOptions")).Return([]string{}, nil)
				mockRepo.On("GetChildren", mock.Anything, "empty").Return([]*domain.Container{}, nil)
			},
			expectedInfo: &ContainerTypeInfo{
				ID:            "empty",
				Type:          "BasicContainer",
				Title:         "Empty Container",
				Description:   "",
				MemberCount:   0,
				ChildCount:    0,
				IsEmpty:       true,
				AcceptedTypes: []string{"*/*"},
				Capabilities:  []string{"create", "read", "update", "delete", "list"},
			},
			expectError: false,
		},
		{
			name:        "container with children",
			containerID: "parent",
			setupMocks: func(mockRepo *MockContainerRepository) {
				container := domain.NewContainer("parent", "", domain.BasicContainer)
				container.SetTitle("Parent Container")

				child1 := domain.NewContainer("child1", "parent", domain.BasicContainer)
				child2 := domain.NewContainer("child2", "parent", domain.BasicContainer)

				mockRepo.On("GetContainer", mock.Anything, "parent").Return(container, nil)
				mockRepo.On("ListMembers", mock.Anything, "parent", mock.AnythingOfType("domain.PaginationOptions")).Return([]string{}, nil)
				mockRepo.On("GetChildren", mock.Anything, "parent").Return([]*domain.Container{child1, child2}, nil)
			},
			expectedInfo: &ContainerTypeInfo{
				ID:            "parent",
				Type:          "BasicContainer",
				Title:         "Parent Container",
				Description:   "",
				MemberCount:   0,
				ChildCount:    2,
				IsEmpty:       true,
				AcceptedTypes: []string{"*/*"},
				Capabilities:  []string{"create", "read", "update", "delete", "list"},
			},
			expectError: false,
		},
		{
			name:        "container not found",
			containerID: "nonexistent",
			setupMocks: func(mockRepo *MockContainerRepository) {
				mockRepo.On("GetContainer", mock.Anything, "nonexistent").Return(nil, domain.ErrResourceNotFound)
			},
			expectedInfo: nil,
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			mockRepo := new(MockContainerRepository)
			tt.setupMocks(mockRepo)

			service := NewContainerService(mockRepo, nil, nil)
			ctx := context.Background()

			// Act
			info, err := service.GetContainerTypeInfo(ctx, tt.containerID)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, info)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedInfo, info)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

// TestContainerService_GenerateStructureInfo tests machine-readable structure information generation
func TestContainerService_GenerateStructureInfo(t *testing.T) {
	tests := []struct {
		name         string
		containerID  string
		depth        int
		setupMocks   func(*MockContainerRepository)
		expectedInfo *ContainerStructureInfo
		expectError  bool
	}{
		{
			name:        "single level structure",
			containerID: "root",
			depth:       1,
			setupMocks: func(mockRepo *MockContainerRepository) {
				container := domain.NewContainer("root", "", domain.BasicContainer)
				container.SetTitle("Root Container")
				container.AddMember("resource1")

				child1 := domain.NewContainer("child1", "root", domain.BasicContainer)
				child1.SetTitle("Child 1")

				mockRepo.On("GetContainer", mock.Anything, "root").Return(container, nil)
				mockRepo.On("ListMembers", mock.Anything, "root", mock.AnythingOfType("domain.PaginationOptions")).Return([]string{"resource1"}, nil)
				mockRepo.On("GetChildren", mock.Anything, "root").Return([]*domain.Container{child1}, nil)
				// Mocks for child1 container
				mockRepo.On("GetContainer", mock.Anything, "child1").Return(child1, nil)
				mockRepo.On("ListMembers", mock.Anything, "child1", mock.AnythingOfType("domain.PaginationOptions")).Return([]string{}, nil)
				// Note: GetChildren for child1 is not called because depth limit is reached
			},
			expectedInfo: &ContainerStructureInfo{
				Container: ContainerInfo{
					ID:       "root",
					Title:    "Root Container",
					Type:     "BasicContainer",
					ParentID: "",
				},
				Members: []MemberInfo{
					{ID: "resource1", Type: "Resource"},
				},
				Children: []ContainerStructureInfo{
					{
						Container: ContainerInfo{
							ID:       "child1",
							Title:    "Child 1",
							Type:     "BasicContainer",
							ParentID: "root",
						},
						Members:  []MemberInfo{},
						Children: []ContainerStructureInfo{},
						Depth:    1,
					},
				},
				Depth: 0,
			},
			expectError: false,
		},
		{
			name:        "deep structure with depth limit",
			containerID: "root",
			depth:       2,
			setupMocks: func(mockRepo *MockContainerRepository) {
				container := domain.NewContainer("root", "", domain.BasicContainer)
				container.SetTitle("Root")

				child := domain.NewContainer("child", "root", domain.BasicContainer)
				child.SetTitle("Child")

				grandchild := domain.NewContainer("grandchild", "child", domain.BasicContainer)
				grandchild.SetTitle("Grandchild")

				mockRepo.On("GetContainer", mock.Anything, "root").Return(container, nil)
				mockRepo.On("ListMembers", mock.Anything, "root", mock.AnythingOfType("domain.PaginationOptions")).Return([]string{}, nil)
				mockRepo.On("GetChildren", mock.Anything, "root").Return([]*domain.Container{child}, nil)
				mockRepo.On("GetContainer", mock.Anything, "child").Return(child, nil)
				mockRepo.On("ListMembers", mock.Anything, "child", mock.AnythingOfType("domain.PaginationOptions")).Return([]string{}, nil)
				mockRepo.On("GetChildren", mock.Anything, "child").Return([]*domain.Container{grandchild}, nil)
				mockRepo.On("GetContainer", mock.Anything, "grandchild").Return(grandchild, nil)
				mockRepo.On("ListMembers", mock.Anything, "grandchild", mock.AnythingOfType("domain.PaginationOptions")).Return([]string{}, nil)
				// Note: GetChildren for grandchild is not called because depth limit is reached
			},
			expectedInfo: &ContainerStructureInfo{
				Container: ContainerInfo{
					ID:       "root",
					Title:    "Root",
					Type:     "BasicContainer",
					ParentID: "",
				},
				Members: []MemberInfo{},
				Children: []ContainerStructureInfo{
					{
						Container: ContainerInfo{
							ID:       "child",
							Title:    "Child",
							Type:     "BasicContainer",
							ParentID: "root",
						},
						Members: []MemberInfo{},
						Children: []ContainerStructureInfo{
							{
								Container: ContainerInfo{
									ID:       "grandchild",
									Title:    "Grandchild",
									Type:     "BasicContainer",
									ParentID: "child",
								},
								Members:  []MemberInfo{},
								Children: []ContainerStructureInfo{},
								Depth:    2,
							},
						},
						Depth: 1,
					},
				},
				Depth: 0,
			},
			expectError: false,
		},
		{
			name:        "container not found",
			containerID: "nonexistent",
			depth:       1,
			setupMocks: func(mockRepo *MockContainerRepository) {
				mockRepo.On("GetContainer", mock.Anything, "nonexistent").Return(nil, domain.ErrResourceNotFound)
			},
			expectedInfo: nil,
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			mockRepo := new(MockContainerRepository)
			tt.setupMocks(mockRepo)

			service := NewContainerService(mockRepo, nil, nil)
			ctx := context.Background()

			// Act
			info, err := service.GenerateStructureInfo(ctx, tt.containerID, tt.depth)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, info)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, info)
				assert.Equal(t, tt.expectedInfo.Container, info.Container)
				assert.Equal(t, tt.expectedInfo.Depth, info.Depth)
				assert.Len(t, info.Members, len(tt.expectedInfo.Members))
				assert.Len(t, info.Children, len(tt.expectedInfo.Children))
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

// TestContainerService_NavigationErrorHandling tests error handling with clear recovery messages
func TestContainerService_NavigationErrorHandling(t *testing.T) {
	tests := []struct {
		name          string
		operation     string
		setupMocks    func(*MockContainerRepository)
		expectedError string
		expectedCode  string
	}{
		{
			name:      "breadcrumbs - container not found",
			operation: "breadcrumbs",
			setupMocks: func(mockRepo *MockContainerRepository) {
				mockRepo.On("GetContainer", mock.Anything, "missing").Return(nil, domain.ErrResourceNotFound)
			},
			expectedError: "resource not found",
			expectedCode:  "RESOURCE_NOT_FOUND",
		},
		{
			name:      "path resolution - invalid path format",
			operation: "path_resolution",
			setupMocks: func(mockRepo *MockContainerRepository) {
				// No mocks needed for validation error
			},
			expectedError: "path cannot be empty",
			expectedCode:  "INVALID_ID",
		},
		{
			name:      "structure info - repository error",
			operation: "structure_info",
			setupMocks: func(mockRepo *MockContainerRepository) {
				mockRepo.On("GetContainer", mock.Anything, "error").Return(nil, domain.NewStorageError("DB_ERROR", "database connection failed"))
			},
			expectedError: "failed to retrieve container",
			expectedCode:  "STORAGE_OPERATION_FAILED",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			mockRepo := new(MockContainerRepository)
			tt.setupMocks(mockRepo)

			service := NewContainerService(mockRepo, nil, nil)
			ctx := context.Background()

			var err error

			// Act
			switch tt.operation {
			case "breadcrumbs":
				_, err = service.GenerateBreadcrumbs(ctx, "missing")
			case "path_resolution":
				_, err = service.ResolveContainerPath(ctx, "")
			case "structure_info":
				_, err = service.GenerateStructureInfo(ctx, "error", 1)
			}

			// Assert
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedError)

			// Check if it's a domain error with the expected code
			if domainErr, ok := err.(*domain.StorageError); ok {
				assert.Equal(t, tt.expectedCode, domainErr.Code)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

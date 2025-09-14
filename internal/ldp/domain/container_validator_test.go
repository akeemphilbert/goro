package domain

import (
	"context"
	"strings"
	"testing"
)

func TestContainerValidator_ValidateContainerID(t *testing.T) {
	validator := NewContainerValidator()

	tests := []struct {
		name        string
		id          string
		expectError bool
		errorCode   string
	}{
		{
			name:        "valid ID",
			id:          "valid-container-123",
			expectError: false,
		},
		{
			name:        "empty ID",
			id:          "",
			expectError: true,
			errorCode:   "EMPTY_CONTAINER_ID",
		},
		{
			name:        "ID too long",
			id:          strings.Repeat("a", 256),
			expectError: true,
			errorCode:   "CONTAINER_ID_TOO_LONG",
		},
		{
			name:        "ID with invalid characters",
			id:          "container/with/slashes",
			expectError: true,
			errorCode:   "INVALID_CONTAINER_ID_CHARS",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateContainerID(tt.id)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
					return
				}
				if storageErr, ok := GetStorageError(err); ok {
					if storageErr.Code != tt.errorCode {
						t.Errorf("expected error code %s, got %s", tt.errorCode, storageErr.Code)
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

func TestContainerValidator_ValidateMembers(t *testing.T) {
	validator := NewContainerValidator()

	tests := []struct {
		name        string
		members     []string
		expectError bool
		errorCode   string
	}{
		{
			name:        "valid members",
			members:     []string{"member1", "member2", "member3"},
			expectError: false,
		},
		{
			name:        "empty members list",
			members:     []string{},
			expectError: false,
		},
		{
			name:        "duplicate members",
			members:     []string{"member1", "member2", "member1"},
			expectError: true,
			errorCode:   "DUPLICATE_MEMBER",
		},
		{
			name:        "empty member ID",
			members:     []string{"member1", "", "member3"},
			expectError: true,
			errorCode:   "EMPTY_MEMBER_ID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateMembers(tt.members)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
					return
				}
				if storageErr, ok := GetStorageError(err); ok {
					if storageErr.Code != tt.errorCode {
						t.Errorf("expected error code %s, got %s", tt.errorCode, storageErr.Code)
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

func TestContainerValidator_ValidateContainer(t *testing.T) {
	validator := NewContainerValidator()
	ctx := context.Background()

	tests := []struct {
		name        string
		container   *Container
		expectError bool
		errorCode   string
	}{
		{
			name:        "nil container",
			container:   nil,
			expectError: true,
			errorCode:   "INVALID_CONTAINER",
		},
		{
			name:        "valid container",
			container:   NewContainer("valid-container", "", BasicContainer),
			expectError: false,
		},
		{
			name: "invalid container type",
			container: func() *Container {
				c := NewContainer("test-container", "", BasicContainer)
				c.ContainerType = ContainerType("InvalidType")
				return c
			}(),
			expectError: true,
			errorCode:   "INVALID_CONTAINER_TYPE",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateContainer(ctx, tt.container)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
					return
				}
				if storageErr, ok := GetStorageError(err); ok {
					if storageErr.Code != tt.errorCode {
						t.Errorf("expected error code %s, got %s", tt.errorCode, storageErr.Code)
					}
				} else {
					t.Errorf("expected StorageError, got %T", err)
				}
			} else {
				if err != nil {
					t.Errorf("expected no error but got: %v", err)
				}
			}
		})
	}
}

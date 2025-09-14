package application

import (
	"context"
	"fmt"

	"github.com/akeemphilbert/goro/internal/user/domain"
)

// FileStorage interface for user file operations
type FileStorage interface {
	WriteUserProfile(ctx context.Context, userID string, profile domain.UserProfile) error
	WriteWebIDDocument(ctx context.Context, userID, webID, document string) error
	DeleteUserFiles(ctx context.Context, userID string) error
	ReadUserProfile(ctx context.Context, userID string) (domain.UserProfile, error)
	ReadWebIDDocument(ctx context.Context, userID string) (string, error)
	UserExists(ctx context.Context, userID string) (bool, error)
}

// UserEventHandler handles user-related domain events for persistence operations
type UserEventHandler struct {
	userRepo    domain.UserWriteRepository
	fileStorage FileStorage
}

// NewUserEventHandler creates a new user event handler
func NewUserEventHandler(userRepo domain.UserWriteRepository, fileStorage FileStorage) *UserEventHandler {
	return &UserEventHandler{
		userRepo:    userRepo,
		fileStorage: fileStorage,
	}
}

// HandleUserRegistered handles user registration events by persisting to database and file storage
func (h *UserEventHandler) HandleUserRegistered(ctx context.Context, event *domain.UserRegisteredEventData) error {
	// First persist to database
	if err := h.userRepo.Create(ctx, event.User); err != nil {
		return fmt.Errorf("failed to persist user registration: %w", err)
	}

	// Then write user profile to file storage
	if err := h.fileStorage.WriteUserProfile(ctx, event.UserID, event.User.Profile); err != nil {
		return fmt.Errorf("failed to write user profile: %w", err)
	}

	return nil
}

// HandleUserProfileUpdated handles user profile update events by updating database and file storage
func (h *UserEventHandler) HandleUserProfileUpdated(ctx context.Context, event *domain.UserProfileUpdatedEventData) error {
	// First update database
	if err := h.userRepo.Update(ctx, event.User); err != nil {
		return fmt.Errorf("failed to update user profile: %w", err)
	}

	// Then update file storage
	if err := h.fileStorage.WriteUserProfile(ctx, event.UserID, event.NewProfile); err != nil {
		return fmt.Errorf("failed to write updated user profile: %w", err)
	}

	return nil
}

// HandleUserDeleted handles user deletion events by cleaning up files and removing from database
func (h *UserEventHandler) HandleUserDeleted(ctx context.Context, event *domain.UserDeletedEventData) error {
	// First cleanup user files
	if err := h.fileStorage.DeleteUserFiles(ctx, event.UserID); err != nil {
		return fmt.Errorf("failed to cleanup user files: %w", err)
	}

	// Then remove from database
	if err := h.userRepo.Delete(ctx, event.UserID); err != nil {
		return fmt.Errorf("failed to delete user from database: %w", err)
	}

	return nil
}

// HandleWebIDGenerated handles WebID generation events by writing WebID document to file storage
func (h *UserEventHandler) HandleWebIDGenerated(ctx context.Context, event *domain.WebIDGeneratedEventData) error {
	// Generate WebID document content (this would typically use a WebID generator)
	webIDDocument := h.generateWebIDDocument(event.User, event.WebID)

	// Write WebID document to file storage
	if err := h.fileStorage.WriteWebIDDocument(ctx, event.UserID, event.WebID, webIDDocument); err != nil {
		return fmt.Errorf("failed to write WebID document: %w", err)
	}

	return nil
}

// generateWebIDDocument generates a Turtle format WebID document
func (h *UserEventHandler) generateWebIDDocument(user *domain.User, webID string) string {
	// This is a simplified WebID document generation
	// In a real implementation, this would use a proper RDF library
	return fmt.Sprintf(`@prefix foaf: <http://xmlns.com/foaf/0.1/> .
@prefix solid: <http://www.w3.org/ns/solid/terms#> .

<%s> a foaf:Person ;
    foaf:name "%s" ;
    foaf:mbox <mailto:%s> .`,
		webID,
		user.Profile.Name,
		user.Email)
}

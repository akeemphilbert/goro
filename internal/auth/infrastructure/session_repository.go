package infrastructure

import (
	"context"
	"fmt"
	"time"

	"github.com/akeemphilbert/goro/internal/auth/domain"
	"gorm.io/gorm"
)

// GormSessionRepository implements SessionRepository using GORM
type GormSessionRepository struct {
	db *gorm.DB
}

// NewGormSessionRepository creates a new GORM session repository
func NewGormSessionRepository(db *gorm.DB) domain.SessionRepository {
	return &GormSessionRepository{db: db}
}

// Save creates or updates a session
func (r *GormSessionRepository) Save(ctx context.Context, session *domain.Session) error {
	if session == nil {
		return fmt.Errorf("session cannot be nil")
	}

	if !session.IsValid() {
		return fmt.Errorf("session is invalid")
	}

	model := &SessionModel{
		ID:           session.ID,
		UserID:       session.UserID,
		WebID:        session.WebID,
		AccountID:    session.AccountID,
		RoleID:       session.RoleID,
		TokenHash:    session.TokenHash,
		ExpiresAt:    session.ExpiresAt,
		CreatedAt:    session.CreatedAt,
		LastActivity: session.LastActivity,
	}

	err := r.db.WithContext(ctx).Save(model).Error
	if err != nil {
		return fmt.Errorf("failed to save session: %w", err)
	}

	return nil
}

// FindByID retrieves a session by its ID
func (r *GormSessionRepository) FindByID(ctx context.Context, id string) (*domain.Session, error) {
	if id == "" {
		return nil, fmt.Errorf("session ID cannot be empty")
	}

	var model SessionModel
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&model).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, domain.ErrSessionNotFound
		}
		return nil, fmt.Errorf("failed to find session by ID: %w", err)
	}

	return &domain.Session{
		ID:           model.ID,
		UserID:       model.UserID,
		WebID:        model.WebID,
		AccountID:    model.AccountID,
		RoleID:       model.RoleID,
		TokenHash:    model.TokenHash,
		ExpiresAt:    model.ExpiresAt,
		CreatedAt:    model.CreatedAt,
		LastActivity: model.LastActivity,
	}, nil
}

// FindByUserID retrieves all sessions for a user
func (r *GormSessionRepository) FindByUserID(ctx context.Context, userID string) ([]*domain.Session, error) {
	if userID == "" {
		return nil, fmt.Errorf("user ID cannot be empty")
	}

	var models []SessionModel
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&models).Error
	if err != nil {
		return nil, fmt.Errorf("failed to find sessions by user ID: %w", err)
	}

	sessions := make([]*domain.Session, len(models))
	for i, model := range models {
		sessions[i] = &domain.Session{
			ID:           model.ID,
			UserID:       model.UserID,
			WebID:        model.WebID,
			AccountID:    model.AccountID,
			RoleID:       model.RoleID,
			TokenHash:    model.TokenHash,
			ExpiresAt:    model.ExpiresAt,
			CreatedAt:    model.CreatedAt,
			LastActivity: model.LastActivity,
		}
	}

	return sessions, nil
}

// Delete removes a session by ID
func (r *GormSessionRepository) Delete(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("session ID cannot be empty")
	}

	result := r.db.WithContext(ctx).Where("id = ?", id).Delete(&SessionModel{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete session: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return domain.ErrSessionNotFound
	}

	return nil
}

// DeleteByUserID removes all sessions for a user
func (r *GormSessionRepository) DeleteByUserID(ctx context.Context, userID string) error {
	if userID == "" {
		return fmt.Errorf("user ID cannot be empty")
	}

	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Delete(&SessionModel{}).Error
	if err != nil {
		return fmt.Errorf("failed to delete sessions by user ID: %w", err)
	}

	return nil
}

// DeleteExpired removes all expired sessions
func (r *GormSessionRepository) DeleteExpired(ctx context.Context) error {
	now := time.Now()
	err := r.db.WithContext(ctx).Where("expires_at < ?", now).Delete(&SessionModel{}).Error
	if err != nil {
		return fmt.Errorf("failed to delete expired sessions: %w", err)
	}

	return nil
}

// UpdateActivity updates the last activity timestamp for a session
func (r *GormSessionRepository) UpdateActivity(ctx context.Context, sessionID string) error {
	if sessionID == "" {
		return fmt.Errorf("session ID cannot be empty")
	}

	now := time.Now()
	result := r.db.WithContext(ctx).Model(&SessionModel{}).
		Where("id = ?", sessionID).
		Update("last_activity", now)

	if result.Error != nil {
		return fmt.Errorf("failed to update session activity: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return domain.ErrSessionNotFound
	}

	return nil
}

// FindByAccountID retrieves all sessions for users in a specific account
func (r *GormSessionRepository) FindByAccountID(ctx context.Context, accountID string) ([]*domain.Session, error) {
	if accountID == "" {
		return nil, fmt.Errorf("account ID cannot be empty")
	}

	var models []SessionModel
	err := r.db.WithContext(ctx).Where("account_id = ?", accountID).Find(&models).Error
	if err != nil {
		return nil, fmt.Errorf("failed to find sessions by account ID: %w", err)
	}

	sessions := make([]*domain.Session, len(models))
	for i, model := range models {
		sessions[i] = &domain.Session{
			ID:           model.ID,
			UserID:       model.UserID,
			WebID:        model.WebID,
			AccountID:    model.AccountID,
			RoleID:       model.RoleID,
			TokenHash:    model.TokenHash,
			ExpiresAt:    model.ExpiresAt,
			CreatedAt:    model.CreatedAt,
			LastActivity: model.LastActivity,
		}
	}

	return sessions, nil
}

// FindByUserIDAndAccountID retrieves sessions for a user in a specific account
func (r *GormSessionRepository) FindByUserIDAndAccountID(ctx context.Context, userID, accountID string) ([]*domain.Session, error) {
	if userID == "" {
		return nil, fmt.Errorf("user ID cannot be empty")
	}
	if accountID == "" {
		return nil, fmt.Errorf("account ID cannot be empty")
	}

	var models []SessionModel
	err := r.db.WithContext(ctx).Where("user_id = ? AND account_id = ?", userID, accountID).Find(&models).Error
	if err != nil {
		return nil, fmt.Errorf("failed to find sessions by user ID and account ID: %w", err)
	}

	sessions := make([]*domain.Session, len(models))
	for i, model := range models {
		sessions[i] = &domain.Session{
			ID:           model.ID,
			UserID:       model.UserID,
			WebID:        model.WebID,
			AccountID:    model.AccountID,
			RoleID:       model.RoleID,
			TokenHash:    model.TokenHash,
			ExpiresAt:    model.ExpiresAt,
			CreatedAt:    model.CreatedAt,
			LastActivity: model.LastActivity,
		}
	}

	return sessions, nil
}

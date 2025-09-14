package infrastructure

import (
	"context"
	"fmt"
	"strings"

	"gorm.io/gorm"

	"github.com/akeemphilbert/goro/internal/user/domain"
	pericarpdomain "github.com/akeemphilbert/pericarp/pkg/domain"
)

// OptimizedGormUserRepository implements domain.UserRepository with performance optimizations
type OptimizedGormUserRepository struct {
	db    *gorm.DB
	cache Cache
}

// NewOptimizedGormUserRepository creates a new optimized GORM-based user repository
func NewOptimizedGormUserRepository(db *gorm.DB, cache Cache) domain.UserRepository {
	return &OptimizedGormUserRepository{
		db:    db,
		cache: cache,
	}
}

// GetByID retrieves a user by ID with caching and optimized queries
func (r *OptimizedGormUserRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("user ID cannot be empty")
	}

	// Try cache first
	if user, found := r.cache.GetUser(id); found {
		return user, nil
	}

	var userModel UserModel

	// Use optimized query with prepared statement and specific field selection
	err := r.db.WithContext(ctx).
		Select("id", "web_id", "email", "name", "status", "created_at", "updated_at").
		Where("id = ?", id).
		First(&userModel).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get user by ID %s: %w", id, err)
	}

	user, err := r.modelToDomain(&userModel)
	if err != nil {
		return nil, err
	}

	// Store in cache
	r.cache.SetUser(id, user)

	return user, nil
}

// GetByWebID retrieves a user by WebID with optimized indexing
func (r *OptimizedGormUserRepository) GetByWebID(ctx context.Context, webid string) (*domain.User, error) {
	if strings.TrimSpace(webid) == "" {
		return nil, fmt.Errorf("WebID cannot be empty")
	}

	var userModel UserModel

	// Use optimized query with index hint
	err := r.db.WithContext(ctx).
		Select("id", "web_id", "email", "name", "status", "created_at", "updated_at").
		Where("web_id = ?", webid).
		First(&userModel).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found with WebID: %s", webid)
		}
		return nil, fmt.Errorf("failed to get user by WebID %s: %w", webid, err)
	}

	user, err := r.modelToDomain(&userModel)
	if err != nil {
		return nil, err
	}

	// Store in cache by ID for future lookups
	r.cache.SetUser(user.ID(), user)

	return user, nil
}

// GetByEmail retrieves a user by email with optimized indexing
func (r *OptimizedGormUserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	if strings.TrimSpace(email) == "" {
		return nil, fmt.Errorf("email cannot be empty")
	}

	var userModel UserModel

	// Use optimized query with index hint
	err := r.db.WithContext(ctx).
		Select("id", "web_id", "email", "name", "status", "created_at", "updated_at").
		Where("email = ?", email).
		First(&userModel).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found with email: %s", email)
		}
		return nil, fmt.Errorf("failed to get user by email %s: %w", email, err)
	}

	user, err := r.modelToDomain(&userModel)
	if err != nil {
		return nil, err
	}

	// Store in cache by ID for future lookups
	r.cache.SetUser(user.ID(), user)

	return user, nil
}

// List retrieves users with filtering and optimized pagination
func (r *OptimizedGormUserRepository) List(ctx context.Context, filter domain.UserFilter) ([]*domain.User, error) {
	if err := filter.Validate(); err != nil {
		return nil, fmt.Errorf("invalid filter: %w", err)
	}

	query := r.db.WithContext(ctx).Model(&UserModel{}).
		Select("id", "web_id", "email", "name", "status", "created_at", "updated_at")

	// Apply filters with optimized indexing
	if filter.Status != "" {
		query = query.Where("status = ?", string(filter.Status))
	}

	if filter.EmailPattern != "" {
		// Use prefix matching for better index utilization
		query = query.Where("email LIKE ?", filter.EmailPattern+"%")
	}

	// Optimize ordering for pagination
	query = query.Order("created_at DESC, id ASC")

	// Apply pagination
	if filter.Limit > 0 {
		query = query.Limit(filter.Limit)
	}

	if filter.Offset > 0 {
		query = query.Offset(filter.Offset)
	}

	var userModels []UserModel
	err := query.Find(&userModels).Error
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	users := make([]*domain.User, len(userModels))
	for i, model := range userModels {
		user, err := r.modelToDomain(&model)
		if err != nil {
			return nil, fmt.Errorf("failed to convert user model to domain: %w", err)
		}
		users[i] = user

		// Cache users from list results
		r.cache.SetUser(user.ID(), user)
	}

	return users, nil
}

// Exists checks if a user exists by ID with caching
func (r *OptimizedGormUserRepository) Exists(ctx context.Context, id string) (bool, error) {
	if strings.TrimSpace(id) == "" {
		return false, fmt.Errorf("user ID cannot be empty")
	}

	// Try cache first
	if _, found := r.cache.GetUser(id); found {
		return true, nil
	}

	// Use optimized existence check
	var count int64
	err := r.db.WithContext(ctx).
		Model(&UserModel{}).
		Select("1").
		Where("id = ?", id).
		Limit(1).
		Count(&count).Error

	if err != nil {
		return false, fmt.Errorf("failed to check user existence for ID %s: %w", id, err)
	}

	return count > 0, nil
}

// modelToDomain converts a UserModel to a domain.User
func (r *OptimizedGormUserRepository) modelToDomain(model *UserModel) (*domain.User, error) {
	// Create user profile from name (simplified for now)
	profile := domain.UserProfile{
		Name:        model.Name,
		Bio:         "",
		Avatar:      "",
		Preferences: make(map[string]interface{}),
	}

	user := &domain.User{
		BasicEntity: pericarpdomain.NewEntity(model.ID),
		WebID:       model.WebID,
		Email:       model.Email,
		Profile:     profile,
		Status:      domain.UserStatus(model.Status),
		CreatedAt:   model.CreatedAt,
		UpdatedAt:   model.UpdatedAt,
	}

	return user, nil
}

// OptimizedGormAccountMemberRepository implements optimized membership queries
type OptimizedGormAccountMemberRepository struct {
	db *gorm.DB
}

// NewOptimizedGormAccountMemberRepository creates an optimized account member repository
func NewOptimizedGormAccountMemberRepository(db *gorm.DB) domain.AccountMemberRepository {
	return &OptimizedGormAccountMemberRepository{db: db}
}

// GetByID retrieves an account member by ID with optimized query
func (r *OptimizedGormAccountMemberRepository) GetByID(ctx context.Context, id string) (*domain.AccountMember, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("member ID cannot be empty")
	}

	var memberModel AccountMemberModel
	err := r.db.WithContext(ctx).
		Select("id", "account_id", "user_id", "role_id", "invited_by", "joined_at", "created_at", "updated_at").
		Where("id = ?", id).
		First(&memberModel).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("account member not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get account member by ID %s: %w", id, err)
	}

	return r.modelToDomain(&memberModel), nil
}

// GetByAccountAndUser retrieves membership with compound index optimization
func (r *OptimizedGormAccountMemberRepository) GetByAccountAndUser(ctx context.Context, accountID, userID string) (*domain.AccountMember, error) {
	if strings.TrimSpace(accountID) == "" {
		return nil, fmt.Errorf("account ID cannot be empty")
	}
	if strings.TrimSpace(userID) == "" {
		return nil, fmt.Errorf("user ID cannot be empty")
	}

	var memberModel AccountMemberModel

	// Use compound index for optimal performance
	err := r.db.WithContext(ctx).
		Select("id", "account_id", "user_id", "role_id", "invited_by", "joined_at", "created_at", "updated_at").
		Where("account_id = ? AND user_id = ?", accountID, userID).
		First(&memberModel).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("account member not found for account %s and user %s", accountID, userID)
		}
		return nil, fmt.Errorf("failed to get account member by account %s and user %s: %w", accountID, userID, err)
	}

	return r.modelToDomain(&memberModel), nil
}

// ListByAccount retrieves all members with optimized indexing and batching
func (r *OptimizedGormAccountMemberRepository) ListByAccount(ctx context.Context, accountID string) ([]*domain.AccountMember, error) {
	if strings.TrimSpace(accountID) == "" {
		return nil, fmt.Errorf("account ID cannot be empty")
	}

	var memberModels []AccountMemberModel

	// Use optimized query with proper indexing and ordering
	err := r.db.WithContext(ctx).
		Select("id", "account_id", "user_id", "role_id", "invited_by", "joined_at", "created_at", "updated_at").
		Where("account_id = ?", accountID).
		Order("joined_at ASC, id ASC").
		Find(&memberModels).Error

	if err != nil {
		return nil, fmt.Errorf("failed to list account members for account %s: %w", accountID, err)
	}

	members := make([]*domain.AccountMember, len(memberModels))
	for i, model := range memberModels {
		members[i] = r.modelToDomain(&model)
	}

	return members, nil
}

// ListByUser retrieves all memberships for a user with optimized indexing
func (r *OptimizedGormAccountMemberRepository) ListByUser(ctx context.Context, userID string) ([]*domain.AccountMember, error) {
	if strings.TrimSpace(userID) == "" {
		return nil, fmt.Errorf("user ID cannot be empty")
	}

	var memberModels []AccountMemberModel

	// Use optimized query with user index
	err := r.db.WithContext(ctx).
		Select("id", "account_id", "user_id", "role_id", "invited_by", "joined_at", "created_at", "updated_at").
		Where("user_id = ?", userID).
		Order("joined_at ASC, id ASC").
		Find(&memberModels).Error

	if err != nil {
		return nil, fmt.Errorf("failed to list account memberships for user %s: %w", userID, err)
	}

	members := make([]*domain.AccountMember, len(memberModels))
	for i, model := range memberModels {
		members[i] = r.modelToDomain(&model)
	}

	return members, nil
}

// modelToDomain converts an AccountMemberModel to a domain.AccountMember
func (r *OptimizedGormAccountMemberRepository) modelToDomain(model *AccountMemberModel) *domain.AccountMember {
	member := &domain.AccountMember{
		BasicEntity: pericarpdomain.NewEntity(model.ID),
		AccountID:   model.AccountID,
		UserID:      model.UserID,
		RoleID:      model.RoleID,
		InvitedBy:   model.InvitedBy,
		JoinedAt:    model.JoinedAt,
		CreatedAt:   model.CreatedAt,
		UpdatedAt:   model.UpdatedAt,
	}

	return member
}

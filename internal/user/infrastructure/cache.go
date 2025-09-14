package infrastructure

import (
	"context"
	"sync"
	"time"

	"github.com/akeemphilbert/goro/internal/user/domain"
)

// Cache interface for user management caching
type Cache interface {
	GetUser(id string) (*domain.User, bool)
	SetUser(id string, user *domain.User)
	DeleteUser(id string)
	GetRole(id string) (*domain.Role, bool)
	SetRole(id string, role *domain.Role)
	DeleteRole(id string)
	Clear()
}

// InMemoryCache provides a simple in-memory cache implementation
type InMemoryCache struct {
	users     map[string]*cacheEntry
	roles     map[string]*cacheEntry
	userMutex sync.RWMutex
	roleMutex sync.RWMutex
	ttl       time.Duration
}

type cacheEntry struct {
	data      interface{}
	timestamp time.Time
}

// NewInMemoryCache creates a new in-memory cache with TTL
func NewInMemoryCache(ttl time.Duration) *InMemoryCache {
	cache := &InMemoryCache{
		users: make(map[string]*cacheEntry),
		roles: make(map[string]*cacheEntry),
		ttl:   ttl,
	}

	// Start cleanup goroutine
	go cache.cleanup()

	return cache
}

// GetUser retrieves a user from cache
func (c *InMemoryCache) GetUser(id string) (*domain.User, bool) {
	c.userMutex.RLock()
	defer c.userMutex.RUnlock()

	entry, exists := c.users[id]
	if !exists {
		return nil, false
	}

	// Check if entry has expired
	if time.Since(entry.timestamp) > c.ttl {
		// Remove expired entry (will be cleaned up by cleanup goroutine)
		return nil, false
	}

	user, ok := entry.data.(*domain.User)
	return user, ok
}

// SetUser stores a user in cache
func (c *InMemoryCache) SetUser(id string, user *domain.User) {
	c.userMutex.Lock()
	defer c.userMutex.Unlock()

	c.users[id] = &cacheEntry{
		data:      user,
		timestamp: time.Now(),
	}
}

// DeleteUser removes a user from cache
func (c *InMemoryCache) DeleteUser(id string) {
	c.userMutex.Lock()
	defer c.userMutex.Unlock()

	delete(c.users, id)
}

// GetRole retrieves a role from cache
func (c *InMemoryCache) GetRole(id string) (*domain.Role, bool) {
	c.roleMutex.RLock()
	defer c.roleMutex.RUnlock()

	entry, exists := c.roles[id]
	if !exists {
		return nil, false
	}

	// Check if entry has expired
	if time.Since(entry.timestamp) > c.ttl {
		return nil, false
	}

	role, ok := entry.data.(*domain.Role)
	return role, ok
}

// SetRole stores a role in cache
func (c *InMemoryCache) SetRole(id string, role *domain.Role) {
	c.roleMutex.Lock()
	defer c.roleMutex.Unlock()

	c.roles[id] = &cacheEntry{
		data:      role,
		timestamp: time.Now(),
	}
}

// DeleteRole removes a role from cache
func (c *InMemoryCache) DeleteRole(id string) {
	c.roleMutex.Lock()
	defer c.roleMutex.Unlock()

	delete(c.roles, id)
}

// Clear removes all entries from cache
func (c *InMemoryCache) Clear() {
	c.userMutex.Lock()
	c.roleMutex.Lock()
	defer c.userMutex.Unlock()
	defer c.roleMutex.Unlock()

	c.users = make(map[string]*cacheEntry)
	c.roles = make(map[string]*cacheEntry)
}

// cleanup removes expired entries periodically
func (c *InMemoryCache) cleanup() {
	ticker := time.NewTicker(c.ttl / 2) // Cleanup twice per TTL period
	defer ticker.Stop()

	for range ticker.C {
		c.cleanupExpired()
	}
}

// cleanupExpired removes expired entries
func (c *InMemoryCache) cleanupExpired() {
	now := time.Now()

	// Cleanup users
	c.userMutex.Lock()
	for id, entry := range c.users {
		if now.Sub(entry.timestamp) > c.ttl {
			delete(c.users, id)
		}
	}
	c.userMutex.Unlock()

	// Cleanup roles
	c.roleMutex.Lock()
	for id, entry := range c.roles {
		if now.Sub(entry.timestamp) > c.ttl {
			delete(c.roles, id)
		}
	}
	c.roleMutex.Unlock()
}

// CachedUserRepository wraps a user repository with caching
type CachedUserRepository struct {
	repo  domain.UserRepository
	cache Cache
}

// NewCachedUserRepository creates a cached user repository
func NewCachedUserRepository(repo domain.UserRepository, cache Cache) domain.UserRepository {
	return &CachedUserRepository{
		repo:  repo,
		cache: cache,
	}
}

// GetByID retrieves a user by ID with caching
func (r *CachedUserRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	// Try cache first
	if user, found := r.cache.GetUser(id); found {
		return user, nil
	}

	// Cache miss, get from repository
	user, err := r.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Store in cache
	r.cache.SetUser(id, user)

	return user, nil
}

// GetByWebID retrieves a user by WebID (no caching for now, could be added)
func (r *CachedUserRepository) GetByWebID(ctx context.Context, webid string) (*domain.User, error) {
	return r.repo.GetByWebID(ctx, webid)
}

// GetByEmail retrieves a user by email (no caching for now, could be added)
func (r *CachedUserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	return r.repo.GetByEmail(ctx, email)
}

// List retrieves users with filtering (no caching for list operations)
func (r *CachedUserRepository) List(ctx context.Context, filter domain.UserFilter) ([]*domain.User, error) {
	return r.repo.List(ctx, filter)
}

// Exists checks if a user exists by ID
func (r *CachedUserRepository) Exists(ctx context.Context, id string) (bool, error) {
	// Try cache first
	if _, found := r.cache.GetUser(id); found {
		return true, nil
	}

	return r.repo.Exists(ctx, id)
}

// CachedRoleRepository wraps a role repository with caching
type CachedRoleRepository struct {
	repo  domain.RoleRepository
	cache Cache
}

// NewCachedRoleRepository creates a cached role repository
func NewCachedRoleRepository(repo domain.RoleRepository, cache Cache) domain.RoleRepository {
	return &CachedRoleRepository{
		repo:  repo,
		cache: cache,
	}
}

// GetByID retrieves a role by ID with caching
func (r *CachedRoleRepository) GetByID(ctx context.Context, id string) (*domain.Role, error) {
	// Try cache first
	if role, found := r.cache.GetRole(id); found {
		return role, nil
	}

	// Cache miss, get from repository
	role, err := r.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Store in cache
	r.cache.SetRole(id, role)

	return role, nil
}

// List retrieves all roles (cache system roles)
func (r *CachedRoleRepository) List(ctx context.Context) ([]*domain.Role, error) {
	return r.repo.List(ctx)
}

// GetSystemRoles retrieves system roles with caching
func (r *CachedRoleRepository) GetSystemRoles(ctx context.Context) ([]*domain.Role, error) {
	// System roles are frequently accessed, so we cache them
	systemRoleIDs := []string{"owner", "admin", "member", "viewer"}
	var roles []*domain.Role

	for _, roleID := range systemRoleIDs {
		role, err := r.GetByID(ctx, roleID)
		if err != nil {
			// If not in cache, fall back to repository
			return r.repo.GetSystemRoles(ctx)
		}
		roles = append(roles, role)
	}

	return roles, nil
}

package infrastructure_test

import (
	"context"
	"fmt"
	"github.com/akeemphilbert/goro/internal/user/infrastructure"
	"sync"
	"testing"
	"time"

	"github.com/akeemphilbert/goro/internal/user/domain"
	pericarpdomain "github.com/akeemphilbert/pericarp/pkg/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Performance targets that should initially fail without optimizations
const (
	// Database query performance targets
	MaxUserLookupTime       = 500 * time.Nanosecond // Single user lookup should be < 500ns (achievable with caching)
	MaxBulkUserQueryTime    = 10 * time.Millisecond // 1000 users query should be < 10ms (achievable with indexing)
	MaxMembershipQueryTime  = 1 * time.Millisecond  // Account membership query should be < 1ms (achievable with indexing)
	MaxConcurrentOperations = 100                   // Should handle 100 concurrent operations

	// Caching performance targets
	MaxCacheHitTime  = 250 * time.Nanosecond // Cache hits should be < 250ns (achievable with in-memory cache)
	MinCacheHitRatio = 0.8                   // Cache hit ratio should be > 80% (achievable with proper caching)

	// Load testing targets
	MaxConcurrentRegistrations = 50                     // Should handle 50 concurrent registrations
	MaxRegistrationTime        = 100 * time.Millisecond // Registration should be < 100ms
)

// TestDatabaseQueryPerformance tests database query performance that should initially fail
func TestDatabaseQueryPerformance(t *testing.T) {
	db := setupPerformanceTestDB(t)
	defer cleanupPerformanceTestDB(db)

	// Seed test data
	seedPerformanceTestData(t, db, 1000) // 1000 users for performance testing

	// Use optimized repository with caching
	cache := infrastructure.NewInMemoryCache(5 * time.Minute)
	userRepo := infrastructure.NewOptimizedGormUserRepository(db, cache)
	ctx := context.Background()

	t.Run("UserLookupPerformance", func(t *testing.T) {
		// Test single user lookup performance
		userID := "perf-user-500" // Middle user for realistic lookup

		start := time.Now()
		user, err := userRepo.GetByID(ctx, userID)
		elapsed := time.Since(start)

		require.NoError(t, err)
		require.NotNil(t, user)

		// This should initially fail without proper indexing
		assert.Less(t, elapsed, MaxUserLookupTime,
			"User lookup took %v, expected less than %v. Need database indexing optimization.",
			elapsed, MaxUserLookupTime)
	})

	t.Run("BulkUserQueryPerformance", func(t *testing.T) {
		// Test bulk user query performance
		filter := domain.UserFilter{
			Status: domain.UserStatusActive,
			Limit:  1000,
		}

		start := time.Now()
		users, err := userRepo.List(ctx, filter)
		elapsed := time.Since(start)

		require.NoError(t, err)
		require.Len(t, users, 1000)

		// This should initially fail without proper indexing and query optimization
		assert.Less(t, elapsed, MaxBulkUserQueryTime,
			"Bulk user query took %v, expected less than %v. Need query optimization.",
			elapsed, MaxBulkUserQueryTime)
	})

	t.Run("EmailLookupPerformance", func(t *testing.T) {
		// Test email-based lookup performance
		email := "perf-user-500@example.com"

		start := time.Now()
		user, err := userRepo.GetByEmail(ctx, email)
		elapsed := time.Since(start)

		require.NoError(t, err)
		require.NotNil(t, user)

		// This should initially fail without email indexing
		assert.Less(t, elapsed, MaxUserLookupTime,
			"Email lookup took %v, expected less than %v. Need email index optimization.",
			elapsed, MaxUserLookupTime)
	})

	t.Run("WebIDLookupPerformance", func(t *testing.T) {
		// Test WebID-based lookup performance
		webID := "https://example.com/profile/perf-user-500"

		start := time.Now()
		user, err := userRepo.GetByWebID(ctx, webID)
		elapsed := time.Since(start)

		require.NoError(t, err)
		require.NotNil(t, user)

		// This should initially fail without WebID indexing
		assert.Less(t, elapsed, MaxUserLookupTime,
			"WebID lookup took %v, expected less than %v. Need WebID index optimization.",
			elapsed, MaxUserLookupTime)
	})
}

// TestMembershipQueryPerformance tests membership query performance that should initially fail
func TestMembershipQueryPerformance(t *testing.T) {
	db := setupPerformanceTestDB(t)
	defer cleanupPerformanceTestDB(db)

	// Seed test data with accounts and memberships
	seedMembershipTestData(t, db, 100, 50) // 100 accounts, 50 members each

	// Use optimized membership repository
	memberRepo := infrastructure.NewOptimizedGormAccountMemberRepository(db)
	ctx := context.Background()

	t.Run("AccountMembershipQueryPerformance", func(t *testing.T) {
		// Test querying all members of a large account
		accountID := "perf-account-50" // Account with many members

		start := time.Now()
		members, err := memberRepo.ListByAccount(ctx, accountID)
		elapsed := time.Since(start)

		require.NoError(t, err)
		require.Len(t, members, 50)

		// This should initially fail without proper indexing on account_id
		assert.Less(t, elapsed, MaxMembershipQueryTime,
			"Account membership query took %v, expected less than %v. Need membership indexing.",
			elapsed, MaxMembershipQueryTime)
	})

	t.Run("UserMembershipQueryPerformance", func(t *testing.T) {
		// Test querying all accounts for a user with many memberships
		userID := "perf-user-25" // User with memberships in multiple accounts

		start := time.Now()
		memberships, err := memberRepo.ListByUser(ctx, userID)
		elapsed := time.Since(start)

		require.NoError(t, err)
		require.Greater(t, len(memberships), 10) // Should have multiple memberships

		// This should initially fail without proper indexing on user_id
		assert.Less(t, elapsed, MaxMembershipQueryTime,
			"User membership query took %v, expected less than %v. Need user membership indexing.",
			elapsed, MaxMembershipQueryTime)
	})

	t.Run("MembershipLookupPerformance", func(t *testing.T) {
		// Test specific account-user membership lookup
		accountID := "perf-account-25"
		userID := "perf-user-25"

		start := time.Now()
		member, err := memberRepo.GetByAccountAndUser(ctx, accountID, userID)
		elapsed := time.Since(start)

		require.NoError(t, err)
		require.NotNil(t, member)

		// This should initially fail without compound indexing
		maxCompoundLookupTime := 100 * time.Microsecond // More realistic for compound queries
		assert.Less(t, elapsed, maxCompoundLookupTime,
			"Membership lookup took %v, expected less than %v. Need compound index optimization.",
			elapsed, maxCompoundLookupTime)
	})
}

// TestCachingEffectiveness tests caching performance that should initially fail
func TestCachingEffectiveness(t *testing.T) {
	db := setupPerformanceTestDB(t)
	defer cleanupPerformanceTestDB(db)

	// Seed test data
	seedPerformanceTestData(t, db, 100)

	// This test will fail initially because no caching is implemented
	t.Run("UserCachePerformance", func(t *testing.T) {
		// Use optimized repository with caching
		cache := infrastructure.NewInMemoryCache(5 * time.Minute)
		userRepo := infrastructure.NewOptimizedGormUserRepository(db, cache)
		ctx := context.Background()
		userID := "perf-user-50"

		// First lookup (should hit database)
		start := time.Now()
		user1, err := userRepo.GetByID(ctx, userID)
		firstLookup := time.Since(start)
		require.NoError(t, err)
		require.NotNil(t, user1)

		// Second lookup (should hit cache if implemented)
		start = time.Now()
		user2, err := userRepo.GetByID(ctx, userID)
		secondLookup := time.Since(start)
		require.NoError(t, err)
		require.NotNil(t, user2)

		// This should initially fail because no caching is implemented
		assert.Less(t, secondLookup, MaxCacheHitTime,
			"Second lookup took %v, expected less than %v. Need caching implementation.",
			secondLookup, MaxCacheHitTime)

		// Cache should be significantly faster than database
		cacheSpeedup := float64(firstLookup) / float64(secondLookup)
		assert.Greater(t, cacheSpeedup, 5.0,
			"Cache speedup was %.2fx, expected at least 5x. Need effective caching.",
			cacheSpeedup)
	})

	t.Run("RoleCachePerformance", func(t *testing.T) {
		// Use cached role repository
		cache := infrastructure.NewInMemoryCache(5 * time.Minute)
		baseRoleRepo := infrastructure.NewGormRoleRepository(db)
		roleRepo := infrastructure.NewCachedRoleRepository(baseRoleRepo, cache)
		ctx := context.Background()

		// Test role caching (roles are frequently accessed)
		var hitTimes []time.Duration
		var missTimes []time.Duration

		// Perform multiple lookups to test cache effectiveness
		for i := 0; i < 10; i++ {
			roleID := "owner" // System role that should be cached

			start := time.Now()
			role, err := roleRepo.GetByID(ctx, roleID)
			elapsed := time.Since(start)

			require.NoError(t, err)
			require.NotNil(t, role)

			if i == 0 {
				missTimes = append(missTimes, elapsed) // First lookup is cache miss
			} else {
				hitTimes = append(hitTimes, elapsed) // Subsequent lookups should be cache hits
			}
		}

		// Calculate average cache hit time
		var totalHitTime time.Duration
		for _, hitTime := range hitTimes {
			totalHitTime += hitTime
		}
		avgHitTime := totalHitTime / time.Duration(len(hitTimes))

		// This should initially fail because no role caching is implemented
		assert.Less(t, avgHitTime, MaxCacheHitTime,
			"Average cache hit time was %v, expected less than %v. Need role caching.",
			avgHitTime, MaxCacheHitTime)
	})
}

// TestConcurrentOperationPerformance tests concurrent operation performance
func TestConcurrentOperationPerformance(t *testing.T) {
	db := setupPerformanceTestDB(t)
	defer cleanupPerformanceTestDB(db)

	// Seed initial data
	seedPerformanceTestData(t, db, 100)

	t.Run("ConcurrentUserLookups", func(t *testing.T) {
		// Use optimized repository with caching
		cache := infrastructure.NewInMemoryCache(5 * time.Minute)
		userRepo := infrastructure.NewOptimizedGormUserRepository(db, cache)
		ctx := context.Background()

		var wg sync.WaitGroup
		var mu sync.Mutex
		var errors []error
		var totalTime time.Duration

		start := time.Now()

		// Perform concurrent user lookups
		for i := 0; i < MaxConcurrentOperations; i++ {
			wg.Add(1)
			go func(userIndex int) {
				defer wg.Done()

				userID := fmt.Sprintf("perf-user-%d", userIndex%100) // Cycle through users
				opStart := time.Now()

				user, err := userRepo.GetByID(ctx, userID)
				opTime := time.Since(opStart)

				mu.Lock()
				if err != nil {
					errors = append(errors, err)
				} else if user == nil {
					errors = append(errors, fmt.Errorf("user not found: %s", userID))
				}
				totalTime += opTime
				mu.Unlock()
			}(i)
		}

		wg.Wait()
		elapsed := time.Since(start)

		// Check for errors
		require.Empty(t, errors, "Concurrent operations should not produce errors")

		// Calculate average operation time
		avgOpTime := totalTime / MaxConcurrentOperations

		// This should initially fail without proper connection pooling and indexing
		assert.Less(t, avgOpTime, MaxUserLookupTime*2, // Allow 2x normal time for concurrent operations
			"Average concurrent operation time was %v, expected less than %v. Need connection pooling optimization.",
			avgOpTime, MaxUserLookupTime*2)

		// Total time should be reasonable for concurrent operations
		maxTotalTime := MaxUserLookupTime * 10 // Should complete much faster than sequential
		assert.Less(t, elapsed, maxTotalTime,
			"Total concurrent operation time was %v, expected less than %v. Need concurrency optimization.",
			elapsed, maxTotalTime)
	})

	t.Run("ConcurrentMembershipQueries", func(t *testing.T) {
		// Seed membership data
		seedMembershipTestData(t, db, 50, 20)

		memberRepo := infrastructure.NewGormAccountMemberRepository(db)
		ctx := context.Background()

		var wg sync.WaitGroup
		var mu sync.Mutex
		var errors []error

		start := time.Now()

		// Perform concurrent membership queries
		for i := 0; i < MaxConcurrentOperations/2; i++ { // Fewer operations for membership queries
			wg.Add(1)
			go func(accountIndex int) {
				defer wg.Done()

				accountID := fmt.Sprintf("perf-account-%d", accountIndex%50)

				members, err := memberRepo.ListByAccount(ctx, accountID)

				mu.Lock()
				if err != nil {
					errors = append(errors, err)
				} else if len(members) == 0 {
					errors = append(errors, fmt.Errorf("no members found for account: %s", accountID))
				}
				mu.Unlock()
			}(i)
		}

		wg.Wait()
		elapsed := time.Since(start)

		// Check for errors
		require.Empty(t, errors, "Concurrent membership queries should not produce errors")

		// This should initially fail without proper indexing and connection handling
		maxConcurrentMembershipTime := MaxMembershipQueryTime * 5
		assert.Less(t, elapsed, maxConcurrentMembershipTime,
			"Concurrent membership queries took %v, expected less than %v. Need membership query optimization.",
			elapsed, maxConcurrentMembershipTime)
	})
}

// TestLoadAndScalabilityPerformance tests system performance under load
func TestLoadAndScalabilityPerformance(t *testing.T) {
	db := setupPerformanceTestDB(t)
	defer cleanupPerformanceTestDB(db)

	t.Run("ConcurrentUserRegistrations", func(t *testing.T) {
		userWriteRepo := infrastructure.NewGormUserWriteRepository(db)
		ctx := context.Background()

		var wg sync.WaitGroup
		var mu sync.Mutex
		var errors []error
		var registrationTimes []time.Duration

		// Perform concurrent user registrations
		for i := 0; i < MaxConcurrentRegistrations; i++ {
			wg.Add(1)
			go func(userIndex int) {
				defer wg.Done()

				start := time.Now()

				user := &domain.User{
					BasicEntity: pericarpdomain.NewEntity(fmt.Sprintf("load-user-%d", userIndex)),
					WebID:       fmt.Sprintf("https://example.com/profile/load-user-%d", userIndex),
					Email:       fmt.Sprintf("load-user-%d@example.com", userIndex),
					Profile: domain.UserProfile{
						Name: fmt.Sprintf("Load User %d", userIndex),
					},
					Status:    domain.UserStatusActive,
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				}

				err := userWriteRepo.Create(ctx, user)
				elapsed := time.Since(start)

				mu.Lock()
				if err != nil {
					errors = append(errors, err)
				}
				registrationTimes = append(registrationTimes, elapsed)
				mu.Unlock()
			}(i)
		}

		wg.Wait()

		// Check for errors
		require.Empty(t, errors, "Concurrent registrations should not produce errors")

		// Calculate average registration time
		var totalTime time.Duration
		for _, regTime := range registrationTimes {
			totalTime += regTime
		}
		avgRegistrationTime := totalTime / time.Duration(len(registrationTimes))

		// This should initially fail without proper database optimization
		assert.Less(t, avgRegistrationTime, MaxRegistrationTime,
			"Average registration time was %v, expected less than %v. Need registration optimization.",
			avgRegistrationTime, MaxRegistrationTime)
	})

	t.Run("LargeDatasetQueryPerformance", func(t *testing.T) {
		// Seed large dataset
		seedPerformanceTestData(t, db, 5000) // 5000 users for scalability testing

		// Use optimized repository with caching
		cache := infrastructure.NewInMemoryCache(5 * time.Minute)
		userRepo := infrastructure.NewOptimizedGormUserRepository(db, cache)
		ctx := context.Background()

		// Test pagination performance with large dataset
		filter := domain.UserFilter{
			Status: domain.UserStatusActive,
			Limit:  100,
			Offset: 2500, // Query from middle of dataset
		}

		start := time.Now()
		users, err := userRepo.List(ctx, filter)
		elapsed := time.Since(start)

		require.NoError(t, err)
		require.Len(t, users, 100)

		// This should initially fail without proper indexing and pagination optimization
		maxPaginationTime := MaxBulkUserQueryTime * 2 // Allow more time for offset queries
		assert.Less(t, elapsed, maxPaginationTime,
			"Large dataset pagination took %v, expected less than %v. Need pagination optimization.",
			elapsed, maxPaginationTime)
	})

	t.Run("ComplexMembershipScalability", func(t *testing.T) {
		// Seed complex membership data
		seedMembershipTestData(t, db, 200, 100) // 200 accounts, 100 members each

		memberRepo := infrastructure.NewGormAccountMemberRepository(db)
		ctx := context.Background()

		// Test querying memberships for accounts with many members
		accountID := "perf-account-100" // Large account

		start := time.Now()
		members, err := memberRepo.ListByAccount(ctx, accountID)
		elapsed := time.Since(start)

		require.NoError(t, err)
		require.Len(t, members, 100)

		// This should initially fail without proper indexing for large membership queries
		maxLargeMembershipTime := MaxMembershipQueryTime * 3
		assert.Less(t, elapsed, maxLargeMembershipTime,
			"Large membership query took %v, expected less than %v. Need scalability optimization.",
			elapsed, maxLargeMembershipTime)
	})
}

// Helper functions for test setup and data seeding

func setupPerformanceTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent), // Reduce noise in performance tests
	})
	require.NoError(t, err)

	// Migrate models
	err = infrastructure.MigrateUserModels(db)
	require.NoError(t, err)

	return db
}

func cleanupPerformanceTestDB(db *gorm.DB) {
	// Clean up is automatic with in-memory SQLite
}

func seedPerformanceTestData(t *testing.T, db *gorm.DB, userCount int) {
	// Seed system roles first
	roles := []infrastructure.RoleModel{
		{ID: "owner", Name: "Owner", Description: "Full access", Permissions: `[]`, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: "admin", Name: "Admin", Description: "Admin access", Permissions: `[]`, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: "member", Name: "Member", Description: "Member access", Permissions: `[]`, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: "viewer", Name: "Viewer", Description: "View access", Permissions: `[]`, CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}

	for _, role := range roles {
		err := db.Create(&role).Error
		require.NoError(t, err)
	}

	// Seed users
	users := make([]infrastructure.UserModel, userCount)
	for i := 0; i < userCount; i++ {
		users[i] = infrastructure.UserModel{
			ID:        fmt.Sprintf("perf-user-%d", i),
			WebID:     fmt.Sprintf("https://example.com/profile/perf-user-%d", i),
			Email:     fmt.Sprintf("perf-user-%d@example.com", i),
			Name:      fmt.Sprintf("Performance User %d", i),
			Status:    "active",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
	}

	// Batch insert users for better performance
	err := db.CreateInBatches(users, 100).Error
	require.NoError(t, err)
}

func seedMembershipTestData(t *testing.T, db *gorm.DB, accountCount, membersPerAccount int) {
	// First seed enough users for all memberships
	totalUsers := accountCount + membersPerAccount*2 // Extra users to ensure coverage
	seedPerformanceTestData(t, db, totalUsers)

	// Seed accounts
	accounts := make([]infrastructure.AccountModel, accountCount)
	for i := 0; i < accountCount; i++ {
		accounts[i] = infrastructure.AccountModel{
			ID:          fmt.Sprintf("perf-account-%d", i),
			OwnerID:     fmt.Sprintf("perf-user-%d", i%100), // Cycle through users as owners
			Name:        fmt.Sprintf("Performance Account %d", i),
			Description: fmt.Sprintf("Test account %d for performance testing", i),
			Settings:    `{}`,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
	}

	err := db.CreateInBatches(accounts, 50).Error
	require.NoError(t, err)

	// Seed account memberships
	var members []infrastructure.AccountMemberModel
	memberID := 0

	for accountIndex := 0; accountIndex < accountCount; accountIndex++ {
		for memberIndex := 0; memberIndex < membersPerAccount; memberIndex++ {
			// Create overlapping memberships so users belong to multiple accounts
			userIndex := memberIndex % totalUsers

			member := infrastructure.AccountMemberModel{
				ID:        fmt.Sprintf("perf-member-%d", memberID),
				AccountID: fmt.Sprintf("perf-account-%d", accountIndex),
				UserID:    fmt.Sprintf("perf-user-%d", userIndex),
				RoleID:    "member", // Most members have member role
				InvitedBy: fmt.Sprintf("perf-user-%d", accountIndex%100),
				JoinedAt:  time.Now(),
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}

			members = append(members, member)
			memberID++
		}
	}

	err = db.CreateInBatches(members, 100).Error
	require.NoError(t, err)
}

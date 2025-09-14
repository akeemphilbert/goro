package infrastructure

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/akeemphilbert/goro/internal/user/domain"
	pericarpdomain "github.com/akeemphilbert/pericarp/pkg/domain"
)

// Load testing performance targets that should initially fail
const (
	MaxThroughputUsers        = 1000                   // Should handle 1000 users/second
	MaxLatencyP95             = 100 * time.Millisecond // 95th percentile latency < 100ms
	MaxLatencyP99             = 200 * time.Millisecond // 99th percentile latency < 200ms
	MaxErrorRate              = 0.01                   // Error rate should be < 1%
	MaxMemoryGrowthRate       = 1.5                    // Memory growth should be < 1.5x
	MaxConnectionPoolSize     = 50                     // Should work with limited connections
	MaxConcurrentTransactions = 100                    // Should handle 100 concurrent transactions
)

// TestHighThroughputOperations tests system performance under high load
func TestHighThroughputOperations(t *testing.T) {
	db := setupLoadTestDB(t)
	defer cleanupPerformanceTestDB(db)

	t.Run("HighThroughputUserRegistration", func(t *testing.T) {
		userWriteRepo := NewGormUserWriteRepository(db)
		ctx := context.Background()

		var wg sync.WaitGroup
		var mu sync.Mutex
		var errors []error
		var latencies []time.Duration
		var successCount int64

		totalOperations := 1000
		startTime := time.Now()

		// Perform high-throughput user registrations
		for i := 0; i < totalOperations; i++ {
			wg.Add(1)
			go func(userIndex int) {
				defer wg.Done()

				opStart := time.Now()

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
				latency := time.Since(opStart)

				mu.Lock()
				if err != nil {
					errors = append(errors, err)
				} else {
					atomic.AddInt64(&successCount, 1)
				}
				latencies = append(latencies, latency)
				mu.Unlock()
			}(i)
		}

		wg.Wait()
		totalTime := time.Since(startTime)

		// Calculate throughput
		throughput := float64(successCount) / totalTime.Seconds()

		// This should initially fail without proper optimization
		assert.Greater(t, throughput, float64(MaxThroughputUsers),
			"Throughput was %.2f users/second, expected greater than %d. Need throughput optimization.",
			throughput, MaxThroughputUsers)

		// Calculate error rate
		errorRate := float64(len(errors)) / float64(totalOperations)
		assert.Less(t, errorRate, MaxErrorRate,
			"Error rate was %.4f, expected less than %.4f. Need error handling optimization.",
			errorRate, MaxErrorRate)

		// Calculate latency percentiles
		percentiles := calculatePercentiles(latencies, 0.95, 0.99)
		p95, p99 := percentiles[0], percentiles[1]

		// These should initially fail without latency optimization
		assert.Less(t, p95, MaxLatencyP95,
			"95th percentile latency was %v, expected less than %v. Need latency optimization.",
			p95, MaxLatencyP95)

		assert.Less(t, p99, MaxLatencyP99,
			"99th percentile latency was %v, expected less than %v. Need latency optimization.",
			p99, MaxLatencyP99)
	})

	t.Run("HighThroughputUserLookups", func(t *testing.T) {
		// Seed data for lookups
		seedPerformanceTestData(t, db, 1000)

		userRepo := NewGormUserRepository(db)
		ctx := context.Background()

		var wg sync.WaitGroup
		var mu sync.Mutex
		var errors []error
		var latencies []time.Duration
		var successCount int64

		totalLookups := 5000 // More lookups than users to test caching
		startTime := time.Now()

		// Perform high-throughput user lookups
		for i := 0; i < totalLookups; i++ {
			wg.Add(1)
			go func(lookupIndex int) {
				defer wg.Done()

				opStart := time.Now()

				// Cycle through users to create cache hits
				userID := fmt.Sprintf("perf-user-%d", lookupIndex%1000)

				user, err := userRepo.GetByID(ctx, userID)
				latency := time.Since(opStart)

				mu.Lock()
				if err != nil {
					errors = append(errors, err)
				} else if user == nil {
					errors = append(errors, fmt.Errorf("user not found: %s", userID))
				} else {
					atomic.AddInt64(&successCount, 1)
				}
				latencies = append(latencies, latency)
				mu.Unlock()
			}(i)
		}

		wg.Wait()
		totalTime := time.Since(startTime)

		// Calculate lookup throughput
		throughput := float64(successCount) / totalTime.Seconds()

		// This should initially fail without caching and indexing
		assert.Greater(t, throughput, float64(MaxThroughputUsers*5), // Lookups should be much faster
			"Lookup throughput was %.2f lookups/second, expected greater than %d. Need lookup optimization.",
			throughput, MaxThroughputUsers*5)

		// Calculate error rate
		errorRate := float64(len(errors)) / float64(totalLookups)
		assert.Less(t, errorRate, MaxErrorRate,
			"Lookup error rate was %.4f, expected less than %.4f. Need lookup reliability.",
			errorRate, MaxErrorRate)
	})
}

// TestScalabilityWithLargeDatasets tests performance with large amounts of data
func TestScalabilityWithLargeDatasets(t *testing.T) {
	db := setupLoadTestDB(t)
	defer cleanupPerformanceTestDB(db)

	t.Run("LargeUserDatasetPerformance", func(t *testing.T) {
		// Seed large dataset
		largeDatasetSize := 10000
		seedPerformanceTestData(t, db, largeDatasetSize)

		userRepo := NewGormUserRepository(db)
		ctx := context.Background()

		// Test query performance with large dataset
		filter := domain.UserFilter{
			Status: domain.UserStatusActive,
			Limit:  100,
			Offset: 5000, // Query from middle of large dataset
		}

		start := time.Now()
		users, err := userRepo.List(ctx, filter)
		elapsed := time.Since(start)

		require.NoError(t, err)
		require.Len(t, users, 100)

		// This should initially fail without proper indexing for large datasets
		maxLargeDatasetQueryTime := MaxBulkUserQueryTime * 3
		assert.Less(t, elapsed, maxLargeDatasetQueryTime,
			"Large dataset query took %v, expected less than %v. Need large dataset optimization.",
			elapsed, maxLargeDatasetQueryTime)

		// Test search performance
		searchFilter := domain.UserFilter{
			EmailPattern: "perf-user-5",
			Limit:        50,
		}

		start = time.Now()
		searchResults, err := userRepo.List(ctx, searchFilter)
		searchTime := time.Since(start)

		require.NoError(t, err)
		require.Greater(t, len(searchResults), 0)

		// This should initially fail without search indexing
		maxSearchTime := MaxBulkUserQueryTime * 2
		assert.Less(t, searchTime, maxSearchTime,
			"Search query took %v, expected less than %v. Need search optimization.",
			searchTime, maxSearchTime)
	})

	t.Run("LargeMembershipDatasetPerformance", func(t *testing.T) {
		// Seed large membership dataset
		seedMembershipTestData(t, db, 500, 200) // 500 accounts, 200 members each

		memberRepo := NewGormAccountMemberRepository(db)
		ctx := context.Background()

		// Test membership queries with large datasets
		accountID := "perf-account-250" // Account with many members

		start := time.Now()
		members, err := memberRepo.ListByAccount(ctx, accountID)
		elapsed := time.Since(start)

		require.NoError(t, err)
		require.Len(t, members, 200)

		// This should initially fail without membership indexing optimization
		maxLargeMembershipTime := MaxMembershipQueryTime * 5
		assert.Less(t, elapsed, maxLargeMembershipTime,
			"Large membership query took %v, expected less than %v. Need membership scalability optimization.",
			elapsed, maxLargeMembershipTime)

		// Test user membership queries across many accounts
		userID := "perf-user-100" // User with memberships in multiple accounts

		start = time.Now()
		userMemberships, err := memberRepo.ListByUser(ctx, userID)
		userMembershipTime := time.Since(start)

		require.NoError(t, err)
		require.Greater(t, len(userMemberships), 10) // Should have multiple memberships

		// This should initially fail without user membership indexing
		assert.Less(t, userMembershipTime, maxLargeMembershipTime,
			"User membership query took %v, expected less than %v. Need user membership optimization.",
			userMembershipTime, maxLargeMembershipTime)
	})
}

// TestConcurrentTransactionPerformance tests database transaction performance under load
func TestConcurrentTransactionPerformance(t *testing.T) {
	db := setupLoadTestDB(t)
	defer cleanupPerformanceTestDB(db)

	t.Run("ConcurrentUserTransactions", func(t *testing.T) {
		userWriteRepo := NewGormUserWriteRepository(db)
		userRepo := NewGormUserRepository(db)
		ctx := context.Background()

		// Seed initial users
		seedPerformanceTestData(t, db, 100)

		var wg sync.WaitGroup
		var mu sync.Mutex
		var errors []error
		var transactionTimes []time.Duration

		// Perform concurrent read/write transactions
		for i := 0; i < MaxConcurrentTransactions; i++ {
			wg.Add(1)
			go func(txIndex int) {
				defer wg.Done()

				start := time.Now()

				userID := fmt.Sprintf("perf-user-%d", txIndex%100)

				// Read user
				user, err := userRepo.GetByID(ctx, userID)
				if err != nil {
					mu.Lock()
					errors = append(errors, fmt.Errorf("read error: %w", err))
					mu.Unlock()
					return
				}

				// Update user
				user.Profile.Name = fmt.Sprintf("Updated User %d", txIndex)
				err = userWriteRepo.Update(ctx, user)

				elapsed := time.Since(start)

				mu.Lock()
				if err != nil {
					errors = append(errors, fmt.Errorf("write error: %w", err))
				}
				transactionTimes = append(transactionTimes, elapsed)
				mu.Unlock()
			}(i)
		}

		wg.Wait()

		// Check error rate
		errorRate := float64(len(errors)) / float64(MaxConcurrentTransactions)
		assert.Less(t, errorRate, MaxErrorRate,
			"Transaction error rate was %.4f, expected less than %.4f. Need transaction optimization.",
			errorRate, MaxErrorRate)

		// Calculate average transaction time
		var totalTime time.Duration
		for _, txTime := range transactionTimes {
			totalTime += txTime
		}
		avgTransactionTime := totalTime / time.Duration(len(transactionTimes))

		// This should initially fail without transaction optimization
		maxTransactionTime := MaxRegistrationTime * 2 // Allow more time for read+write
		assert.Less(t, avgTransactionTime, maxTransactionTime,
			"Average transaction time was %v, expected less than %v. Need transaction performance optimization.",
			avgTransactionTime, maxTransactionTime)
	})

	t.Run("DatabaseConnectionPoolPerformance", func(t *testing.T) {
		userRepo := NewGormUserRepository(db)
		ctx := context.Background()

		// Seed data
		seedPerformanceTestData(t, db, 100)

		var wg sync.WaitGroup
		var mu sync.Mutex
		var errors []error
		var connectionTimes []time.Duration

		// Test connection pool under high concurrency
		highConcurrency := MaxConnectionPoolSize * 2 // More requests than pool size

		for i := 0; i < highConcurrency; i++ {
			wg.Add(1)
			go func(connIndex int) {
				defer wg.Done()

				start := time.Now()

				userID := fmt.Sprintf("perf-user-%d", connIndex%100)
				user, err := userRepo.GetByID(ctx, userID)

				elapsed := time.Since(start)

				mu.Lock()
				if err != nil {
					errors = append(errors, err)
				} else if user == nil {
					errors = append(errors, fmt.Errorf("user not found: %s", userID))
				}
				connectionTimes = append(connectionTimes, elapsed)
				mu.Unlock()
			}(i)
		}

		wg.Wait()

		// Check connection pool error rate
		errorRate := float64(len(errors)) / float64(highConcurrency)
		assert.Less(t, errorRate, MaxErrorRate,
			"Connection pool error rate was %.4f, expected less than %.4f. Need connection pool optimization.",
			errorRate, MaxErrorRate)

		// Calculate connection acquisition time
		var totalTime time.Duration
		for _, connTime := range connectionTimes {
			totalTime += connTime
		}
		avgConnectionTime := totalTime / time.Duration(len(connectionTimes))

		// This should initially fail without proper connection pooling
		maxConnectionTime := MaxUserLookupTime * 3 // Allow overhead for connection acquisition
		assert.Less(t, avgConnectionTime, maxConnectionTime,
			"Average connection time was %v, expected less than %v. Need connection pool optimization.",
			avgConnectionTime, maxConnectionTime)
	})
}

// TestMemoryAndResourceUsage tests resource usage under load
func TestMemoryAndResourceUsage(t *testing.T) {
	db := setupLoadTestDB(t)
	defer cleanupPerformanceTestDB(db)

	t.Run("MemoryUsageUnderLoad", func(t *testing.T) {
		userRepo := NewGormUserRepository(db)
		ctx := context.Background()

		// Seed large dataset
		seedPerformanceTestData(t, db, 1000)

		// Simulate memory usage test (in real implementation, use runtime.MemStats)
		initialMemoryEstimate := 10 * 1024 * 1024 // 10MB baseline

		// Perform operations that might cause memory growth
		for i := 0; i < 1000; i++ {
			userID := fmt.Sprintf("perf-user-%d", i)
			_, err := userRepo.GetByID(ctx, userID)
			require.NoError(t, err)
		}

		// Estimate memory after operations
		finalMemoryEstimate := initialMemoryEstimate * 2 // Assume 2x growth for test

		memoryGrowthRatio := float64(finalMemoryEstimate) / float64(initialMemoryEstimate)

		// This should initially fail without memory optimization
		assert.Less(t, memoryGrowthRatio, MaxMemoryGrowthRate,
			"Memory growth ratio was %.2f, expected less than %.2f. Need memory optimization.",
			memoryGrowthRatio, MaxMemoryGrowthRate)
	})

	t.Run("ResourceCleanupPerformance", func(t *testing.T) {
		userWriteRepo := NewGormUserWriteRepository(db)
		ctx := context.Background()

		// Create and delete users to test resource cleanup
		var wg sync.WaitGroup
		var mu sync.Mutex
		var errors []error

		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func(userIndex int) {
				defer wg.Done()

				userID := fmt.Sprintf("cleanup-user-%d", userIndex)

				// Create user
				user := &domain.User{
					BasicEntity: pericarpdomain.NewEntity(userID),
					WebID:       fmt.Sprintf("https://example.com/profile/%s", userID),
					Email:       fmt.Sprintf("%s@example.com", userID),
					Profile: domain.UserProfile{
						Name: fmt.Sprintf("Cleanup User %d", userIndex),
					},
					Status:    domain.UserStatusActive,
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				}

				err := userWriteRepo.Create(ctx, user)
				if err != nil {
					mu.Lock()
					errors = append(errors, fmt.Errorf("create error: %w", err))
					mu.Unlock()
					return
				}

				// Delete user
				err = userWriteRepo.Delete(ctx, userID)
				if err != nil {
					mu.Lock()
					errors = append(errors, fmt.Errorf("delete error: %w", err))
					mu.Unlock()
				}
			}(i)
		}

		wg.Wait()

		// Check cleanup error rate
		errorRate := float64(len(errors)) / 200.0 // 100 creates + 100 deletes
		assert.Less(t, errorRate, MaxErrorRate,
			"Resource cleanup error rate was %.4f, expected less than %.4f. Need cleanup optimization.",
			errorRate, MaxErrorRate)
	})
}

// Helper functions for load testing

func setupLoadTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
		// Configure for better performance in tests
		PrepareStmt:                              true,
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	require.NoError(t, err)

	// Configure connection pool for load testing
	sqlDB, err := db.DB()
	require.NoError(t, err)

	sqlDB.SetMaxOpenConns(MaxConnectionPoolSize)
	sqlDB.SetMaxIdleConns(MaxConnectionPoolSize / 2)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// Migrate models
	err = MigrateUserModels(db)
	require.NoError(t, err)

	return db
}

func calculatePercentiles(latencies []time.Duration, percentiles ...float64) []time.Duration {
	if len(latencies) == 0 {
		return make([]time.Duration, len(percentiles))
	}

	// Simple percentile calculation (not optimized for performance)
	// In production, use a proper percentile calculation library

	// Sort latencies
	sorted := make([]time.Duration, len(latencies))
	copy(sorted, latencies)

	// Simple bubble sort for test purposes
	for i := 0; i < len(sorted); i++ {
		for j := 0; j < len(sorted)-1-i; j++ {
			if sorted[j] > sorted[j+1] {
				sorted[j], sorted[j+1] = sorted[j+1], sorted[j]
			}
		}
	}

	results := make([]time.Duration, len(percentiles))
	for i, p := range percentiles {
		index := int(float64(len(sorted)-1) * p)
		if index >= len(sorted) {
			index = len(sorted) - 1
		}
		results[i] = sorted[index]
	}

	return results
}

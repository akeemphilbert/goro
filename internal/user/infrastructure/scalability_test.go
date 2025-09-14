package infrastructure_test

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/akeemphilbert/goro/internal/user/domain"
	"github.com/akeemphilbert/goro/internal/user/infrastructure"
	pericarpdomain "github.com/akeemphilbert/pericarp/pkg/domain"
)

// Scalability testing targets
const (
	// Load testing targets
	MinThroughputRegistrations = 100                    // Should handle 100 registrations/second
	MinThroughputLookups       = 1000                   // Should handle 1000 lookups/second
	MaxP95Latency              = 50 * time.Millisecond  // 95th percentile < 50ms
	MaxP99Latency              = 100 * time.Millisecond // 99th percentile < 100ms
	ScalabilityMaxErrorRate    = 0.01                   // Error rate < 1%

	// Scalability targets
	MaxLargeDatasetQueryTime = 100 * time.Millisecond // Large dataset queries < 100ms
	MaxMemoryGrowthFactor    = 2.0                    // Memory growth < 2x baseline
	MinConcurrentUsers       = 200                    // Should handle 200 concurrent users
	MaxDatabaseConnections   = 25                     // Should work with limited connections
)

// TestUserRegistrationThroughput tests user registration performance under load
func TestUserRegistrationThroughput(t *testing.T) {
	db := setupScalabilityDB(t)
	defer cleanupPerformanceTestDB(db)

	userWriteRepo := infrastructure.NewGormUserWriteRepository(db)

	t.Run("HighThroughputRegistrations", func(t *testing.T) {
		var wg sync.WaitGroup
		var mu sync.Mutex
		var errors []error
		var latencies []time.Duration
		var successCount int64

		registrationCount := 500
		startTime := time.Now()

		// Perform concurrent registrations
		for i := 0; i < registrationCount; i++ {
			wg.Add(1)
			go func(userIndex int) {
				defer wg.Done()

				ctx := context.Background()
				opStart := time.Now()

				user := &domain.User{
					BasicEntity: pericarpdomain.NewEntity(fmt.Sprintf("scale-user-%d", userIndex)),
					WebID:       fmt.Sprintf("https://example.com/profile/scale-user-%d", userIndex),
					Email:       fmt.Sprintf("scale-user-%d@example.com", userIndex),
					Profile: domain.UserProfile{
						Name: fmt.Sprintf("Scale User %d", userIndex),
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

		assert.Greater(t, throughput, float64(MinThroughputRegistrations),
			"Registration throughput was %.2f/sec, expected > %d/sec",
			throughput, MinThroughputRegistrations)

		// Calculate error rate
		errorRate := float64(len(errors)) / float64(registrationCount)
		assert.Less(t, errorRate, ScalabilityMaxErrorRate,
			"Registration error rate was %.4f, expected < %.4f",
			errorRate, ScalabilityMaxErrorRate)

		// Calculate latency percentiles
		percentiles := calculatePercentiles(latencies, 0.95, 0.99)
		p95, p99 := percentiles[0], percentiles[1]

		assert.Less(t, p95, MaxP95Latency,
			"95th percentile registration latency was %v, expected < %v",
			p95, MaxP95Latency)

		assert.Less(t, p99, MaxP99Latency,
			"99th percentile registration latency was %v, expected < %v",
			p99, MaxP99Latency)

		t.Logf("Registration Performance: %.2f/sec throughput, %.2f%% error rate, P95: %v, P99: %v",
			throughput, errorRate*100, p95, p99)
	})
}

// TestUserLookupThroughput tests user lookup performance under load
func TestUserLookupThroughput(t *testing.T) {
	db := setupScalabilityDB(t)
	defer cleanupPerformanceTestDB(db)

	// Seed data for lookups
	seedScalabilityTestData(t, db, 1000)

	cache := infrastructure.NewInMemoryCache(5 * time.Minute)
	userRepo := infrastructure.NewOptimizedGormUserRepository(db, cache)

	t.Run("HighThroughputLookups", func(t *testing.T) {
		var wg sync.WaitGroup
		var mu sync.Mutex
		var errors []error
		var latencies []time.Duration
		var successCount int64

		lookupCount := 2000 // More lookups than users to test caching
		startTime := time.Now()

		// Perform concurrent lookups
		for i := 0; i < lookupCount; i++ {
			wg.Add(1)
			go func(lookupIndex int) {
				defer wg.Done()

				ctx := context.Background()
				opStart := time.Now()

				// Cycle through users to create cache hits
				userID := fmt.Sprintf("scale-user-%d", lookupIndex%1000)

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

		// Calculate throughput
		throughput := float64(successCount) / totalTime.Seconds()

		assert.Greater(t, throughput, float64(MinThroughputLookups),
			"Lookup throughput was %.2f/sec, expected > %d/sec",
			throughput, MinThroughputLookups)

		// Calculate error rate
		errorRate := float64(len(errors)) / float64(lookupCount)
		assert.Less(t, errorRate, ScalabilityMaxErrorRate,
			"Lookup error rate was %.4f, expected < %.4f",
			errorRate, ScalabilityMaxErrorRate)

		// Calculate latency percentiles
		percentiles := calculatePercentiles(latencies, 0.95, 0.99)
		p95, p99 := percentiles[0], percentiles[1]

		assert.Less(t, p95, MaxP95Latency,
			"95th percentile lookup latency was %v, expected < %v",
			p95, MaxP95Latency)

		assert.Less(t, p99, MaxP99Latency,
			"99th percentile lookup latency was %v, expected < %v",
			p99, MaxP99Latency)

		t.Logf("Lookup Performance: %.2f/sec throughput, %.2f%% error rate, P95: %v, P99: %v",
			throughput, errorRate*100, p95, p99)
	})
}

// TestLargeDatasetScalability tests performance with large datasets
func TestLargeDatasetScalability(t *testing.T) {
	db := setupScalabilityDB(t)
	defer cleanupPerformanceTestDB(db)

	t.Run("LargeUserDataset", func(t *testing.T) {
		// Test with increasingly large datasets
		dataSizes := []int{1000, 5000, 10000, 25000}

		cache := infrastructure.NewInMemoryCache(5 * time.Minute)
		userRepo := infrastructure.NewOptimizedGormUserRepository(db, cache)
		ctx := context.Background()

		for _, size := range dataSizes {
			t.Run(fmt.Sprintf("Dataset_%d_users", size), func(t *testing.T) {
				// Clear and seed data
				clearTestData(t, db)
				seedScalabilityTestData(t, db, size)

				// Test pagination performance
				filter := domain.UserFilter{
					Status: domain.UserStatusActive,
					Limit:  100,
					Offset: size / 2, // Query from middle of dataset
				}

				start := time.Now()
				users, err := userRepo.List(ctx, filter)
				elapsed := time.Since(start)

				require.NoError(t, err)
				require.Len(t, users, 100)

				assert.Less(t, elapsed, MaxLargeDatasetQueryTime,
					"Dataset size %d: query took %v, expected < %v",
					size, elapsed, MaxLargeDatasetQueryTime)

				// Test search performance
				searchFilter := domain.UserFilter{
					EmailPattern: fmt.Sprintf("scale-user-%d", size/4),
					Limit:        10,
				}

				start = time.Now()
				searchResults, err := userRepo.List(ctx, searchFilter)
				searchTime := time.Since(start)

				require.NoError(t, err)
				require.Greater(t, len(searchResults), 0)

				assert.Less(t, searchTime, MaxLargeDatasetQueryTime,
					"Dataset size %d: search took %v, expected < %v",
					size, searchTime, MaxLargeDatasetQueryTime)

				t.Logf("Dataset %d: pagination %v, search %v", size, elapsed, searchTime)
			})
		}
	})

	t.Run("LargeMembershipDataset", func(t *testing.T) {
		// Test membership queries with large datasets
		membershipSizes := []struct {
			accounts          int
			membersPerAccount int
		}{
			{100, 50},
			{200, 100},
			{500, 200},
		}

		memberRepo := infrastructure.NewOptimizedGormAccountMemberRepository(db)
		ctx := context.Background()

		for _, size := range membershipSizes {
			t.Run(fmt.Sprintf("Accounts_%d_Members_%d", size.accounts, size.membersPerAccount), func(t *testing.T) {
				// Clear and seed membership data
				clearTestData(t, db)
				seedMembershipScalabilityData(t, db, size.accounts, size.membersPerAccount)

				// Test account membership query
				accountID := fmt.Sprintf("scale-account-%d", size.accounts/2)

				start := time.Now()
				members, err := memberRepo.ListByAccount(ctx, accountID)
				elapsed := time.Since(start)

				require.NoError(t, err)
				require.Len(t, members, size.membersPerAccount)

				assert.Less(t, elapsed, MaxLargeDatasetQueryTime,
					"Membership query took %v, expected < %v",
					elapsed, MaxLargeDatasetQueryTime)

				t.Logf("Membership %dx%d: query %v", size.accounts, size.membersPerAccount, elapsed)
			})
		}
	})
}

// TestConcurrentUserLoad tests system behavior under concurrent user load
func TestConcurrentUserLoad(t *testing.T) {
	db := setupScalabilityDB(t)
	defer cleanupPerformanceTestDB(db)

	// Seed initial data
	seedScalabilityTestData(t, db, 500)

	cache := infrastructure.NewInMemoryCache(5 * time.Minute)
	userRepo := infrastructure.NewOptimizedGormUserRepository(db, cache)
	userWriteRepo := infrastructure.NewGormUserWriteRepository(db)

	t.Run("MixedWorkload", func(t *testing.T) {
		var wg sync.WaitGroup
		var mu sync.Mutex
		var readErrors, writeErrors []error
		var readLatencies, writeLatencies []time.Duration
		var readCount, writeCount int64

		concurrentUsers := MinConcurrentUsers
		operationsPerUser := 10

		startTime := time.Now()

		// Simulate concurrent users performing mixed operations
		for userIndex := 0; userIndex < concurrentUsers; userIndex++ {
			wg.Add(1)
			go func(userID int) {
				defer wg.Done()

				ctx := context.Background()

				for opIndex := 0; opIndex < operationsPerUser; opIndex++ {
					// 80% reads, 20% writes
					if opIndex%5 == 0 {
						// Write operation
						opStart := time.Now()

						newUser := &domain.User{
							BasicEntity: pericarpdomain.NewEntity(fmt.Sprintf("concurrent-user-%d-%d", userID, opIndex)),
							WebID:       fmt.Sprintf("https://example.com/profile/concurrent-user-%d-%d", userID, opIndex),
							Email:       fmt.Sprintf("concurrent-user-%d-%d@example.com", userID, opIndex),
							Profile: domain.UserProfile{
								Name: fmt.Sprintf("Concurrent User %d-%d", userID, opIndex),
							},
							Status:    domain.UserStatusActive,
							CreatedAt: time.Now(),
							UpdatedAt: time.Now(),
						}

						err := userWriteRepo.Create(ctx, newUser)
						latency := time.Since(opStart)

						mu.Lock()
						if err != nil {
							writeErrors = append(writeErrors, err)
						} else {
							atomic.AddInt64(&writeCount, 1)
						}
						writeLatencies = append(writeLatencies, latency)
						mu.Unlock()
					} else {
						// Read operation
						opStart := time.Now()

						targetUserID := fmt.Sprintf("scale-user-%d", (userID*operationsPerUser+opIndex)%500)
						_, err := userRepo.GetByID(ctx, targetUserID)
						latency := time.Since(opStart)

						mu.Lock()
						if err != nil {
							readErrors = append(readErrors, err)
						} else {
							atomic.AddInt64(&readCount, 1)
						}
						readLatencies = append(readLatencies, latency)
						mu.Unlock()
					}
				}
			}(userIndex)
		}

		wg.Wait()
		totalTime := time.Since(startTime)

		// Calculate metrics
		readThroughput := float64(readCount) / totalTime.Seconds()
		writeThroughput := float64(writeCount) / totalTime.Seconds()

		readErrorRate := float64(len(readErrors)) / float64(readCount+int64(len(readErrors)))
		writeErrorRate := float64(len(writeErrors)) / float64(writeCount+int64(len(writeErrors)))

		// Assertions
		assert.Less(t, readErrorRate, ScalabilityMaxErrorRate,
			"Read error rate was %.4f, expected < %.4f", readErrorRate, ScalabilityMaxErrorRate)

		assert.Less(t, writeErrorRate, ScalabilityMaxErrorRate,
			"Write error rate was %.4f, expected < %.4f", writeErrorRate, ScalabilityMaxErrorRate)

		// Calculate latency percentiles
		readPercentiles := calculatePercentiles(readLatencies, 0.95, 0.99)
		readP95, readP99 := readPercentiles[0], readPercentiles[1]
		writePercentiles := calculatePercentiles(writeLatencies, 0.95, 0.99)
		writeP95, writeP99 := writePercentiles[0], writePercentiles[1]

		assert.Less(t, readP95, MaxP95Latency,
			"Read P95 latency was %v, expected < %v", readP95, MaxP95Latency)

		assert.Less(t, writeP95, MaxP95Latency*2, // Allow more time for writes
			"Write P95 latency was %v, expected < %v", writeP95, MaxP95Latency*2)

		t.Logf("Mixed Workload: %d concurrent users, %.2f reads/sec, %.2f writes/sec",
			concurrentUsers, readThroughput, writeThroughput)
		t.Logf("Read: %.2f%% errors, P95: %v, P99: %v", readErrorRate*100, readP95, readP99)
		t.Logf("Write: %.2f%% errors, P95: %v, P99: %v", writeErrorRate*100, writeP95, writeP99)
	})
}

// TestDatabaseConnectionScaling tests performance with limited database connections
func TestDatabaseConnectionScaling(t *testing.T) {
	db := setupLimitedConnectionDB(t, MaxDatabaseConnections)
	defer cleanupPerformanceTestDB(db)

	// Seed data
	seedScalabilityTestData(t, db, 200)

	cache := infrastructure.NewInMemoryCache(5 * time.Minute)
	userRepo := infrastructure.NewOptimizedGormUserRepository(db, cache)

	t.Run("ConnectionPoolStress", func(t *testing.T) {
		var wg sync.WaitGroup
		var mu sync.Mutex
		var errors []error
		var latencies []time.Duration
		var successCount int64

		// More concurrent operations than available connections
		concurrentOps := MaxDatabaseConnections * 3

		startTime := time.Now()

		for i := 0; i < concurrentOps; i++ {
			wg.Add(1)
			go func(opIndex int) {
				defer wg.Done()

				ctx := context.Background()
				opStart := time.Now()

				userID := fmt.Sprintf("scale-user-%d", opIndex%200)
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

		// Calculate metrics
		throughput := float64(successCount) / totalTime.Seconds()
		errorRate := float64(len(errors)) / float64(concurrentOps)

		assert.Less(t, errorRate, ScalabilityMaxErrorRate,
			"Connection pool error rate was %.4f, expected < %.4f", errorRate, ScalabilityMaxErrorRate)

		// Should still achieve reasonable throughput with limited connections
		minThroughputWithLimitedConns := float64(MinThroughputLookups) * 0.5 // 50% of normal
		assert.Greater(t, throughput, minThroughputWithLimitedConns,
			"Throughput with limited connections was %.2f/sec, expected > %.2f/sec",
			throughput, minThroughputWithLimitedConns)

		t.Logf("Connection Pool: %d connections, %.2f/sec throughput, %.2f%% error rate",
			MaxDatabaseConnections, throughput, errorRate*100)
	})
}

// Helper functions for scalability testing

func setupScalabilityDB(t *testing.T) *gorm.DB {
	db := setupPerformanceTestDB(t)

	// Configure for scalability testing
	sqlDB, err := db.DB()
	require.NoError(t, err)

	sqlDB.SetMaxOpenConns(50)
	sqlDB.SetMaxIdleConns(25)
	sqlDB.SetConnMaxLifetime(time.Hour)

	return db
}

func setupLimitedConnectionDB(t *testing.T, maxConns int) *gorm.DB {
	db := setupScalabilityDB(t)

	// Limit connections for testing
	sqlDB, err := db.DB()
	require.NoError(t, err)

	sqlDB.SetMaxOpenConns(maxConns)
	sqlDB.SetMaxIdleConns(maxConns / 2)

	return db
}

func clearTestData(t *testing.T, db *gorm.DB) {
	// Clear all test data
	err := db.Exec("DELETE FROM account_member_models").Error
	require.NoError(t, err)

	err = db.Exec("DELETE FROM invitation_models").Error
	require.NoError(t, err)

	err = db.Exec("DELETE FROM account_models").Error
	require.NoError(t, err)

	err = db.Exec("DELETE FROM user_models").Error
	require.NoError(t, err)
}

func seedScalabilityTestData(t *testing.T, db *gorm.DB, userCount int) {
	// Seed system roles first
	roles := []infrastructure.RoleModel{
		{ID: "owner", Name: "Owner", Description: "Full access", Permissions: `[]`, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: "admin", Name: "Admin", Description: "Admin access", Permissions: `[]`, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: "member", Name: "Member", Description: "Member access", Permissions: `[]`, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: "viewer", Name: "Viewer", Description: "View access", Permissions: `[]`, CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}

	for _, role := range roles {
		err := db.FirstOrCreate(&role, "id = ?", role.ID).Error
		require.NoError(t, err)
	}

	// Seed users in batches for better performance
	batchSize := 1000
	for i := 0; i < userCount; i += batchSize {
		end := i + batchSize
		if end > userCount {
			end = userCount
		}

		users := make([]infrastructure.UserModel, end-i)
		for j := i; j < end; j++ {
			users[j-i] = infrastructure.UserModel{
				ID:        fmt.Sprintf("scale-user-%d", j),
				WebID:     fmt.Sprintf("https://example.com/profile/scale-user-%d", j),
				Email:     fmt.Sprintf("scale-user-%d@example.com", j),
				Name:      fmt.Sprintf("Scale User %d", j),
				Status:    "active",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
		}

		err := db.CreateInBatches(users, 100).Error
		require.NoError(t, err)
	}
}

func seedMembershipScalabilityData(t *testing.T, db *gorm.DB, accountCount, membersPerAccount int) {
	// Seed users first
	totalUsers := accountCount + membersPerAccount*2
	seedScalabilityTestData(t, db, totalUsers)

	// Seed accounts
	accounts := make([]infrastructure.AccountModel, accountCount)
	for i := 0; i < accountCount; i++ {
		accounts[i] = infrastructure.AccountModel{
			ID:          fmt.Sprintf("scale-account-%d", i),
			OwnerID:     fmt.Sprintf("scale-user-%d", i),
			Name:        fmt.Sprintf("Scale Account %d", i),
			Description: fmt.Sprintf("Test account %d for scalability testing", i),
			Settings:    `{}`,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
	}

	err := db.CreateInBatches(accounts, 100).Error
	require.NoError(t, err)

	// Seed memberships
	var members []infrastructure.AccountMemberModel
	memberID := 0

	for accountIndex := 0; accountIndex < accountCount; accountIndex++ {
		for memberIndex := 0; memberIndex < membersPerAccount; memberIndex++ {
			userIndex := (accountIndex + memberIndex) % totalUsers

			member := infrastructure.AccountMemberModel{
				ID:        fmt.Sprintf("scale-member-%d", memberID),
				AccountID: fmt.Sprintf("scale-account-%d", accountIndex),
				UserID:    fmt.Sprintf("scale-user-%d", userIndex),
				RoleID:    "member",
				InvitedBy: fmt.Sprintf("scale-user-%d", accountIndex),
				JoinedAt:  time.Now(),
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}

			members = append(members, member)
			memberID++
		}
	}

	err = db.CreateInBatches(members, 200).Error
	require.NoError(t, err)
}

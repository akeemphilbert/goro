package infrastructure

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/akeemphilbert/goro/internal/user/domain"
	pericarpdomain "github.com/akeemphilbert/pericarp/pkg/domain"
	"gorm.io/gorm"
)

// Benchmark tests for establishing performance baselines

// BenchmarkUserLookup benchmarks user lookup operations
func BenchmarkUserLookup(b *testing.B) {
	db := setupBenchmarkDB(b)
	defer cleanupBenchmarkDB(db)

	// Seed test data
	seedBenchmarkData(b, db, 1000)

	b.Run("WithoutCache", func(b *testing.B) {
		userRepo := NewGormUserRepository(db)
		ctx := context.Background()

		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				userID := fmt.Sprintf("bench-user-%d", i%1000)
				_, err := userRepo.GetByID(ctx, userID)
				if err != nil {
					b.Errorf("Lookup failed: %v", err)
				}
				i++
			}
		})
	})

	b.Run("WithCache", func(b *testing.B) {
		cache := NewInMemoryCache(5 * time.Minute)
		userRepo := NewOptimizedGormUserRepository(db, cache)
		ctx := context.Background()

		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				userID := fmt.Sprintf("bench-user-%d", i%1000)
				_, err := userRepo.GetByID(ctx, userID)
				if err != nil {
					b.Errorf("Lookup failed: %v", err)
				}
				i++
			}
		})
	})
}

// BenchmarkUserRegistration benchmarks user registration operations
func BenchmarkUserRegistration(b *testing.B) {
	db := setupBenchmarkDB(b)
	defer cleanupBenchmarkDB(db)

	userWriteRepo := NewGormUserWriteRepository(db)
	ctx := context.Background()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			user := &domain.User{
				BasicEntity: pericarpdomain.NewEntity(fmt.Sprintf("bench-reg-user-%d-%d", b.N, i)),
				WebID:       fmt.Sprintf("https://example.com/profile/bench-reg-user-%d-%d", b.N, i),
				Email:       fmt.Sprintf("bench-reg-user-%d-%d@example.com", b.N, i),
				Profile: domain.UserProfile{
					Name: fmt.Sprintf("Bench User %d-%d", b.N, i),
				},
				Status:    domain.UserStatusActive,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}

			err := userWriteRepo.Create(ctx, user)
			if err != nil {
				b.Errorf("Registration failed: %v", err)
			}
			i++
		}
	})
}

// BenchmarkMembershipQuery benchmarks membership query operations
func BenchmarkMembershipQuery(b *testing.B) {
	db := setupBenchmarkDB(b)
	defer cleanupBenchmarkDB(db)

	// Seed membership data
	seedMembershipBenchmarkData(b, db, 100, 50)

	b.Run("AccountMembers", func(b *testing.B) {
		memberRepo := NewOptimizedGormAccountMemberRepository(db)
		ctx := context.Background()

		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				accountID := fmt.Sprintf("bench-account-%d", i%100)
				_, err := memberRepo.ListByAccount(ctx, accountID)
				if err != nil {
					b.Errorf("Membership query failed: %v", err)
				}
				i++
			}
		})
	})

	b.Run("UserMemberships", func(b *testing.B) {
		memberRepo := NewOptimizedGormAccountMemberRepository(db)
		ctx := context.Background()

		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				userID := fmt.Sprintf("bench-user-%d", i%150)
				_, err := memberRepo.ListByUser(ctx, userID)
				if err != nil {
					b.Errorf("User membership query failed: %v", err)
				}
				i++
			}
		})
	})
}

// BenchmarkBulkOperations benchmarks bulk operations
func BenchmarkBulkOperations(b *testing.B) {
	db := setupBenchmarkDB(b)
	defer cleanupBenchmarkDB(db)

	// Seed data for bulk operations
	seedBenchmarkData(b, db, 5000)

	b.Run("UserList", func(b *testing.B) {
		cache := NewInMemoryCache(5 * time.Minute)
		userRepo := NewOptimizedGormUserRepository(db, cache)
		ctx := context.Background()

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			filter := domain.UserFilter{
				Status: domain.UserStatusActive,
				Limit:  100,
				Offset: (i * 100) % 4900, // Cycle through dataset
			}

			_, err := userRepo.List(ctx, filter)
			if err != nil {
				b.Errorf("Bulk list failed: %v", err)
			}
		}
	})

	b.Run("UserSearch", func(b *testing.B) {
		cache := NewInMemoryCache(5 * time.Minute)
		userRepo := NewOptimizedGormUserRepository(db, cache)
		ctx := context.Background()

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			filter := domain.UserFilter{
				EmailPattern: fmt.Sprintf("bench-user-%d", i%1000),
				Limit:        10,
			}

			_, err := userRepo.List(ctx, filter)
			if err != nil {
				b.Errorf("Search failed: %v", err)
			}
		}
	})
}

// BenchmarkCacheOperations benchmarks cache performance
func BenchmarkCacheOperations(b *testing.B) {
	cache := NewInMemoryCache(5 * time.Minute)

	// Pre-populate cache
	for i := 0; i < 1000; i++ {
		user := &domain.User{
			BasicEntity: pericarpdomain.NewEntity(fmt.Sprintf("cache-user-%d", i)),
			WebID:       fmt.Sprintf("https://example.com/profile/cache-user-%d", i),
			Email:       fmt.Sprintf("cache-user-%d@example.com", i),
			Profile: domain.UserProfile{
				Name: fmt.Sprintf("Cache User %d", i),
			},
			Status:    domain.UserStatusActive,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		cache.SetUser(user.ID(), user)
	}

	b.Run("CacheHit", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				userID := fmt.Sprintf("cache-user-%d", i%1000)
				_, found := cache.GetUser(userID)
				if !found {
					b.Errorf("Cache miss for user: %s", userID)
				}
				i++
			}
		})
	})

	b.Run("CacheMiss", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				userID := fmt.Sprintf("missing-user-%d", i)
				_, found := cache.GetUser(userID)
				if found {
					b.Errorf("Unexpected cache hit for user: %s", userID)
				}
				i++
			}
		})
	})

	b.Run("CacheSet", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				user := &domain.User{
					BasicEntity: pericarpdomain.NewEntity(fmt.Sprintf("new-user-%d-%d", b.N, i)),
					WebID:       fmt.Sprintf("https://example.com/profile/new-user-%d-%d", b.N, i),
					Email:       fmt.Sprintf("new-user-%d-%d@example.com", b.N, i),
					Profile: domain.UserProfile{
						Name: fmt.Sprintf("New User %d-%d", b.N, i),
					},
					Status:    domain.UserStatusActive,
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				}
				cache.SetUser(user.ID(), user)
				i++
			}
		})
	})
}

// Helper functions for benchmarking

func setupBenchmarkDB(tb testing.TB) *gorm.DB {
	return setupPerformanceTestDB(tb.(*testing.T))
}

func cleanupBenchmarkDB(db *gorm.DB) {
	cleanupPerformanceTestDB(db)
}

func seedBenchmarkData(tb testing.TB, db *gorm.DB, userCount int) {
	t := tb.(*testing.T)

	// Seed system roles
	roles := []RoleModel{
		{ID: "owner", Name: "Owner", Description: "Full access", Permissions: `[]`, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: "admin", Name: "Admin", Description: "Admin access", Permissions: `[]`, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: "member", Name: "Member", Description: "Member access", Permissions: `[]`, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: "viewer", Name: "Viewer", Description: "View access", Permissions: `[]`, CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}

	for _, role := range roles {
		err := db.FirstOrCreate(&role, "id = ?", role.ID).Error
		if err != nil {
			tb.Fatalf("Failed to seed role: %v", err)
		}
	}

	// Seed users in batches
	batchSize := 500
	for i := 0; i < userCount; i += batchSize {
		end := i + batchSize
		if end > userCount {
			end = userCount
		}

		users := make([]UserModel, end-i)
		for j := i; j < end; j++ {
			users[j-i] = UserModel{
				ID:        fmt.Sprintf("bench-user-%d", j),
				WebID:     fmt.Sprintf("https://example.com/profile/bench-user-%d", j),
				Email:     fmt.Sprintf("bench-user-%d@example.com", j),
				Name:      fmt.Sprintf("Bench User %d", j),
				Status:    "active",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
		}

		err := db.CreateInBatches(users, 100).Error
		if err != nil {
			t.Fatalf("Failed to seed users: %v", err)
		}
	}
}

func seedMembershipBenchmarkData(tb testing.TB, db *gorm.DB, accountCount, membersPerAccount int) {
	t := tb.(*testing.T)

	// Seed users first
	totalUsers := accountCount + membersPerAccount
	seedBenchmarkData(tb, db, totalUsers)

	// Seed accounts
	accounts := make([]AccountModel, accountCount)
	for i := 0; i < accountCount; i++ {
		accounts[i] = AccountModel{
			ID:          fmt.Sprintf("bench-account-%d", i),
			OwnerID:     fmt.Sprintf("bench-user-%d", i),
			Name:        fmt.Sprintf("Bench Account %d", i),
			Description: fmt.Sprintf("Benchmark account %d", i),
			Settings:    `{}`,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
	}

	err := db.CreateInBatches(accounts, 50).Error
	if err != nil {
		t.Fatalf("Failed to seed accounts: %v", err)
	}

	// Seed memberships
	var members []AccountMemberModel
	memberID := 0

	for accountIndex := 0; accountIndex < accountCount; accountIndex++ {
		for memberIndex := 0; memberIndex < membersPerAccount; memberIndex++ {
			userIndex := (accountIndex + memberIndex) % totalUsers

			member := AccountMemberModel{
				ID:        fmt.Sprintf("bench-member-%d", memberID),
				AccountID: fmt.Sprintf("bench-account-%d", accountIndex),
				UserID:    fmt.Sprintf("bench-user-%d", userIndex),
				RoleID:    "member",
				InvitedBy: fmt.Sprintf("bench-user-%d", accountIndex),
				JoinedAt:  time.Now(),
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}

			members = append(members, member)
			memberID++
		}
	}

	err = db.CreateInBatches(members, 100).Error
	if err != nil {
		t.Fatalf("Failed to seed memberships: %v", err)
	}
}

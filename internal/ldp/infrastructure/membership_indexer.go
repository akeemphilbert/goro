package infrastructure

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// ResourceType represents the type of a resource
type ResourceType string

const (
	ResourceTypeContainer ResourceType = "Container"
	ResourceTypeResource  ResourceType = "Resource"
)

// MemberInfo contains information about a container member
type MemberInfo struct {
	ID          string
	Type        ResourceType
	ContentType string
	Size        int64
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// PaginationOptions contains pagination parameters
type PaginationOptions struct {
	Limit  int
	Offset int
}

// MembershipIndexer defines the interface for container membership indexing
type MembershipIndexer interface {
	IndexMembership(ctx context.Context, containerID, memberID string) error
	RemoveMembership(ctx context.Context, containerID, memberID string) error
	GetMembers(ctx context.Context, containerID string, pagination PaginationOptions) ([]MemberInfo, error)
	GetContainers(ctx context.Context, memberID string) ([]string, error)
	RebuildIndex(ctx context.Context) error
	Close() error
}

// SQLiteMembershipIndexer implements MembershipIndexer using SQLite
type SQLiteMembershipIndexer struct {
	db *sql.DB
}

// NewSQLiteMembershipIndexer creates a new SQLite membership indexer
func NewSQLiteMembershipIndexer(dbPath string) (*SQLiteMembershipIndexer, error) {
	db, err := sql.Open("sqlite3", dbPath+"?_foreign_keys=on")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	indexer := &SQLiteMembershipIndexer{db: db}

	// Apply migrations
	if err := migrateDatabase(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return indexer, nil
}

// IndexMembership adds a membership relationship to the index
func (s *SQLiteMembershipIndexer) IndexMembership(ctx context.Context, containerID, memberID string) error {
	// Determine member type by checking if it's a container
	memberType := ResourceTypeResource
	var exists int
	err := s.db.QueryRowContext(ctx, "SELECT 1 FROM containers WHERE id = ?", memberID).Scan(&exists)
	if err == nil {
		memberType = ResourceTypeContainer
	} else if err != sql.ErrNoRows {
		return fmt.Errorf("failed to check member type: %w", err)
	}

	// Insert membership
	query := `
		INSERT OR REPLACE INTO memberships (container_id, member_id, member_type, created_at) 
		VALUES (?, ?, ?, CURRENT_TIMESTAMP)`

	_, err = s.db.ExecContext(ctx, query, containerID, memberID, string(memberType))
	if err != nil {
		return fmt.Errorf("failed to index membership: %w", err)
	}

	return nil
}

// RemoveMembership removes a membership relationship from the index
func (s *SQLiteMembershipIndexer) RemoveMembership(ctx context.Context, containerID, memberID string) error {
	query := "DELETE FROM memberships WHERE container_id = ? AND member_id = ?"

	result, err := s.db.ExecContext(ctx, query, containerID, memberID)
	if err != nil {
		return fmt.Errorf("failed to remove membership: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("membership not found: container_id=%s, member_id=%s", containerID, memberID)
	}

	return nil
}

// GetMembers retrieves all members of a container with pagination
func (s *SQLiteMembershipIndexer) GetMembers(ctx context.Context, containerID string, pagination PaginationOptions) ([]MemberInfo, error) {
	query := `
		SELECT m.member_id, m.member_type, m.created_at,
			   COALESCE(c.updated_at, m.created_at) as updated_at
		FROM memberships m
		LEFT JOIN containers c ON m.member_id = c.id AND m.member_type = 'Container'
		WHERE m.container_id = ?
		ORDER BY m.created_at`

	args := []interface{}{containerID}

	// Add pagination if specified
	if pagination.Limit > 0 {
		query += " LIMIT ?"
		args = append(args, pagination.Limit)

		if pagination.Offset > 0 {
			query += " OFFSET ?"
			args = append(args, pagination.Offset)
		}
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query members: %w", err)
	}
	defer rows.Close()

	var members []MemberInfo
	for rows.Next() {
		var member MemberInfo
		var memberTypeStr string
		var createdAtStr, updatedAtStr string

		err := rows.Scan(&member.ID, &memberTypeStr, &createdAtStr, &updatedAtStr)
		if err != nil {
			return nil, fmt.Errorf("failed to scan member: %w", err)
		}

		// Parse timestamps
		member.CreatedAt, err = time.Parse("2006-01-02 15:04:05", createdAtStr)
		if err != nil {
			// Try alternative format
			member.CreatedAt, err = time.Parse(time.RFC3339, createdAtStr)
			if err != nil {
				member.CreatedAt = time.Now() // fallback
			}
		}

		member.UpdatedAt, err = time.Parse("2006-01-02 15:04:05", updatedAtStr)
		if err != nil {
			// Try alternative format
			member.UpdatedAt, err = time.Parse(time.RFC3339, updatedAtStr)
			if err != nil {
				member.UpdatedAt = member.CreatedAt // fallback
			}
		}

		member.Type = ResourceType(memberTypeStr)

		// Set default content type based on type
		if member.Type == ResourceTypeContainer {
			member.ContentType = "application/ld+json"
		} else {
			member.ContentType = "application/octet-stream"
		}

		members = append(members, member)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating members: %w", err)
	}

	return members, nil
}

// GetContainers retrieves all containers that contain a specific member
func (s *SQLiteMembershipIndexer) GetContainers(ctx context.Context, memberID string) ([]string, error) {
	query := "SELECT container_id FROM memberships WHERE member_id = ? ORDER BY created_at"

	rows, err := s.db.QueryContext(ctx, query, memberID)
	if err != nil {
		return nil, fmt.Errorf("failed to query containers: %w", err)
	}
	defer rows.Close()

	var containers []string
	for rows.Next() {
		var containerID string
		if err := rows.Scan(&containerID); err != nil {
			return nil, fmt.Errorf("failed to scan container ID: %w", err)
		}
		containers = append(containers, containerID)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating containers: %w", err)
	}

	return containers, nil
}

// RebuildIndex rebuilds the membership index from scratch
func (s *SQLiteMembershipIndexer) RebuildIndex(ctx context.Context) error {
	// Start transaction
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Clear existing memberships
	if _, err := tx.ExecContext(ctx, "DELETE FROM memberships"); err != nil {
		return fmt.Errorf("failed to clear memberships: %w", err)
	}

	// Note: In a real implementation, this would scan the filesystem
	// or other storage to rebuild the index. For now, we just ensure
	// the table is clean and ready for new memberships.

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit rebuild transaction: %w", err)
	}

	return nil
}

// Close closes the database connection
func (s *SQLiteMembershipIndexer) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

// GetDB returns the database connection (for internal use by repository)
func (s *SQLiteMembershipIndexer) GetDB() *sql.DB {
	return s.db
}

// GetMemberCount returns the total number of members in a container
func (s *SQLiteMembershipIndexer) GetMemberCount(ctx context.Context, containerID string) (int, error) {
	var count int
	query := "SELECT COUNT(*) FROM memberships WHERE container_id = ?"

	err := s.db.QueryRowContext(ctx, query, containerID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get member count: %w", err)
	}

	return count, nil
}

// GetContainerStats returns statistics about a container
func (s *SQLiteMembershipIndexer) GetContainerStats(ctx context.Context, containerID string) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Get total member count
	memberCount, err := s.GetMemberCount(ctx, containerID)
	if err != nil {
		return nil, err
	}
	stats["member_count"] = memberCount

	// Get member type breakdown
	query := `
		SELECT member_type, COUNT(*) 
		FROM memberships 
		WHERE container_id = ? 
		GROUP BY member_type`

	rows, err := s.db.QueryContext(ctx, query, containerID)
	if err != nil {
		return nil, fmt.Errorf("failed to query member type stats: %w", err)
	}
	defer rows.Close()

	typeStats := make(map[string]int)
	for rows.Next() {
		var memberType string
		var count int
		if err := rows.Scan(&memberType, &count); err != nil {
			return nil, fmt.Errorf("failed to scan type stats: %w", err)
		}
		typeStats[memberType] = count
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating type stats: %w", err)
	}

	stats["type_breakdown"] = typeStats

	return stats, nil
}

package infrastructure

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

// GenericMembershipIndexer implements MembershipIndexer for multiple database types
type GenericMembershipIndexer struct {
	db       *sql.DB
	driver   string
	provider SchemaProvider
}

// NewGenericMembershipIndexer creates a new database-agnostic membership indexer
func NewGenericMembershipIndexer(config DatabaseConfig) (*GenericMembershipIndexer, error) {
	db, err := NewDatabaseConnection(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create database connection: %w", err)
	}

	provider, err := NewSchemaProvider(config.Driver)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to create schema provider: %w", err)
	}

	indexer := &GenericMembershipIndexer{
		db:       db,
		driver:   strings.ToLower(config.Driver),
		provider: provider,
	}

	// Apply migrations
	if err := MigrateDatabaseWithProvider(db, provider); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return indexer, nil
}

// IndexMembership adds a membership relationship to the index
func (g *GenericMembershipIndexer) IndexMembership(ctx context.Context, containerID, memberID string) error {
	// Determine member type by checking if it's a container
	memberType := ResourceTypeResource
	var exists int
	err := g.db.QueryRowContext(ctx, "SELECT 1 FROM containers WHERE id = "+g.placeholder(1), memberID).Scan(&exists)
	if err == nil {
		memberType = ResourceTypeContainer
	} else if err != sql.ErrNoRows {
		return fmt.Errorf("failed to check member type: %w", err)
	}

	// Insert membership with database-specific syntax
	var query string
	switch g.driver {
	case "sqlite3", "sqlite":
		query = `INSERT OR REPLACE INTO memberships (container_id, member_id, member_type, created_at) 
				 VALUES (?, ?, ?, CURRENT_TIMESTAMP)`
	case "postgres", "postgresql":
		query = `INSERT INTO memberships (container_id, member_id, member_type, created_at) 
				 VALUES ($1, $2, $3, CURRENT_TIMESTAMP)
				 ON CONFLICT (container_id, member_id) DO UPDATE SET
				 member_type = EXCLUDED.member_type, created_at = EXCLUDED.created_at`
	default:
		return fmt.Errorf("unsupported database driver: %s", g.driver)
	}

	_, err = g.db.ExecContext(ctx, query, containerID, memberID, string(memberType))
	if err != nil {
		return fmt.Errorf("failed to index membership: %w", err)
	}

	return nil
}

// RemoveMembership removes a membership relationship from the index
func (g *GenericMembershipIndexer) RemoveMembership(ctx context.Context, containerID, memberID string) error {
	query := "DELETE FROM memberships WHERE container_id = " + g.placeholder(1) + " AND member_id = " + g.placeholder(2)

	result, err := g.db.ExecContext(ctx, query, containerID, memberID)
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
func (g *GenericMembershipIndexer) GetMembers(ctx context.Context, containerID string, pagination PaginationOptions) ([]MemberInfo, error) {
	query := `
		SELECT m.member_id, m.member_type, m.created_at,
			   COALESCE(c.updated_at, m.created_at) as updated_at
		FROM memberships m
		LEFT JOIN containers c ON m.member_id = c.id AND m.member_type = 'Container'
		WHERE m.container_id = ` + g.placeholder(1) + `
		ORDER BY m.created_at`

	args := []interface{}{containerID}

	// Add pagination if specified
	if pagination.Limit > 0 {
		query += " LIMIT " + g.placeholder(2)
		args = append(args, pagination.Limit)

		if pagination.Offset > 0 {
			query += " OFFSET " + g.placeholder(3)
			args = append(args, pagination.Offset)
		}
	}

	rows, err := g.db.QueryContext(ctx, query, args...)
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

		// Parse timestamps based on database type
		member.CreatedAt, err = g.parseTimestamp(createdAtStr)
		if err != nil {
			member.CreatedAt = time.Now() // fallback
		}

		member.UpdatedAt, err = g.parseTimestamp(updatedAtStr)
		if err != nil {
			member.UpdatedAt = member.CreatedAt // fallback
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
func (g *GenericMembershipIndexer) GetContainers(ctx context.Context, memberID string) ([]string, error) {
	query := "SELECT container_id FROM memberships WHERE member_id = " + g.placeholder(1) + " ORDER BY created_at"

	rows, err := g.db.QueryContext(ctx, query, memberID)
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
func (g *GenericMembershipIndexer) RebuildIndex(ctx context.Context) error {
	// Start transaction
	tx, err := g.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Clear existing memberships
	if _, err := tx.ExecContext(ctx, "DELETE FROM memberships"); err != nil {
		return fmt.Errorf("failed to clear memberships: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit rebuild transaction: %w", err)
	}

	return nil
}

// Close closes the database connection
func (g *GenericMembershipIndexer) Close() error {
	if g.db != nil {
		return g.db.Close()
	}
	return nil
}

// GetMemberCount returns the total number of members in a container
func (g *GenericMembershipIndexer) GetMemberCount(ctx context.Context, containerID string) (int, error) {
	var count int
	query := "SELECT COUNT(*) FROM memberships WHERE container_id = " + g.placeholder(1)

	err := g.db.QueryRowContext(ctx, query, containerID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get member count: %w", err)
	}

	return count, nil
}

// GetContainerStats returns statistics about a container
func (g *GenericMembershipIndexer) GetContainerStats(ctx context.Context, containerID string) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Get total member count
	memberCount, err := g.GetMemberCount(ctx, containerID)
	if err != nil {
		return nil, err
	}
	stats["member_count"] = memberCount

	// Get member type breakdown
	query := `
		SELECT member_type, COUNT(*) 
		FROM memberships 
		WHERE container_id = ` + g.placeholder(1) + `
		GROUP BY member_type`

	rows, err := g.db.QueryContext(ctx, query, containerID)
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

// placeholder returns the appropriate placeholder for the database type
func (g *GenericMembershipIndexer) placeholder(n int) string {
	switch g.driver {
	case "sqlite3", "sqlite":
		return "?"
	case "postgres", "postgresql":
		return fmt.Sprintf("$%d", n)
	default:
		return "?"
	}
}

// parseTimestamp parses a timestamp string based on database type
func (g *GenericMembershipIndexer) parseTimestamp(timestampStr string) (time.Time, error) {
	// Try common timestamp formats
	formats := []string{
		"2006-01-02 15:04:05",         // SQLite format
		"2006-01-02T15:04:05Z",        // ISO format
		"2006-01-02T15:04:05.000000Z", // PostgreSQL format
		time.RFC3339,                  // RFC3339 format
		time.RFC3339Nano,              // RFC3339 with nanoseconds
	}

	for _, format := range formats {
		if t, err := time.Parse(format, timestampStr); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse timestamp: %s", timestampStr)
}

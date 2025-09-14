package infrastructure

import (
	"database/sql"
	"fmt"
	"strings"
)

// SchemaProvider handles database schema operations for different database types
type SchemaProvider interface {
	CreateContainerSchema(db *sql.DB) error
	CreateSchemaMigrationsTable(db *sql.DB) error
	GetCurrentSchemaVersion(db *sql.DB) (int, error)
	RecordMigration(db *sql.DB, version int, description string) error
	ValidateSchema(db *sql.DB) error
}

// NewSchemaProvider creates a schema provider for the given database driver
func NewSchemaProvider(driver string) (SchemaProvider, error) {
	switch strings.ToLower(driver) {
	case "sqlite3", "sqlite":
		return &SQLiteSchemaProvider{}, nil
	case "postgres", "postgresql":
		return &PostgreSQLSchemaProvider{}, nil
	default:
		return nil, fmt.Errorf("unsupported database driver: %s", driver)
	}
}

// SQLiteSchemaProvider implements schema operations for SQLite
type SQLiteSchemaProvider struct{}

func (p *SQLiteSchemaProvider) CreateContainerSchema(db *sql.DB) error {
	return createSQLiteContainerSchema(db)
}

func (p *SQLiteSchemaProvider) CreateSchemaMigrationsTable(db *sql.DB) error {
	return createSQLiteSchemaMigrationsTable(db)
}

func (p *SQLiteSchemaProvider) GetCurrentSchemaVersion(db *sql.DB) (int, error) {
	return getCurrentSchemaVersion(db)
}

func (p *SQLiteSchemaProvider) RecordMigration(db *sql.DB, version int, description string) error {
	return recordMigration(db, version, description)
}

func (p *SQLiteSchemaProvider) ValidateSchema(db *sql.DB) error {
	return validateSQLiteSchema(db)
}

// PostgreSQLSchemaProvider implements schema operations for PostgreSQL
type PostgreSQLSchemaProvider struct{}

func (p *PostgreSQLSchemaProvider) CreateContainerSchema(db *sql.DB) error {
	return createPostgreSQLContainerSchema(db)
}

func (p *PostgreSQLSchemaProvider) CreateSchemaMigrationsTable(db *sql.DB) error {
	return createPostgreSQLSchemaMigrationsTable(db)
}

func (p *PostgreSQLSchemaProvider) GetCurrentSchemaVersion(db *sql.DB) (int, error) {
	return getCurrentSchemaVersion(db)
}

func (p *PostgreSQLSchemaProvider) RecordMigration(db *sql.DB, version int, description string) error {
	return recordMigration(db, version, description)
}

func (p *PostgreSQLSchemaProvider) ValidateSchema(db *sql.DB) error {
	return validatePostgreSQLSchema(db)
}

// MigrateDatabase applies all pending migrations using the appropriate schema provider
func MigrateDatabaseWithProvider(db *sql.DB, provider SchemaProvider) error {
	// Create schema migrations table first
	if err := provider.CreateSchemaMigrationsTable(db); err != nil {
		return err
	}

	// Get current version
	currentVersion, err := provider.GetCurrentSchemaVersion(db)
	if err != nil {
		return err
	}

	// Apply migrations
	migrations := []struct {
		version     int
		description string
		apply       func(*sql.DB, SchemaProvider) error
	}{
		{
			version:     1,
			description: "Initial container schema",
			apply: func(db *sql.DB, p SchemaProvider) error {
				return p.CreateContainerSchema(db)
			},
		},
	}

	for _, migration := range migrations {
		if migration.version > currentVersion {
			if err := migration.apply(db, provider); err != nil {
				return fmt.Errorf("failed to apply migration %d: %w", migration.version, err)
			}

			if err := provider.RecordMigration(db, migration.version, migration.description); err != nil {
				return err
			}
		}
	}

	return nil
}

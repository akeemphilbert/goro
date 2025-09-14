package infrastructure

import (
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/lib/pq"           // PostgreSQL driver
	_ "github.com/mattn/go-sqlite3" // SQLite driver
)

// DatabaseConfig holds database connection configuration
type DatabaseConfig struct {
	Driver   string // "sqlite3" or "postgres"
	Host     string // PostgreSQL host (ignored for SQLite)
	Port     int    // PostgreSQL port (ignored for SQLite)
	Database string // Database name or SQLite file path
	Username string // PostgreSQL username (ignored for SQLite)
	Password string // PostgreSQL password (ignored for SQLite)
	SSLMode  string // PostgreSQL SSL mode (ignored for SQLite)
}

// NewDatabaseConnection creates a database connection based on configuration
func NewDatabaseConnection(config DatabaseConfig) (*sql.DB, error) {
	var dsn string
	var err error

	switch strings.ToLower(config.Driver) {
	case "sqlite3", "sqlite":
		dsn = config.Database + "?_foreign_keys=on"
	case "postgres", "postgresql":
		dsn, err = buildPostgresDSN(config)
		if err != nil {
			return nil, fmt.Errorf("failed to build PostgreSQL DSN: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported database driver: %s", config.Driver)
	}

	db, err := sql.Open(config.Driver, dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

// buildPostgresDSN constructs a PostgreSQL connection string
func buildPostgresDSN(config DatabaseConfig) (string, error) {
	if config.Host == "" {
		config.Host = "localhost"
	}
	if config.Port == 0 {
		config.Port = 5432
	}
	if config.SSLMode == "" {
		config.SSLMode = "disable"
	}

	dsn := fmt.Sprintf("host=%s port=%d dbname=%s user=%s password=%s sslmode=%s",
		config.Host, config.Port, config.Database, config.Username, config.Password, config.SSLMode)

	return dsn, nil
}

// DefaultSQLiteConfig returns a default SQLite configuration
func DefaultSQLiteConfig(dbPath string) DatabaseConfig {
	return DatabaseConfig{
		Driver:   "sqlite3",
		Database: dbPath,
	}
}

// DefaultPostgresConfig returns a default PostgreSQL configuration
func DefaultPostgresConfig(host, database, username, password string) DatabaseConfig {
	return DatabaseConfig{
		Driver:   "postgres",
		Host:     host,
		Port:     5432,
		Database: database,
		Username: username,
		Password: password,
		SSLMode:  "disable",
	}
}

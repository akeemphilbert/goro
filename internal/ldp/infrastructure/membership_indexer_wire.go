package infrastructure

import (
	"path/filepath"

	"github.com/google/wire"
)

// MembershipIndexerProvider creates a SQLiteMembershipIndexer instance (legacy)
func MembershipIndexerProvider(basePath string) (*SQLiteMembershipIndexer, error) {
	// Create database path for membership index
	dbPath := filepath.Join(basePath, "membership.db")

	indexer, err := NewSQLiteMembershipIndexer(dbPath)
	if err != nil {
		return nil, err
	}

	return indexer, nil
}

// GenericMembershipIndexerProvider creates a GenericMembershipIndexer instance
func GenericMembershipIndexerProvider(config DatabaseConfig) (*GenericMembershipIndexer, error) {
	indexer, err := NewGenericMembershipIndexer(config)
	if err != nil {
		return nil, err
	}

	return indexer, nil
}

// SQLiteMembershipIndexerProvider creates a SQLite-specific membership indexer
func SQLiteMembershipIndexerProvider(basePath string) (*GenericMembershipIndexer, error) {
	dbPath := filepath.Join(basePath, "membership.db")
	config := DefaultSQLiteConfig(dbPath)

	return NewGenericMembershipIndexer(config)
}

// PostgreSQLMembershipIndexerProvider creates a PostgreSQL-specific membership indexer
func PostgreSQLMembershipIndexerProvider(host, database, username, password string) (*GenericMembershipIndexer, error) {
	config := DefaultPostgresConfig(host, database, username, password)

	return NewGenericMembershipIndexer(config)
}

// MembershipIndexerSet provides the membership indexer for Wire dependency injection
var MembershipIndexerSet = wire.NewSet(
	MembershipIndexerProvider,
	wire.Bind(new(MembershipIndexer), new(*SQLiteMembershipIndexer)),
)

// GenericMembershipIndexerSet provides the generic membership indexer for Wire dependency injection
var GenericMembershipIndexerSet = wire.NewSet(
	GenericMembershipIndexerProvider,
	wire.Bind(new(MembershipIndexer), new(*GenericMembershipIndexer)),
)

// SQLiteMembershipIndexerSet provides SQLite-specific membership indexer
var SQLiteMembershipIndexerSet = wire.NewSet(
	SQLiteMembershipIndexerProvider,
	wire.Bind(new(MembershipIndexer), new(*GenericMembershipIndexer)),
)

// PostgreSQLMembershipIndexerSet provides PostgreSQL-specific membership indexer
var PostgreSQLMembershipIndexerSet = wire.NewSet(
	PostgreSQLMembershipIndexerProvider,
	wire.Bind(new(MembershipIndexer), new(*GenericMembershipIndexer)),
)

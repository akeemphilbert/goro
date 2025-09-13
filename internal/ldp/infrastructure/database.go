package infrastructure

import (
	"github.com/akeemphilbert/pericarp/pkg/infrastructure"
	"gorm.io/gorm"
)

// DatabaseProvider creates a GORM database instance using pericarp's database setup
func DatabaseProvider() (*gorm.DB, error) {
	// Use SQLite for development/testing - in production this would come from config
	config := infrastructure.DefaultSQLiteConfig()

	// Create database using pericarp's NewDatabase function
	db, err := infrastructure.NewDatabase(config)
	if err != nil {
		return nil, err
	}

	return db, nil
}

// EventStoreProvider creates a GormEventStore using pericarp's implementation
func EventStoreProvider(db *gorm.DB) (*infrastructure.GormEventStore, error) {
	eventStore, err := infrastructure.NewGormEventStore(db)
	if err != nil {
		return nil, err
	}

	return eventStore, nil
}

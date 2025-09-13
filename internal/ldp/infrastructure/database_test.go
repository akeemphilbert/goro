package infrastructure

import (
	"testing"

	"github.com/akeemphilbert/pericarp/pkg/infrastructure"
)

func TestDatabaseProvider(t *testing.T) {
	// Test database creation
	db, err := DatabaseProvider()
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	if db == nil {
		t.Fatal("Database should not be nil")
	}

	// Test that we can create an event store with the database
	eventStore, err := EventStoreProvider(db)
	if err != nil {
		t.Fatalf("Failed to create event store: %v", err)
	}
	if eventStore == nil {
		t.Fatal("Event store should not be nil")
	}

	// Verify the event store implements the correct interface
	// The type assertion is done implicitly by the EventStoreProvider function
}

func TestEventStoreProvider(t *testing.T) {
	// Create database first
	db, err := DatabaseProvider()
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}

	// Test event store creation
	eventStore, err := EventStoreProvider(db)
	if err != nil {
		t.Fatalf("Failed to create event store: %v", err)
	}
	if eventStore == nil {
		t.Fatal("Event store should not be nil")
	}

	// Test that the event store is properly initialized
	// The GormEventStore should have created the events table
	// This is verified by the fact that no error was returned
}

func TestInfrastructureIntegration(t *testing.T) {
	// Test the full infrastructure stack
	db, err := DatabaseProvider()
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}

	eventStore, err := EventStoreProvider(db)
	if err != nil {
		t.Fatalf("Failed to create event store: %v", err)
	}

	eventDispatcher, err := NewEventDispatcher()
	if err != nil {
		t.Fatalf("Failed to create event dispatcher: %v", err)
	}

	// Test that we can create a unit of work with these components
	unitOfWork := infrastructure.UnitOfWorkProvider(eventStore, eventDispatcher)
	if unitOfWork == nil {
		t.Fatal("Unit of work should not be nil")
	}

	// Cleanup
	if closer, ok := eventDispatcher.(*infrastructure.WatermillEventDispatcher); ok {
		err = closer.Close()
		if err != nil {
			t.Errorf("Failed to close event dispatcher: %v", err)
		}
	}
}

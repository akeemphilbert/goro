package application

import (
	"context"
	"testing"

	"github.com/akeemphilbert/goro/internal/user/infrastructure"
	pericarpdomain "github.com/akeemphilbert/pericarp/pkg/domain"
)

// SimpleMockEventDispatcher for testing
type SimpleMockEventDispatcher struct{}

func (m *SimpleMockEventDispatcher) Dispatch(ctx context.Context, envelopes []pericarpdomain.Envelope) error {
	return nil
}

func (m *SimpleMockEventDispatcher) Subscribe(eventType string, handler pericarpdomain.EventHandler) error {
	return nil
}

// TestSimpleWireProviders tests the basic Wire providers
func TestSimpleWireProviders(t *testing.T) {
	t.Run("ProvideInvitationGenerator", func(t *testing.T) {
		generator := ProvideInvitationGenerator()
		if generator == nil {
			t.Error("ProvideInvitationGenerator returned nil generator")
		}

		// Test basic functionality
		token := generator.GenerateToken()
		if token == "" {
			t.Error("GenerateToken returned empty string")
		}

		invitationID := generator.GenerateInvitationID()
		if invitationID == "" {
			t.Error("GenerateInvitationID returned empty string")
		}
	})

	t.Run("ProvideFileStorageAdapter", func(t *testing.T) {
		// Create a mock domain file storage
		tempDir := t.TempDir()
		domainStorage, err := infrastructure.ProvideFileStorage(tempDir)
		if err != nil {
			t.Fatalf("Failed to create domain file storage: %v", err)
		}

		adapter := ProvideFileStorageAdapter(domainStorage)
		if adapter == nil {
			t.Error("ProvideFileStorageAdapter returned nil adapter")
		}
	})

	t.Run("ProvideUnitOfWorkFactory", func(t *testing.T) {
		// Create mock dependencies
		eventDispatcher := &SimpleMockEventDispatcher{}

		factory := ProvideUnitOfWorkFactory(nil, eventDispatcher)
		if factory == nil {
			t.Error("ProvideUnitOfWorkFactory returned nil factory")
		}

		// Test that factory creates unit of work
		uow := factory()
		if uow == nil {
			t.Error("UnitOfWork factory returned nil unit of work")
		}
	})
}

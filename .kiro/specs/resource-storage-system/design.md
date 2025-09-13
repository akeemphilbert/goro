# Resource Storage System Design

## Overview

The Resource Storage System provides the foundational data storage capabilities for the Solid pod server, implementing a clean architecture approach using the Kratos framework. The system handles both RDF (structured) and binary (unstructured) data storage with support for multiple serialization formats and efficient file management.

The design leverages the pericarp library (https://github.com/akeemphilbert/pericarp) for domain entity management and implements an event-driven architecture to ensure data consistency and enable extensibility.

## Architecture

### Core Components

The system follows clean architecture principles with three distinct layers:

1. **Domain Layer** - Contains the core business logic and entities
2. **Application Layer** - Orchestrates use cases and business workflows  
3. **Infrastructure Layer** - Handles external dependencies and persistence

### Event-Driven Design

The system uses domain events to maintain consistency and enable loose coupling:
- Resource creation/modification triggers domain events
- Event handlers persist changes to the file system
- Events enable audit trails and future extensibility (notifications, indexing, etc.)

## Components and Interfaces

### Domain Layer (`internal/ldp/domain/`)

#### Resource Entity
```go
// Resource represents a stored resource in the pod
type Resource struct {
    // Core fields managed by pericarp
    ID          string
    ContentType string
    Data        []byte
    Metadata    map[string]interface{}
    CreatedAt   time.Time
    UpdatedAt   time.Time
}

// Methods for RDF format support
func (r *Resource) FromJSONLD(data []byte) error
func (r *Resource) FromRDF(data []byte) error  
func (r *Resource) FromXML(data []byte) error
func (r *Resource) ToFormat(format string) ([]byte, error)
```

#### Repository Interface
```go
type ResourceRepository interface {
    Store(ctx context.Context, resource *Resource) error
    Retrieve(ctx context.Context, id string) (*Resource, error)
    Delete(ctx context.Context, id string) error
    Exists(ctx context.Context, id string) (bool, error)
}
```

#### Domain Events
```go
// Using generic EntityEvent from pericarp
type EntityEvent struct {
    EntityID    string
    EntityType  string
    EventType   string
    Data        map[string]interface{}
    Timestamp   time.Time
}

// Event types
const (
    EventTypeResourceCreated = "resource.created"
    EventTypeResourceUpdated = "resource.updated"
    EventTypeResourceDeleted = "resource.deleted"
)
```

### Application Layer (`internal/ldp/application/`)

#### Storage Service
```go
type StorageService struct {
    repo ResourceRepository
    eventBus EventBus
}

func (s *StorageService) StoreResource(ctx context.Context, data []byte, contentType string) (*Resource, error)
func (s *StorageService) RetrieveResource(ctx context.Context, id string, acceptFormat string) (*Resource, error)
func (s *StorageService) DeleteResource(ctx context.Context, id string) error
```

#### Event Handlers
```go
type ResourceEventHandler struct {
    fileStorage FileStorage
}

func (h *ResourceEventHandler) HandleEntityEvent(event EntityEvent) error
```

### Infrastructure Layer (`internal/ldp/infrastructure/`)

#### File System Repository
```go
type FileSystemRepository struct {
    basePath string
    indexer  ResourceIndexer
}
```

#### Format Converter
```go
type RDFConverter struct{}

func (c *RDFConverter) Convert(data []byte, fromFormat, toFormat string) ([]byte, error)
func (c *RDFConverter) ValidateFormat(format string) bool
```

## Data Models

### Resource Storage Structure
```
pod-data/
├── resources/
│   ├── {resource-id}/
│   │   ├── content          # Raw resource data
│   │   ├── metadata.json    # Resource metadata
│   │   └── index.json       # Search index data
└── events/
    └── {date}/
        └── events.log       # Event audit trail
```

### Metadata Schema
```json
{
  "id": "resource-uuid",
  "contentType": "text/turtle",
  "originalFormat": "application/ld+json",
  "size": 1024,
  "checksum": "sha256-hash",
  "createdAt": "2024-01-01T00:00:00Z",
  "updatedAt": "2024-01-01T00:00:00Z",
  "tags": ["rdf", "public"]
}
```

## Error Handling

### Error Types
```go
type StorageError struct {
    Code    string
    Message string
    Cause   error
}

// Specific error types
var (
    ErrResourceNotFound     = &StorageError{Code: "RESOURCE_NOT_FOUND"}
    ErrUnsupportedFormat    = &StorageError{Code: "UNSUPPORTED_FORMAT"}
    ErrInsufficientStorage  = &StorageError{Code: "INSUFFICIENT_STORAGE"}
    ErrDataCorruption       = &StorageError{Code: "DATA_CORRUPTION"}
    ErrFormatConversion     = &StorageError{Code: "FORMAT_CONVERSION_FAILED"}
)
```

### Error Handling Strategy
- **Validation Errors**: Return 400 Bad Request with detailed error messages
- **Format Errors**: Return 406 Not Acceptable for unsupported formats
- **Storage Errors**: Return 507 Insufficient Storage when space is limited
- **Corruption Detection**: Verify checksums and reject corrupted data
- **Graceful Degradation**: Maintain system stability during partial failures

## Testing Strategy

### Unit Testing
- Test domain entities with pericarp integration
- Mock repository interfaces for service testing
- Test format conversion edge cases
- Validate error handling scenarios

### Integration Testing
- Test file system operations end-to-end
- Verify event handling and persistence
- Test concurrent access scenarios
- Validate data integrity across operations

### Performance Testing
- Benchmark large file operations
- Test streaming upload/download performance
- Measure concurrent access performance
- Validate sub-second response times for frequent data

### Behavior-Driven Testing
Use Gherkin scenarios to test requirements:
```gherkin
Scenario: Store RDF data in multiple formats
  Given a valid JSON-LD resource
  When I store the resource
  Then it should be accessible in Turtle format
  And it should be accessible in RDF/XML format
  And the semantic meaning should be preserved
```

## Design Decisions & Rationales

### 1. Pericarp Library Integration
**Decision**: Use pericarp for domain entity management
**Rationale**: Provides proven patterns for domain modeling and reduces boilerplate code while maintaining clean architecture principles

### 2. Event-Driven Architecture
**Decision**: Implement domain events for resource operations
**Rationale**: Enables loose coupling, audit trails, and future extensibility for features like real-time notifications and search indexing

### 3. File System Storage
**Decision**: Use file system as primary storage backend
**Rationale**: Provides direct access, simplicity, and aligns with Solid pod requirements for user data ownership

### 4. Format Conversion Strategy
**Decision**: Convert formats on-demand rather than storing multiple copies
**Rationale**: Reduces storage overhead while maintaining semantic integrity; acceptable performance trade-off for typical usage patterns

### 5. Streaming Support
**Decision**: Implement streaming for large file operations
**Rationale**: Essential for handling large binary files efficiently and meeting performance requirements

### 6. Checksum Validation
**Decision**: Generate and verify checksums for all stored data
**Rationale**: Ensures data integrity and enables corruption detection as required by the specifications
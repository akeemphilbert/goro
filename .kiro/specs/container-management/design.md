# Design Document

## Overview

The Container Management system extends the existing LDP server architecture to support hierarchical organization of resources through LDP BasicContainer implementation. This design builds upon the current clean architecture with domain/application/infrastructure layers, adding container-specific functionality while maintaining compatibility with existing resource storage and streaming capabilities.

The system implements the W3C Linked Data Platform (LDP) BasicContainer specification, enabling clients to create, navigate, and manage hierarchical collections of resources. Containers serve as organizational units that can contain both RDF resources and binary files, with automatic membership management and efficient querying capabilities.

## Architecture

### Core Components

The container management system integrates with the existing architecture through these key components:

1. **Container Domain Entity** - Extends the current Resource model to support container-specific behavior
2. **Container Repository** - Implements hierarchical storage with membership tracking
3. **Container Service** - Orchestrates container operations with LDP compliance
4. **Container HTTP Handlers** - Provides LDP-compliant REST endpoints
5. **Membership Indexer** - Maintains efficient container-resource relationships

### Integration Points

The design leverages existing infrastructure:
- **StreamingResourceRepository** - Extended to support container metadata
- **StorageService** - Enhanced with container-aware operations  
- **Event System** - Container lifecycle events for audit and integration
- **RDF Converter** - Container metadata serialization in multiple formats
- **HTTP Middleware** - Content negotiation for container representations

## Components and Interfaces

### Domain Layer (`internal/ldp/domain/`)

#### Container Entity
```go
type Container struct {
    *Resource                    // Inherits from existing Resource
    Members      []string        // Resource IDs contained in this container
    ParentID     string         // Parent container ID (empty for root)
    ContainerType ContainerType  // BasicContainer, DirectContainer, etc.
}

type ContainerType string
const (
    BasicContainer  ContainerType = "BasicContainer"
    DirectContainer ContainerType = "DirectContainer"
)
```

#### Container Repository Interface
```go
type ContainerRepository interface {
    ResourceRepository                    // Inherits base repository operations
    
    // Container-specific operations
    CreateContainer(ctx context.Context, container *Container) error
    AddMember(ctx context.Context, containerID, memberID string) error
    RemoveMember(ctx context.Context, containerID, memberID string) error
    ListMembers(ctx context.Context, containerID string, pagination PaginationOptions) ([]string, error)
    GetChildren(ctx context.Context, containerID string) ([]*Container, error)
    GetParent(ctx context.Context, containerID string) (*Container, error)
    
    // Hierarchy navigation
    GetPath(ctx context.Context, containerID string) ([]string, error)
    FindByPath(ctx context.Context, path string) (*Container, error)
}
```

#### Container Events
```go
const (
    EventTypeContainerCreated     = "container_created"
    EventTypeContainerUpdated     = "container_updated" 
    EventTypeContainerDeleted     = "container_deleted"
    EventTypeMemberAdded          = "member_added"
    EventTypeMemberRemoved        = "member_removed"
)
```

### Application Layer (`internal/ldp/application/`)

#### Container Service
```go
type ContainerService struct {
    containerRepo    ContainerRepository
    resourceRepo     StreamingResourceRepository
    membershipIndex  MembershipIndexer
    eventDispatcher  EventDispatcher
    unitOfWorkFactory UnitOfWorkFactory
}

// Core operations
func (s *ContainerService) CreateContainer(ctx context.Context, id, parentID string) (*Container, error)
func (s *ContainerService) GetContainer(ctx context.Context, id string, acceptFormat string) (*Container, error)
func (s *ContainerService) DeleteContainer(ctx context.Context, id string) error
func (s *ContainerService) AddResource(ctx context.Context, containerID, resourceID string) error
func (s *ContainerService) RemoveResource(ctx context.Context, containerID, resourceID string) error

// Navigation and discovery
func (s *ContainerService) ListContainerMembers(ctx context.Context, containerID string, pagination PaginationOptions) (*ContainerListing, error)
func (s *ContainerService) GetContainerPath(ctx context.Context, containerID string) ([]string, error)
func (s *ContainerService) FindContainerByPath(ctx context.Context, path string) (*Container, error)
```

#### Membership Indexer
```go
type MembershipIndexer interface {
    IndexMembership(ctx context.Context, containerID, memberID string) error
    RemoveMembership(ctx context.Context, containerID, memberID string) error
    GetMembers(ctx context.Context, containerID string, pagination PaginationOptions) ([]MemberInfo, error)
    GetContainers(ctx context.Context, memberID string) ([]string, error)
    RebuildIndex(ctx context.Context) error
}

type MemberInfo struct {
    ID           string
    Type         ResourceType  // Container or Resource
    ContentType  string
    Size         int64
    CreatedAt    time.Time
    UpdatedAt    time.Time
}
```

### Infrastructure Layer (`internal/ldp/infrastructure/`)

#### Container Repository Implementation
```go
type FileSystemContainerRepository struct {
    *FileSystemRepository        // Inherits base filesystem operations
    indexer    MembershipIndexer
    basePath   string
}

// Implements hierarchical storage using filesystem directories
// Container metadata stored as .container files
// Membership tracked in SQLite index for performance
```

#### SQLite Membership Indexer
```go
type SQLiteMembershipIndexer struct {
    db *sql.DB
}

// Tables:
// - containers (id, parent_id, type, created_at, updated_at)
// - memberships (container_id, member_id, member_type, created_at)
// - container_metadata (container_id, key, value)
```

### Transport Layer (`internal/infrastructure/transport/http/`)

#### Container HTTP Handlers
```go
type ContainerHandler struct {
    containerService *ContainerService
    resourceService  *StorageService
}

// LDP-compliant endpoints:
// GET    /containers/{id}           - Retrieve container with members
// POST   /containers/{id}           - Create resource in container  
// PUT    /containers/{id}           - Update container metadata
// DELETE /containers/{id}           - Delete container
// HEAD   /containers/{id}           - Container metadata headers
// OPTIONS /containers/{id}          - Supported operations
```

## Data Models

### Container RDF Representation

Containers are represented as RDF resources with LDP-specific properties:

```turtle
@prefix ldp: <http://www.w3.org/ns/ldp#> .
@prefix dcterms: <http://purl.org/dc/terms/> .

<http://example.org/container/documents> a ldp:BasicContainer ;
    dcterms:title "Documents Container" ;
    dcterms:created "2025-09-13T10:00:00Z" ;
    dcterms:modified "2025-09-13T10:00:00Z" ;
    ldp:contains <http://example.org/resource/doc1> ,
                 <http://example.org/resource/doc2> ,
                 <http://example.org/container/images> .
```

### Container Metadata Schema

```json
{
  "id": "container-id",
  "type": "BasicContainer", 
  "parentId": "parent-container-id",
  "title": "Container Title",
  "description": "Container Description",
  "createdAt": "2025-09-13T10:00:00Z",
  "updatedAt": "2025-09-13T10:00:00Z",
  "memberCount": 42,
  "totalSize": 1048576,
  "members": [
    {
      "id": "resource-1",
      "type": "Resource",
      "contentType": "text/turtle",
      "size": 1024,
      "createdAt": "2025-09-13T10:00:00Z"
    }
  ]
}
```

### Database Schema

```sql
-- Container hierarchy and metadata
CREATE TABLE containers (
    id TEXT PRIMARY KEY,
    parent_id TEXT REFERENCES containers(id),
    type TEXT NOT NULL DEFAULT 'BasicContainer',
    title TEXT,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Container membership relationships
CREATE TABLE memberships (
    container_id TEXT NOT NULL REFERENCES containers(id),
    member_id TEXT NOT NULL,
    member_type TEXT NOT NULL, -- 'Container' or 'Resource'
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (container_id, member_id)
);

-- Indexes for efficient queries
CREATE INDEX idx_containers_parent ON containers(parent_id);
CREATE INDEX idx_memberships_container ON memberships(container_id);
CREATE INDEX idx_memberships_member ON memberships(member_id);
```

## Error Handling

### Container-Specific Errors

```go
var (
    ErrContainerNotFound     = NewDomainError("CONTAINER_NOT_FOUND", "container not found")
    ErrContainerNotEmpty     = NewDomainError("CONTAINER_NOT_EMPTY", "container contains resources")
    ErrCircularReference     = NewDomainError("CIRCULAR_REFERENCE", "circular container reference")
    ErrInvalidHierarchy      = NewDomainError("INVALID_HIERARCHY", "invalid container hierarchy")
    ErrMembershipConflict    = NewDomainError("MEMBERSHIP_CONFLICT", "membership already exists")
    ErrInvalidContainerType  = NewDomainError("INVALID_CONTAINER_TYPE", "unsupported container type")
)
```

### Error Context and Recovery

- **Hierarchy Validation** - Prevent circular references during container creation
- **Membership Consistency** - Ensure referential integrity between containers and resources
- **Concurrent Access** - Handle race conditions in membership updates
- **Index Corruption** - Automatic index rebuilding on detection of inconsistencies

## Testing Strategy

### BDD Scenarios

Container functionality will be tested using Gherkin scenarios covering:

1. **Container Creation** - Basic container creation with hierarchy validation
2. **Resource Management** - Adding/removing resources from containers
3. **Navigation** - Path resolution and breadcrumb generation
4. **LDP Compliance** - Standard LDP operations and responses
5. **Performance** - Large container handling and pagination
6. **Concurrency** - Concurrent container operations

### Test Structure

```
features/
├── container_creation.feature
├── container_membership.feature  
├── container_navigation.feature
├── container_ldp_compliance.feature
├── container_performance.feature
└── container_concurrency.feature
```

### Integration Testing

- **End-to-End Workflows** - Complete container lifecycle testing
- **HTTP API Testing** - LDP-compliant endpoint validation
- **Database Integration** - Membership index consistency testing
- **Event Processing** - Container event handling verification
- **Performance Testing** - Large container and deep hierarchy handling

### Unit Testing

- **Domain Logic** - Container entity behavior and validation
- **Repository Operations** - Storage and retrieval functionality
- **Service Orchestration** - Business logic coordination
- **Index Management** - Membership tracking accuracy
- **Format Conversion** - RDF serialization in multiple formats

The testing approach ensures comprehensive coverage of container functionality while maintaining compatibility with existing resource management features.
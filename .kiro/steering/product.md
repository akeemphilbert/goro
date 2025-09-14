---
inclusion: always
---

# LDP Server Product Guidelines

**Goro** is a Linked Data Platform (LDP) server built in Go with clean architecture, Google Wire DI, and comprehensive user management. This document guides AI assistants working with the codebase.

## Core Domain Concepts

### LDP Resources & Containers
- **Resources**: RDF data in Turtle (.ttl), JSON-LD (.jsonld), RDF/XML (.rdf) formats
- **Containers**: Collections of resources with membership management
- **Binary Files**: Streamed via `io.Reader`/`io.Writer`, never loaded to memory
- **Content Negotiation**: Automatic format conversion based on HTTP Accept headers

### User Management System
- **Users**: Core entities with WebID identifiers and profile management
- **Accounts**: Multi-user organizations with role-based access control
- **Roles**: Permission sets (Owner, Admin, Member, Viewer) with inheritance
- **Invitations**: Workflow for account membership with expiration and validation

## Mandatory Code Patterns

### Layer Separation (STRICT)
```go
// Domain layer - NO external imports except standard library
type User struct { /* pure business logic */ }
type UserRepository interface { /* contracts only */ }

// Application layer - orchestrates domain
type UserService struct { /* use cases */ }

// Infrastructure layer - external concerns
type FileUserRepository struct { /* implements UserRepository */ }
```

### Event-Driven Architecture
- Domain events for all state changes: `UserCreated`, `AccountMemberAdded`, `ResourceUpdated`
- Event handlers in application layer for side effects
- Event persistence for audit trails and system integration

### Repository Pattern
- Interfaces defined in `internal/*/domain/repository.go`
- Read repositories for queries, Write repositories for mutations
- Filesystem-based implementations with atomic operations
- Caching layer for performance optimization

### Wire Dependency Injection
- Provider functions in `wire.go` files at each layer
- **CRITICAL**: Run `wire ./cmd/server` after ANY provider changes
- Group related dependencies in provider sets
- Never edit `wire_gen.go` files manually

## Testing Requirements

### BDD-First Development
1. Write Gherkin scenarios in `/features/*.feature` BEFORE implementation
2. Implement step definitions in `/features/*_test.go`
3. Use real filesystem with temporary directories for integration tests
4. Clean up resources with `defer` statements

### Test Categories
- **Unit Tests**: Mock dependencies, test single components
- **Integration Tests**: Real storage, test component interactions
- **Performance Tests**: Validate streaming, memory usage, concurrency
- **BDD Tests**: End-to-end scenarios matching business requirements

## File Organization Rules

### Domain Boundaries
```
internal/
├── user/           # User management domain
├── ldp/            # LDP resource management domain
└── infrastructure/ # Shared transport layer
```

### Naming Conventions
- **Services**: `UserService`, `AccountService` (application layer)
- **Repositories**: `UserRepository`, `AccountRepository` (domain interfaces)
- **Events**: Past tense - `UserCreated`, `AccountMemberRemoved`
- **Handlers**: `UserEventHandler`, `AccountEventHandler`

## Performance Guidelines
- Stream large files with `io.Copy()`, never `ioutil.ReadAll()`
- Use caching for frequently accessed data
- Implement pagination for large result sets
- Optimize database queries with proper indexing
- Use goroutines for concurrent operations with proper context handling

## Error Handling Standards
- Wrap errors with context: `fmt.Errorf("failed to create user: %w", err)`
- Define domain-specific errors in domain layer
- Return structured errors from HTTP handlers
- Log errors with appropriate levels and context
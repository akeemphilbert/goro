---
inclusion: always
---

# LDP Server Project Structure

## Directory Layout
This LDP (Linked Data Platform) server follows clean architecture with strict layer separation:

```
goro/
├── cmd/server/             # Main application entry point with Wire DI
├── configs/                # YAML configuration files
├── features/               # BDD test scenarios (Gherkin) and step definitions
├── internal/               # Private application code (clean architecture layers)
│   ├── conf/              # Configuration loading and validation
│   ├── infrastructure/    # External concerns (HTTP transport, middleware)
│   └── ldp/               # LDP domain implementation
│       ├── application/   # Use cases and application services
│       ├── domain/        # Pure business logic (NO external dependencies)
│       ├── infrastructure/# Repositories, RDF converters, event storage
│       └── integration/   # End-to-end integration tests
├── web/                   # Client applications (admin UI, browser extension)
├── go.mod & go.sum        # Go module dependencies
└── server                 # Compiled binary
```

## Clean Architecture Layers (STRICT)

### Domain Layer (`internal/ldp/domain/`)
- **Pure business logic** - NO external package imports allowed
- **Entities**: Resource, Event domain objects
- **Interfaces**: Repository contracts defined here, implemented elsewhere
- **Domain Events**: Business events for audit and integration
- **Domain Errors**: Business-specific error types

### Application Layer (`internal/ldp/application/`)
- **Use Cases**: StorageService orchestrating domain operations
- **Event Handlers**: Processing domain events for side effects
- **Application Services**: Coordinating multiple domain operations
- **Wire Providers**: Dependency injection configuration

### Infrastructure Layer (`internal/ldp/infrastructure/`)
- **Repository Implementations**: FilesystemRepository, OptimizedFilesystemRepository
- **External Services**: RDF converters, event dispatchers, caching
- **Database**: SQLite for event storage and resource indexing
- **Performance Optimizations**: Resource cache, indexer

### Transport Layer (`internal/infrastructure/transport/http/`)
- **HTTP Server**: Native Go HTTP server with middleware chain
- **Handlers**: Resource CRUD operations, health checks, error handling
- **Middleware**: CORS, logging, timeout, content negotiation
- **Protocol Compliance**: Full LDP specification support

## File Organization Rules

### Wire Dependency Injection
- **Wire providers** in `wire.go` files at each layer
- **ALWAYS run** `wire ./cmd/server` after provider changes
- **Generated code** in `wire_gen.go` (never edit manually)

### Testing Structure
- **BDD scenarios** in `/features/*.feature` (Gherkin syntax)
- **Step definitions** in `/features/*_test.go`
- **Unit tests** alongside source files (`*_test.go`) with a package name suffixed with _test
- **Integration tests** use temporary directories and in-memory SQLite
- **Performance tests** validate streaming and memory efficiency

### Configuration Management
- **Main config** in `configs/config.yaml`
- **TLS example** in `configs/https-example.yaml`
- **Config struct** in `internal/conf/conf.go`
- **Environment overrides** supported

## Naming Conventions
- **Packages**: lowercase, domain-focused names
- **Interfaces**: Repository, Service, Handler suffixes
- **Implementations**: Concrete types in infrastructure layer
- **Events**: Past tense (ResourceCreated, ResourceUpdated)
- **Errors**: Domain-specific with context wrapping
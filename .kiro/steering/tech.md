---
inclusion: always
---

---
inclusion: always
---

# LDP Server Technology Stack & Development Guidelines

## Core Technologies & Versions
- **Go 1.23.3** - Primary language with strict clean architecture
- **Google Wire v0.6.0** - Compile-time dependency injection
- **SQLite** - Event storage and resource indexing via `database/sql`
- **File System** - Primary storage with atomic operations
- **Native HTTP Server** - Go's `net/http` with middleware chain

## Architecture Layers (STRICT Separation)
1. **Domain** (`internal/ldp/domain/`) - Pure business logic, ZERO external imports
2. **Application** (`internal/ldp/application/`) - Use cases orchestrating domain
3. **Infrastructure** (`internal/ldp/infrastructure/`) - External concerns (DB, filesystem, RDF)
4. **Transport** (`internal/infrastructure/transport/http/`) - HTTP protocol handling

## Critical Wire Dependency Injection Rules
- **ALWAYS run `wire ./cmd/server`** after ANY provider changes
- Wire providers in `wire.go` files, never edit `wire_gen.go`
- Group related dependencies in provider sets
- Use `wire.Build()` to compose dependency graphs

## LDP Protocol Implementation
- **RDF Formats**: Turtle (.ttl), JSON-LD (.jsonld), RDF/XML (.rdf)
- **Content Negotiation**: Middleware processes Accept headers before handlers
- **Binary Files**: Stream via `io.Reader`/`io.Writer`, metadata in SQLite
- **Resource Indexing**: SQLite with cache layer for performance
- **Event System**: Domain events persisted for audit/replay

## Mandatory Code Patterns
- **Repository Pattern**: Interfaces in domain, implementations in infrastructure
- **Error Wrapping**: `fmt.Errorf("operation failed: %w", err)` at each boundary
- **Resource Cleanup**: Always `defer` file/DB connection cleanup
- **Streaming I/O**: NEVER `ioutil.ReadAll()` for large files
- **HTTP Methods**: Full LDP compliance (GET, POST, PUT, DELETE, HEAD, OPTIONS)

## Testing Strategy (BDD-First)
1. **Write Gherkin scenarios** in `/features/*.feature` BEFORE coding
2. **Step definitions** in `/features/*_test.go`
3. **Unit tests** with mocks for each layer
4. **Integration tests** use temp directories and in-memory SQLite
5. **Performance tests** validate streaming and memory efficiency

## Essential Development Commands
```bash
# CRITICAL: Run after provider changes
wire ./cmd/server

# Build and run server
go build ./cmd/server && ./server

# Run all tests
go test ./...

# BDD feature tests
go test ./features/

# Custom configuration
./server -conf ./configs/config.yaml
```

## AI Assistant Critical Rules
- **Wire Regeneration**: MANDATORY after any provider modification
- **Layer Boundaries**: Domain layer imports ONLY standard library
- **Streaming Operations**: Use `io.Copy()`, never load large files to memory
- **Repository Interfaces**: Define in domain, implement in infrastructure
- **Error Context**: Wrap errors with operation context at each layer
- **BDD First**: Write feature scenarios before implementation
- **Test Isolation**: Use temporary storage, clean up with `defer`
- **Content Types**: Validate RDF media types in middleware layer
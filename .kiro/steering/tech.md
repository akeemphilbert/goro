---
inclusion: always
---

# LDP Server Technical Guidelines

## Technology Stack
- **Go 1.23.3** with clean architecture
- **Google Wire v0.6.0** for dependency injection
- **SQLite** for events and indexing
- **Filesystem** for primary storage
- **Native HTTP** with middleware chain

## Architecture Layers (STRICT)
```
internal/
├── {domain}/domain/        # Pure business logic (NO external imports)
├── {domain}/application/   # Use cases and orchestration
├── {domain}/infrastructure/# External systems (DB, filesystem, RDF)
└── infrastructure/transport/# HTTP handlers and middleware
```

## Critical Wire DI Rules
- **MANDATORY**: Run `wire ./cmd/server` after ANY provider changes
- Define providers in `wire.go`, NEVER edit `wire_gen.go`
- Group dependencies in provider sets with `wire.NewSet()`

## Code Patterns (REQUIRED)
```go
// Repository interfaces in domain
type UserRepository interface { Save(User) error }

// Implementations in infrastructure  
type FileUserRepository struct {}

// Error wrapping at boundaries
return fmt.Errorf("save user failed: %w", err)

// Streaming for large files
io.Copy(dst, src) // NOT ioutil.ReadAll()

// Resource cleanup
defer file.Close()
```

## Testing Requirements
1. **BDD First**: Write `.feature` files before implementation
2. **Step definitions**: Implement in `*_test.go` files
3. **Isolation**: Use temp directories, clean with `defer`
4. **Performance**: Validate streaming and memory usage

## Development Workflow
```bash
# After provider changes (CRITICAL)
wire ./cmd/server

# Build and test
go build ./cmd/server && go test ./...

# Run BDD tests
go test ./features/
```

## AI Assistant Rules
- **Wire regeneration** is MANDATORY after provider modifications
- **Domain purity**: Only standard library imports in domain layer
- **Stream large data**: Use `io.Copy()`, never load to memory
- **Repository pattern**: Interfaces in domain, implementations in infrastructure
- **Error context**: Wrap with operation details at each boundary
- **Test-driven**: Write failing tests first, then implement
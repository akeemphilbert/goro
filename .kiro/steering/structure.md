# Project Structure

## Directory Layout
The project follows Go-Kratos conventions with clean architecture principles:

```
goro/
├── api/                    # Protocol Buffer definitions and generated code
├── cmd/                    # Application entry points and main packages
├── configs/                # Configuration files (YAML, JSON, etc.)
├── internal/               # Private application code (not importable)
│   ├── application/        # Application layer (use cases, services)
│   ├── domain/            # Domain layer (entities, value objects, interfaces)
│   └── infrastructure/    # Infrastructure layer (repositories, external services)
├── pkg/                   # Public packages (importable by other projects)
├── web/                   # Web assets, static files, templates
├── go.mod                 # Go module definition
└── go.sum                 # Go module checksums
```

## Architecture Layers

### Internal Package Structure
- **Domain Layer** (`internal/domain/`) - Core business logic, entities, and domain interfaces
- **Application Layer** (`internal/application/`) - Use cases, application services, and orchestration
- **Infrastructure Layer** (`internal/infrastructure/`) - External dependencies, databases, HTTP clients

### Public Interfaces
- **API** (`api/`) - Protocol Buffer definitions for gRPC services
- **CMD** (`cmd/`) - Application entry points and CLI commands
- **PKG** (`pkg/`) - Reusable packages that can be imported by other projects

## Naming Conventions
- Use lowercase package names
- Follow Go naming conventions (PascalCase for exported, camelCase for unexported)
- Domain entities should be in singular form
- Repository interfaces in domain, implementations in infrastructure
- Use descriptive names that reflect business concepts

## File Organization
- Group related functionality in packages
- Keep interfaces close to their usage
- Place Wire providers in dedicated files (`wire.go`)
- Configuration structs in `configs/` directory
- Tests alongside source files (`*_test.go`)
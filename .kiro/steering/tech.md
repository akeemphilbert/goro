# Technology Stack

## Core Framework & Language
- **Go 1.23.3** - Primary programming language
- **Kratos v2.8.0** - Microservice framework for Go
- **gRPC** - Remote procedure call framework
- **Protocol Buffers** - Interface definition and serialization

## Key Dependencies
- **Google Wire v0.6.0** - Dependency injection code generation
- **Uber AutoMaxProcs v1.5.1** - Automatic GOMAXPROCS configuration
- **Google API Proto** - Google API protocol buffer definitions

## Build System
- **Go Modules** - Dependency management (go.mod/go.sum)
- Standard Go toolchain for building and testing

## Common Commands
```bash
# Build the application
go build ./cmd/...

# Run tests
go test ./...

# Generate Wire dependency injection
wire ./...

# Generate Protocol Buffer files
protoc --go_out=. --go-grpc_out=. api/**/*.proto

# Tidy dependencies
go mod tidy

# Download dependencies
go mod download

# Run with auto CPU configuration
go run ./cmd/server
```

## Development Tools
- Wire for compile-time dependency injection
- Protocol Buffer compiler for API generation
- Standard Go testing framework
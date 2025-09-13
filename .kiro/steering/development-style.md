---
inclusion: always
---

# Development Style Guidelines

## Test-Driven Development (TDD)
- Follow TDD principles: write failing tests first, then implement to make them pass
- Use Cucumber framework with Gherkin syntax for behavior-driven requirements
- Write unit tests before implementation for all components
- Ensure tests are isolated, fast, and deterministic

## Code Style & Conventions
- Follow standard Go formatting with `gofmt` and `goimports`
- Use meaningful variable and function names that reflect business domain
- Keep functions small and focused on single responsibility
- Prefer composition over inheritance
- Use interfaces for abstraction, especially in domain layer

## Architecture Patterns
- Implement clean architecture with strict layer separation
- Domain layer should have no external dependencies
- Use dependency injection via Google Wire for loose coupling
- Repository pattern for data access with interfaces in domain layer
- Service layer for business logic orchestration

## Error Handling
- Use Go's idiomatic error handling with explicit error returns
- Wrap errors with context using `fmt.Errorf` or error wrapping libraries
- Define custom error types for domain-specific errors
- Handle errors at appropriate levels, don't ignore them

## gRPC & Protocol Buffers
- Define clear, versioned API contracts in `.proto` files
- Use semantic versioning for API changes
- Include proper field validation and documentation
- Generate code using `protoc` with Go plugins

## Dependency Management
- Keep `go.mod` clean and up-to-date with `go mod tidy`
- Pin major versions for stability
- Regularly update dependencies for security patches
- Use `go mod vendor` for reproducible builds when needed

## Performance & Concurrency
- Use goroutines and channels appropriately for concurrent operations
- Implement proper context cancellation for request timeouts
- Profile code for performance bottlenecks when needed
- Follow Go's memory management best practices

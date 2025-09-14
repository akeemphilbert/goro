# Container Integration Testing Guide

This document provides comprehensive guidance for testing container functionality in the LDP server.

## Overview

The container integration test suite covers all aspects of container functionality including:

- **BDD Feature Tests** - Behavior-driven tests using Gherkin scenarios
- **End-to-End Integration Tests** - Complete workflow testing
- **Performance Tests** - Load and performance validation
- **Concurrency Tests** - Race condition and concurrent access testing
- **Event Processing Tests** - Event emission and handling validation

## Test Structure

### BDD Feature Files

Located in `features/` directory:

- `container_creation.feature` - Container creation and hierarchy management
- `container_membership.feature` - Resource membership operations
- `container_http_api.feature` - HTTP API compliance testing
- `container_performance.feature` - Performance requirements validation
- `container_concurrency.feature` - Concurrent operation testing
- `container_events.feature` - Event processing validation

### Test Implementation Files

- `container_integration_test.go` - BDD step definitions and test context
- `container_end_to_end_test.go` - Complete workflow integration tests
- `container_performance_integration_test.go` - Performance and load tests

## Running Tests

### Quick Start

```bash
# Run all container tests
./features/run_container_tests.sh all

# Run specific test categories
./features/run_container_tests.sh bdd      # BDD feature tests only
./features/run_container_tests.sh e2e      # End-to-end tests only
./features/run_container_tests.sh perf     # Performance tests only
./features/run_container_tests.sh unit     # Unit tests only
```

### Individual Test Execution

```bash
# BDD tests
go test -v ./features -run TestContainerCreation
go test -v ./features -run TestContainerMembership
go test -v ./features -run TestContainerHTTPAPI
go test -v ./features -run TestContainerPerformance
go test -v ./features -run TestContainerConcurrency
go test -v ./features -run TestContainerEvents

# Integration tests
go test -v ./features -run TestContainerEndToEnd

# Performance tests (with extended timeout)
go test -v ./features -run TestContainerPerformance -timeout 10m
```

## Test Categories

### 1. BDD Feature Tests

**Purpose**: Validate that container functionality meets business requirements using natural language scenarios.

**Coverage**:
- Container creation and validation
- Hierarchical container structures
- Resource membership management
- LDP protocol compliance
- HTTP API endpoints
- Performance requirements
- Concurrent operations
- Event processing

**Example Scenario**:
```gherkin
Scenario: Create nested container hierarchy
  Given a container "documents" exists
  When I create a container "images" inside "documents"
  Then the container "images" should be created successfully
  And the container "images" should have parent "documents"
  And the container "documents" should contain "images"
```

### 2. End-to-End Integration Tests

**Purpose**: Test complete workflows from HTTP requests to data persistence.

**Coverage**:
- Complete container lifecycle (create, update, delete)
- Resource management within containers
- HTTP API compliance with real server
- Cross-layer integration validation
- Error handling and recovery

**Key Tests**:
- `TestCompleteContainerLifecycle` - Full CRUD operations
- `TestContainerHTTPAPICompliance` - HTTP protocol validation
- `TestContainerConcurrency` - Concurrent operation safety
- `TestContainerPerformance` - Basic performance validation

### 3. Performance Tests

**Purpose**: Validate performance requirements and identify bottlenecks.

**Coverage**:
- Large container operations (1K-10K resources)
- Pagination performance
- Concurrent access patterns
- Deep hierarchy navigation
- Memory usage under load

**Performance Targets**:
- Container listing: < 5 seconds for 10K resources
- Pagination: < 1 second for any page
- Concurrent operations: > 50 ops/sec
- Memory usage: Constant with streaming

**Key Metrics**:
- Response times
- Throughput (operations/second)
- Memory usage patterns
- Error rates under load

### 4. Concurrency Tests

**Purpose**: Ensure data integrity and performance under concurrent access.

**Coverage**:
- Concurrent container creation
- Simultaneous resource additions/removals
- Race condition prevention
- Membership index consistency
- Event ordering

**Test Scenarios**:
- Multiple clients creating containers simultaneously
- Bulk resource operations from different goroutines
- Container deletion while resources are being added
- Membership index consistency under load

### 5. Event Processing Tests

**Purpose**: Validate event emission and processing for audit and integration.

**Coverage**:
- Event emission for all container operations
- Event ordering and consistency
- Event handler processing
- Event replay capability
- Event filtering and querying

**Event Types Tested**:
- `container_created`
- `container_updated`
- `container_deleted`
- `member_added`
- `member_removed`

## Test Data Management

### Temporary Storage

All tests use temporary directories that are automatically cleaned up:

```go
tempDir, err := os.MkdirTemp("", "container-test-*")
// ... test execution ...
defer os.RemoveAll(tempDir)
```

### Test Isolation

Each test scenario runs in isolation with:
- Fresh temporary storage
- Clean database state
- Independent server instances
- Separate event streams

### Resource Cleanup

Automatic cleanup ensures:
- No test data persistence
- No resource leaks
- Clean state for subsequent tests

## Performance Benchmarks

### Expected Performance Characteristics

| Operation | Target Performance | Test Validation |
|-----------|-------------------|-----------------|
| Container Creation | < 10ms | ✓ |
| Resource Addition | < 5ms | ✓ |
| Container Listing (1K items) | < 100ms | ✓ |
| Container Listing (10K items) | < 5s | ✓ |
| Pagination (any page) | < 1s | ✓ |
| Path Resolution (20 levels) | < 100ms | ✓ |
| Concurrent Operations | > 50 ops/sec | ✓ |

### Load Testing Parameters

- **Small Load**: 100-500 resources
- **Medium Load**: 1K-5K resources  
- **Large Load**: 10K+ resources
- **Concurrent Users**: 10-50 simultaneous operations
- **Deep Hierarchy**: Up to 20 levels

## Troubleshooting

### Common Test Failures

1. **Timeout Errors**
   - Increase test timeout for performance tests
   - Check system resources (CPU, memory, disk)
   - Verify no resource leaks

2. **Concurrency Issues**
   - Check for race conditions in test setup
   - Ensure proper synchronization in test code
   - Verify database connection limits

3. **Memory Issues**
   - Monitor memory usage during large tests
   - Check for resource leaks in cleanup
   - Verify streaming implementation

### Debug Mode

Enable verbose logging for debugging:

```bash
# Run with verbose output
go test -v ./features -run TestContainer

# Run with race detection
go test -race ./features -run TestContainer

# Run with memory profiling
go test -memprofile=mem.prof ./features -run TestContainer
```

### Test Environment

Ensure test environment has:
- Sufficient disk space for temporary files
- Adequate memory for large container tests
- No conflicting processes on test ports

## Contributing

### Adding New Tests

1. **BDD Scenarios**: Add to appropriate `.feature` file
2. **Step Definitions**: Implement in `container_integration_test.go`
3. **Integration Tests**: Add to `container_end_to_end_test.go`
4. **Performance Tests**: Add to `container_performance_integration_test.go`

### Test Naming Conventions

- BDD scenarios: Use natural language descriptions
- Test functions: `Test[Component][Functionality]`
- Helper functions: `[action][Object]` (e.g., `createContainer`)

### Documentation

Update this guide when:
- Adding new test categories
- Changing performance targets
- Modifying test structure
- Adding new test utilities

## Requirements Coverage

This test suite validates all container management requirements:

### Requirement 1.5 - Container Validation
- ✓ Circular reference detection
- ✓ Container hierarchy validation
- ✓ Error handling and recovery

### Requirement 2.5 - LDP Compliance
- ✓ HTTP method support
- ✓ Content negotiation
- ✓ Error response formats
- ✓ LDP header compliance

### Requirement 3.4 - Performance
- ✓ Large container handling
- ✓ Pagination performance
- ✓ Memory efficiency
- ✓ Response time targets

### Requirement 4.5 - Metadata Management
- ✓ Timestamp accuracy
- ✓ Dublin Core properties
- ✓ Metadata corruption detection
- ✓ Recovery mechanisms

### Requirement 5.5 - Discovery and Navigation
- ✓ Path resolution accuracy
- ✓ Breadcrumb generation
- ✓ Error handling
- ✓ Performance under load

## Continuous Integration

### CI Pipeline Integration

```yaml
# Example GitHub Actions workflow
- name: Run Container Tests
  run: |
    ./features/run_container_tests.sh all
    
- name: Run Performance Tests
  run: |
    ./features/run_container_tests.sh perf
  timeout-minutes: 15
```

### Test Reporting

Tests generate reports in multiple formats:
- Console output with colored status
- JUnit XML for CI integration
- Performance metrics for monitoring
- Coverage reports for code quality
# Behavior-Driven Development Tests

This directory contains BDD tests for the Resource Storage System using Gherkin scenarios and Go test implementations.

## Overview

The BDD tests cover all major requirements of the Resource Storage System:

1. **RDF Format Support** - Testing multiple RDF serialization formats (Turtle, JSON-LD, RDF/XML)
2. **Binary File Storage** - Testing storage and retrieval of binary files with integrity verification
3. **Error Handling** - Testing comprehensive error scenarios and appropriate responses
4. **Performance Requirements** - Testing response times, streaming, and concurrent access
5. **Data Integrity** - Testing data consistency, corruption detection, and atomic operations

## Test Structure

### Feature Files (`.feature`)
- `rdf_format_support.feature` - RDF format scenarios
- `binary_file_storage.feature` - Binary file storage scenarios  
- `error_handling.feature` - Error handling scenarios
- `performance_requirements.feature` - Performance testing scenarios
- `data_integrity.feature` - Data integrity scenarios

### Test Implementation Files (`*_test.go`)
- `step_definitions_test.go` - Core BDD test context and step definitions
- `rdf_format_support_test.go` - RDF format test implementations
- `binary_file_storage_test.go` - Binary file test implementations
- `error_handling_test.go` - Error handling test implementations
- `performance_requirements_test.go` - Performance test implementations
- `data_integrity_test.go` - Data integrity test implementations

## Running the Tests

### Run All BDD Tests
```bash
go test ./features/...
```

### Run Specific Test Categories
```bash
# RDF Format Support Tests
go test ./features/ -run TestRDFFormatSupport

# Binary File Storage Tests
go test ./features/ -run TestBinaryFileStorage

# Error Handling Tests
go test ./features/ -run TestErrorHandling

# Performance Tests
go test ./features/ -run TestPerformanceRequirements

# Data Integrity Tests
go test ./features/ -run TestDataIntegrity
```

### Run with Verbose Output
```bash
go test ./features/... -v
```

## Test Coverage

The BDD tests cover the following requirements:

### Requirement 1.1 - RDF Format Support
- ✅ Store and retrieve Turtle format
- ✅ Store and retrieve JSON-LD format  
- ✅ Store and retrieve RDF/XML format
- ✅ Content negotiation between formats
- ✅ Semantic meaning preservation

### Requirement 1.2 - Content Negotiation
- ✅ Accept header processing
- ✅ Format conversion on demand
- ✅ Error handling for unsupported formats

### Requirement 1.3 - Format Conversion
- ✅ Conversion between all supported formats
- ✅ Semantic integrity preservation
- ✅ Conversion error handling

### Requirement 2.1 - Binary File Storage
- ✅ Store binary files without modification
- ✅ Retrieve exact original content
- ✅ MIME type preservation

### Requirement 2.2 - File Integrity
- ✅ Checksum verification
- ✅ Size validation
- ✅ Corruption detection

### Requirement 4.4 - Error Handling
- ✅ Resource not found (404)
- ✅ Unsupported format (406)
- ✅ Invalid data (400)
- ✅ Storage limitations (507)

### Requirement 4.5 - Data Consistency
- ✅ Atomic operations
- ✅ Consistency during failures
- ✅ Recovery from errors

## Test Data

The tests use realistic sample data:

### RDF Test Data
- **Turtle**: Person entity with name and age properties
- **JSON-LD**: Same entity in JSON-LD format with context
- **RDF/XML**: Same entity in RDF/XML format

### Binary Test Data
- **JPEG**: Minimal JPEG header for image testing
- **PDF**: Minimal PDF header for document testing
- **Large Files**: Generated binary data for streaming tests

## Architecture

The BDD test framework uses:

- **BDDTestContext**: Central test context managing test state
- **Step Definitions**: Reusable step implementations for Gherkin scenarios
- **Test Server**: HTTP test server for integration testing
- **Temporary Storage**: Isolated file system for each test run

## Best Practices

1. **Isolation**: Each test runs in isolation with its own storage
2. **Cleanup**: Automatic cleanup of test resources
3. **Realistic Data**: Use realistic test data that matches production scenarios
4. **Error Testing**: Comprehensive error scenario coverage
5. **Performance**: Performance tests with realistic expectations

## Extending Tests

To add new BDD scenarios:

1. Add Gherkin scenarios to appropriate `.feature` file
2. Implement step definitions in `step_definitions_test.go`
3. Create test implementations in corresponding `*_test.go` file
4. Update this README with new coverage information

## Dependencies

The BDD tests depend on:
- Standard Go testing framework
- testify/assert and testify/require for assertions
- HTTP test server for integration testing
- Temporary file system for isolated testing
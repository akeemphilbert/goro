---
inclusion: always
---

# LDP Server Product Overview

**Goro** is a Linked Data Platform (LDP) server implementation built in Go, providing HTTP-based storage and retrieval of RDF resources and binary files. The project follows clean architecture principles with strict layer separation and uses Google Wire for dependency injection.

## Core Purpose
An LDP-compliant server that stores, manages, and serves RDF resources in multiple formats (Turtle, JSON-LD, RDF/XML) alongside binary files, with full content negotiation and streaming capabilities.

## Key Characteristics
- **LDP Protocol Compliance** - Full HTTP-based Linked Data Platform specification support
- **Multi-Format RDF Support** - Native handling of Turtle (.ttl), JSON-LD (.jsonld), and RDF/XML (.rdf)
- **Binary File Storage** - Streaming storage for large files with integrity verification
- **Clean Architecture** - Strict domain/application/infrastructure layer separation
- **Event-Driven Design** - Domain events for audit trails and system integration
- **Performance Optimized** - Streaming I/O, caching, and concurrent request handling
- **BDD Testing** - Comprehensive Gherkin scenarios covering all requirements
- **Web Clients** - Admin UI (Nuxt.js) and browser extension for management

## Architecture Principles
- **Domain-First Design** - Pure business logic with no external dependencies in domain layer
- **Repository Pattern** - Abstract storage interfaces with filesystem and optimized implementations
- **Content Negotiation** - Automatic format conversion based on HTTP Accept headers
- **Streaming Operations** - Never load large files into memory, use io.Reader/Writer
- **Atomic Operations** - Ensure data consistency with proper transaction boundaries
- **Error Context** - Rich error wrapping with context at each layer boundary

## Target Use Cases
- **Semantic Web Applications** - Store and serve RDF data with full LDP compliance
- **Knowledge Graphs** - Manage linked data resources with proper relationships
- **Document Management** - Handle mixed RDF metadata and binary file content
- **Research Data Platforms** - Academic and scientific data storage with provenance
- **Content Management Systems** - Semantic content with rich metadata support
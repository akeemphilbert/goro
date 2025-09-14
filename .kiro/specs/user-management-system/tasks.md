# Implementation Plan

- [x] 1. Write tests and implement user domain structure
  - [x] 1.1 Write failing tests for core domain entities
    - Create test files for User, Account, AccountMember, Role, Invitation entities
    - Write tests for entity validation, business rules, and domain methods
    - Write tests for atomic entity design and relationship constraints
    - _Requirements: 1.1, 2.1, 3.1, 5.1_

  - [x] 1.2 Implement domain entities to make tests pass
    - Create the `internal/user` package with clean architecture layers (domain/application/infrastructure)
    - Implement core domain entities with atomic design to satisfy failing tests
    - Define domain interfaces for repositories (read-only and write operations)
    - _Requirements: 1.1, 2.1, 3.1, 5.1_

- [x] 2. Write tests and implement domain events
  - [x] 2.1 Write failing tests for user lifecycle events
    - Create test files for UserRegisteredEvent, UserProfileUpdatedEvent, UserDeletedEvent, WebIDGeneratedEvent
    - Write tests for event serialization, validation, and structure
    - Test event emission and handling patterns
    - _Requirements: 5.1, 6.1_

  - [x] 2.2 Implement user events to make tests pass
    - Define UserRegisteredEvent, UserProfileUpdatedEvent, UserDeletedEvent, WebIDGeneratedEvent
    - Implement event structures with proper serialization support to satisfy tests
    - _Requirements: 5.1, 6.1_

  - [x] 2.3 Write failing tests for account and membership events
    - Create test files for AccountCreatedEvent, MemberInvitedEvent, InvitationAcceptedEvent
    - Write tests for MemberAddedEvent, MemberRemovedEvent, MemberRoleUpdatedEvent
    - Test atomic entity relationship events following RDF-like principles
    - _Requirements: 3.1, 3.2, 3.3_

  - [x] 2.4 Implement account events to make tests pass
    - Define all account and membership domain events to satisfy failing tests
    - Implement atomic entity relationship events following RDF-like principles
    - _Requirements: 3.1, 3.2, 3.3_

- [x] 3. Write tests and implement GORM models
  - [x] 3.1 Write failing tests for GORM models
    - Create test files for UserModel, RoleModel, AccountModel, AccountMemberModel, InvitationModel
    - Write tests for GORM model validation, constraints, and relationships
    - Test JSON serialization for complex fields (permissions, settings)
    - Write tests for database migration and model creation
    - _Requirements: 1.1, 3.1, 4.1, 5.1, 6.1, 7.1_

  - [x] 3.2 Implement GORM models to make tests pass
    - Create UserModel, RoleModel, AccountModel, AccountMemberModel, InvitationModel
    - Define proper GORM tags, constraints, and indexes to satisfy failing tests
    - Implement JSON serialization for complex fields to pass serialization tests
    - _Requirements: 1.1, 3.1, 4.1, 5.1, 6.1, 7.1_

  - [x] 3.3 Write failing tests for database migration and seeding
    - Write tests for MigrateUserModels function and auto-migration
    - Create tests for system role seeding (Owner, Admin, Member, Viewer) with permissions
    - Test Wire provider integration with existing GORM instance
    - _Requirements: 3.3, 4.1_

  - [x] 3.4 Implement migration and seeding to make tests pass
    - Implement MigrateUserModels function for auto-migration to satisfy tests
    - Create system role seeding with proper permissions to pass seeding tests
    - Add Wire provider for database initialization with existing GORM instance
    - _Requirements: 3.3, 4.1_

- [x] 4. Write tests and implement repository layer
  - [x] 4.1 Write failing tests for read-only repositories
    - Create test files for GormUserRepository with query method tests (GetByID, GetByWebID, GetByEmail, List)
    - Write tests for GormAccountRepository with account query scenarios
    - Create tests for GormAccountMemberRepository for membership projection queries
    - Write tests for GormRoleRepository with system role query scenarios
    - Use temporary database for repository integration tests
    - _Requirements: 1.2, 2.1, 3.1, 5.1, 7.2_

  - [x] 4.2 Implement read-only repositories to make tests pass
    - Implement GormUserRepository with query methods to satisfy failing tests
    - Implement GormAccountRepository with account queries to pass tests
    - Implement GormAccountMemberRepository for membership projections to satisfy tests
    - Implement GormRoleRepository with system role queries to pass tests
    - _Requirements: 1.2, 2.1, 3.1, 5.1, 7.2_

  - [x] 4.3 Write failing tests for write repositories
    - Create test files for GormUserWriteRepository with persistence operation tests
    - Write tests for GormAccountWriteRepository with account persistence scenarios
    - Create tests for GormAccountMemberWriteRepository for membership projection writes
    - Write tests for GormInvitationWriteRepository with invitation management scenarios
    - _Requirements: 1.1, 3.2, 3.3, 5.1, 6.1_

  - [x] 4.4 Implement write repositories to make tests pass
    - Implement GormUserWriteRepository for user persistence operations to satisfy tests
    - Implement GormAccountWriteRepository for account persistence to pass tests
    - Implement GormAccountMemberWriteRepository for membership projections to satisfy tests
    - Implement GormInvitationWriteRepository for invitation management to pass tests
    - _Requirements: 1.1, 3.2, 3.3, 5.1, 6.1_

- [x] 5. Write tests and implement WebID generation and file storage
  - [x] 5.1 Write failing tests for WebID generator
    - Create test files for WebIDGenerator interface with URI generation scenarios
    - Write tests for Turtle format WebID document generation and validation
    - Test WebID format validation and uniqueness constraints
    - _Requirements: 1.1, 5.2, 5.3_

  - [x] 5.2 Implement WebID generator to make tests pass
    - Implement WebIDGenerator interface with URI generation logic to satisfy tests
    - Create Turtle format WebID document generation to pass format tests
    - Add validation for WebID format and uniqueness to satisfy validation tests
    - _Requirements: 1.1, 5.2, 5.3_

  - [x] 5.3 Write failing tests for file storage
    - Create test files for user data file storage with temporary directories
    - Write tests for file storage structure for user profiles and WebID documents
    - Test atomic file operations for user data persistence and cleanup
    - _Requirements: 2.2, 6.3, 6.4_

  - [x] 5.4 Implement file storage to make tests pass
    - Create file storage structure for user profiles and WebID documents to satisfy tests
    - Implement atomic file operations for user data persistence to pass operation tests
    - Add file cleanup for user deletion operations to satisfy cleanup tests
    - _Requirements: 2.2, 6.3, 6.4_

- [ ] 6. Write tests and implement application services
  - [ ] 6.1 Write failing tests for UserService
    - Create test files for UserService with mocked dependencies
    - Write tests for RegisterUser method with validation and WebID generation scenarios
    - Test UpdateProfile method with change detection and validation
    - Write tests for DeleteAccount method with proper cleanup validation
    - Test query methods (GetUserByID, GetUserByWebID) using mocked read repositories
    - Test domain event emission for all operations
    - _Requirements: 1.1, 1.2, 2.1, 2.2, 5.1, 5.2, 5.3, 6.1, 6.2_

  - [ ] 6.2 Implement UserService to make tests pass
    - Create RegisterUser method with validation and WebID generation to satisfy tests
    - Implement UpdateProfile method with change detection to pass update tests
    - Create DeleteAccount method with proper cleanup validation to satisfy deletion tests
    - Add query methods using read repositories to pass query tests
    - Emit appropriate domain events for all operations to satisfy event tests
    - _Requirements: 1.1, 1.2, 2.1, 2.2, 5.1, 5.2, 5.3, 6.1, 6.2_

  - [ ] 6.3 Write failing tests for AccountService
    - Create test files for AccountService with mocked dependencies
    - Write tests for CreateAccount method with owner assignment scenarios
    - Test InviteUser method with role validation and token generation
    - Write tests for AcceptInvitation method with invitation validation and member creation
    - Test UpdateMemberRole method with permission checks and validation
    - Test atomic entity methods using Account domain methods
    - Test domain event emission for all membership operations
    - _Requirements: 3.1, 3.2, 3.3, 3.4, 3.5_

  - [ ] 6.4 Implement AccountService to make tests pass
    - Create CreateAccount method with owner assignment to satisfy tests
    - Implement InviteUser method with role validation and token generation to pass tests
    - Create AcceptInvitation method with validation and member creation to satisfy tests
    - Implement UpdateMemberRole method with permission checks to pass validation tests
    - Add atomic entity methods using Account domain methods to satisfy entity tests
    - Emit domain events for all membership operations to pass event tests
    - _Requirements: 3.1, 3.2, 3.3, 3.4, 3.5_

- [ ] 7. Write tests and implement event handlers
  - [ ] 7.1 Write failing tests for user event handlers
    - Create test files for UserEventHandler with mocked write repositories
    - Write tests for database persistence (UserRegistered, UserProfileUpdated, UserDeleted)
    - Test file event handlers for WebID document creation and user file cleanup
    - Test error handling and transaction management for consistency
    - _Requirements: 1.4, 2.4, 5.4, 6.4_

  - [ ] 7.2 Implement user event handlers to make tests pass
    - Implement UserEventHandler for database persistence to satisfy persistence tests
    - Create file event handlers for WebID document creation and cleanup to pass file tests
    - Add error handling and transaction management to satisfy consistency tests
    - _Requirements: 1.4, 2.4, 5.4, 6.4_

  - [ ] 7.3 Write failing tests for account event handlers
    - Create test files for AccountEventHandler with mocked write repositories
    - Write tests for account persistence operations and validation
    - Test membership projection handlers (MemberAdded, MemberRemoved, MemberRoleUpdated)
    - Write tests for invitation lifecycle handlers (MemberInvited, InvitationAccepted)
    - Test atomic entity relationship management through events
    - _Requirements: 3.4, 3.5_

  - [ ] 7.4 Implement account event handlers to make tests pass
    - Implement AccountEventHandler for account persistence operations to satisfy tests
    - Create membership projection handlers to pass projection tests
    - Implement invitation lifecycle handlers to satisfy invitation tests
    - Add atomic entity relationship management through events to pass relationship tests
    - _Requirements: 3.4, 3.5_

- [ ] 8. Write tests and implement HTTP handlers
  - [ ] 8.1 Write failing tests for user management HTTP handlers
    - Create test files for user registration endpoint with input validation scenarios
    - Write tests for user profile management endpoints (GET, PUT) with various inputs
    - Test user deletion endpoint with confirmation requirements and edge cases
    - Test proper HTTP status codes and error responses for all scenarios
    - _Requirements: 1.1, 1.2, 2.1, 2.2, 6.1, 6.2_

  - [ ] 8.2 Implement user HTTP handlers to make tests pass
    - Create user registration endpoint with input validation to satisfy validation tests
    - Implement user profile management endpoints (GET, PUT) to pass profile tests
    - Create user deletion endpoint with confirmation requirements to satisfy deletion tests
    - Add proper HTTP status codes and error responses to pass response tests
    - _Requirements: 1.1, 1.2, 2.1, 2.2, 6.1, 6.2_

  - [ ] 8.3 Write failing tests for account management HTTP handlers
    - Create test files for account creation and management endpoint scenarios
    - Write tests for invitation endpoints (create, accept, list) with validation
    - Test membership management endpoints (add, remove, update role) with permissions
    - Test role-based access control validation for all protected endpoints
    - _Requirements: 3.1, 3.2, 3.3, 3.4, 3.5_

  - [ ] 8.4 Implement account HTTP handlers to make tests pass
    - Create account creation and management endpoints to satisfy creation tests
    - Implement invitation endpoints (create, accept, list) to pass invitation tests
    - Create membership management endpoints to satisfy membership tests
    - Add role-based access control validation to pass authorization tests
    - _Requirements: 3.1, 3.2, 3.3, 3.4, 3.5_

- [ ] 9. Write tests and implement Wire dependency injection
  - [ ] 9.1 Write failing tests for Wire providers
    - Create test files for Wire provider functionality and dependency resolution
    - Write tests for all repositories and services provider creation
    - Test provider sets for user management components integration
    - Test database migration integration in Wire setup
    - _Requirements: 7.3_

  - [ ] 9.2 Implement Wire providers to make tests pass
    - Implement Wire providers for all repositories and services to satisfy provider tests
    - Create provider sets for user management components to pass integration tests
    - Add database migration to Wire setup to satisfy migration tests
    - Integrate with existing server Wire configuration to pass configuration tests
    - _Requirements: 7.3_

  - [ ] 9.3 Write failing tests for server integration
    - Create test files for main server Wire configuration with user management
    - Write tests for HTTP route registration for user management endpoints
    - Test proper dependency injection for all components in server context
    - _Requirements: 7.3_

  - [ ] 9.4 Implement server integration to make tests pass
    - Update main server Wire configuration to include user management to satisfy tests
    - Add HTTP route registration for user management endpoints to pass routing tests
    - Ensure proper dependency injection for all components to satisfy injection tests
    - Run `wire ./cmd/server` to generate dependency injection code
    - _Requirements: 7.3_

- [ ] 10. Write comprehensive integration and end-to-end tests
  - [ ] 10.1 Write integration tests for complete workflows
    - Create integration tests for GORM repository implementations with temporary database
    - Write integration tests for event handlers with database operations and file storage
    - Test complete user registration workflow from HTTP request to database persistence
    - Test complete account and membership workflows with event projections
    - Verify atomic entity relationships through event projections
    - _Requirements: 1.5, 2.5, 3.5, 7.4_

  - [ ] 10.2 Write end-to-end HTTP workflow tests
    - Write end-to-end tests for complete user workflows (registration, profile update, deletion)
    - Create end-to-end tests for account workflows (creation, invitation, membership management)
    - Test authentication and authorization for protected endpoints
    - Test error scenarios and edge cases across the entire system
    - _Requirements: 1.5, 2.5, 3.5, 6.5_

- [ ] 11. Write BDD feature tests following TDD
  - [ ] 11.1 Write failing BDD scenarios for user lifecycle
    - Create Gherkin scenarios for user registration with WebID creation that initially fail
    - Write scenarios for user profile management and self-deletion that fail initially
    - Create scenarios for user status transitions and validation rules that fail
    - Write step definitions that fail until implementation is complete
    - _Requirements: 1.1, 1.2, 2.1, 2.2, 5.1, 5.2, 6.1, 6.2_

  - [ ] 11.2 Write failing BDD scenarios for account management
    - Create Gherkin scenarios for account creation and ownership that initially fail
    - Write invitation workflow scenarios (invite, accept, expire) that fail initially
    - Create membership management scenarios (add, remove, role updates) that fail
    - Write scenarios for role-based permissions and access control that fail initially
    - Write step definitions that fail until full implementation is complete
    - _Requirements: 3.1, 3.2, 3.3, 3.4, 3.5_

- [ ] 12. Write performance tests and implement optimizations
  - [ ] 12.1 Write failing performance tests
    - Create performance tests for database query patterns that initially fail performance targets
    - Write tests for caching effectiveness that fail without caching implementation
    - Create tests for membership query performance that fail without optimization
    - Write load tests for concurrent operations that fail without proper indexing
    - _Requirements: 7.1, 7.2, 7.3_

  - [ ] 12.2 Implement optimizations to make performance tests pass
    - Add database indexes for common query patterns to satisfy query performance tests
    - Implement caching for frequently accessed data (users, roles) to pass caching tests
    - Optimize membership queries with proper projections to satisfy membership performance tests
    - _Requirements: 7.1, 7.2, 7.3_

  - [ ] 12.3 Write comprehensive load and scalability tests
    - Create load tests for concurrent user registration and authentication operations
    - Write scalability tests for system performance with large numbers of users and accounts
    - Test membership query performance with many members per account
    - Create scalability benchmarks and performance baselines for future optimization
    - _Requirements: 7.1, 7.2, 7.4, 7.5_
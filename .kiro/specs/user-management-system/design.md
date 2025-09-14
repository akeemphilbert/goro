# User Management System Design

## Overview

The User Management System provides comprehensive user and account lifecycle management as a separate domain within the server architecture. The system manages user accounts, WebID creation, account organization with role-based membership, and self-service operations while integrating with the existing LDP resource management and HTTP transport layers.

The design follows clean architecture principles with the user management domain (`internal/user`) operating independently from the LDP domain (`internal/ldp`), allowing for clear separation of concerns and maintainable code organization.

## Architecture

### User Management Domain
The user management system is implemented as a separate domain with its own clean architecture:

```
internal/user/
├── domain/
│   ├── user.go              # User entity with WebID and profile
│   ├── account.go           # Account entity for pod organization  
│   ├── invitation.go        # Invitation entity for user onboarding
│   ├── role.go              # Role entity with permissions
│   ├── user_repository.go   # User storage interface (read-only queries)
│   ├── account_repository.go # Account storage interface (read-only queries)
│   ├── role_repository.go   # Role storage interface (read-only queries)
│   ├── user_events.go       # User lifecycle domain events
│   └── account_events.go    # Account lifecycle domain events
├── application/
│   ├── user_service.go           # User lifecycle operations (event emission)
│   ├── account_service.go        # Account and invitation management (event emission)
│   ├── registration_service.go   # User registration workflow
│   ├── user_event_handlers.go    # Database persistence event handlers
│   └── file_event_handlers.go    # File system persistence event handlers
└── infrastructure/
    ├── user_repository_impl.go      # GORM-based user storage
    ├── account_repository_impl.go   # GORM-based account persistence
    ├── webid_generator.go           # WebID URI generation
    ├── database_models.go           # GORM model definitions
    ├── migration.go                 # Auto-migration for user models
    └── wire.go                      # Wire providers for dependency injection
```

### Transport Layer Integration
HTTP handlers extend the existing transport layer for user management endpoints:

```
internal/infrastructure/transport/http/handlers/
├── user_handlers.go         # User CRUD operations
├── account_handlers.go      # Account management
└── registration_handlers.go # User registration endpoints
```

## Components and Interfaces

### Core Domain Entities

#### User Entity
```go
type User struct {
    ID        string
    WebID     string
    Email     string
    Profile   UserProfile
    Status    UserStatus
    CreatedAt time.Time
    UpdatedAt time.Time
}

type UserProfile struct {
    Name        string
    Bio         string
    Avatar      string
    Preferences map[string]interface{}
}

type UserStatus string
const (
    UserStatusActive    UserStatus = "active"
    UserStatusSuspended UserStatus = "suspended" 
    UserStatusDeleted   UserStatus = "deleted"
)
```

#### Role Entity
```go
type Role struct {
    ID          string
    Name        string
    Description string
    Permissions []Permission
    CreatedAt   time.Time
    UpdatedAt   time.Time
}

type Permission struct {
    Resource string // "user", "account", "resource", etc.
    Action   string // "create", "read", "update", "delete"
    Scope    string // "own", "account", "global"
}

// Predefined system roles
var (
    RoleOwner = Role{
        ID:   "owner",
        Name: "Owner",
        Description: "Full access to account and all resources",
        Permissions: []Permission{
            {Resource: "*", Action: "*", Scope: "account"},
        },
    }
    
    RoleAdmin = Role{
        ID:   "admin", 
        Name: "Administrator",
        Description: "Administrative access to account management",
        Permissions: []Permission{
            {Resource: "user", Action: "*", Scope: "account"},
            {Resource: "account", Action: "read,update", Scope: "account"},
            {Resource: "resource", Action: "*", Scope: "account"},
        },
    }
    
    RoleMember = Role{
        ID:   "member",
        Name: "Member", 
        Description: "Standard member access",
        Permissions: []Permission{
            {Resource: "user", Action: "read", Scope: "account"},
            {Resource: "resource", Action: "create,read,update", Scope: "own"},
        },
    }
    
    RoleViewer = Role{
        ID:   "viewer",
        Name: "Viewer",
        Description: "Read-only access",
        Permissions: []Permission{
            {Resource: "user", Action: "read", Scope: "account"},
            {Resource: "resource", Action: "read", Scope: "account"},
        },
    }
)
```

#### Account Entity
```go
type Account struct {
    ID          string
    OwnerID     string
    Name        string
    Description string
    Settings    AccountSettings
    CreatedAt   time.Time
}

// Domain methods for account operations
func (a *Account) AddMember(user *User, role *Role, invitation *Invitation)  error {
    if role == nil {
        a.AddError(fmt.Error("role must be specified"))
    }
    // Business logic validation
   a.AddEvent(new MemberAddedEvent(user,role,invitation))
}

func (a *Account) RemoveMember(userID string) error {
    // Business logic validation
    return nil
}

func (a *Account) UpdateMemberRole(userID, newRoleID string) error {
    // Business logic validation
    return nil
}

type AccountSettings struct {
    AllowInvitations bool
    DefaultRoleID    string
    MaxMembers       int
}
```

#### Invitation Entity
```go
type Invitation struct {
    ID        string
    AccountID string
    Email     string
    RoleID    string // References Role.ID
    Token     string
    InvitedBy string
    Status    InvitationStatus
    ExpiresAt time.Time
    CreatedAt time.Time
}

type InvitationStatus string
const (
    InvitationStatusPending  InvitationStatus = "pending"
    InvitationStatusAccepted InvitationStatus = "accepted"
    InvitationStatusExpired  InvitationStatus = "expired"
    InvitationStatusRevoked  InvitationStatus = "revoked"
)
```



### Domain Events

The user management system follows event-driven architecture where services emit domain events and event handlers perform persistence operations:

#### User Domain Events
```go
type UserRegisteredEvent struct {
    UserID    string
    WebID     string
    Email     string
    Profile   UserProfile
    Timestamp time.Time
}

type UserProfileUpdatedEvent struct {
    UserID    string
    Profile   UserProfile
    Timestamp time.Time
}

type UserDeletedEvent struct {
    UserID    string
    Timestamp time.Time
}

type WebIDGeneratedEvent struct {
    UserID     string
    WebID      string
    WebIDDoc   string // Turtle format
    Timestamp  time.Time
}
```

#### Account Domain Events
```go
type AccountCreatedEvent struct {
    AccountID   string
    OwnerID     string
    Name        string
    Description string
    Timestamp   time.Time
}

type MemberInvitedEvent struct {
    InvitationID string
    AccountID    string
    Email        string
    RoleID       string
    InvitedBy    string
    Token        string
    ExpiresAt    time.Time
    Timestamp    time.Time
}

type InvitationAcceptedEvent struct {
    InvitationID string
    AccountID    string
    UserID       string
    RoleID       string
    Timestamp    time.Time
}

type MemberAddedEvent struct {
    MemberID  string
    AccountID string
    UserID    string
    RoleID    string
    InvitedBy string
    Timestamp time.Time
}

type MemberRemovedEvent struct {
    MemberID  string
    AccountID string
    UserID    string
    Timestamp time.Time
}

type MemberRoleUpdatedEvent struct {
    MemberID  string
    AccountID string
    UserID    string
    OldRoleID string
    NewRoleID string
    UpdatedBy string
    Timestamp time.Time
}
```

### Repository Interfaces

#### User Repository (Read-Only)
```go
type UserRepository interface {
    GetByID(ctx context.Context, id string) (*User, error)
    GetByWebID(ctx context.Context, webid string) (*User, error)
    GetByEmail(ctx context.Context, email string) (*User, error)
    List(ctx context.Context, filter UserFilter) ([]*User, error)
    Exists(ctx context.Context, id string) (bool, error)
}
```

#### Account Repository (Read-Only)
```go
type AccountRepository interface {
    GetByID(ctx context.Context, id string) (*Account, error)
    GetByOwner(ctx context.Context, ownerID string) ([]*Account, error)
}
```

#### AccountMember Repository (Read-Only - Projection)
```go
type AccountMemberRepository interface {
    GetByID(ctx context.Context, id string) (*AccountMember, error)
    GetByAccountAndUser(ctx context.Context, accountID, userID string) (*AccountMember, error)
    ListByAccount(ctx context.Context, accountID string) ([]*AccountMember, error)
    ListByUser(ctx context.Context, userID string) ([]*AccountMember, error)
}
```

#### Invitation Repository (Read-Only)
```go
type InvitationRepository interface {
    GetByID(ctx context.Context, id string) (*Invitation, error)
    GetByToken(ctx context.Context, token string) (*Invitation, error)
    ListByAccount(ctx context.Context, accountID string) ([]*Invitation, error)
    ListByEmail(ctx context.Context, email string) ([]*Invitation, error)
}
```

#### Role Repository (Read-Only)
```go
type RoleRepository interface {
    GetByID(ctx context.Context, id string) (*Role, error)
    List(ctx context.Context) ([]*Role, error)
    GetSystemRoles(ctx context.Context) ([]*Role, error)
}
```

#### Event-Driven Repository Interfaces (Write Operations)
```go
type UserWriteRepository interface {
    Create(ctx context.Context, user *User) error
    Update(ctx context.Context, user *User) error
    Delete(ctx context.Context, id string) error
}

type AccountWriteRepository interface {
    Create(ctx context.Context, account *Account) error
    Update(ctx context.Context, account *Account) error
    Delete(ctx context.Context, id string) error
}

type AccountMemberWriteRepository interface {
    Create(ctx context.Context, member *AccountMember) error
    Update(ctx context.Context, member *AccountMember) error
    Delete(ctx context.Context, id string) error
}

type InvitationWriteRepository interface {
    Create(ctx context.Context, invitation *Invitation) error
    Update(ctx context.Context, invitation *Invitation) error
    Delete(ctx context.Context, id string) error
}

type RoleWriteRepository interface {
    Create(ctx context.Context, role *Role) error
    Update(ctx context.Context, role *Role) error
    Delete(ctx context.Context, id string) error
    SeedSystemRoles(ctx context.Context) error
}
```

### Application Services

#### User Service
Orchestrates user lifecycle operations by emitting domain events for persistence:

```go
type UserService struct {
    eventBus     EventDispatcher
    webidGen     WebIDGenerator
    userRepo     UserRepository // Read-only queries
}

func (s *UserService) RegisterUser(ctx context.Context, req RegisterUserRequest) (*User, error)
func (s *UserService) UpdateProfile(ctx context.Context, userID string, profile UserProfile) error
func (s *UserService) DeleteAccount(ctx context.Context, userID string) error
func (s *UserService) GetUserByID(ctx context.Context, userID string) (*User, error)
func (s *UserService) GetUserByWebID(ctx context.Context, webID string) (*User, error)
```

#### Event Handlers
Handle persistence operations in response to domain events:

```go
type UserEventHandler struct {
    userRepo    UserRepository
    fileStorage FileStorage
}

func (h *UserEventHandler) HandleUserRegistered(ctx context.Context, event UserRegisteredEvent) error
func (h *UserEventHandler) HandleUserProfileUpdated(ctx context.Context, event UserProfileUpdatedEvent) error
func (h *UserEventHandler) HandleUserDeleted(ctx context.Context, event UserDeletedEvent) error
func (h *UserEventHandler) HandleWebIDGenerated(ctx context.Context, event WebIDGeneratedEvent) error

type AccountEventHandler struct {
    accountRepo       AccountWriteRepository
    memberRepo        AccountMemberWriteRepository
    invitationRepo    InvitationWriteRepository
    fileStorage       FileStorage
}

func (h *AccountEventHandler) HandleAccountCreated(ctx context.Context, event AccountCreatedEvent) error
func (h *AccountEventHandler) HandleMemberInvited(ctx context.Context, event MemberInvitedEvent) error
func (h *AccountEventHandler) HandleInvitationAccepted(ctx context.Context, event InvitationAcceptedEvent) error
func (h *AccountEventHandler) HandleMemberAdded(ctx context.Context, event MemberAddedEvent) error
func (h *AccountEventHandler) HandleMemberRemoved(ctx context.Context, event MemberRemovedEvent) error
func (h *AccountEventHandler) HandleMemberRoleUpdated(ctx context.Context, event MemberRoleUpdatedEvent) error
```

#### Account Service
Manages account creation, invitations, and member management by emitting domain events:

```go
type AccountService struct {
    eventBus    EventDispatcher
    inviteGen   InvitationGenerator
    accountRepo AccountRepository // Read-only queries
    userRepo    UserRepository    // Read-only queries
}

func (s *AccountService) CreateAccount(ctx context.Context, ownerID string, name string) (*Account, error)
func (s *AccountService) InviteUser(ctx context.Context, accountID, inviterID, email string, roleID string) (*Invitation, error)
func (s *AccountService) AcceptInvitation(ctx context.Context, token string, userID string) error
func (s *AccountService) UpdateMemberRole(ctx context.Context, accountID, userID string, roleID string) error
```

## Data Models

### File System Storage Structure
User data is stored in the existing file system structure with new directories:

```
data/pod-storage/
├── users/
│   ├── {user-id}/
│   │   ├── profile.json      # User profile and metadata
│   │   └── webid.ttl         # WebID document in Turtle format
├── accounts/
│   ├── {account-id}/
│   │   ├── metadata.json     # Account information
│   │   ├── members.json      # Account membership data
│   │   └── invitations.json  # Pending invitations
└── index/
    └── users.db             # GORM database (SQLite/Postgres/MySQL)
```

### WebID Document Format
Each user gets a WebID document in Turtle format:

```turtle
@prefix foaf: <http://xmlns.com/foaf/0.1/> .
@prefix solid: <http://www.w3.org/ns/solid/terms#> .
@prefix ldp: <http://www.w3.org/ns/ldp#> .

<{webid-uri}> a foaf:Person ;
    foaf:name "{user-name}" ;
    foaf:mbox <mailto:{email}> ;
    solid:account <{account-uri}> ;
    solid:privateTypeIndex <{private-index-uri}> ;
    solid:publicTypeIndex <{public-index-uri}> .
```

### Database Models (GORM)
User indexing uses GORM models with auto-migration for database abstraction:

```go
// GORM model for users table
type UserModel struct {
    ID        string    `gorm:"primaryKey;type:varchar(255)"`
    WebID     string    `gorm:"uniqueIndex;not null;type:varchar(500)"`
    Email     string    `gorm:"uniqueIndex;not null;type:varchar(255)"`
    Name      string    `gorm:"type:varchar(255)"`
    Status    string    `gorm:"not null;type:varchar(50)"`
    CreatedAt time.Time `gorm:"not null"`
    UpdatedAt time.Time `gorm:"not null"`
}

// GORM model for accounts table
type AccountModel struct {
    ID          string    `gorm:"primaryKey;type:varchar(255)"`
    OwnerID     string    `gorm:"not null;type:varchar(255);index"`
    Name        string    `gorm:"not null;type:varchar(255)"`
    Description string    `gorm:"type:text"`
    Settings    string    `gorm:"type:text"` // JSON serialized AccountSettings
    CreatedAt   time.Time `gorm:"not null"`
    UpdatedAt   time.Time `gorm:"not null"`
}

// GORM model for roles
type RoleModel struct {
    ID          string    `gorm:"primaryKey;type:varchar(255)"`
    Name        string    `gorm:"not null;type:varchar(255)"`
    Description string    `gorm:"type:text"`
    Permissions string    `gorm:"type:text"` // JSON serialized permissions
    CreatedAt   time.Time `gorm:"not null"`
    UpdatedAt   time.Time `gorm:"not null"`
}

// GORM model for account membership (projection from events)
type AccountMemberModel struct {
    ID        string    `gorm:"primaryKey;type:varchar(255)"`
    AccountID string    `gorm:"not null;type:varchar(255);index"`
    UserID    string    `gorm:"not null;type:varchar(255);index"`
    RoleID    string    `gorm:"not null;type:varchar(255)"`
    InvitedBy string    `gorm:"type:varchar(255)"`
    JoinedAt  time.Time `gorm:"not null"`
    CreatedAt time.Time `gorm:"not null"`
    
    // Unique constraint on account+user combination
    // This is a projection table built from events
}

// GORM model for invitations
type InvitationModel struct {
    ID        string    `gorm:"primaryKey;type:varchar(255)"`
    AccountID string    `gorm:"not null;type:varchar(255);index"`
    Email     string    `gorm:"not null;type:varchar(255)"`
    RoleID    string    `gorm:"not null;type:varchar(255)"`
    Token     string    `gorm:"uniqueIndex;not null;type:varchar(255)"`
    InvitedBy string    `gorm:"not null;type:varchar(255)"`
    Status    string    `gorm:"not null;type:varchar(50)"` // pending, accepted, expired
    ExpiresAt time.Time `gorm:"not null"`
    CreatedAt time.Time `gorm:"not null"`
}
```

### Database Migration
The user package uses the existing GORM DB instance from the dependency container and performs auto-migration:

```go
// Migration function for user management models
func MigrateUserModels(db *gorm.DB) error {
    return db.AutoMigrate(
        &UserModel{},
        &AccountModel{},
        &AccountMemberModel{},
        &InvitationModel{},
    )
}

// Wire provider for user repository with GORM
func ProvideUserRepository(db *gorm.DB) UserRepository {
    // Ensure models are migrated
    if err := MigrateUserModels(db); err != nil {
        panic(fmt.Sprintf("failed to migrate user models: %v", err))
    }
    return NewGormUserRepository(db)
}
```

## Error Handling

### Domain Errors
User management introduces specific domain errors:

```go
var (
    ErrUserNotFound         = errors.New("user not found")
    ErrUserAlreadyExists    = errors.New("user already exists")
    ErrInvalidWebID         = errors.New("invalid WebID format")
    ErrAccountNotFound      = errors.New("account not found")
    ErrInsufficientPermissions = errors.New("insufficient permissions")
    ErrInvitationExpired    = errors.New("invitation has expired")
)
```

### Error Context Wrapping
Errors are wrapped with context at each layer boundary:

```go
// Domain layer
func (r *userRepository) GetByID(ctx context.Context, id string) (*User, error) {
    user, err := r.storage.Load(id)
    if err != nil {
        return nil, fmt.Errorf("failed to load user %s: %w", id, err)
    }
    return user, nil
}

// Application layer  
func (s *UserService) UpdateProfile(ctx context.Context, userID string, profile UserProfile) error {
    user, err := s.userRepo.GetByID(ctx, userID)
    if err != nil {
        return fmt.Errorf("profile update failed: %w", err)
    }
    // ... update logic
}
```

### HTTP Error Responses
Transport layer maps domain errors to appropriate HTTP status codes:

```go
func mapUserError(err error) (int, string) {
    switch {
    case errors.Is(err, ErrUserNotFound):
        return http.StatusNotFound, "User not found"
    case errors.Is(err, ErrUserAlreadyExists):
        return http.StatusConflict, "User already exists"
    case errors.Is(err, ErrInsufficientPermissions):
        return http.StatusForbidden, "Insufficient permissions"
    default:
        return http.StatusInternalServerError, "Internal server error"
    }
}
```

## Testing Strategy

### BDD Feature Scenarios
User management features are tested with Gherkin scenarios:

```gherkin
Feature: User Registration
  Scenario: Successful user registration
    Given the system is running
    When I register with email "user@example.com" and name "John Doe"
    Then a new user should be created
    And a WebID should be generated
    And the user should receive a confirmation

Feature: Account Management  
  Scenario: Account owner invites new member
    Given I am an account owner
    When I invite "member@example.com" with role "member"
    Then an invitation should be sent
    And the invitation should be pending
```

### Unit Testing Strategy
Each component has comprehensive unit tests:

- **Domain entities**: Test validation, state transitions, and business rules
- **Repository implementations**: Test CRUD operations with temporary storage
- **Application services**: Test use case orchestration with mocked dependencies
- **HTTP handlers**: Test request/response handling and error scenarios

### Integration Testing
End-to-end tests validate complete user workflows:

- User registration and WebID creation
- Account creation and member invitation
- External identity linking and authentication
- User profile updates and account deletion
- Performance testing with concurrent user operations

### Performance Testing
Scalability tests ensure the system handles growth:

- User lookup performance with large user bases
- Concurrent registration and authentication operations
- Account membership queries with many members
- WebID resolution and caching effectiveness
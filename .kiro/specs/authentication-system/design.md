# Authentication System Design

## Overview

The Authentication System provides secure, multi-method authentication for the Solid pod server using clean architecture principles. The system supports WebID-OIDC, username/password, and third-party OAuth providers with optional multi-factor authentication. It integrates with the existing user management system and follows the established domain-driven design patterns.

## Architecture

### Layer Structure
Following the established clean architecture pattern:

```
internal/auth/
├── domain/           # Pure business logic - authentication entities and rules
├── application/      # Use cases and orchestration services  
├── infrastructure/   # External integrations (OAuth, OIDC, storage)
└── integration/      # End-to-end integration tests
```

### Integration Points
- **User Management**: Leverages existing `internal/user` domain for user entities
- **HTTP Transport**: Extends `internal/infrastructure/transport/http` with auth handlers
- **Event System**: Publishes authentication events to existing event infrastructure
- **Configuration**: Integrates with `internal/conf` for provider configuration
- **Email Service**: Uses shared `internal/infrastructure/email` service for transactional emails

## Shared Infrastructure Components

### Email Service (`internal/infrastructure/email/`)
A shared email service interface that can be used across all domains for transactional emails:

```go
// Email service interface for transactional emails
type Service interface {
    SendEmail(ctx context.Context, email *Email) error
    SendTemplatedEmail(ctx context.Context, template string, data interface{}, recipients ...string) error
}

// Email represents a transactional email
type Email struct {
    To          []string
    CC          []string
    BCC         []string
    Subject     string
    TextBody    string
    HTMLBody    string
    Attachments []Attachment
}

// Template-based email for common use cases
type TemplateData struct {
    UserName    string
    ResetURL    string
    ExpiryTime  time.Time
    SupportURL  string
}
```

#### Implementation Options

```go
// AWS SES implementation
type SESEmailService struct {
    client    *ses.Client
    fromEmail string
    templates map[string]*template.Template
}

// Local SMTP implementation
type SMTPEmailService struct {
    host      string
    port      int
    username  string
    password  string
    fromEmail string
    templates map[string]*template.Template
}

// Development/testing implementation
type MockEmailService struct {
    sentEmails []Email
}
```

## Components and Interfaces

### Domain Layer (`internal/auth/domain/`)

#### Core Entities
```go
// Session represents an authenticated user session
type Session struct {
    ID           string
    UserID       string
    WebID        string
    TokenHash    string
    ExpiresAt    time.Time
    CreatedAt    time.Time
    LastActivity time.Time
}

// AuthenticationMethod represents supported auth methods
type AuthenticationMethod string
const (
    MethodWebIDOIDC    AuthenticationMethod = "webid-oidc"
    MethodPassword     AuthenticationMethod = "password"
    MethodOAuth        AuthenticationMethod = "oauth"
)

// PasswordCredential represents user password credentials
type PasswordCredential struct {
    UserID       string
    PasswordHash string
    Salt         string
    CreatedAt    time.Time
    UpdatedAt    time.Time
}

// PasswordResetToken represents password reset tokens
type PasswordResetToken struct {
    Token     string
    UserID    string
    Email     string
    ExpiresAt time.Time
    CreatedAt time.Time
    Used      bool
}
```

#### Repository Interfaces
```go
type SessionRepository interface {
    Save(ctx context.Context, session *Session) error
    FindByID(ctx context.Context, id string) (*Session, error)
    FindByUserID(ctx context.Context, userID string) ([]*Session, error)
    Delete(ctx context.Context, id string) error
    DeleteExpired(ctx context.Context) error
}

type PasswordRepository interface {
    Save(ctx context.Context, credential *PasswordCredential) error
    FindByUserID(ctx context.Context, userID string) (*PasswordCredential, error)
    Update(ctx context.Context, credential *PasswordCredential) error
    Delete(ctx context.Context, userID string) error
}

type PasswordResetRepository interface {
    Save(ctx context.Context, token *PasswordResetToken) error
    FindByToken(ctx context.Context, token string) (*PasswordResetToken, error)
    MarkAsUsed(ctx context.Context, token string) error
    DeleteExpired(ctx context.Context) error
}

type ExternalIdentityRepository interface {
    LinkIdentity(ctx context.Context, userID, provider, externalID string) error
    FindByExternalID(ctx context.Context, provider, externalID string) (string, error)
    GetLinkedIdentities(ctx context.Context, userID string) ([]ExternalIdentity, error)
}
```

### Application Layer (`internal/auth/application/`)

#### Authentication Service
```go
type AuthenticationService struct {
    userRepo     user.UserRepository
    sessionRepo  SessionRepository
    tokenManager TokenManager
    eventBus     EventBus
}

// Core authentication methods
func (s *AuthenticationService) AuthenticateWithPassword(ctx context.Context, username, password string) (*Session, error)
func (s *AuthenticationService) AuthenticateWithWebID(ctx context.Context, webID string, oidcToken string) (*Session, error)
func (s *AuthenticationService) AuthenticateWithOAuth(ctx context.Context, provider string, oauthToken string) (*Session, error)
func (s *AuthenticationService) ValidateSession(ctx context.Context, sessionID string) (*Session, error)
func (s *AuthenticationService) RefreshSession(ctx context.Context, sessionID string) (*Session, error)
func (s *AuthenticationService) Logout(ctx context.Context, sessionID string) error
```

#### Registration Service
```go
type RegistrationService struct {
    userService    user.UserService
    identityRepo   ExternalIdentityRepository
    webidGenerator WebIDGenerator
}

func (s *RegistrationService) RegisterWithExternalIdentity(ctx context.Context, provider string, profile ExternalProfile) (*user.User, error)
func (s *RegistrationService) LinkExternalIdentity(ctx context.Context, userID, provider string, profile ExternalProfile) error
```

#### Password Management Service
```go
type PasswordService struct {
    userRepo        user.UserRepository
    passwordRepo    PasswordRepository
    resetRepo       PasswordResetRepository
    emailService    email.Service  // Shared infrastructure service
    tokenGenerator  SecureTokenGenerator
    hasher          PasswordHasher
}

func (s *PasswordService) SetPassword(ctx context.Context, userID, password string) error
func (s *PasswordService) ChangePassword(ctx context.Context, userID, currentPassword, newPassword string) error
func (s *PasswordService) InitiatePasswordReset(ctx context.Context, email string) error
func (s *PasswordService) CompletePasswordReset(ctx context.Context, token, newPassword string) error
func (s *PasswordService) ValidatePassword(ctx context.Context, userID, password string) error
```

### Infrastructure Layer (`internal/auth/infrastructure/`)

#### OAuth Providers
```go
type OAuthProvider interface {
    GetAuthURL(state string) string
    ExchangeCode(ctx context.Context, code string) (*OAuthToken, error)
    GetUserProfile(ctx context.Context, token *OAuthToken) (*ExternalProfile, error)
}

// Concrete implementations
type GoogleOAuthProvider struct {
    clientID     string
    clientSecret string
    redirectURL  string
}

type GitHubOAuthProvider struct {
    clientID     string
    clientSecret string
    redirectURL  string
}
```

#### WebID-OIDC Provider
```go
type WebIDOIDCProvider struct {
    httpClient *http.Client
    validator  JWTValidator
}

func (p *WebIDOIDCProvider) ValidateWebIDToken(ctx context.Context, token string) (*WebIDClaims, error)
func (p *WebIDOIDCProvider) DiscoverProvider(ctx context.Context, webID string) (*OIDCConfiguration, error)
```

#### Token Management
```go
type JWTTokenManager struct {
    signingKey []byte
    issuer     string
    expiry     time.Duration
}

func (tm *JWTTokenManager) GenerateToken(ctx context.Context, session *Session) (string, error)
func (tm *JWTTokenManager) ValidateToken(ctx context.Context, token string) (*TokenClaims, error)
func (tm *JWTTokenManager) RefreshToken(ctx context.Context, token string) (string, error)
```

#### Password Security
```go
type PasswordHasher interface {
    Hash(password string) (hash, salt string, err error)
    Verify(password, hash, salt string) bool
}

type BCryptPasswordHasher struct {
    cost int
}



type SecureTokenGenerator interface {
    GenerateToken() (string, error)
}
```

#### Repository Implementations
```go
// GORM-based session repository (database agnostic)
type GormSessionRepository struct {
    db *gorm.DB
}

func (r *GormSessionRepository) Save(ctx context.Context, session *Session) error {
    model := &SessionModel{
        ID:           session.ID,
        UserID:       session.UserID,
        WebID:        session.WebID,
        TokenHash:    session.TokenHash,
        ExpiresAt:    session.ExpiresAt,
        CreatedAt:    session.CreatedAt,
        LastActivity: session.LastActivity,

    }
    return r.db.WithContext(ctx).Save(model).Error
}

// GORM-based password repository
type GormPasswordRepository struct {
    db     *gorm.DB
    hasher PasswordHasher
}

// GORM-based password reset repository
type GormPasswordResetRepository struct {
    db *gorm.DB
}

// GORM-based external identity repository  
type GormExternalIdentityRepository struct {
    db *gorm.DB
}
```

## Data Models

### GORM Data Models
Following the existing codebase pattern for database agnosticism:

```go
// Session model for GORM
type SessionModel struct {
    ID           string    `gorm:"primaryKey;type:varchar(255)"`
    UserID       string    `gorm:"not null;type:varchar(255);index"`
    WebID        string    `gorm:"not null;type:varchar(255)"`
    TokenHash    string    `gorm:"not null;type:varchar(255)"`
    ExpiresAt    time.Time `gorm:"not null;index"`
    CreatedAt    time.Time `gorm:"not null"`
    LastActivity time.Time `gorm:"not null"`

}

func (SessionModel) TableName() string {
    return "sessions"
}

// Password credential model for GORM
type PasswordCredentialModel struct {
    UserID       string    `gorm:"primaryKey;type:varchar(255)"`
    PasswordHash string    `gorm:"not null;type:varchar(255)"`
    Salt         string    `gorm:"not null;type:varchar(255)"`
    CreatedAt    time.Time `gorm:"not null"`
    UpdatedAt    time.Time `gorm:"not null"`
}

func (PasswordCredentialModel) TableName() string {
    return "password_credentials"
}

// Password reset token model for GORM
type PasswordResetTokenModel struct {
    Token     string    `gorm:"primaryKey;type:varchar(255)"`
    UserID    string    `gorm:"not null;type:varchar(255);index"`
    Email     string    `gorm:"not null;type:varchar(255)"`
    ExpiresAt time.Time `gorm:"not null;index"`
    CreatedAt time.Time `gorm:"not null"`
    Used      bool      `gorm:"default:false"`
}

func (PasswordResetTokenModel) TableName() string {
    return "password_reset_tokens"
}

// External identity linking model for GORM
type ExternalIdentityModel struct {
    ID         uint      `gorm:"primaryKey;autoIncrement"`
    UserID     string    `gorm:"not null;type:varchar(255);index"`
    Provider   string    `gorm:"not null;type:varchar(100)"`
    ExternalID string    `gorm:"not null;type:varchar(255)"`
    CreatedAt  time.Time `gorm:"not null"`
}

func (ExternalIdentityModel) TableName() string {
    return "external_identities"
}

// Add unique constraint for provider + external_id combination
func (ExternalIdentityModel) Indexes() []gorm.Index {
    return []gorm.Index{
        {
            Name:    "idx_provider_external_id",
            Fields:  []string{"provider", "external_id"},
            Unique:  true,
        },
    }
}
```

### Configuration Structure
```yaml
auth:
  session:
    expiry: "24h"
    refresh_threshold: "1h"
  jwt:
    signing_key: "${JWT_SIGNING_KEY}"
    issuer: "solid-pod-server"
  oauth:
    google:
      client_id: "${GOOGLE_CLIENT_ID}"
      client_secret: "${GOOGLE_CLIENT_SECRET}"
      redirect_url: "${BASE_URL}/auth/oauth/google/callback"
    github:
      client_id: "${GITHUB_CLIENT_ID}"
      client_secret: "${GITHUB_CLIENT_SECRET}"
      redirect_url: "${BASE_URL}/auth/oauth/github/callback"
  webid_oidc:
    timeout: "30s"
    cache_ttl: "1h"
  password:
    bcrypt_cost: 12
    reset_token_expiry: "1h"
    min_length: 8
email:
  provider: "${EMAIL_PROVIDER}" # "ses", "smtp", or "mock"
  from_address: "${EMAIL_FROM_ADDRESS}"
  
  # SMTP configuration (when provider = "smtp")
  smtp:
    host: "${SMTP_HOST}"
    port: "${SMTP_PORT}"
    username: "${SMTP_USERNAME}"
    password: "${SMTP_PASSWORD}"
    tls: true
  
  # AWS SES configuration (when provider = "ses")
  ses:
    region: "${AWS_REGION}"
    access_key_id: "${AWS_ACCESS_KEY_ID}"
    secret_access_key: "${AWS_SECRET_ACCESS_KEY}"
  
  # Email templates
  templates:
    password_reset: "templates/password_reset.html"
    welcome: "templates/welcome.html"
    account_verification: "templates/account_verification.html"
```

## Error Handling

### Domain Errors
```go
var (
    ErrInvalidCredentials      = errors.New("invalid credentials")
    ErrSessionExpired          = errors.New("session expired")
    ErrSessionNotFound         = errors.New("session not found")

    ErrExternalAuthFailed     = errors.New("external authentication failed")
    ErrWebIDValidationFailed  = errors.New("WebID validation failed")
    ErrPasswordTooWeak        = errors.New("password does not meet security requirements")
    ErrPasswordResetExpired   = errors.New("password reset token expired")
    ErrPasswordResetInvalid   = errors.New("invalid password reset token")
    ErrPasswordResetUsed      = errors.New("password reset token already used")
    ErrCurrentPasswordInvalid = errors.New("current password is incorrect")
)
```

### HTTP Error Responses
- `401 Unauthorized`: Invalid credentials, expired sessions
- `403 Forbidden`: Insufficient permissions
- `400 Bad Request`: Malformed authentication requests
- `500 Internal Server Error`: Provider configuration issues, system errors

## Testing Strategy

### Unit Tests
- Domain entity validation and business rules
- Service layer authentication flows
- Repository implementations with mock dependencies
- Token generation and validation logic

### Integration Tests
- End-to-end authentication flows for each method
- OAuth provider integration with test servers
- WebID-OIDC validation with mock identity providers
- Session management across multiple requests

### BDD Scenarios
```gherkin
Feature: Multi-Method Authentication
  Scenario: User authenticates with username and password
    Given a user exists with username "alice" and password "secure123"
    When the user submits valid credentials
    Then a session should be created
    And the user should receive a valid JWT token

  Scenario: External identity registration
    Given a valid Google OAuth token
    When a new user registers with Google
    Then a new WebID should be created
    And the Google identity should be linked

  Scenario: Password setup for new user
    Given a user without a password
    When the user sets up their password
    Then the password should be securely hashed and stored
    And the user should be able to authenticate with it

  Scenario: Password reset flow
    Given a user with a forgotten password
    When the user requests a password reset
    Then a reset email should be sent
    And the user should be able to reset their password with the token
```

### Performance Tests
- Session validation performance under load
- OAuth callback handling with concurrent requests
- Token refresh performance
- Email delivery performance and reliability

### Email Service Testing
- Mock email service for unit tests
- SMTP integration tests with test mail servers
- AWS SES integration tests (when configured)
- Template rendering and validation tests
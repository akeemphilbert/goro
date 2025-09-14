# Email Service

The email service provides a unified interface for sending transactional emails across different providers (SMTP, AWS SES, and Mock for testing).

## Features

- **Multiple Providers**: Support for SMTP, AWS SES, and Mock services
- **Template System**: Built-in email templates with customizable data
- **Configuration-Driven**: Easy configuration via YAML or environment variables
- **Testing Support**: Mock service for unit and integration tests
- **Thread-Safe**: All implementations are safe for concurrent use

## Quick Start

### Basic Usage

```go
import "github.com/akeemphilbert/goro/internal/infrastructure/email"

// Create a mock service for testing
service := email.NewMockEmailServiceForTesting()

// Send a simple email
email := &email.Email{
    To:       []string{"user@example.com"},
    Subject:  "Welcome",
    TextBody: "Welcome to Solid Pod!",
    HTMLBody: "<h1>Welcome to Solid Pod!</h1>",
}

ctx := context.Background()
err := service.SendEmail(ctx, email)
```

### Templated Emails

```go
// Send a password reset email using built-in template
data := email.TemplateData{
    UserName:   "John Doe",
    ResetURL:   "https://example.com/reset/token123",
    ExpiryTime: time.Now().Add(1 * time.Hour),
    SupportURL: "https://example.com/support",
}

err := service.SendTemplatedEmail(ctx, "password_reset", data, "john@example.com")
```

## Configuration

### SMTP Configuration

```yaml
email:
  provider: "smtp"
  from_address: "noreply@example.com"
  smtp:
    host: "smtp.gmail.com"
    port: 587
    username: "your-email@gmail.com"
    password: "your-app-password"
    tls: true
  templates:
    password_reset: "/path/to/templates/password_reset.html"
    welcome: "/path/to/templates/welcome.html"
```

### AWS SES Configuration

```yaml
email:
  provider: "ses"
  from_address: "noreply@example.com"
  ses:
    region: "us-east-1"
    access_key_id: "${AWS_ACCESS_KEY_ID}"
    secret_access_key: "${AWS_SECRET_ACCESS_KEY}"
  templates:
    password_reset: ""  # Use default templates
    welcome: ""
```

### Mock Configuration (for testing)

```yaml
email:
  provider: "mock"
  from_address: "test@example.com"
```

## Built-in Templates

The service includes default templates for common use cases:

- **password_reset**: Password reset emails with reset link and expiry
- **welcome**: Welcome emails for new users
- **account_verification**: Account verification emails

### Template Data Structure

```go
type TemplateData struct {
    UserName    string    // User's display name
    ResetURL    string    // Password reset URL
    ExpiryTime  time.Time // Token expiry time
    SupportURL  string    // Support/help URL
}
```

## Testing

### Mock Service

The mock service records all sent emails for testing:

```go
mockService := email.NewMockEmailServiceForTesting()

// Send emails...
mockService.SendEmail(ctx, email)

// Verify emails were sent
sentEmails := mockService.GetSentEmails()
assert.Equal(t, 1, len(sentEmails))

// Find specific emails
resetEmails := mockService.FindEmailsByTemplate("password_reset")
userEmails := mockService.FindEmailsByRecipient("user@example.com")
```

### Test Utilities

```go
// Configure mock to fail
mockService.SetShouldFail(true, errors.New("network error"))

// Clear sent emails
mockService.Clear()

// Get last sent email
lastEmail := mockService.GetLastSentEmail()
```

## Environment Variables

For production deployments, use environment variables:

```bash
# SMTP Configuration
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=your-email@gmail.com
SMTP_PASSWORD=your-app-password

# AWS SES Configuration
AWS_REGION=us-east-1
AWS_ACCESS_KEY_ID=your-access-key
AWS_SECRET_ACCESS_KEY=your-secret-key

# General Configuration
EMAIL_FROM_ADDRESS=noreply@yourdomain.com
```

## Error Handling

All email service methods return errors for:

- Invalid configuration
- Network failures
- Authentication failures
- Template rendering errors
- Missing recipients

```go
err := service.SendEmail(ctx, email)
if err != nil {
    log.Printf("Failed to send email: %v", err)
    // Handle error appropriately
}
```

## Security Considerations

- Store SMTP passwords and AWS credentials securely
- Use environment variables or secret management systems
- Enable TLS for SMTP connections
- Validate email addresses before sending
- Implement rate limiting for email sending
- Use proper IAM roles for AWS SES access

## Integration with Authentication System

This email service is designed to integrate with the authentication system for:

- Password reset emails
- Account verification emails
- Welcome emails for new users
- Security notifications

See the authentication system documentation for specific integration examples.
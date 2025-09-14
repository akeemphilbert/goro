package email

import (
	"context"
	"time"
)

// Service defines the interface for sending transactional emails
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

// Attachment represents an email attachment
type Attachment struct {
	Filename    string
	ContentType string
	Data        []byte
}

// TemplateData represents common template data for emails
type TemplateData struct {
	UserName   string
	ResetURL   string
	ExpiryTime time.Time
	SupportURL string
}

// Provider represents different email service providers
type Provider string

const (
	ProviderSMTP Provider = "smtp"
	ProviderSES  Provider = "ses"
	ProviderMock Provider = "mock"
)

// Config holds email service configuration
type Config struct {
	Provider    Provider `yaml:"provider"`
	FromAddress string   `yaml:"from_address"`

	// SMTP configuration
	SMTP SMTPConfig `yaml:"smtp"`

	// AWS SES configuration
	SES SESConfig `yaml:"ses"`

	// Template configuration
	Templates TemplateConfig `yaml:"templates"`
}

// SMTPConfig holds SMTP server configuration
type SMTPConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	TLS      bool   `yaml:"tls"`
}

// SESConfig holds AWS SES configuration
type SESConfig struct {
	Region          string `yaml:"region"`
	AccessKeyID     string `yaml:"access_key_id"`
	SecretAccessKey string `yaml:"secret_access_key"`
}

// TemplateConfig holds email template configuration
type TemplateConfig struct {
	PasswordReset       string `yaml:"password_reset"`
	Welcome             string `yaml:"welcome"`
	AccountVerification string `yaml:"account_verification"`
}

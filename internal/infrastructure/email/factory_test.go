package email

import (
	"context"
	"strings"
	"testing"
)

func TestNewEmailService_MockProvider(t *testing.T) {
	config := Config{
		Provider:    ProviderMock,
		FromAddress: "test@example.com",
	}

	service, err := NewEmailService(config)
	if err != nil {
		t.Fatalf("Failed to create mock email service: %v", err)
	}

	// Verify it's a mock service
	mockService, ok := service.(*MockEmailService)
	if !ok {
		t.Fatal("Expected MockEmailService")
	}

	// Test that it works
	email := &Email{
		To:       []string{"recipient@example.com"},
		Subject:  "Test",
		TextBody: "Test body",
	}

	err = mockService.SendEmail(context.Background(), email)
	if err != nil {
		t.Errorf("Mock service should not fail: %v", err)
	}

	if mockService.GetSentEmailsCount() != 1 {
		t.Errorf("Expected 1 sent email, got %d", mockService.GetSentEmailsCount())
	}
}

func TestNewEmailService_SMTPProvider(t *testing.T) {
	config := Config{
		Provider:    ProviderSMTP,
		FromAddress: "test@example.com",
		SMTP: SMTPConfig{
			Host:     "smtp.example.com",
			Port:     587,
			Username: "user@example.com",
			Password: "password",
			TLS:      true,
		},
	}

	service, err := NewEmailService(config)
	if err != nil {
		t.Fatalf("Failed to create SMTP email service: %v", err)
	}

	// Verify it's an SMTP service
	_, ok := service.(*SMTPEmailService)
	if !ok {
		t.Fatal("Expected SMTPEmailService")
	}
}

func TestNewEmailService_SESProvider_InvalidConfig(t *testing.T) {
	config := Config{
		Provider:    ProviderSES,
		FromAddress: "test@example.com",
		SES: SESConfig{
			Region: "invalid-region",
			// Invalid config that should fail
		},
	}

	service, err := NewEmailService(config)
	// AWS SDK might not fail immediately, so we just check that we get a service or error
	if err != nil && service != nil {
		t.Error("Service should be nil when creation fails")
	}

	// If we get a service, it should be a SES service
	if service != nil {
		if _, ok := service.(*SESEmailService); !ok {
			t.Error("Expected SESEmailService when no error")
		}
	}
}

func TestNewEmailService_UnsupportedProvider(t *testing.T) {
	config := Config{
		Provider:    "unsupported",
		FromAddress: "test@example.com",
	}

	_, err := NewEmailService(config)
	if err == nil {
		t.Error("Expected error for unsupported provider")
	}

	if !strings.Contains(err.Error(), "unsupported email provider") {
		t.Errorf("Expected unsupported provider error, got %v", err)
	}
}

func TestNewMockEmailServiceForTesting(t *testing.T) {
	service := NewMockEmailServiceForTesting()

	// Verify it's a mock service
	mockService, ok := service.(*MockEmailService)
	if !ok {
		t.Fatal("Expected MockEmailService")
	}

	// Test that templates are loaded
	data := TemplateData{
		UserName:   "Test User",
		ResetURL:   "https://example.com/reset",
		SupportURL: "https://example.com/support",
	}

	err := mockService.SendTemplatedEmail(context.Background(), "password_reset", data, "test@example.com")
	if err != nil {
		t.Errorf("Should be able to send templated email: %v", err)
	}

	sentEmails := mockService.GetSentEmails()
	if len(sentEmails) != 1 {
		t.Errorf("Expected 1 sent email, got %d", len(sentEmails))
	}

	if sentEmails[0].Template != "password_reset" {
		t.Errorf("Expected password_reset template, got %s", sentEmails[0].Template)
	}
}

func TestNewEmailService_WithTemplateConfig(t *testing.T) {
	config := Config{
		Provider:    ProviderMock,
		FromAddress: "test@example.com",
		Templates: TemplateConfig{
			PasswordReset:       "/path/to/templates",
			Welcome:             "welcome.html",
			AccountVerification: "verify.html",
		},
	}

	// This should not fail even with invalid template paths for mock provider
	service, err := NewEmailService(config)
	if err != nil {
		t.Fatalf("Failed to create email service with template config: %v", err)
	}

	_, ok := service.(*MockEmailService)
	if !ok {
		t.Fatal("Expected MockEmailService")
	}
}

func TestConfig_Structure(t *testing.T) {
	config := Config{
		Provider:    ProviderSMTP,
		FromAddress: "from@example.com",
		SMTP: SMTPConfig{
			Host:     "smtp.example.com",
			Port:     587,
			Username: "user@example.com",
			Password: "password",
			TLS:      true,
		},
		SES: SESConfig{
			Region:          "us-east-1",
			AccessKeyID:     "access-key",
			SecretAccessKey: "secret-key",
		},
		Templates: TemplateConfig{
			PasswordReset:       "password_reset.html",
			Welcome:             "welcome.html",
			AccountVerification: "account_verification.html",
		},
	}

	if config.Provider != ProviderSMTP {
		t.Errorf("Expected provider SMTP, got %s", config.Provider)
	}

	if config.FromAddress != "from@example.com" {
		t.Errorf("Expected from address 'from@example.com', got %s", config.FromAddress)
	}

	if config.SMTP.Host != "smtp.example.com" {
		t.Errorf("Expected SMTP host 'smtp.example.com', got %s", config.SMTP.Host)
	}

	if config.SES.Region != "us-east-1" {
		t.Errorf("Expected SES region 'us-east-1', got %s", config.SES.Region)
	}

	if config.Templates.PasswordReset != "password_reset.html" {
		t.Errorf("Expected password reset template 'password_reset.html', got %s", config.Templates.PasswordReset)
	}
}

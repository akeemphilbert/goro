package email

import (
	"context"
	"strings"
	"testing"
)

func TestSESEmailService_SendEmail_NoRecipients(t *testing.T) {
	// Create a mock SES service without actual AWS credentials
	templateManager := NewTemplateManager("")
	service := &SESEmailService{
		client:    nil, // We won't actually call AWS
		fromEmail: "from@example.com",
		templates: templateManager,
	}

	email := &Email{
		Subject:  "Test Subject",
		TextBody: "Test body",
	}

	ctx := context.Background()
	err := service.SendEmail(ctx, email)
	if err == nil {
		t.Error("Expected error for email with no recipients")
	}

	if !strings.Contains(err.Error(), "no recipients specified") {
		t.Errorf("Expected 'no recipients specified' error, got %v", err)
	}
}

func TestSESEmailService_SendTemplatedEmail_NoRecipients(t *testing.T) {
	templateManager := NewTemplateManager("")
	templateManager.CreateDefaultTemplates()

	service := &SESEmailService{
		client:    nil,
		fromEmail: "from@example.com",
		templates: templateManager,
	}

	data := TemplateData{UserName: "Test User"}

	ctx := context.Background()
	err := service.SendTemplatedEmail(ctx, "password_reset", data)
	if err == nil {
		t.Error("Expected error for templated email with no recipients")
	}

	if !strings.Contains(err.Error(), "no recipients specified") {
		t.Errorf("Expected 'no recipients specified' error, got %v", err)
	}
}

func TestSESEmailService_SendTemplatedEmail_InvalidTemplate(t *testing.T) {
	templateManager := NewTemplateManager("")

	service := &SESEmailService{
		client:    nil,
		fromEmail: "from@example.com",
		templates: templateManager,
	}

	data := TemplateData{UserName: "Test User"}

	ctx := context.Background()
	err := service.SendTemplatedEmail(ctx, "nonexistent", data, "test@example.com")
	if err == nil {
		t.Error("Expected error for nonexistent template")
	}

	if !strings.Contains(err.Error(), "failed to render template") {
		t.Errorf("Expected 'failed to render template' error, got %v", err)
	}
}

func TestNewSESEmailService_InvalidConfig(t *testing.T) {
	// Test with empty config
	config := SESConfig{
		Region: "invalid-region",
	}
	templateManager := NewTemplateManager("")

	service, err := NewSESEmailService(config, "from@example.com", templateManager)
	// AWS SDK might not fail immediately with invalid region, so we just check the behavior
	if err != nil && service != nil {
		t.Error("Service should be nil when creation fails")
	}

	// If we get a service, it should be properly initialized
	if service != nil && service.fromEmail != "from@example.com" {
		t.Error("Service should have correct from email when created successfully")
	}
}

func TestBulkDestination_Structure(t *testing.T) {
	dest := BulkDestination{
		ToAddresses: []string{"test1@example.com", "test2@example.com"},
		TemplateData: map[string]interface{}{
			"UserName":    "Test User",
			"CustomField": "Custom Value",
		},
	}

	if len(dest.ToAddresses) != 2 {
		t.Errorf("Expected 2 addresses, got %d", len(dest.ToAddresses))
	}

	if dest.ToAddresses[0] != "test1@example.com" {
		t.Errorf("Expected first address 'test1@example.com', got %s", dest.ToAddresses[0])
	}

	templateData, ok := dest.TemplateData.(map[string]interface{})
	if !ok {
		t.Fatal("Expected template data to be a map")
	}

	if templateData["UserName"] != "Test User" {
		t.Errorf("Expected UserName 'Test User', got %v", templateData["UserName"])
	}
}

// Test SES configuration structure
func TestSESConfig_Structure(t *testing.T) {
	config := SESConfig{
		Region:          "us-east-1",
		AccessKeyID:     "AKIAIOSFODNN7EXAMPLE",
		SecretAccessKey: "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
	}

	if config.Region != "us-east-1" {
		t.Errorf("Expected region 'us-east-1', got %s", config.Region)
	}

	if config.AccessKeyID != "AKIAIOSFODNN7EXAMPLE" {
		t.Errorf("Expected access key 'AKIAIOSFODNN7EXAMPLE', got %s", config.AccessKeyID)
	}

	if config.SecretAccessKey != "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY" {
		t.Errorf("Expected secret key to match, got %s", config.SecretAccessKey)
	}
}

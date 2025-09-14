package email

import (
	"strings"
	"testing"
	"time"
)

func TestTemplateManager_CreateDefaultTemplates(t *testing.T) {
	tm := NewTemplateManager("")

	err := tm.CreateDefaultTemplates()
	if err != nil {
		t.Fatalf("Failed to create default templates: %v", err)
	}

	// Test password reset template
	data := TemplateData{
		UserName:   "John Doe",
		ResetURL:   "https://example.com/reset/token123",
		ExpiryTime: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
		SupportURL: "https://example.com/support",
	}

	subject, textBody, htmlBody, err := tm.RenderTemplate("password_reset", data)
	if err != nil {
		t.Fatalf("Failed to render password reset template: %v", err)
	}

	if subject != "Password Reset Request" {
		t.Errorf("Expected subject 'Password Reset Request', got %s", subject)
	}

	if !strings.Contains(textBody, "John Doe") {
		t.Error("Text body should contain user name")
	}

	if !strings.Contains(textBody, "https://example.com/reset/token123") {
		t.Error("Text body should contain reset URL")
	}

	if !strings.Contains(htmlBody, "John Doe") {
		t.Error("HTML body should contain user name")
	}

	if !strings.Contains(htmlBody, "https://example.com/reset/token123") {
		t.Error("HTML body should contain reset URL")
	}
}

func TestTemplateManager_RenderTemplate_Welcome(t *testing.T) {
	tm := NewTemplateManager("")
	tm.CreateDefaultTemplates()

	data := TemplateData{
		UserName:   "Alice Smith",
		SupportURL: "https://example.com/help",
	}

	subject, textBody, htmlBody, err := tm.RenderTemplate("welcome", data)
	if err != nil {
		t.Fatalf("Failed to render welcome template: %v", err)
	}

	if subject != "Welcome to Solid Pod" {
		t.Errorf("Expected subject 'Welcome to Solid Pod', got %s", subject)
	}

	if !strings.Contains(textBody, "Alice Smith") {
		t.Error("Text body should contain user name")
	}

	if !strings.Contains(htmlBody, "Alice Smith") {
		t.Error("HTML body should contain user name")
	}

	if !strings.Contains(textBody, "https://example.com/help") {
		t.Error("Text body should contain support URL")
	}
}

func TestTemplateManager_RenderTemplate_NotFound(t *testing.T) {
	tm := NewTemplateManager("")

	_, _, _, err := tm.RenderTemplate("nonexistent", nil)
	if err == nil {
		t.Error("Expected error for nonexistent template")
	}

	if !strings.Contains(err.Error(), "template nonexistent not found") {
		t.Errorf("Expected 'template not found' error, got %v", err)
	}
}

func TestTemplateManager_ExtractSubject(t *testing.T) {
	tm := NewTemplateManager("")

	tests := []struct {
		name         string
		templateName string
		data         interface{}
		expected     string
	}{
		{
			name:         "password reset default",
			templateName: "password_reset",
			data:         nil,
			expected:     "Password Reset Request",
		},
		{
			name:         "welcome default",
			templateName: "welcome",
			data:         nil,
			expected:     "Welcome to Solid Pod",
		},
		{
			name:         "custom subject from data",
			templateName: "custom",
			data:         map[string]interface{}{"Subject": "Custom Subject"},
			expected:     "Custom Subject",
		},
		{
			name:         "unknown template",
			templateName: "unknown_template",
			data:         nil,
			expected:     "Unknown Template",
		},
		{
			name:         "template with underscores",
			templateName: "account_verification",
			data:         nil,
			expected:     "Verify Your Account",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tm.extractSubject(tt.templateName, tt.data)
			if result != tt.expected {
				t.Errorf("Expected subject %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestTemplateManager_RenderTemplateDataAsJSON(t *testing.T) {
	tm := NewTemplateManager("")

	data := TemplateData{
		UserName:   "John Doe",
		ResetURL:   "https://example.com/reset",
		SupportURL: "https://example.com/support",
	}

	jsonStr, err := tm.RenderTemplateDataAsJSON(data)
	if err != nil {
		t.Fatalf("Failed to render template data as JSON: %v", err)
	}

	if !strings.Contains(jsonStr, "John Doe") {
		t.Error("JSON should contain user name")
	}

	if !strings.Contains(jsonStr, "https://example.com/reset") {
		t.Error("JSON should contain reset URL")
	}

	// Test with invalid data (circular reference)
	invalidData := make(map[string]interface{})
	invalidData["self"] = invalidData

	_, err = tm.RenderTemplateDataAsJSON(invalidData)
	if err == nil {
		t.Error("Expected error for circular reference data")
	}
}

func TestTemplateManager_RenderTemplate_WithMapData(t *testing.T) {
	tm := NewTemplateManager("")
	tm.CreateDefaultTemplates()

	// Test with map data instead of struct
	data := map[string]interface{}{
		"UserName":   "Map User",
		"ResetURL":   "https://example.com/map-reset",
		"ExpiryTime": time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC),
		"SupportURL": "https://example.com/map-support",
	}

	subject, textBody, htmlBody, err := tm.RenderTemplate("password_reset", data)
	if err != nil {
		t.Fatalf("Failed to render template with map data: %v", err)
	}

	if subject != "Password Reset Request" {
		t.Errorf("Expected subject 'Password Reset Request', got %s", subject)
	}

	if !strings.Contains(textBody, "Map User") {
		t.Error("Text body should contain user name from map")
	}

	if !strings.Contains(htmlBody, "https://example.com/map-reset") {
		t.Error("HTML body should contain reset URL from map")
	}
}

func TestTemplateManager_EmptyTemplateDir(t *testing.T) {
	tm := NewTemplateManager("")

	// Should not fail when template directory is empty
	if tm.templateDir != "" {
		t.Errorf("Expected empty template directory, got %s", tm.templateDir)
	}

	// Should be able to create default templates
	err := tm.CreateDefaultTemplates()
	if err != nil {
		t.Errorf("Should be able to create default templates with empty dir: %v", err)
	}
}

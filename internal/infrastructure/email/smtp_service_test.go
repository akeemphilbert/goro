package email

import (
	"context"
	"strings"
	"testing"
)

func TestSMTPEmailService_NewSMTPEmailService(t *testing.T) {
	config := SMTPConfig{
		Host:     "smtp.example.com",
		Port:     587,
		Username: "user@example.com",
		Password: "password",
		TLS:      true,
	}

	templateManager := NewTemplateManager("")
	templateManager.CreateDefaultTemplates()

	service := NewSMTPEmailService(config, "from@example.com", templateManager)

	if service.host != "smtp.example.com" {
		t.Errorf("Expected host 'smtp.example.com', got %s", service.host)
	}

	if service.port != 587 {
		t.Errorf("Expected port 587, got %d", service.port)
	}

	if service.username != "user@example.com" {
		t.Errorf("Expected username 'user@example.com', got %s", service.username)
	}

	if service.fromEmail != "from@example.com" {
		t.Errorf("Expected fromEmail 'from@example.com', got %s", service.fromEmail)
	}

	if !service.useTLS {
		t.Error("Expected TLS to be enabled")
	}
}

func TestSMTPEmailService_SendEmail_NoRecipients(t *testing.T) {
	config := SMTPConfig{
		Host: "smtp.example.com",
		Port: 587,
	}

	templateManager := NewTemplateManager("")
	service := NewSMTPEmailService(config, "from@example.com", templateManager)

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

func TestSMTPEmailService_SendTemplatedEmail_NoRecipients(t *testing.T) {
	config := SMTPConfig{
		Host: "smtp.example.com",
		Port: 587,
	}

	templateManager := NewTemplateManager("")
	templateManager.CreateDefaultTemplates()
	service := NewSMTPEmailService(config, "from@example.com", templateManager)

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

func TestSMTPEmailService_SendTemplatedEmail_InvalidTemplate(t *testing.T) {
	config := SMTPConfig{
		Host: "smtp.example.com",
		Port: 587,
	}

	templateManager := NewTemplateManager("")
	service := NewSMTPEmailService(config, "from@example.com", templateManager)

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

func TestSMTPEmailService_BuildMessage(t *testing.T) {
	config := SMTPConfig{
		Host: "smtp.example.com",
		Port: 587,
	}

	templateManager := NewTemplateManager("")
	service := NewSMTPEmailService(config, "from@example.com", templateManager)

	tests := []struct {
		name     string
		email    *Email
		contains []string
	}{
		{
			name: "text only email",
			email: &Email{
				To:       []string{"test@example.com"},
				Subject:  "Text Only",
				TextBody: "This is a text email",
			},
			contains: []string{
				"From: from@example.com",
				"To: test@example.com",
				"Subject: Text Only",
				"Content-Type: text/plain; charset=UTF-8",
				"This is a text email",
			},
		},
		{
			name: "html only email",
			email: &Email{
				To:       []string{"test@example.com"},
				Subject:  "HTML Only",
				HTMLBody: "<p>This is an HTML email</p>",
			},
			contains: []string{
				"From: from@example.com",
				"To: test@example.com",
				"Subject: HTML Only",
				"Content-Type: text/html; charset=UTF-8",
				"<p>This is an HTML email</p>",
			},
		},
		{
			name: "multipart email",
			email: &Email{
				To:       []string{"test@example.com"},
				Subject:  "Multipart",
				TextBody: "Text version",
				HTMLBody: "<p>HTML version</p>",
			},
			contains: []string{
				"From: from@example.com",
				"To: test@example.com",
				"Subject: Multipart",
				"Content-Type: multipart/alternative",
				"Text version",
				"<p>HTML version</p>",
			},
		},
		{
			name: "email with CC and BCC",
			email: &Email{
				To:       []string{"to@example.com"},
				CC:       []string{"cc@example.com"},
				BCC:      []string{"bcc@example.com"},
				Subject:  "With CC and BCC",
				TextBody: "Test body",
			},
			contains: []string{
				"To: to@example.com",
				"Cc: cc@example.com",
				"Subject: With CC and BCC",
			},
		},
		{
			name: "multiple recipients",
			email: &Email{
				To:       []string{"to1@example.com", "to2@example.com"},
				Subject:  "Multiple Recipients",
				TextBody: "Test body",
			},
			contains: []string{
				"To: to1@example.com, to2@example.com",
				"Subject: Multiple Recipients",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			message := service.buildMessage(tt.email)

			for _, expected := range tt.contains {
				if !strings.Contains(message, expected) {
					t.Errorf("Message should contain '%s'\nMessage:\n%s", expected, message)
				}
			}

			// Check that message ends with CRLF
			if !strings.HasSuffix(message, "\r\n") && !strings.HasSuffix(message, "\n") {
				t.Error("Message should end with line terminator")
			}
		})
	}
}

func TestSMTPEmailService_BuildMessage_EmptyBodies(t *testing.T) {
	config := SMTPConfig{
		Host: "smtp.example.com",
		Port: 587,
	}

	templateManager := NewTemplateManager("")
	service := NewSMTPEmailService(config, "from@example.com", templateManager)

	email := &Email{
		To:      []string{"test@example.com"},
		Subject: "Empty Body",
		// No TextBody or HTMLBody
	}

	message := service.buildMessage(email)

	// Should still contain headers
	if !strings.Contains(message, "From: from@example.com") {
		t.Error("Message should contain From header")
	}

	if !strings.Contains(message, "To: test@example.com") {
		t.Error("Message should contain To header")
	}

	if !strings.Contains(message, "Subject: Empty Body") {
		t.Error("Message should contain Subject header")
	}

	// Should default to text/plain
	if !strings.Contains(message, "Content-Type: text/plain; charset=UTF-8") {
		t.Error("Message should default to text/plain content type")
	}
}

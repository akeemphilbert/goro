package email

import (
	"context"
	"errors"
	"testing"
)

func TestMockEmailService_SendEmail(t *testing.T) {
	templateManager := NewTemplateManager("")
	templateManager.CreateDefaultTemplates()
	service := NewMockEmailService(templateManager)

	email := &Email{
		To:       []string{"test@example.com"},
		Subject:  "Test Subject",
		TextBody: "Test body",
	}

	ctx := context.Background()
	err := service.SendEmail(ctx, email)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	sentEmails := service.GetSentEmails()
	if len(sentEmails) != 1 {
		t.Fatalf("Expected 1 sent email, got %d", len(sentEmails))
	}

	sentEmail := sentEmails[0]
	if sentEmail.Email.Subject != "Test Subject" {
		t.Errorf("Expected subject 'Test Subject', got %s", sentEmail.Email.Subject)
	}

	if len(sentEmail.Email.To) != 1 || sentEmail.Email.To[0] != "test@example.com" {
		t.Errorf("Expected recipient 'test@example.com', got %v", sentEmail.Email.To)
	}
}

func TestMockEmailService_SendEmail_NoRecipients(t *testing.T) {
	templateManager := NewTemplateManager("")
	service := NewMockEmailService(templateManager)

	email := &Email{
		Subject:  "Test Subject",
		TextBody: "Test body",
	}

	ctx := context.Background()
	err := service.SendEmail(ctx, email)
	if err == nil {
		t.Fatal("Expected error for email with no recipients")
	}

	if service.GetSentEmailsCount() != 0 {
		t.Errorf("Expected 0 sent emails, got %d", service.GetSentEmailsCount())
	}
}

func TestMockEmailService_SendTemplatedEmail(t *testing.T) {
	templateManager := NewTemplateManager("")
	templateManager.CreateDefaultTemplates()
	service := NewMockEmailService(templateManager)

	data := TemplateData{
		UserName:   "John Doe",
		ResetURL:   "https://example.com/reset",
		SupportURL: "https://example.com/support",
	}

	ctx := context.Background()
	err := service.SendTemplatedEmail(ctx, "password_reset", data, "test@example.com")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	sentEmails := service.GetSentEmails()
	if len(sentEmails) != 1 {
		t.Fatalf("Expected 1 sent email, got %d", len(sentEmails))
	}

	sentEmail := sentEmails[0]
	if sentEmail.Template != "password_reset" {
		t.Errorf("Expected template 'password_reset', got %s", sentEmail.Template)
	}

	if sentEmail.Email.Subject != "Password Reset Request" {
		t.Errorf("Expected subject 'Password Reset Request', got %s", sentEmail.Email.Subject)
	}
}

func TestMockEmailService_GetLastSentEmail(t *testing.T) {
	templateManager := NewTemplateManager("")
	service := NewMockEmailService(templateManager)

	// No emails sent yet
	lastEmail := service.GetLastSentEmail()
	if lastEmail != nil {
		t.Error("Expected nil for last email when no emails sent")
	}

	// Send first email
	email1 := &Email{
		To:      []string{"test1@example.com"},
		Subject: "First Email",
	}
	service.SendEmail(context.Background(), email1)

	// Send second email
	email2 := &Email{
		To:      []string{"test2@example.com"},
		Subject: "Second Email",
	}
	service.SendEmail(context.Background(), email2)

	lastEmail = service.GetLastSentEmail()
	if lastEmail == nil {
		t.Fatal("Expected last email to not be nil")
	}

	if lastEmail.Email.Subject != "Second Email" {
		t.Errorf("Expected last email subject 'Second Email', got %s", lastEmail.Email.Subject)
	}
}

func TestMockEmailService_FindEmailsByRecipient(t *testing.T) {
	templateManager := NewTemplateManager("")
	service := NewMockEmailService(templateManager)

	// Send emails to different recipients
	email1 := &Email{
		To:      []string{"alice@example.com"},
		Subject: "Email to Alice",
	}
	service.SendEmail(context.Background(), email1)

	email2 := &Email{
		To:      []string{"bob@example.com"},
		Subject: "Email to Bob",
	}
	service.SendEmail(context.Background(), email2)

	email3 := &Email{
		To:      []string{"alice@example.com"},
		Subject: "Another email to Alice",
	}
	service.SendEmail(context.Background(), email3)

	// Find emails for Alice
	aliceEmails := service.FindEmailsByRecipient("alice@example.com")
	if len(aliceEmails) != 2 {
		t.Errorf("Expected 2 emails for Alice, got %d", len(aliceEmails))
	}

	// Find emails for Bob
	bobEmails := service.FindEmailsByRecipient("bob@example.com")
	if len(bobEmails) != 1 {
		t.Errorf("Expected 1 email for Bob, got %d", len(bobEmails))
	}

	// Find emails for non-existent recipient
	charlieEmails := service.FindEmailsByRecipient("charlie@example.com")
	if len(charlieEmails) != 0 {
		t.Errorf("Expected 0 emails for Charlie, got %d", len(charlieEmails))
	}
}

func TestMockEmailService_FindEmailsByTemplate(t *testing.T) {
	templateManager := NewTemplateManager("")
	templateManager.CreateDefaultTemplates()
	service := NewMockEmailService(templateManager)

	data := TemplateData{UserName: "Test User"}

	// Send templated emails
	service.SendTemplatedEmail(context.Background(), "password_reset", data, "test1@example.com")
	service.SendTemplatedEmail(context.Background(), "welcome", data, "test2@example.com")
	service.SendTemplatedEmail(context.Background(), "password_reset", data, "test3@example.com")

	// Find password reset emails
	resetEmails := service.FindEmailsByTemplate("password_reset")
	if len(resetEmails) != 2 {
		t.Errorf("Expected 2 password reset emails, got %d", len(resetEmails))
	}

	// Find welcome emails
	welcomeEmails := service.FindEmailsByTemplate("welcome")
	if len(welcomeEmails) != 1 {
		t.Errorf("Expected 1 welcome email, got %d", len(welcomeEmails))
	}
}

func TestMockEmailService_SetShouldFail(t *testing.T) {
	templateManager := NewTemplateManager("")
	service := NewMockEmailService(templateManager)

	testError := errors.New("test error")
	service.SetShouldFail(true, testError)

	email := &Email{
		To:      []string{"test@example.com"},
		Subject: "Test",
	}

	err := service.SendEmail(context.Background(), email)
	if err != testError {
		t.Errorf("Expected test error, got %v", err)
	}

	if service.GetSentEmailsCount() != 0 {
		t.Errorf("Expected 0 sent emails when failing, got %d", service.GetSentEmailsCount())
	}

	// Reset failure state
	service.SetShouldFail(false, nil)

	err = service.SendEmail(context.Background(), email)
	if err != nil {
		t.Errorf("Expected no error after resetting failure state, got %v", err)
	}

	if service.GetSentEmailsCount() != 1 {
		t.Errorf("Expected 1 sent email after success, got %d", service.GetSentEmailsCount())
	}
}

func TestMockEmailService_Clear(t *testing.T) {
	templateManager := NewTemplateManager("")
	service := NewMockEmailService(templateManager)

	// Send some emails
	email := &Email{
		To:      []string{"test@example.com"},
		Subject: "Test",
	}
	service.SendEmail(context.Background(), email)
	service.SendEmail(context.Background(), email)

	if service.GetSentEmailsCount() != 2 {
		t.Errorf("Expected 2 sent emails before clear, got %d", service.GetSentEmailsCount())
	}

	// Clear emails
	service.Clear()

	if service.GetSentEmailsCount() != 0 {
		t.Errorf("Expected 0 sent emails after clear, got %d", service.GetSentEmailsCount())
	}

	if service.GetLastSentEmail() != nil {
		t.Error("Expected nil last email after clear")
	}
}

func TestMockEmailService_ConcurrentAccess(t *testing.T) {
	templateManager := NewTemplateManager("")
	service := NewMockEmailService(templateManager)

	// Test concurrent access
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(i int) {
			email := &Email{
				To:      []string{"test@example.com"},
				Subject: "Concurrent Test",
			}
			service.SendEmail(context.Background(), email)
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	if service.GetSentEmailsCount() != 10 {
		t.Errorf("Expected 10 sent emails from concurrent access, got %d", service.GetSentEmailsCount())
	}
}

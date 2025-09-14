package email

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// MockEmailService implements the Service interface for testing
type MockEmailService struct {
	mu         sync.RWMutex
	sentEmails []SentEmail
	templates  *TemplateManager
	shouldFail bool
	failError  error
}

// SentEmail represents an email that was sent through the mock service
type SentEmail struct {
	Email     *Email
	Template  string
	Data      interface{}
	Timestamp time.Time
}

// NewMockEmailService creates a new mock email service
func NewMockEmailService(templates *TemplateManager) *MockEmailService {
	return &MockEmailService{
		sentEmails: make([]SentEmail, 0),
		templates:  templates,
	}
}

// SendEmail records the email as sent
func (m *MockEmailService) SendEmail(ctx context.Context, email *Email) error {
	if m.shouldFail {
		return m.failError
	}

	if len(email.To) == 0 {
		return fmt.Errorf("no recipients specified")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.sentEmails = append(m.sentEmails, SentEmail{
		Email:     email,
		Timestamp: time.Now(),
	})

	return nil
}

// SendTemplatedEmail records the templated email as sent
func (m *MockEmailService) SendTemplatedEmail(ctx context.Context, template string, data interface{}, recipients ...string) error {
	if m.shouldFail {
		return m.failError
	}

	if len(recipients) == 0 {
		return fmt.Errorf("no recipients specified")
	}

	subject, textBody, htmlBody, err := m.templates.RenderTemplate(template, data)
	if err != nil {
		return fmt.Errorf("failed to render template: %w", err)
	}

	email := &Email{
		To:       recipients,
		Subject:  subject,
		TextBody: textBody,
		HTMLBody: htmlBody,
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.sentEmails = append(m.sentEmails, SentEmail{
		Email:     email,
		Template:  template,
		Data:      data,
		Timestamp: time.Now(),
	})

	return nil
}

// GetSentEmails returns all emails that were sent
func (m *MockEmailService) GetSentEmails() []SentEmail {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Return a copy to prevent race conditions
	emails := make([]SentEmail, len(m.sentEmails))
	copy(emails, m.sentEmails)
	return emails
}

// GetLastSentEmail returns the last email that was sent
func (m *MockEmailService) GetLastSentEmail() *SentEmail {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if len(m.sentEmails) == 0 {
		return nil
	}

	return &m.sentEmails[len(m.sentEmails)-1]
}

// GetSentEmailsCount returns the number of emails sent
func (m *MockEmailService) GetSentEmailsCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return len(m.sentEmails)
}

// Clear removes all sent emails from the mock
func (m *MockEmailService) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.sentEmails = m.sentEmails[:0]
}

// SetShouldFail configures the mock to fail with the given error
func (m *MockEmailService) SetShouldFail(shouldFail bool, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.shouldFail = shouldFail
	m.failError = err
}

// FindEmailsByRecipient finds emails sent to a specific recipient
func (m *MockEmailService) FindEmailsByRecipient(recipient string) []SentEmail {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var found []SentEmail
	for _, email := range m.sentEmails {
		for _, to := range email.Email.To {
			if to == recipient {
				found = append(found, email)
				break
			}
		}
	}

	return found
}

// FindEmailsByTemplate finds emails sent using a specific template
func (m *MockEmailService) FindEmailsByTemplate(template string) []SentEmail {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var found []SentEmail
	for _, email := range m.sentEmails {
		if email.Template == template {
			found = append(found, email)
		}
	}

	return found
}

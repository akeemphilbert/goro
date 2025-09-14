package email

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/smtp"
	"strings"
)

// SMTPEmailService implements the Service interface using SMTP
type SMTPEmailService struct {
	host      string
	port      int
	username  string
	password  string
	fromEmail string
	useTLS    bool
	templates *TemplateManager
}

// NewSMTPEmailService creates a new SMTP email service
func NewSMTPEmailService(config SMTPConfig, fromEmail string, templates *TemplateManager) *SMTPEmailService {
	return &SMTPEmailService{
		host:      config.Host,
		port:      config.Port,
		username:  config.Username,
		password:  config.Password,
		fromEmail: fromEmail,
		useTLS:    config.TLS,
		templates: templates,
	}
}

// SendEmail sends an email using SMTP
func (s *SMTPEmailService) SendEmail(ctx context.Context, email *Email) error {
	if len(email.To) == 0 {
		return fmt.Errorf("no recipients specified")
	}

	// Create SMTP client
	addr := fmt.Sprintf("%s:%d", s.host, s.port)

	var client *smtp.Client
	var err error

	if s.useTLS {
		tlsConfig := &tls.Config{
			ServerName: s.host,
		}
		conn, err := tls.Dial("tcp", addr, tlsConfig)
		if err != nil {
			return fmt.Errorf("failed to connect with TLS: %w", err)
		}
		defer conn.Close()

		client, err = smtp.NewClient(conn, s.host)
		if err != nil {
			return fmt.Errorf("failed to create SMTP client: %w", err)
		}
	} else {
		client, err = smtp.Dial(addr)
		if err != nil {
			return fmt.Errorf("failed to connect to SMTP server: %w", err)
		}
	}
	defer client.Close()

	// Authenticate if credentials are provided
	if s.username != "" && s.password != "" {
		auth := smtp.PlainAuth("", s.username, s.password, s.host)
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("SMTP authentication failed: %w", err)
		}
	}

	// Set sender
	if err := client.Mail(s.fromEmail); err != nil {
		return fmt.Errorf("failed to set sender: %w", err)
	}

	// Set recipients
	allRecipients := append(email.To, email.CC...)
	allRecipients = append(allRecipients, email.BCC...)

	for _, recipient := range allRecipients {
		if err := client.Rcpt(recipient); err != nil {
			return fmt.Errorf("failed to set recipient %s: %w", recipient, err)
		}
	}

	// Send email data
	writer, err := client.Data()
	if err != nil {
		return fmt.Errorf("failed to get data writer: %w", err)
	}
	defer writer.Close()

	// Build email message
	message := s.buildMessage(email)
	if _, err := writer.Write([]byte(message)); err != nil {
		return fmt.Errorf("failed to write email data: %w", err)
	}

	return nil
}

// SendTemplatedEmail sends an email using a template
func (s *SMTPEmailService) SendTemplatedEmail(ctx context.Context, template string, data interface{}, recipients ...string) error {
	if len(recipients) == 0 {
		return fmt.Errorf("no recipients specified")
	}

	subject, textBody, htmlBody, err := s.templates.RenderTemplate(template, data)
	if err != nil {
		return fmt.Errorf("failed to render template: %w", err)
	}

	email := &Email{
		To:       recipients,
		Subject:  subject,
		TextBody: textBody,
		HTMLBody: htmlBody,
	}

	return s.SendEmail(ctx, email)
}

// buildMessage constructs the email message with headers
func (s *SMTPEmailService) buildMessage(email *Email) string {
	var message strings.Builder

	// Headers
	message.WriteString(fmt.Sprintf("From: %s\r\n", s.fromEmail))
	message.WriteString(fmt.Sprintf("To: %s\r\n", strings.Join(email.To, ", ")))

	if len(email.CC) > 0 {
		message.WriteString(fmt.Sprintf("Cc: %s\r\n", strings.Join(email.CC, ", ")))
	}

	message.WriteString(fmt.Sprintf("Subject: %s\r\n", email.Subject))
	message.WriteString("MIME-Version: 1.0\r\n")

	// Content type based on available body types
	if email.HTMLBody != "" && email.TextBody != "" {
		boundary := "boundary-mixed-content"
		message.WriteString(fmt.Sprintf("Content-Type: multipart/alternative; boundary=%s\r\n", boundary))
		message.WriteString("\r\n")

		// Text part
		message.WriteString(fmt.Sprintf("--%s\r\n", boundary))
		message.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
		message.WriteString("\r\n")
		message.WriteString(email.TextBody)
		message.WriteString("\r\n")

		// HTML part
		message.WriteString(fmt.Sprintf("--%s\r\n", boundary))
		message.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
		message.WriteString("\r\n")
		message.WriteString(email.HTMLBody)
		message.WriteString("\r\n")

		message.WriteString(fmt.Sprintf("--%s--\r\n", boundary))
	} else if email.HTMLBody != "" {
		message.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
		message.WriteString("\r\n")
		message.WriteString(email.HTMLBody)
		message.WriteString("\r\n")
	} else {
		message.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
		message.WriteString("\r\n")
		message.WriteString(email.TextBody)
		message.WriteString("\r\n")
	}

	return message.String()
}

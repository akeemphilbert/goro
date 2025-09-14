package email

import (
	"context"
	"fmt"
	"time"
)

// ExampleUsage demonstrates how to use the email service
func ExampleUsage() {
	// Example 1: Using Mock Email Service for testing
	mockService := NewMockEmailServiceForTesting()

	// Send a simple email
	email := &Email{
		To:       []string{"user@example.com"},
		Subject:  "Welcome to Solid Pod",
		TextBody: "Welcome to your new Solid Pod account!",
		HTMLBody: "<h1>Welcome to your new Solid Pod account!</h1>",
	}

	ctx := context.Background()
	err := mockService.SendEmail(ctx, email)
	if err != nil {
		fmt.Printf("Failed to send email: %v\n", err)
		return
	}

	fmt.Printf("Email sent successfully! Total emails sent: %d\n",
		mockService.(*MockEmailService).GetSentEmailsCount())

	// Example 2: Using templated email
	templateData := TemplateData{
		UserName:   "John Doe",
		ResetURL:   "https://mypod.example.com/reset/abc123",
		ExpiryTime: time.Now().Add(1 * time.Hour),
		SupportURL: "https://mypod.example.com/support",
	}

	err = mockService.SendTemplatedEmail(ctx, "password_reset", templateData, "john@example.com")
	if err != nil {
		fmt.Printf("Failed to send templated email: %v\n", err)
		return
	}

	fmt.Printf("Templated email sent successfully!\n")

	// Example 3: Creating email service from configuration
	config := Config{
		Provider:    ProviderMock,
		FromAddress: "noreply@mypod.example.com",
		Templates: TemplateConfig{
			PasswordReset:       "", // Use default templates
			Welcome:             "",
			AccountVerification: "",
		},
	}

	service, err := NewEmailService(config)
	if err != nil {
		fmt.Printf("Failed to create email service: %v\n", err)
		return
	}

	// Use the service
	err = service.SendTemplatedEmail(ctx, "welcome", templateData, "newuser@example.com")
	if err != nil {
		fmt.Printf("Failed to send welcome email: %v\n", err)
		return
	}

	fmt.Printf("Welcome email sent via configured service!\n")
}

// ExampleSMTPConfiguration shows how to configure SMTP email service
func ExampleSMTPConfiguration() Config {
	return Config{
		Provider:    ProviderSMTP,
		FromAddress: "noreply@mypod.example.com",
		SMTP: SMTPConfig{
			Host:     "smtp.gmail.com",
			Port:     587,
			Username: "your-email@gmail.com",
			Password: "your-app-password",
			TLS:      true,
		},
		Templates: TemplateConfig{
			PasswordReset:       "/path/to/templates/password_reset.html",
			Welcome:             "/path/to/templates/welcome.html",
			AccountVerification: "/path/to/templates/verify.html",
		},
	}
}

// ExampleSESConfiguration shows how to configure AWS SES email service
func ExampleSESConfiguration() Config {
	return Config{
		Provider:    ProviderSES,
		FromAddress: "noreply@mypod.example.com",
		SES: SESConfig{
			Region:          "us-east-1",
			AccessKeyID:     "AKIAIOSFODNN7EXAMPLE",
			SecretAccessKey: "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
		},
		Templates: TemplateConfig{
			PasswordReset:       "", // Use default templates
			Welcome:             "",
			AccountVerification: "",
		},
	}
}

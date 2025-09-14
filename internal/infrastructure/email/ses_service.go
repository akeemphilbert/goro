package email

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/ses"
	"github.com/aws/aws-sdk-go-v2/service/ses/types"
)

// SESEmailService implements the Service interface using AWS SES
type SESEmailService struct {
	client    *ses.Client
	fromEmail string
	templates *TemplateManager
}

// NewSESEmailService creates a new AWS SES email service
func NewSESEmailService(sesConfig SESConfig, fromEmail string, templates *TemplateManager) (*SESEmailService, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(sesConfig.Region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			sesConfig.AccessKeyID,
			sesConfig.SecretAccessKey,
			"",
		)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := ses.NewFromConfig(cfg)

	return &SESEmailService{
		client:    client,
		fromEmail: fromEmail,
		templates: templates,
	}, nil
}

// SendEmail sends an email using AWS SES
func (s *SESEmailService) SendEmail(ctx context.Context, email *Email) error {
	if len(email.To) == 0 {
		return fmt.Errorf("no recipients specified")
	}

	// Prepare destinations
	destinations := &types.Destination{
		ToAddresses: email.To,
	}

	if len(email.CC) > 0 {
		destinations.CcAddresses = email.CC
	}

	if len(email.BCC) > 0 {
		destinations.BccAddresses = email.BCC
	}

	// Prepare message content
	message := &types.Message{
		Subject: &types.Content{
			Data:    aws.String(email.Subject),
			Charset: aws.String("UTF-8"),
		},
	}

	// Set body content
	body := &types.Body{}

	if email.TextBody != "" {
		body.Text = &types.Content{
			Data:    aws.String(email.TextBody),
			Charset: aws.String("UTF-8"),
		}
	}

	if email.HTMLBody != "" {
		body.Html = &types.Content{
			Data:    aws.String(email.HTMLBody),
			Charset: aws.String("UTF-8"),
		}
	}

	message.Body = body

	// Send email
	input := &ses.SendEmailInput{
		Source:      aws.String(s.fromEmail),
		Destination: destinations,
		Message:     message,
	}

	_, err := s.client.SendEmail(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to send email via SES: %w", err)
	}

	return nil
}

// SendTemplatedEmail sends an email using a template
func (s *SESEmailService) SendTemplatedEmail(ctx context.Context, template string, data interface{}, recipients ...string) error {
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

// SendBulkTemplatedEmail sends templated emails to multiple recipients using SES bulk operations
func (s *SESEmailService) SendBulkTemplatedEmail(ctx context.Context, templateName string, defaultTemplateData interface{}, destinations []BulkDestination) error {
	if len(destinations) == 0 {
		return fmt.Errorf("no destinations specified")
	}

	// Convert destinations to SES format
	sesDestinations := make([]types.BulkEmailDestination, len(destinations))
	for i, dest := range destinations {
		templateDataJSON, err := s.templates.RenderTemplateDataAsJSON(dest.TemplateData)
		if err != nil {
			return fmt.Errorf("failed to render template data for destination %d: %w", i, err)
		}

		sesDestinations[i] = types.BulkEmailDestination{
			Destination: &types.Destination{
				ToAddresses: dest.ToAddresses,
			},
			ReplacementTemplateData: aws.String(templateDataJSON),
		}
	}

	// Prepare default template data
	defaultTemplateDataJSON, err := s.templates.RenderTemplateDataAsJSON(defaultTemplateData)
	if err != nil {
		return fmt.Errorf("failed to render default template data: %w", err)
	}

	input := &ses.SendBulkTemplatedEmailInput{
		Source:              aws.String(s.fromEmail),
		Template:            aws.String(templateName),
		DefaultTemplateData: aws.String(defaultTemplateDataJSON),
		Destinations:        sesDestinations,
	}

	_, err = s.client.SendBulkTemplatedEmail(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to send bulk templated email via SES: %w", err)
	}

	return nil
}

// BulkDestination represents a destination for bulk email sending
type BulkDestination struct {
	ToAddresses  []string
	TemplateData interface{}
}

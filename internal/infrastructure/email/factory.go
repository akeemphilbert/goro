package email

import (
	"fmt"
)

// NewEmailService creates an email service based on the configuration
func NewEmailService(config Config) (Service, error) {
	// Create template manager - use empty string for template directory if not specified
	templateDir := ""
	if config.Templates.PasswordReset != "" {
		templateDir = config.Templates.PasswordReset
	}

	templateManager := NewTemplateManager(templateDir)

	// Always create default templates first, then try to load from files if directory is specified
	if err := templateManager.CreateDefaultTemplates(); err != nil {
		return nil, fmt.Errorf("failed to create default templates: %w", err)
	}

	// If template directory is specified, try to load additional templates
	if templateDir != "" {
		// Ignore errors from loading file templates - default templates are sufficient
		templateManager.LoadAllTemplates()
	}

	switch config.Provider {
	case ProviderSMTP:
		return NewSMTPEmailService(config.SMTP, config.FromAddress, templateManager), nil

	case ProviderSES:
		service, err := NewSESEmailService(config.SES, config.FromAddress, templateManager)
		if err != nil {
			return nil, fmt.Errorf("failed to create SES email service: %w", err)
		}
		return service, nil

	case ProviderMock:
		return NewMockEmailService(templateManager), nil

	default:
		return nil, fmt.Errorf("unsupported email provider: %s", config.Provider)
	}
}

// NewMockEmailServiceForTesting creates a mock email service for testing
func NewMockEmailServiceForTesting() Service {
	templateManager := NewTemplateManager("")
	templateManager.CreateDefaultTemplates() // Ignore error for testing
	return NewMockEmailService(templateManager)
}

package email

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"path/filepath"
	"strings"
	textTemplate "text/template"
)

// TemplateManager handles email template rendering
type TemplateManager struct {
	htmlTemplates map[string]*template.Template
	textTemplates map[string]*textTemplate.Template
	templateDir   string
}

// NewTemplateManager creates a new template manager
func NewTemplateManager(templateDir string) *TemplateManager {
	return &TemplateManager{
		htmlTemplates: make(map[string]*template.Template),
		textTemplates: make(map[string]*textTemplate.Template),
		templateDir:   templateDir,
	}
}

// LoadTemplate loads a template from the template directory
func (tm *TemplateManager) LoadTemplate(name string) error {
	// Load HTML template
	htmlPath := filepath.Join(tm.templateDir, name+".html")
	htmlTemplate, err := template.ParseFiles(htmlPath)
	if err == nil {
		tm.htmlTemplates[name] = htmlTemplate
	}

	// Load text template
	textPath := filepath.Join(tm.templateDir, name+".txt")
	textTemplate, err := textTemplate.ParseFiles(textPath)
	if err == nil {
		tm.textTemplates[name] = textTemplate
	}

	// At least one template must exist
	if tm.htmlTemplates[name] == nil && tm.textTemplates[name] == nil {
		return fmt.Errorf("no template files found for %s (checked %s and %s)", name, htmlPath, textPath)
	}

	return nil
}

// LoadAllTemplates loads all templates from the template directory
func (tm *TemplateManager) LoadAllTemplates() error {
	// Common template names
	templateNames := []string{
		"password_reset",
		"welcome",
		"account_verification",
		"invitation",
		"account_created",
	}

	for _, name := range templateNames {
		if err := tm.LoadTemplate(name); err != nil {
			// Log warning but don't fail - templates are optional
			continue
		}
	}

	return nil
}

// RenderTemplate renders a template with the given data
func (tm *TemplateManager) RenderTemplate(name string, data interface{}) (subject, textBody, htmlBody string, err error) {
	// Extract subject from template data if available
	subject = tm.extractSubject(name, data)

	// Render text template
	if textTmpl, exists := tm.textTemplates[name]; exists {
		var textBuf bytes.Buffer
		if err := textTmpl.Execute(&textBuf, data); err != nil {
			return "", "", "", fmt.Errorf("failed to render text template %s: %w", name, err)
		}
		textBody = textBuf.String()
	}

	// Render HTML template
	if htmlTmpl, exists := tm.htmlTemplates[name]; exists {
		var htmlBuf bytes.Buffer
		if err := htmlTmpl.Execute(&htmlBuf, data); err != nil {
			return "", "", "", fmt.Errorf("failed to render HTML template %s: %w", name, err)
		}
		htmlBody = htmlBuf.String()
	}

	// If no templates were found, return error
	if textBody == "" && htmlBody == "" {
		return "", "", "", fmt.Errorf("template %s not found", name)
	}

	return subject, textBody, htmlBody, nil
}

// RenderTemplateDataAsJSON converts template data to JSON for SES bulk operations
func (tm *TemplateManager) RenderTemplateDataAsJSON(data interface{}) (string, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("failed to marshal template data to JSON: %w", err)
	}
	return string(jsonData), nil
}

// extractSubject extracts the subject from template data or generates a default
func (tm *TemplateManager) extractSubject(templateName string, data interface{}) string {
	// Try to extract subject from data if it has a Subject field
	if dataMap, ok := data.(map[string]interface{}); ok {
		if subject, exists := dataMap["Subject"]; exists {
			if subjectStr, ok := subject.(string); ok {
				return subjectStr
			}
		}
	}

	// Generate default subject based on template name
	switch templateName {
	case "password_reset":
		return "Password Reset Request"
	case "welcome":
		return "Welcome to Solid Pod"
	case "account_verification":
		return "Verify Your Account"
	case "invitation":
		return "You've been invited to join an account"
	case "account_created":
		return "Your account has been created"
	default:
		return strings.Title(strings.ReplaceAll(templateName, "_", " "))
	}
}

// CreateDefaultTemplates creates default email templates in memory
func (tm *TemplateManager) CreateDefaultTemplates() error {
	// Default password reset template
	passwordResetHTML := `
<!DOCTYPE html>
<html>
<head>
    <title>Password Reset</title>
</head>
<body>
    <h2>Password Reset Request</h2>
    <p>Hello {{.UserName}},</p>
    <p>You have requested to reset your password. Click the link below to reset it:</p>
    <p><a href="{{.ResetURL}}">Reset Password</a></p>
    <p>This link will expire at {{.ExpiryTime.Format "2006-01-02 15:04:05 UTC"}}.</p>
    <p>If you did not request this reset, please ignore this email.</p>
    <p>Need help? Contact us at <a href="{{.SupportURL}}">{{.SupportURL}}</a></p>
</body>
</html>`

	passwordResetText := `Password Reset Request

Hello {{.UserName}},

You have requested to reset your password. Visit the following link to reset it:

{{.ResetURL}}

This link will expire at {{.ExpiryTime.Format "2006-01-02 15:04:05 UTC"}}.

If you did not request this reset, please ignore this email.

Need help? Visit {{.SupportURL}}`

	// Parse and store templates
	htmlTmpl, err := template.New("password_reset").Parse(passwordResetHTML)
	if err != nil {
		return fmt.Errorf("failed to parse password reset HTML template: %w", err)
	}
	tm.htmlTemplates["password_reset"] = htmlTmpl

	textTmpl, err := textTemplate.New("password_reset").Parse(passwordResetText)
	if err != nil {
		return fmt.Errorf("failed to parse password reset text template: %w", err)
	}
	tm.textTemplates["password_reset"] = textTmpl

	// Default welcome template
	welcomeHTML := `
<!DOCTYPE html>
<html>
<head>
    <title>Welcome</title>
</head>
<body>
    <h2>Welcome to Solid Pod!</h2>
    <p>Hello {{.UserName}},</p>
    <p>Welcome to your new Solid Pod! Your account has been successfully created.</p>
    <p>You can now start using your pod to store and manage your data securely.</p>
    <p>Need help getting started? Visit our support at <a href="{{.SupportURL}}">{{.SupportURL}}</a></p>
</body>
</html>`

	welcomeText := `Welcome to Solid Pod!

Hello {{.UserName}},

Welcome to your new Solid Pod! Your account has been successfully created.

You can now start using your pod to store and manage your data securely.

Need help getting started? Visit our support at {{.SupportURL}}`

	htmlTmpl, err = template.New("welcome").Parse(welcomeHTML)
	if err != nil {
		return fmt.Errorf("failed to parse welcome HTML template: %w", err)
	}
	tm.htmlTemplates["welcome"] = htmlTmpl

	textTmpl, err = textTemplate.New("welcome").Parse(welcomeText)
	if err != nil {
		return fmt.Errorf("failed to parse welcome text template: %w", err)
	}
	tm.textTemplates["welcome"] = textTmpl

	return nil
}

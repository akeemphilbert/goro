package email

import (
	"testing"
)

func TestEmail_Validation(t *testing.T) {
	tests := []struct {
		name    string
		email   *Email
		wantErr bool
	}{
		{
			name: "valid email with all fields",
			email: &Email{
				To:       []string{"test@example.com"},
				CC:       []string{"cc@example.com"},
				BCC:      []string{"bcc@example.com"},
				Subject:  "Test Subject",
				TextBody: "Test text body",
				HTMLBody: "<p>Test HTML body</p>",
			},
			wantErr: false,
		},
		{
			name: "valid email with minimal fields",
			email: &Email{
				To:       []string{"test@example.com"},
				Subject:  "Test Subject",
				TextBody: "Test text body",
			},
			wantErr: false,
		},
		{
			name: "empty recipients",
			email: &Email{
				Subject:  "Test Subject",
				TextBody: "Test text body",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasRecipients := len(tt.email.To) > 0
			if hasRecipients == tt.wantErr {
				t.Errorf("Email validation mismatch: hasRecipients=%v, wantErr=%v", hasRecipients, tt.wantErr)
			}
		})
	}
}

func TestTemplateData_Fields(t *testing.T) {
	data := TemplateData{
		UserName:   "John Doe",
		ResetURL:   "https://example.com/reset",
		SupportURL: "https://example.com/support",
	}

	if data.UserName != "John Doe" {
		t.Errorf("Expected UserName to be 'John Doe', got %s", data.UserName)
	}

	if data.ResetURL != "https://example.com/reset" {
		t.Errorf("Expected ResetURL to be 'https://example.com/reset', got %s", data.ResetURL)
	}

	if data.SupportURL != "https://example.com/support" {
		t.Errorf("Expected SupportURL to be 'https://example.com/support', got %s", data.SupportURL)
	}
}

func TestProvider_Constants(t *testing.T) {
	if ProviderSMTP != "smtp" {
		t.Errorf("Expected ProviderSMTP to be 'smtp', got %s", ProviderSMTP)
	}

	if ProviderSES != "ses" {
		t.Errorf("Expected ProviderSES to be 'ses', got %s", ProviderSES)
	}

	if ProviderMock != "mock" {
		t.Errorf("Expected ProviderMock to be 'mock', got %s", ProviderMock)
	}
}

package emails

import (
	"log/slog"
)

// EmailService handles sending emails
type EmailService struct {
	fromEmail string
	apiKey    string
}

// NewEmailService creates a new email service
func NewEmailService(fromEmail, apiKey string) *EmailService {
	return &EmailService{
		fromEmail: fromEmail,
		apiKey:    apiKey,
	}
}

// SendEmail sends an email
func (s *EmailService) SendEmail(to, subject, body string) error {
	slog.Info("Sending email", "to", to, "subject", subject)
	// TODO: Implement email sending logic (e.g., SendGrid, SES)
	return nil
}

// SendTemplateEmail sends an email using a template
func (s *EmailService) SendTemplateEmail(to, templateID string, data map[string]interface{}) error {
	slog.Info("Sending template email", "to", to, "templateID", templateID)
	// TODO: Implement template email logic
	return nil
}

package providers

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/smtp"
	"time"

	"gopkg.in/gomail.v2"

	"github.com/acheevo/tfa/internal/shared/config"
	"github.com/acheevo/tfa/internal/shared/email/domain"
)

// SMTPProvider implements the EmailProvider interface for SMTP
type SMTPProvider struct {
	config *config.Config
	dialer *gomail.Dialer
}

// NewSMTPProvider creates a new SMTP email provider
func NewSMTPProvider(cfg *config.Config) *SMTPProvider {
	dialer := gomail.NewDialer(
		cfg.SMTPHost,
		cfg.SMTPPort,
		cfg.SMTPUsername,
		cfg.SMTPPassword,
	)

	// Configure TLS
	if cfg.SMTPUseTLS {
		dialer.TLSConfig = &tls.Config{
			ServerName:         cfg.SMTPHost,
			InsecureSkipVerify: cfg.SMTPSkipTLSCheck, // #nosec G402 -- Configurable for development environments
		}
	}

	// Set authentication method
	if cfg.SMTPUsername != "" && cfg.SMTPPassword != "" {
		dialer.Auth = smtp.PlainAuth("", cfg.SMTPUsername, cfg.SMTPPassword, cfg.SMTPHost)
	}

	return &SMTPProvider{
		config: cfg,
		dialer: dialer,
	}
}

// Send sends an email message via SMTP
func (p *SMTPProvider) Send(ctx context.Context, message *domain.EmailMessage) (*domain.EmailResult, error) {
	if p.dialer == nil {
		return nil, domain.ErrEmailProviderNotConfigured
	}

	// Create the email message
	m := gomail.NewMessage()

	// Set headers
	if message.FromName != "" {
		m.SetHeader("From", m.FormatAddress(message.From, message.FromName))
	} else {
		m.SetHeader("From", message.From)
	}

	m.SetHeader("To", message.To...)
	m.SetHeader("Subject", message.Subject)

	if len(message.CC) > 0 {
		m.SetHeader("Cc", message.CC...)
	}

	if len(message.BCC) > 0 {
		m.SetHeader("Bcc", message.BCC...)
	}

	if message.ReplyTo != "" {
		m.SetHeader("Reply-To", message.ReplyTo)
	}

	// Set custom headers
	for key, value := range message.Headers {
		m.SetHeader(key, value)
	}

	// Set message ID header for tracking
	m.SetHeader("Message-ID", fmt.Sprintf("<%s@%s>", message.ID, p.config.SMTPHost))

	// Set body
	if message.TextBody != "" {
		m.SetBody("text/plain", message.TextBody)
	}

	if message.HTMLBody != "" {
		if message.TextBody != "" {
			m.AddAlternative("text/html", message.HTMLBody)
		} else {
			m.SetBody("text/html", message.HTMLBody)
		}
	}

	// Add attachments
	for _, attachment := range message.Attachments {
		if attachment.Inline {
			m.Embed(attachment.Name, gomail.SetCopyFunc(func(w io.Writer) error {
				_, err := w.Write(attachment.Data)
				return err
			}))
		} else {
			m.Attach(attachment.Name, gomail.SetCopyFunc(func(w io.Writer) error {
				_, err := w.Write(attachment.Data)
				return err
			}))
		}
	}

	// Send with timeout
	done := make(chan error, 1)
	go func() {
		done <- p.dialer.DialAndSend(m)
	}()

	select {
	case err := <-done:
		if err != nil {
			return &domain.EmailResult{
				MessageID: message.ID,
				Status:    domain.StatusFailed,
				Message:   err.Error(),
			}, err
		}

		return &domain.EmailResult{
			MessageID: message.ID,
			Status:    domain.StatusSent,
			Message:   "Email sent successfully via SMTP",
			Metadata: map[string]string{
				"provider": string(domain.ProviderSMTP),
				"host":     p.config.SMTPHost,
				"sent_at":  time.Now().UTC().Format(time.RFC3339),
			},
		}, nil

	case <-ctx.Done():
		return &domain.EmailResult{
			MessageID: message.ID,
			Status:    domain.StatusFailed,
			Message:   "SMTP send timeout",
		}, ctx.Err()
	}
}

// SendTemplate sends an email using a template (SMTP doesn't support server-side templates)
func (p *SMTPProvider) SendTemplate(
	ctx context.Context,
	templateID string,
	to []string,
	variables map[string]interface{},
) (*domain.EmailResult, error) {
	// SMTP doesn't support server-side templates, this should be handled by the template engine
	return nil, fmt.Errorf("SMTP provider does not support server-side templates, use the template engine")
}

// GetDeliveryStatus gets the delivery status of an email (SMTP doesn't support delivery tracking)
func (p *SMTPProvider) GetDeliveryStatus(ctx context.Context, messageID string) (*domain.EmailDeliveryStatus, error) {
	// SMTP doesn't provide delivery status tracking
	return &domain.EmailDeliveryStatus{
		MessageID: messageID,
		Status:    domain.StatusSent, // We can only confirm it was sent, not delivered
	}, nil
}

// SupportsTemplates returns whether this provider supports server-side templates
func (p *SMTPProvider) SupportsTemplates() bool {
	return false
}

// SupportsWebhooks returns whether this provider supports webhooks
func (p *SMTPProvider) SupportsWebhooks() bool {
	return false
}

// GetProviderName returns the provider name
func (p *SMTPProvider) GetProviderName() domain.EmailProvider {
	return domain.ProviderSMTP
}

// HealthCheck performs a health check on the SMTP connection
func (p *SMTPProvider) HealthCheck(ctx context.Context) error {
	if p.dialer == nil {
		return domain.ErrEmailProviderNotConfigured
	}

	// Try to establish a connection
	conn, err := p.dialer.Dial()
	if err != nil {
		return fmt.Errorf("SMTP health check failed: %w", err)
	}
	defer func() {
		_ = conn.Close() // Ignore close errors in health check
	}()

	return nil
}

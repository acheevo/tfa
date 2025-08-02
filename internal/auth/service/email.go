package service

import (
	"bytes"
	"fmt"
	"html/template"
	"log/slog"

	"gopkg.in/gomail.v2"

	"github.com/acheevo/tfa/internal/shared/config"
)

// EmailService handles email sending operations
type EmailService struct {
	config *config.Config
	logger *slog.Logger
	dialer *gomail.Dialer
}

// NewEmailService creates a new email service
func NewEmailService(config *config.Config, logger *slog.Logger) *EmailService {
	var dialer *gomail.Dialer
	if config.SMTPUsername != "" && config.SMTPPassword != "" {
		dialer = gomail.NewDialer(config.SMTPHost, config.SMTPPort, config.SMTPUsername, config.SMTPPassword)
	}

	return &EmailService{
		config: config,
		logger: logger,
		dialer: dialer,
	}
}

// SendEmailVerification sends an email verification email
func (e *EmailService) SendEmailVerification(email, token, firstName string) error {
	if e.dialer == nil {
		e.logger.Warn("email service not configured, skipping email verification", "email", email)
		return nil
	}

	verificationURL := fmt.Sprintf("%s/verify-email?token=%s", e.config.FrontendURL, token)

	subject := "Verify your email address"
	htmlBody, err := e.renderEmailVerificationTemplate(firstName, verificationURL)
	if err != nil {
		return fmt.Errorf("failed to render email template: %w", err)
	}

	textBody := fmt.Sprintf(`Hi %s,

Please verify your email address by clicking the link below:
%s

If you didn't create an account, you can safely ignore this email.

Best regards,
%s Team`, firstName, verificationURL, e.config.EmailFromName)

	return e.sendEmail(email, subject, htmlBody, textBody)
}

// SendPasswordReset sends a password reset email
func (e *EmailService) SendPasswordReset(email, token, firstName string) error {
	if e.dialer == nil {
		e.logger.Warn("email service not configured, skipping password reset email", "email", email)
		return nil
	}

	resetURL := fmt.Sprintf("%s/reset-password?token=%s", e.config.FrontendURL, token)

	subject := "Reset your password"
	htmlBody, err := e.renderPasswordResetTemplate(firstName, resetURL)
	if err != nil {
		return fmt.Errorf("failed to render email template: %w", err)
	}

	textBody := fmt.Sprintf(`Hi %s,

You requested to reset your password. Click the link below to reset it:
%s

This link will expire in 24 hours. If you didn't request this, you can safely ignore this email.

Best regards,
%s Team`, firstName, resetURL, e.config.EmailFromName)

	return e.sendEmail(email, subject, htmlBody, textBody)
}

// SendWelcomeEmail sends a welcome email to new users
func (e *EmailService) SendWelcomeEmail(email, firstName string) error {
	if e.dialer == nil {
		e.logger.Warn("email service not configured, skipping welcome email", "email", email)
		return nil
	}

	subject := fmt.Sprintf("Welcome to %s!", e.config.EmailFromName)
	htmlBody, err := e.renderWelcomeTemplate(firstName)
	if err != nil {
		return fmt.Errorf("failed to render email template: %w", err)
	}

	textBody := fmt.Sprintf(`Hi %s,

Welcome to %s! Your account has been successfully created and verified.

You can now access all the features of our platform.

Best regards,
%s Team`, firstName, e.config.EmailFromName, e.config.EmailFromName)

	return e.sendEmail(email, subject, htmlBody, textBody)
}

// sendEmail sends an email with both HTML and text content
func (e *EmailService) sendEmail(to, subject, htmlBody, textBody string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", m.FormatAddress(e.config.EmailFrom, e.config.EmailFromName))
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/plain", textBody)
	m.AddAlternative("text/html", htmlBody)

	if err := e.dialer.DialAndSend(m); err != nil {
		e.logger.Error("failed to send email", "to", to, "subject", subject, "error", err)
		return fmt.Errorf("failed to send email: %w", err)
	}

	e.logger.Info("email sent successfully", "to", to, "subject", subject)
	return nil
}

// renderEmailVerificationTemplate renders the email verification template
func (e *EmailService) renderEmailVerificationTemplate(firstName, verificationURL string) (string, error) {
	tmpl := `<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Verify your email</title>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { text-align: center; margin-bottom: 30px; }
        .button { display: inline-block; padding: 12px 24px; background-color: #007bff; color: white; 
                  text-decoration: none; border-radius: 4px; margin: 20px 0; }
        .footer { margin-top: 30px; font-size: 12px; color: #666; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Verify your email address</h1>
        </div>
        <p>Hi {{.FirstName}},</p>
        <p>Thank you for creating an account! Please verify your email address by clicking the button below:</p>
        <p style="text-align: center;">
            <a href="{{.VerificationURL}}" class="button">Verify Email Address</a>
        </p>
        <p>If the button doesn't work, you can copy and paste this link into your browser:</p>
        <p><a href="{{.VerificationURL}}">{{.VerificationURL}}</a></p>
        <p>If you didn't create an account, you can safely ignore this email.</p>
        <div class="footer">
            <p>Best regards,<br>{{.AppName}} Team</p>
        </div>
    </div>
</body>
</html>`

	t, err := template.New("email_verification").Parse(tmpl)
	if err != nil {
		return "", err
	}

	data := struct {
		FirstName       string
		VerificationURL string
		AppName         string
	}{
		FirstName:       firstName,
		VerificationURL: verificationURL,
		AppName:         e.config.EmailFromName,
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// renderPasswordResetTemplate renders the password reset template
func (e *EmailService) renderPasswordResetTemplate(firstName, resetURL string) (string, error) {
	tmpl := `<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Reset your password</title>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { text-align: center; margin-bottom: 30px; }
        .button { display: inline-block; padding: 12px 24px; background-color: #dc3545; color: white; 
                  text-decoration: none; border-radius: 4px; margin: 20px 0; }
        .footer { margin-top: 30px; font-size: 12px; color: #666; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Reset your password</h1>
        </div>
        <p>Hi {{.FirstName}},</p>
        <p>You requested to reset your password. Click the button below to reset it:</p>
        <p style="text-align: center;">
            <a href="{{.ResetURL}}" class="button">Reset Password</a>
        </p>
        <p>If the button doesn't work, you can copy and paste this link into your browser:</p>
        <p><a href="{{.ResetURL}}">{{.ResetURL}}</a></p>
        <p><strong>This link will expire in 24 hours.</strong></p>
        <p>If you didn't request this password reset, you can safely ignore this email.</p>
        <div class="footer">
            <p>Best regards,<br>{{.AppName}} Team</p>
        </div>
    </div>
</body>
</html>`

	t, err := template.New("password_reset").Parse(tmpl)
	if err != nil {
		return "", err
	}

	data := struct {
		FirstName string
		ResetURL  string
		AppName   string
	}{
		FirstName: firstName,
		ResetURL:  resetURL,
		AppName:   e.config.EmailFromName,
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// renderWelcomeTemplate renders the welcome email template
func (e *EmailService) renderWelcomeTemplate(firstName string) (string, error) {
	tmpl := `<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Welcome!</title>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { text-align: center; margin-bottom: 30px; }
        .footer { margin-top: 30px; font-size: 12px; color: #666; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Welcome to {{.AppName}}!</h1>
        </div>
        <p>Hi {{.FirstName}},</p>
        <p>Welcome to {{.AppName}}! Your account has been successfully created and verified.</p>
        <p>You can now access all the features of our platform. If you have any questions, 
        feel free to reach out to our support team.</p>
        <p>Thank you for joining us!</p>
        <div class="footer">
            <p>Best regards,<br>{{.AppName}} Team</p>
        </div>
    </div>
</body>
</html>`

	t, err := template.New("welcome").Parse(tmpl)
	if err != nil {
		return "", err
	}

	data := struct {
		FirstName string
		AppName   string
	}{
		FirstName: firstName,
		AppName:   e.config.EmailFromName,
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

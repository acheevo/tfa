package templates

import (
	"bytes"
	"fmt"
	"html/template"
	"log/slog"
	"strings"
	"sync"
	textTemplate "text/template"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/acheevo/tfa/internal/shared/email/domain"
)

// DefaultTemplateEngine implements EmailTemplateEngine
type DefaultTemplateEngine struct {
	templates map[string]*domain.EmailTemplate
	mutex     sync.RWMutex
	logger    *slog.Logger
}

// NewDefaultTemplateEngine creates a new template engine
func NewDefaultTemplateEngine(logger *slog.Logger) *DefaultTemplateEngine {
	engine := &DefaultTemplateEngine{
		templates: make(map[string]*domain.EmailTemplate),
		logger:    logger,
	}

	// Register default templates
	if err := engine.registerDefaultTemplates(); err != nil {
		logger.Error("failed to register default templates", "error", err)
	}

	return engine
}

// Render renders a template with the given variables
func (e *DefaultTemplateEngine) Render(
	templateID string,
	variables map[string]interface{},
) (*domain.RenderedTemplate, error) {
	e.mutex.RLock()
	tmpl, exists := e.templates[templateID]
	e.mutex.RUnlock()

	if !exists {
		return nil, domain.ErrTemplateNotFound
	}

	// Validate required variables
	if err := e.validateVariables(tmpl, variables); err != nil {
		return nil, err
	}

	// Render subject
	subject, err := e.renderText(tmpl.Subject, variables)
	if err != nil {
		return nil, fmt.Errorf("failed to render subject: %w", err)
	}

	// Render HTML body
	htmlBody, err := e.renderHTML(tmpl.HTMLBody, variables)
	if err != nil {
		return nil, fmt.Errorf("failed to render HTML body: %w", err)
	}

	// Render text body
	textBody, err := e.renderText(tmpl.TextBody, variables)
	if err != nil {
		return nil, fmt.Errorf("failed to render text body: %w", err)
	}

	return &domain.RenderedTemplate{
		Subject:  subject,
		HTMLBody: htmlBody,
		TextBody: textBody,
	}, nil
}

// RegisterTemplate registers a new template
func (e *DefaultTemplateEngine) RegisterTemplate(tmpl *domain.EmailTemplate) error {
	if err := e.ValidateTemplate(tmpl); err != nil {
		return err
	}

	e.mutex.Lock()
	e.templates[tmpl.ID] = tmpl
	e.mutex.Unlock()

	e.logger.Info("template registered", "template_id", tmpl.ID, "name", tmpl.Name)
	return nil
}

// GetTemplate retrieves a template by ID
func (e *DefaultTemplateEngine) GetTemplate(templateID string) (*domain.EmailTemplate, error) {
	e.mutex.RLock()
	tmpl, exists := e.templates[templateID]
	e.mutex.RUnlock()

	if !exists {
		return nil, domain.ErrTemplateNotFound
	}

	return tmpl, nil
}

// ListTemplates returns all registered templates
func (e *DefaultTemplateEngine) ListTemplates() ([]*domain.EmailTemplate, error) {
	e.mutex.RLock()
	defer e.mutex.RUnlock()

	templates := make([]*domain.EmailTemplate, 0, len(e.templates))
	for _, tmpl := range e.templates {
		templates = append(templates, tmpl)
	}

	return templates, nil
}

// ValidateTemplate validates a template
func (e *DefaultTemplateEngine) ValidateTemplate(tmpl *domain.EmailTemplate) error {
	if tmpl.ID == "" {
		return fmt.Errorf("%w: template ID is required", domain.ErrTemplateInvalid)
	}

	if tmpl.Name == "" {
		return fmt.Errorf("%w: template name is required", domain.ErrTemplateInvalid)
	}

	if tmpl.Subject == "" {
		return fmt.Errorf("%w: template subject is required", domain.ErrTemplateInvalid)
	}

	if tmpl.HTMLBody == "" && tmpl.TextBody == "" {
		return fmt.Errorf("%w: template must have either HTML or text body", domain.ErrTemplateInvalid)
	}

	// Validate template syntax
	if tmpl.HTMLBody != "" {
		_, err := template.New("test").Parse(tmpl.HTMLBody)
		if err != nil {
			return fmt.Errorf("%w: HTML template syntax error: %v", domain.ErrTemplateInvalid, err)
		}
	}

	if tmpl.TextBody != "" {
		_, err := textTemplate.New("test").Parse(tmpl.TextBody)
		if err != nil {
			return fmt.Errorf("%w: text template syntax error: %v", domain.ErrTemplateInvalid, err)
		}
	}

	if tmpl.Subject != "" {
		_, err := textTemplate.New("test").Parse(tmpl.Subject)
		if err != nil {
			return fmt.Errorf("%w: subject template syntax error: %v", domain.ErrTemplateInvalid, err)
		}
	}

	return nil
}

// renderHTML renders an HTML template
func (e *DefaultTemplateEngine) renderHTML(tmplText string, variables map[string]interface{}) (string, error) {
	tmpl, err := template.New("html").Funcs(e.getTemplateFunctions()).Parse(tmplText)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, variables); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// renderText renders a text template
func (e *DefaultTemplateEngine) renderText(tmplText string, variables map[string]interface{}) (string, error) {
	tmpl, err := textTemplate.New("text").Funcs(e.getTextTemplateFunctions()).Parse(tmplText)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, variables); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// validateVariables validates that all required variables are provided
func (e *DefaultTemplateEngine) validateVariables(tmpl *domain.EmailTemplate, variables map[string]interface{}) error {
	missingVars := []string{}

	for _, requiredVar := range tmpl.Variables {
		if _, exists := variables[requiredVar]; !exists {
			missingVars = append(missingVars, requiredVar)
		}
	}

	if len(missingVars) > 0 {
		return fmt.Errorf("%w: %s", domain.ErrTemplateMissingVariables, strings.Join(missingVars, ", "))
	}

	return nil
}

// getTemplateFunctions returns HTML template functions
func (e *DefaultTemplateEngine) getTemplateFunctions() template.FuncMap {
	caser := cases.Title(language.English)
	return template.FuncMap{
		"upper": strings.ToUpper,
		"lower": strings.ToLower,
		"title": caser.String,
		"trim":  strings.TrimSpace,
		"default": func(defaultValue, value interface{}) interface{} {
			if value == nil || value == "" {
				return defaultValue
			}
			return value
		},
	}
}

// getTextTemplateFunctions returns text template functions
func (e *DefaultTemplateEngine) getTextTemplateFunctions() textTemplate.FuncMap {
	caser := cases.Title(language.English)
	return textTemplate.FuncMap{
		"upper": strings.ToUpper,
		"lower": strings.ToLower,
		"title": caser.String,
		"trim":  strings.TrimSpace,
		"default": func(defaultValue, value interface{}) interface{} {
			if value == nil || value == "" {
				return defaultValue
			}
			return value
		},
	}
}

// registerDefaultTemplates registers built-in templates
func (e *DefaultTemplateEngine) registerDefaultTemplates() error {
	// Email verification template
	if err := e.RegisterTemplate(&domain.EmailTemplate{
		ID:        "email_verification",
		Name:      "Email Verification",
		Subject:   "Verify your email address",
		Variables: []string{"user_name", "verification_url", "app_name"},
		HTMLBody: `<!DOCTYPE html>
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
        <p>Hi {{.user_name | default "there"}},</p>
        <p>Thank you for creating an account! Please verify your email address by clicking the button below:</p>
        <p style="text-align: center;">
            <a href="{{.verification_url}}" class="button">Verify Email Address</a>
        </p>
        <p>If the button doesn't work, you can copy and paste this link into your browser:</p>
        <p><a href="{{.verification_url}}">{{.verification_url}}</a></p>
        <p>If you didn't create an account, you can safely ignore this email.</p>
        <div class="footer">
            <p>Best regards,<br>{{.app_name}} Team</p>
        </div>
    </div>
</body>
</html>`,
		TextBody: `Hi {{.user_name | default "there"}},

Thank you for creating an account! Please verify your email address by clicking the link below:

{{.verification_url}}

If you didn't create an account, you can safely ignore this email.

Best regards,
{{.app_name}} Team`,
	}); err != nil {
		return fmt.Errorf("failed to register email verification template: %w", err)
	}

	// Password reset template
	if err := e.RegisterTemplate(&domain.EmailTemplate{
		ID:        "password_reset",
		Name:      "Password Reset",
		Subject:   "Reset your password",
		Variables: []string{"user_name", "reset_url", "app_name"},
		HTMLBody: `<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Reset your password</title>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { text-align: center; margin-bottom: 30px; }
        .button { 
            display: inline-block; padding: 12px 24px; background-color: #dc3545; 
            color: white; text-decoration: none; border-radius: 4px; margin: 20px 0; 
        }
        .footer { margin-top: 30px; font-size: 12px; color: #666; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Reset your password</h1>
        </div>
        <p>Hi {{.user_name | default "there"}},</p>
        <p>You requested to reset your password. Click the button below to reset it:</p>
        <p style="text-align: center;">
            <a href="{{.reset_url}}" class="button">Reset Password</a>
        </p>
        <p>If the button doesn't work, you can copy and paste this link into your browser:</p>
        <p><a href="{{.reset_url}}">{{.reset_url}}</a></p>
        <p><strong>This link will expire in 24 hours.</strong></p>
        <p>If you didn't request this password reset, you can safely ignore this email.</p>
        <div class="footer">
            <p>Best regards,<br>{{.app_name}} Team</p>
        </div>
    </div>
</body>
</html>`,
		TextBody: `Hi {{.user_name | default "there"}},

You requested to reset your password. Click the link below to reset it:

{{.reset_url}}

This link will expire in 24 hours.

If you didn't request this password reset, you can safely ignore this email.

Best regards,
{{.app_name}} Team`,
	}); err != nil {
		return fmt.Errorf("failed to register password reset template: %w", err)
	}

	// Welcome email template
	if err := e.RegisterTemplate(&domain.EmailTemplate{
		ID:        "welcome",
		Name:      "Welcome Email",
		Subject:   "Welcome to {{.app_name}}!",
		Variables: []string{"user_name", "app_name"},
		HTMLBody: `<!DOCTYPE html>
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
            <h1>Welcome to {{.app_name}}!</h1>
        </div>
        <p>Hi {{.user_name | default "there"}},</p>
        <p>Welcome to {{.app_name}}! Your account has been successfully created and verified.</p>
        <p>You can now access all the features of our platform. If you have any questions, 
        feel free to reach out to our support team.</p>
        <p>Thank you for joining us!</p>
        <div class="footer">
            <p>Best regards,<br>{{.app_name}} Team</p>
        </div>
    </div>
</body>
</html>`,
		TextBody: `Hi {{.user_name | default "there"}},

Welcome to {{.app_name}}! Your account has been successfully created and verified.

You can now access all the features of our platform. If you have any questions, 
feel free to reach out to our support team.

Thank you for joining us!

Best regards,
{{.app_name}} Team`,
	}); err != nil {
		return fmt.Errorf("failed to register welcome template: %w", err)
	}

	return nil
}

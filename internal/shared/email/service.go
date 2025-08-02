package email

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/acheevo/tfa/internal/shared/config"
	"github.com/acheevo/tfa/internal/shared/email/domain"
	"github.com/acheevo/tfa/internal/shared/email/providers"
	"github.com/acheevo/tfa/internal/shared/email/queue"
	"github.com/acheevo/tfa/internal/shared/email/templates"
)

// Service is the main email service implementation
type Service struct {
	config         *config.Config
	logger         *slog.Logger
	provider       domain.EmailProviderInterface
	queue          domain.EmailQueueInterface
	templateEngine domain.EmailTemplateEngine
}

// NewService creates a new email service
func NewService(
	cfg *config.Config,
	logger *slog.Logger,
	db interface{}, // Can be *gorm.DB or other database interface
	templateEngine domain.EmailTemplateEngine,
) (*Service, error) {
	// Create email provider based on configuration
	provider, err := createProvider(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create email provider: %w", err)
	}

	// Create queue (assuming database queue for now)
	var emailQueue domain.EmailQueueInterface
	if gormDB, ok := db.(interface{ DB() interface{} }); ok {
		// Extract gorm.DB from the wrapper
		if actualDB, ok := gormDB.DB().(*gorm.DB); ok {
			emailQueue = queue.NewDatabaseQueue(actualDB, logger)
		}
	}

	if emailQueue == nil {
		return nil, fmt.Errorf("failed to create email queue: unsupported database type")
	}

	// Use provided template engine or create default one
	if templateEngine == nil {
		templateEngine = templates.NewDefaultTemplateEngine(logger)
	}

	service := &Service{
		config:         cfg,
		logger:         logger,
		provider:       provider,
		queue:          emailQueue,
		templateEngine: templateEngine,
	}

	return service, nil
}

// Send queues an email for asynchronous sending
func (s *Service) Send(ctx context.Context, message *domain.EmailMessage) error {
	// Set default values
	if message.ID == "" {
		message.ID = uuid.New().String()
	}

	if message.From == "" {
		message.From = s.config.EmailFrom
	}

	if message.FromName == "" {
		message.FromName = s.config.EmailFromName
	}

	if message.Priority == 0 {
		message.Priority = domain.PriorityNormal
	}

	message.CreatedAt = time.Now()

	// Validate email
	if err := s.validateMessage(message); err != nil {
		return fmt.Errorf("message validation failed: %w", err)
	}

	// Enqueue the message
	if err := s.queue.Enqueue(ctx, message); err != nil {
		s.logger.Error("failed to enqueue email", "error", err, "message_id", message.ID)
		return fmt.Errorf("failed to enqueue email: %w", err)
	}

	s.logger.Info("email queued successfully",
		"message_id", message.ID,
		"to", message.To,
		"subject", message.Subject,
	)

	return nil
}

// SendTemplate sends an email using a template
func (s *Service) SendTemplate(
	ctx context.Context,
	templateID string,
	to []string,
	variables map[string]interface{},
) error {
	// Render the template
	rendered, err := s.templateEngine.Render(templateID, variables)
	if err != nil {
		return fmt.Errorf("failed to render template: %w", err)
	}

	// Create email message
	message := &domain.EmailMessage{
		ID:         uuid.New().String(),
		To:         to,
		Subject:    rendered.Subject,
		HTMLBody:   rendered.HTMLBody,
		TextBody:   rendered.TextBody,
		TemplateID: templateID,
		Variables:  variables,
		Priority:   domain.PriorityNormal,
		Metadata: map[string]string{
			"template_id": templateID,
		},
	}

	return s.Send(ctx, message)
}

// SendImmediate sends an email immediately without queuing
func (s *Service) SendImmediate(ctx context.Context, message *domain.EmailMessage) (*domain.EmailResult, error) {
	// Set default values
	if message.ID == "" {
		message.ID = uuid.New().String()
	}

	if message.From == "" {
		message.From = s.config.EmailFrom
	}

	if message.FromName == "" {
		message.FromName = s.config.EmailFromName
	}

	// Validate email
	if err := s.validateMessage(message); err != nil {
		return nil, fmt.Errorf("message validation failed: %w", err)
	}

	// Send immediately using provider
	result, err := s.provider.Send(ctx, message)
	if err != nil {
		s.logger.Error("failed to send email immediately", "error", err, "message_id", message.ID)
		return result, fmt.Errorf("failed to send email: %w", err)
	}

	s.logger.Info("email sent immediately",
		"message_id", message.ID,
		"to", message.To,
		"subject", message.Subject,
		"result", result.Status,
	)

	return result, nil
}

// Schedule schedules an email for future sending
func (s *Service) Schedule(ctx context.Context, message *domain.EmailMessage, scheduledAt time.Time) error {
	message.ScheduledAt = &scheduledAt
	return s.Send(ctx, message)
}

// RegisterTemplate registers a new email template
func (s *Service) RegisterTemplate(template *domain.EmailTemplate) error {
	return s.templateEngine.RegisterTemplate(template)
}

// GetTemplate retrieves a template by ID
func (s *Service) GetTemplate(templateID string) (*domain.EmailTemplate, error) {
	return s.templateEngine.GetTemplate(templateID)
}

// ProcessQueue processes emails in the queue
func (s *Service) ProcessQueue(ctx context.Context) error {
	batchSize := 10 // Process 10 emails at a time

	emails, err := s.queue.Dequeue(ctx, batchSize)
	if err != nil {
		return fmt.Errorf("failed to dequeue emails: %w", err)
	}

	if len(emails) == 0 {
		return nil // No emails to process
	}

	s.logger.Info("processing email queue", "batch_size", len(emails))

	for _, queuedEmail := range emails {
		// Convert queued email back to message
		message, err := s.queuedEmailToMessage(queuedEmail)
		if err != nil {
			s.logger.Error("failed to convert queued email to message",
				"error", err,
				"email_id", queuedEmail.ID,
			)
			if markErr := s.queue.MarkFailed(ctx, queuedEmail.ID, err); markErr != nil {
				s.logger.Error("failed to mark email as failed", "error", markErr, "email_id", queuedEmail.ID)
			}
			continue
		}

		// Send the email
		result, err := s.provider.Send(ctx, message)
		if err != nil {
			s.logger.Error("failed to send email from queue",
				"error", err,
				"email_id", queuedEmail.ID,
				"message_id", message.ID,
			)
			if markErr := s.queue.MarkFailed(ctx, queuedEmail.ID, err); markErr != nil {
				s.logger.Error("failed to mark email as failed", "error", markErr, "email_id", queuedEmail.ID)
			}
			continue
		}

		// Mark as sent
		if err := s.queue.MarkSent(ctx, queuedEmail.ID, result); err != nil {
			s.logger.Error("failed to mark email as sent",
				"error", err,
				"email_id", queuedEmail.ID,
			)
		}
	}

	return nil
}

// GetQueueStats returns queue statistics
func (s *Service) GetQueueStats(ctx context.Context) (*domain.QueueStats, error) {
	return s.queue.GetStats(ctx)
}

// GetDeliveryStatus gets the delivery status of an email
func (s *Service) GetDeliveryStatus(ctx context.Context, messageID string) (*domain.EmailDeliveryStatus, error) {
	return s.provider.GetDeliveryStatus(ctx, messageID)
}

// GetEmailStats returns email statistics (placeholder for now)
func (s *Service) GetEmailStats(ctx context.Context) (*domain.EmailStats, error) {
	// This would typically query a database for delivery events
	// For now, return empty stats
	return &domain.EmailStats{}, nil
}

// HealthCheck performs a health check on the email service
func (s *Service) HealthCheck(ctx context.Context) error {
	// Check provider health
	if healthChecker, ok := s.provider.(interface{ HealthCheck(context.Context) error }); ok {
		if err := healthChecker.HealthCheck(ctx); err != nil {
			return fmt.Errorf("provider health check failed: %w", err)
		}
	}

	// Check queue health (basic stats query)
	_, err := s.queue.GetStats(ctx)
	if err != nil {
		return fmt.Errorf("queue health check failed: %w", err)
	}

	return nil
}

// Convenience methods for common email types

// SendEmailVerification sends an email verification email
func (s *Service) SendEmailVerification(ctx context.Context, email, userName, verificationURL string) error {
	variables := map[string]interface{}{
		"user_name":        userName,
		"verification_url": verificationURL,
		"app_name":         s.config.AppName,
	}

	return s.SendTemplate(ctx, "email_verification", []string{email}, variables)
}

// SendPasswordReset sends a password reset email
func (s *Service) SendPasswordReset(ctx context.Context, email, userName, resetURL string) error {
	variables := map[string]interface{}{
		"user_name": userName,
		"reset_url": resetURL,
		"app_name":  s.config.AppName,
	}

	return s.SendTemplate(ctx, "password_reset", []string{email}, variables)
}

// SendWelcomeEmail sends a welcome email
func (s *Service) SendWelcomeEmail(ctx context.Context, email, userName string) error {
	variables := map[string]interface{}{
		"user_name": userName,
		"app_name":  s.config.AppName,
	}

	return s.SendTemplate(ctx, "welcome", []string{email}, variables)
}

// Helper methods

// validateMessage validates an email message
func (s *Service) validateMessage(message *domain.EmailMessage) error {
	if len(message.To) == 0 {
		return domain.ErrInvalidEmailAddress
	}

	if message.Subject == "" {
		return fmt.Errorf("subject is required")
	}

	if message.HTMLBody == "" && message.TextBody == "" {
		return fmt.Errorf("email body is required")
	}

	// Additional validations can be added here
	return nil
}

// queuedEmailToMessage converts a queued email back to a message
func (s *Service) queuedEmailToMessage(queuedEmail *domain.QueuedEmail) (*domain.EmailMessage, error) {
	// This conversion logic should be in the queue implementation
	// For now, we'll implement a basic conversion
	if dbQueue, ok := s.queue.(*queue.DatabaseQueue); ok {
		return dbQueue.QueuedEmailToMessage(queuedEmail)
	}

	return nil, fmt.Errorf("unsupported queue type for message conversion")
}

// createProvider creates an email provider based on configuration
func createProvider(cfg *config.Config) (domain.EmailProviderInterface, error) {
	switch cfg.EmailProvider {
	case "smtp":
		return providers.NewSMTPProvider(cfg), nil
	case "sendgrid":
		// TODO: Implement SendGrid provider
		return nil, fmt.Errorf("sendGrid provider not implemented yet")
	case "postmark":
		// TODO: Implement Postmark provider
		return nil, fmt.Errorf("postmark provider not implemented yet")
	case "mailgun":
		// TODO: Implement Mailgun provider
		return nil, fmt.Errorf("mailgun provider not implemented yet")
	default:
		return nil, fmt.Errorf("unsupported email provider: %s", cfg.EmailProvider)
	}
}

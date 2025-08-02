package domain

import (
	"context"
	"time"
)

// EmailProvider represents the different email service providers
type EmailProvider string

const (
	ProviderSMTP     EmailProvider = "smtp"
	ProviderSendGrid EmailProvider = "sendgrid"
	ProviderPostmark EmailProvider = "postmark"
	ProviderMailgun  EmailProvider = "mailgun"
)

// EmailPriority represents the priority of an email
type EmailPriority int

const (
	PriorityLow EmailPriority = iota
	PriorityNormal
	PriorityHigh
	PriorityCritical
)

// EmailStatus represents the status of an email
type EmailStatus string

const (
	StatusPending   EmailStatus = "pending"
	StatusSending   EmailStatus = "sending"
	StatusSent      EmailStatus = "sent"
	StatusFailed    EmailStatus = "failed"
	StatusRetrying  EmailStatus = "retrying"
	StatusCancelled EmailStatus = "canceled"
)

// EmailTemplate represents an email template
type EmailTemplate struct {
	ID        string            `json:"id"`
	Name      string            `json:"name"`
	Subject   string            `json:"subject"`
	HTMLBody  string            `json:"html_body"`
	TextBody  string            `json:"text_body"`
	Variables []string          `json:"variables"`
	Metadata  map[string]string `json:"metadata"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
}

// EmailMessage represents an email message
type EmailMessage struct {
	ID          string                 `json:"id"`
	From        string                 `json:"from"`
	FromName    string                 `json:"from_name"`
	To          []string               `json:"to"`
	CC          []string               `json:"cc,omitempty"`
	BCC         []string               `json:"bcc,omitempty"`
	ReplyTo     string                 `json:"reply_to,omitempty"`
	Subject     string                 `json:"subject"`
	HTMLBody    string                 `json:"html_body,omitempty"`
	TextBody    string                 `json:"text_body,omitempty"`
	TemplateID  string                 `json:"template_id,omitempty"`
	Variables   map[string]interface{} `json:"variables,omitempty"`
	Attachments []EmailAttachment      `json:"attachments,omitempty"`
	Headers     map[string]string      `json:"headers,omitempty"`
	Tags        []string               `json:"tags,omitempty"`
	Metadata    map[string]string      `json:"metadata,omitempty"`
	Priority    EmailPriority          `json:"priority"`
	ScheduledAt *time.Time             `json:"scheduled_at,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
}

// EmailAttachment represents an email attachment
type EmailAttachment struct {
	Name        string `json:"name"`
	ContentType string `json:"content_type"`
	Data        []byte `json:"data"`
	Inline      bool   `json:"inline"`
	ContentID   string `json:"content_id,omitempty"`
}

// QueuedEmail represents an email in the queue
type QueuedEmail struct {
	ID           string        `json:"id" gorm:"primarykey"`
	MessageID    string        `json:"message_id" gorm:"uniqueIndex;not null"`
	From         string        `json:"from" gorm:"not null"`
	FromName     string        `json:"from_name"`
	To           string        `json:"to" gorm:"not null"` // JSON array as string
	CC           string        `json:"cc"`                 // JSON array as string
	BCC          string        `json:"bcc"`                // JSON array as string
	ReplyTo      string        `json:"reply_to"`
	Subject      string        `json:"subject" gorm:"not null"`
	HTMLBody     string        `json:"html_body" gorm:"type:text"`
	TextBody     string        `json:"text_body" gorm:"type:text"`
	TemplateID   string        `json:"template_id"`
	Variables    string        `json:"variables" gorm:"type:text"`   // JSON as string
	Attachments  string        `json:"attachments" gorm:"type:text"` // JSON as string
	Headers      string        `json:"headers" gorm:"type:text"`     // JSON as string
	Tags         string        `json:"tags"`                         // JSON array as string
	Metadata     string        `json:"metadata" gorm:"type:text"`    // JSON as string
	Priority     EmailPriority `json:"priority" gorm:"default:1"`
	Status       EmailStatus   `json:"status" gorm:"default:'pending'"`
	Provider     EmailProvider `json:"provider"`
	AttemptCount int           `json:"attempt_count" gorm:"default:0"`
	MaxRetries   int           `json:"max_retries" gorm:"default:3"`
	LastError    string        `json:"last_error" gorm:"type:text"`
	ScheduledAt  *time.Time    `json:"scheduled_at"`
	SentAt       *time.Time    `json:"sent_at"`
	CreatedAt    time.Time     `json:"created_at"`
	UpdatedAt    time.Time     `json:"updated_at"`
}

// EmailDeliveryEvent represents an email delivery event
type EmailDeliveryEvent struct {
	ID        string        `json:"id" gorm:"primarykey"`
	EmailID   string        `json:"email_id" gorm:"not null;index"`
	Event     string        `json:"event" gorm:"not null"` // sent, delivered, opened, clicked, bounced, complained
	Data      string        `json:"data" gorm:"type:text"` // JSON data specific to event
	Provider  EmailProvider `json:"provider"`
	Timestamp time.Time     `json:"timestamp"`
	CreatedAt time.Time     `json:"created_at"`
}

// EmailStats represents email statistics
type EmailStats struct {
	TotalSent      int64   `json:"total_sent"`
	TotalDelivered int64   `json:"total_delivered"`
	TotalOpened    int64   `json:"total_opened"`
	TotalClicked   int64   `json:"total_clicked"`
	TotalBounced   int64   `json:"total_bounced"`
	TotalFailed    int64   `json:"total_failed"`
	DeliveryRate   float64 `json:"delivery_rate"`
	OpenRate       float64 `json:"open_rate"`
	ClickRate      float64 `json:"click_rate"`
	BounceRate     float64 `json:"bounce_rate"`
}

// EmailProvider interface defines the contract for email providers
type EmailProviderInterface interface {
	Send(ctx context.Context, message *EmailMessage) (*EmailResult, error)
	SendTemplate(
		ctx context.Context,
		templateID string,
		to []string,
		variables map[string]interface{},
	) (*EmailResult, error)
	GetDeliveryStatus(ctx context.Context, messageID string) (*EmailDeliveryStatus, error)
	SupportsTemplates() bool
	SupportsWebhooks() bool
	GetProviderName() EmailProvider
}

// EmailResult represents the result of sending an email
type EmailResult struct {
	MessageID  string            `json:"message_id"`
	ProviderID string            `json:"provider_id"`
	Status     EmailStatus       `json:"status"`
	Message    string            `json:"message,omitempty"`
	Metadata   map[string]string `json:"metadata,omitempty"`
}

// EmailDeliveryStatus represents the delivery status of an email
type EmailDeliveryStatus struct {
	MessageID   string               `json:"message_id"`
	Status      EmailStatus          `json:"status"`
	DeliveredAt *time.Time           `json:"delivered_at,omitempty"`
	OpenedAt    *time.Time           `json:"opened_at,omitempty"`
	ClickedAt   *time.Time           `json:"clicked_at,omitempty"`
	BouncedAt   *time.Time           `json:"bounced_at,omitempty"`
	Error       string               `json:"error,omitempty"`
	Events      []EmailDeliveryEvent `json:"events,omitempty"`
}

// EmailQueue interface defines the contract for email queuing
type EmailQueueInterface interface {
	Enqueue(ctx context.Context, message *EmailMessage) error
	Dequeue(ctx context.Context, limit int) ([]*QueuedEmail, error)
	MarkSent(ctx context.Context, emailID string, result *EmailResult) error
	MarkFailed(ctx context.Context, emailID string, err error) error
	RetryFailed(ctx context.Context, maxRetries int) error
	GetStats(ctx context.Context) (*QueueStats, error)
	PurgeOld(ctx context.Context, olderThan time.Duration) error
}

// QueueStats represents queue statistics
type QueueStats struct {
	Pending   int64 `json:"pending"`
	Sending   int64 `json:"sending"`
	Sent      int64 `json:"sent"`
	Failed    int64 `json:"failed"`
	Retrying  int64 `json:"retrying"`
	Scheduled int64 `json:"scheduled"`
}

// EmailTemplateEngine interface defines the contract for template engines
type EmailTemplateEngine interface {
	Render(templateID string, variables map[string]interface{}) (*RenderedTemplate, error)
	RegisterTemplate(template *EmailTemplate) error
	GetTemplate(templateID string) (*EmailTemplate, error)
	ListTemplates() ([]*EmailTemplate, error)
	ValidateTemplate(template *EmailTemplate) error
}

// RenderedTemplate represents a rendered email template
type RenderedTemplate struct {
	Subject  string `json:"subject"`
	HTMLBody string `json:"html_body"`
	TextBody string `json:"text_body"`
}

// EmailService interface defines the main email service contract
type EmailServiceInterface interface {
	// Basic sending
	Send(ctx context.Context, message *EmailMessage) error
	SendTemplate(ctx context.Context, templateID string, to []string, variables map[string]interface{}) error
	SendImmediate(ctx context.Context, message *EmailMessage) (*EmailResult, error)

	// Scheduling
	Schedule(ctx context.Context, message *EmailMessage, scheduledAt time.Time) error

	// Template management
	RegisterTemplate(template *EmailTemplate) error
	GetTemplate(templateID string) (*EmailTemplate, error)

	// Queue management
	ProcessQueue(ctx context.Context) error
	GetQueueStats(ctx context.Context) (*QueueStats, error)

	// Delivery tracking
	GetDeliveryStatus(ctx context.Context, messageID string) (*EmailDeliveryStatus, error)
	GetEmailStats(ctx context.Context) (*EmailStats, error)

	// Health check
	HealthCheck(ctx context.Context) error
}

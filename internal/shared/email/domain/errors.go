package domain

import "errors"

var (
	// ErrEmailProviderNotConfigured is returned when the email provider is not properly configured
	ErrEmailProviderNotConfigured = errors.New("email provider not configured")

	// ErrTemplateNotFound is returned when a template is not found
	ErrTemplateNotFound = errors.New("email template not found")

	// ErrTemplateInvalid is returned when a template is invalid
	ErrTemplateInvalid = errors.New("email template is invalid")

	// ErrTemplateMissingVariables is returned when required template variables are missing
	ErrTemplateMissingVariables = errors.New("required template variables are missing")

	// ErrInvalidEmailAddress is returned when an email address is invalid
	ErrInvalidEmailAddress = errors.New("invalid email address")

	// ErrEmailQueueFull is returned when the email queue is full
	ErrEmailQueueFull = errors.New("email queue is full")

	// ErrEmailNotFound is returned when an email is not found in the queue
	ErrEmailNotFound = errors.New("email not found")

	// ErrMaxRetriesExceeded is returned when maximum retry attempts are exceeded
	ErrMaxRetriesExceeded = errors.New("maximum retry attempts exceeded")

	// ErrEmailTooLarge is returned when an email exceeds size limits
	ErrEmailTooLarge = errors.New("email exceeds size limits")

	// ErrAttachmentTooLarge is returned when an attachment exceeds size limits
	ErrAttachmentTooLarge = errors.New("attachment exceeds size limits")

	// ErrProviderRateLimit is returned when the provider rate limit is exceeded
	ErrProviderRateLimit = errors.New("provider rate limit exceeded")

	// ErrProviderTemporaryFailure is returned when the provider has a temporary failure
	ErrProviderTemporaryFailure = errors.New("provider temporary failure")

	// ErrProviderPermanentFailure is returned when the provider has a permanent failure
	ErrProviderPermanentFailure = errors.New("provider permanent failure")

	// ErrWebhookSignatureInvalid is returned when a webhook signature is invalid
	ErrWebhookSignatureInvalid = errors.New("webhook signature is invalid")

	// ErrDeliveryTracking is returned when delivery tracking fails
	ErrDeliveryTracking = errors.New("delivery tracking failed")
)

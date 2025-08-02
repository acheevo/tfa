package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/acheevo/tfa/internal/shared/email/domain"
)

// DatabaseQueue implements EmailQueueInterface using database storage
type DatabaseQueue struct {
	db     *gorm.DB
	logger *slog.Logger
}

// NewDatabaseQueue creates a new database-backed email queue
func NewDatabaseQueue(db *gorm.DB, logger *slog.Logger) *DatabaseQueue {
	return &DatabaseQueue{
		db:     db,
		logger: logger,
	}
}

// Enqueue adds an email message to the queue
func (q *DatabaseQueue) Enqueue(ctx context.Context, message *domain.EmailMessage) error {
	// Convert message to queued email
	queuedEmail := q.messageToQueuedEmail(message)

	// Save to database
	if err := q.db.WithContext(ctx).Create(queuedEmail).Error; err != nil {
		q.logger.Error("failed to enqueue email", "error", err, "message_id", message.ID)
		return fmt.Errorf("failed to enqueue email: %w", err)
	}

	q.logger.Info("email enqueued successfully",
		"message_id", message.ID,
		"to", message.To,
		"subject", message.Subject,
		"priority", message.Priority,
	)

	return nil
}

// Dequeue retrieves emails from the queue for processing
func (q *DatabaseQueue) Dequeue(ctx context.Context, limit int) ([]*domain.QueuedEmail, error) {
	var emails []*domain.QueuedEmail

	// Get emails ready for processing (pending or retrying, and scheduled time has passed)
	now := time.Now()
	err := q.db.WithContext(ctx).
		Where("status IN (?, ?) AND (scheduled_at IS NULL OR scheduled_at <= ?)",
			domain.StatusPending, domain.StatusRetrying, now).
		Order("priority DESC, created_at ASC").
		Limit(limit).
		Find(&emails).Error
	if err != nil {
		q.logger.Error("failed to dequeue emails", "error", err)
		return nil, fmt.Errorf("failed to dequeue emails: %w", err)
	}

	// Mark emails as sending to prevent duplicate processing
	if len(emails) > 0 {
		var ids []string
		for _, email := range emails {
			ids = append(ids, email.ID)
		}

		err = q.db.WithContext(ctx).
			Model(&domain.QueuedEmail{}).
			Where("id IN ?", ids).
			Update("status", domain.StatusSending).Error
		if err != nil {
			q.logger.Error("failed to mark emails as sending", "error", err)
			return nil, fmt.Errorf("failed to mark emails as sending: %w", err)
		}
	}

	q.logger.Debug("dequeued emails for processing", "count", len(emails))
	return emails, nil
}

// MarkSent marks an email as successfully sent
func (q *DatabaseQueue) MarkSent(ctx context.Context, emailID string, result *domain.EmailResult) error {
	updates := map[string]interface{}{
		"status":  domain.StatusSent,
		"sent_at": time.Now(),
	}

	if result != nil {
		if result.Metadata != nil {
			metadataJSON, _ := json.Marshal(result.Metadata)
			updates["metadata"] = string(metadataJSON)
		}
	}

	err := q.db.WithContext(ctx).
		Model(&domain.QueuedEmail{}).
		Where("id = ?", emailID).
		Updates(updates).Error
	if err != nil {
		q.logger.Error("failed to mark email as sent", "error", err, "email_id", emailID)
		return fmt.Errorf("failed to mark email as sent: %w", err)
	}

	q.logger.Info("email marked as sent", "email_id", emailID)
	return nil
}

// MarkFailed marks an email as failed
func (q *DatabaseQueue) MarkFailed(ctx context.Context, emailID string, failureErr error) error {
	var queuedEmail domain.QueuedEmail
	if err := q.db.WithContext(ctx).Where("id = ?", emailID).First(&queuedEmail).Error; err != nil {
		return fmt.Errorf("failed to find email: %w", err)
	}

	queuedEmail.AttemptCount++
	queuedEmail.LastError = failureErr.Error()

	// Check if we should retry or mark as permanently failed
	if queuedEmail.AttemptCount >= queuedEmail.MaxRetries {
		queuedEmail.Status = domain.StatusFailed
		q.logger.Warn("email permanently failed after max retries",
			"email_id", emailID,
			"attempts", queuedEmail.AttemptCount,
			"error", failureErr.Error(),
		)
	} else {
		queuedEmail.Status = domain.StatusRetrying
		// Calculate exponential backoff for next retry
		backoffSeconds := calculateBackoff(queuedEmail.AttemptCount)
		nextRetry := time.Now().Add(time.Duration(backoffSeconds) * time.Second)
		queuedEmail.ScheduledAt = &nextRetry

		q.logger.Info("email scheduled for retry",
			"email_id", emailID,
			"attempt", queuedEmail.AttemptCount,
			"next_retry", nextRetry,
			"error", failureErr.Error(),
		)
	}

	if err := q.db.WithContext(ctx).Save(&queuedEmail).Error; err != nil {
		q.logger.Error("failed to update email failure status", "error", err, "email_id", emailID)
		return fmt.Errorf("failed to update email failure status: %w", err)
	}

	return nil
}

// RetryFailed retries failed emails that haven't exceeded max retries
func (q *DatabaseQueue) RetryFailed(ctx context.Context, maxRetries int) error {
	result := q.db.WithContext(ctx).
		Model(&domain.QueuedEmail{}).
		Where("status = ? AND attempt_count < ?", domain.StatusFailed, maxRetries).
		Updates(map[string]interface{}{
			"status":       domain.StatusPending,
			"scheduled_at": nil,
		})

	if result.Error != nil {
		q.logger.Error("failed to retry failed emails", "error", result.Error)
		return fmt.Errorf("failed to retry failed emails: %w", result.Error)
	}

	q.logger.Info("retried failed emails", "count", result.RowsAffected)
	return nil
}

// GetStats returns queue statistics
func (q *DatabaseQueue) GetStats(ctx context.Context) (*domain.QueueStats, error) {
	stats := &domain.QueueStats{}

	// Get counts for each status
	statusCounts := []struct {
		Status string
		Count  int64
	}{}

	err := q.db.WithContext(ctx).
		Model(&domain.QueuedEmail{}).
		Select("status, COUNT(*) as count").
		Group("status").
		Find(&statusCounts).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get queue stats: %w", err)
	}

	// Map counts to stats struct
	for _, sc := range statusCounts {
		switch domain.EmailStatus(sc.Status) {
		case domain.StatusPending:
			stats.Pending = sc.Count
		case domain.StatusSending:
			stats.Sending = sc.Count
		case domain.StatusSent:
			stats.Sent = sc.Count
		case domain.StatusFailed:
			stats.Failed = sc.Count
		case domain.StatusRetrying:
			stats.Retrying = sc.Count
		}
	}

	// Count scheduled emails
	err = q.db.WithContext(ctx).
		Model(&domain.QueuedEmail{}).
		Where("scheduled_at > ?", time.Now()).
		Count(&stats.Scheduled).Error
	if err != nil {
		return nil, fmt.Errorf("failed to count scheduled emails: %w", err)
	}

	return stats, nil
}

// PurgeOld removes old emails from the queue
func (q *DatabaseQueue) PurgeOld(ctx context.Context, olderThan time.Duration) error {
	cutoff := time.Now().Add(-olderThan)

	result := q.db.WithContext(ctx).
		Where("created_at < ? AND status IN (?, ?)", cutoff, domain.StatusSent, domain.StatusFailed).
		Delete(&domain.QueuedEmail{})

	if result.Error != nil {
		q.logger.Error("failed to purge old emails", "error", result.Error)
		return fmt.Errorf("failed to purge old emails: %w", result.Error)
	}

	q.logger.Info("purged old emails", "count", result.RowsAffected, "older_than", olderThan)
	return nil
}

// messageToQueuedEmail converts an EmailMessage to a QueuedEmail
func (q *DatabaseQueue) messageToQueuedEmail(message *domain.EmailMessage) *domain.QueuedEmail {
	// Generate ID if not provided
	if message.ID == "" {
		message.ID = uuid.New().String()
	}

	// Marshal complex fields to JSON
	toJSON, _ := json.Marshal(message.To)
	ccJSON, _ := json.Marshal(message.CC)
	bccJSON, _ := json.Marshal(message.BCC)
	variablesJSON, _ := json.Marshal(message.Variables)
	attachmentsJSON, _ := json.Marshal(message.Attachments)
	headersJSON, _ := json.Marshal(message.Headers)
	tagsJSON, _ := json.Marshal(message.Tags)
	metadataJSON, _ := json.Marshal(message.Metadata)

	queuedEmail := &domain.QueuedEmail{
		ID:          uuid.New().String(),
		MessageID:   message.ID,
		From:        message.From,
		FromName:    message.FromName,
		To:          string(toJSON),
		CC:          string(ccJSON),
		BCC:         string(bccJSON),
		ReplyTo:     message.ReplyTo,
		Subject:     message.Subject,
		HTMLBody:    message.HTMLBody,
		TextBody:    message.TextBody,
		TemplateID:  message.TemplateID,
		Variables:   string(variablesJSON),
		Attachments: string(attachmentsJSON),
		Headers:     string(headersJSON),
		Tags:        string(tagsJSON),
		Metadata:    string(metadataJSON),
		Priority:    message.Priority,
		Status:      domain.StatusPending,
		MaxRetries:  3, // Default max retries
		ScheduledAt: message.ScheduledAt,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	return queuedEmail
}

// QueuedEmailToMessage converts a QueuedEmail back to an EmailMessage
func (q *DatabaseQueue) QueuedEmailToMessage(queuedEmail *domain.QueuedEmail) (*domain.EmailMessage, error) {
	message := &domain.EmailMessage{
		ID:          queuedEmail.MessageID,
		From:        queuedEmail.From,
		FromName:    queuedEmail.FromName,
		ReplyTo:     queuedEmail.ReplyTo,
		Subject:     queuedEmail.Subject,
		HTMLBody:    queuedEmail.HTMLBody,
		TextBody:    queuedEmail.TextBody,
		TemplateID:  queuedEmail.TemplateID,
		Priority:    queuedEmail.Priority,
		ScheduledAt: queuedEmail.ScheduledAt,
		CreatedAt:   queuedEmail.CreatedAt,
	}

	// Unmarshal JSON fields
	if queuedEmail.To != "" {
		if err := json.Unmarshal([]byte(queuedEmail.To), &message.To); err != nil {
			q.logger.Error("failed to unmarshal To field", "error", err, "email_id", queuedEmail.ID)
		}
	}
	if queuedEmail.CC != "" {
		if err := json.Unmarshal([]byte(queuedEmail.CC), &message.CC); err != nil {
			q.logger.Error("failed to unmarshal CC field", "error", err, "email_id", queuedEmail.ID)
		}
	}
	if queuedEmail.BCC != "" {
		if err := json.Unmarshal([]byte(queuedEmail.BCC), &message.BCC); err != nil {
			q.logger.Error("failed to unmarshal BCC field", "error", err, "email_id", queuedEmail.ID)
		}
	}
	if queuedEmail.Variables != "" {
		if err := json.Unmarshal([]byte(queuedEmail.Variables), &message.Variables); err != nil {
			q.logger.Error("failed to unmarshal Variables field", "error", err, "email_id", queuedEmail.ID)
		}
	}
	if queuedEmail.Attachments != "" {
		if err := json.Unmarshal([]byte(queuedEmail.Attachments), &message.Attachments); err != nil {
			q.logger.Error("failed to unmarshal Attachments field", "error", err, "email_id", queuedEmail.ID)
		}
	}
	if queuedEmail.Headers != "" {
		if err := json.Unmarshal([]byte(queuedEmail.Headers), &message.Headers); err != nil {
			q.logger.Error("failed to unmarshal Headers field", "error", err, "email_id", queuedEmail.ID)
		}
	}
	if queuedEmail.Tags != "" {
		if err := json.Unmarshal([]byte(queuedEmail.Tags), &message.Tags); err != nil {
			q.logger.Error("failed to unmarshal Tags field", "error", err, "email_id", queuedEmail.ID)
		}
	}
	if queuedEmail.Metadata != "" {
		if err := json.Unmarshal([]byte(queuedEmail.Metadata), &message.Metadata); err != nil {
			q.logger.Error("failed to unmarshal Metadata field", "error", err, "email_id", queuedEmail.ID)
		}
	}

	return message, nil
}

// calculateBackoff calculates exponential backoff delay in seconds
func calculateBackoff(attempt int) int {
	// Exponential backoff: 2^attempt minutes, capped at 60 minutes
	// Ensure attempt is within safe bounds to prevent overflow
	if attempt > 6 {
		attempt = 6 // Cap at 2^6 = 64 minutes to prevent overflow
	}
	// Use a safe conversion approach
	var backoff int
	if attempt >= 0 && attempt <= 6 {
		backoff = 1 << attempt // 2^attempt
	} else {
		backoff = 64 // fallback to max value
	}
	if backoff > 60 {
		backoff = 60
	}
	return backoff * 60 // Convert to seconds
}

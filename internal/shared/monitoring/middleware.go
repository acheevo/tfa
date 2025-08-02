package monitoring

import (
	"log/slog"
	"runtime"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/acheevo/tfa/internal/shared/config"
	"github.com/acheevo/tfa/internal/shared/monitoring/metrics"
)

// MonitoringMiddleware provides comprehensive monitoring for HTTP requests
func MonitoringMiddleware(
	config *config.Config,
	metricsCollector metrics.MetricsCollector,
	logger *slog.Logger,
) gin.HandlerFunc {
	if !config.MetricsEnabled {
		return func(c *gin.Context) { c.Next() }
	}

	defaultMetrics := metrics.GetDefaultMetrics()

	return gin.HandlerFunc(func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		// Increment in-flight requests
		labels := map[string]string{
			"method":   method,
			"endpoint": path,
		}
		_ = metricsCollector.IncrementGauge(defaultMetrics.HTTP.RequestsInFlight, labels)

		// Process request
		c.Next()

		// Calculate metrics after request processing
		duration := time.Since(start)
		status := strconv.Itoa(c.Writer.Status())

		// Update labels with response status
		labels["status"] = status

		// Record metrics
		_ = metricsCollector.IncrementCounter(defaultMetrics.HTTP.RequestsTotal, labels)
		_ = metricsCollector.ObserveHistogram(defaultMetrics.HTTP.RequestDuration, duration.Seconds(), labels)

		// Record request/response sizes if available
		if c.Request.ContentLength > 0 {
			_ = metricsCollector.ObserveHistogram(defaultMetrics.HTTP.RequestSize, float64(c.Request.ContentLength), labels)
		}

		responseSize := c.Writer.Size()
		if responseSize > 0 {
			_ = metricsCollector.ObserveHistogram(defaultMetrics.HTTP.ResponseSize, float64(responseSize), labels)
		}

		// Decrement in-flight requests
		_ = metricsCollector.DecrementGauge(defaultMetrics.HTTP.RequestsInFlight, map[string]string{
			"method":   method,
			"endpoint": path,
		})

		// Enhanced logging for monitoring
		logLevel := slog.LevelInfo
		if c.Writer.Status() >= 400 {
			logLevel = slog.LevelWarn
		}
		if c.Writer.Status() >= 500 {
			logLevel = slog.LevelError
		}

		logger.Log(c.Request.Context(), logLevel, "HTTP request completed",
			"method", method,
			"path", path,
			"status", status,
			"duration", duration.String(),
			"request_size", c.Request.ContentLength,
			"response_size", responseSize,
			"ip", c.ClientIP(),
			"user_agent", c.Request.UserAgent(),
			"trace_id", c.GetString("trace_id"),
			"request_id", c.GetString("request_id"),
		)
	})
}

// DatabaseMetricsMiddleware provides database monitoring
func DatabaseMetricsMiddleware(metricsCollector metrics.MetricsCollector) func(operation, table string) func() {
	defaultMetrics := metrics.GetDefaultMetrics()

	return func(operation, table string) func() {
		start := time.Now()

		return func() {
			duration := time.Since(start)
			labels := map[string]string{
				"operation": operation,
				"table":     table,
			}

			_ = metricsCollector.IncrementCounter(defaultMetrics.Database.QueriesTotal, labels)
			_ = metricsCollector.ObserveHistogram(defaultMetrics.Database.QueryDuration, duration.Seconds(), labels)
		}
	}
}

// EmailMetricsRecorder provides email monitoring
type EmailMetricsRecorder struct {
	metricsCollector metrics.MetricsCollector
	defaultMetrics   *metrics.DefaultMetrics
}

// NewEmailMetricsRecorder creates a new email metrics recorder
func NewEmailMetricsRecorder(metricsCollector metrics.MetricsCollector) *EmailMetricsRecorder {
	return &EmailMetricsRecorder{
		metricsCollector: metricsCollector,
		defaultMetrics:   metrics.GetDefaultMetrics(),
	}
}

// RecordEmailSent records a successful email send
func (e *EmailMetricsRecorder) RecordEmailSent(provider, template string, duration time.Duration) {
	labels := map[string]string{
		"provider": provider,
		"template": template,
	}

	_ = e.metricsCollector.IncrementCounter(e.defaultMetrics.Email.EmailsSent, labels)
	_ = e.metricsCollector.ObserveHistogram(e.defaultMetrics.Email.EmailDeliveryTime, duration.Seconds(), labels)
}

// RecordEmailFailed records a failed email send
func (e *EmailMetricsRecorder) RecordEmailFailed(provider, template, reason string) {
	labels := map[string]string{
		"provider": provider,
		"template": template,
		"reason":   reason,
	}

	_ = e.metricsCollector.IncrementCounter(e.defaultMetrics.Email.EmailsFailed, labels)
}

// RecordEmailQueued records an email being queued
func (e *EmailMetricsRecorder) RecordEmailQueued(priority string) {
	labels := map[string]string{
		"priority": priority,
	}

	_ = e.metricsCollector.IncrementGauge(e.defaultMetrics.Email.EmailsQueued, labels)
}

// RecordEmailDequeued records an email being dequeued
func (e *EmailMetricsRecorder) RecordEmailDequeued(priority string) {
	labels := map[string]string{
		"priority": priority,
	}

	_ = e.metricsCollector.DecrementGauge(e.defaultMetrics.Email.EmailsQueued, labels)
}

// AuthMetricsRecorder provides authentication monitoring
type AuthMetricsRecorder struct {
	metricsCollector metrics.MetricsCollector
	defaultMetrics   *metrics.DefaultMetrics
}

// NewAuthMetricsRecorder creates a new auth metrics recorder
func NewAuthMetricsRecorder(metricsCollector metrics.MetricsCollector) *AuthMetricsRecorder {
	return &AuthMetricsRecorder{
		metricsCollector: metricsCollector,
		defaultMetrics:   metrics.GetDefaultMetrics(),
	}
}

// RecordLoginAttempt records a login attempt
func (a *AuthMetricsRecorder) RecordLoginAttempt(method, result string) {
	labels := map[string]string{
		"method": method,
		"result": result,
	}

	_ = a.metricsCollector.IncrementCounter(a.defaultMetrics.Auth.LoginAttempts, labels)

	// Also record specific success/failure counters
	if result == "success" {
		_ = a.metricsCollector.IncrementCounter(a.defaultMetrics.Auth.LoginSuccesses, map[string]string{"method": method})
	} else {
		_ = a.metricsCollector.IncrementCounter(a.defaultMetrics.Auth.LoginFailures, map[string]string{"method": method})
	}
}

// RecordTokenIssued records a token being issued
func (a *AuthMetricsRecorder) RecordTokenIssued(tokenType string) {
	labels := map[string]string{
		"type": tokenType,
	}

	_ = a.metricsCollector.IncrementCounter(a.defaultMetrics.Auth.TokensIssued, labels)
}

// RecordTokenValidated records a token validation
func (a *AuthMetricsRecorder) RecordTokenValidated(tokenType, result string) {
	labels := map[string]string{
		"type":   tokenType,
		"result": result,
	}

	_ = a.metricsCollector.IncrementCounter(a.defaultMetrics.Auth.TokensValidated, labels)
}

// RecordPasswordReset records a password reset request
func (a *AuthMetricsRecorder) RecordPasswordReset(method string) {
	labels := map[string]string{
		"method": method,
	}

	_ = a.metricsCollector.IncrementCounter(a.defaultMetrics.Auth.PasswordResets, labels)
}

// BusinessMetricsRecorder provides business metrics recording
type BusinessMetricsRecorder struct {
	metricsCollector metrics.MetricsCollector
	defaultMetrics   *metrics.DefaultMetrics
}

// NewBusinessMetricsRecorder creates a new business metrics recorder
func NewBusinessMetricsRecorder(metricsCollector metrics.MetricsCollector) *BusinessMetricsRecorder {
	return &BusinessMetricsRecorder{
		metricsCollector: metricsCollector,
		defaultMetrics:   metrics.GetDefaultMetrics(),
	}
}

// RecordUserRegistration records a user registration
func (b *BusinessMetricsRecorder) RecordUserRegistration(source string) {
	labels := map[string]string{
		"source": source,
	}

	_ = b.metricsCollector.IncrementCounter(b.defaultMetrics.Business.UsersRegistered, labels)
}

// RecordActiveUsers records the number of active users
func (b *BusinessMetricsRecorder) RecordActiveUsers(count float64) {
	_ = b.metricsCollector.SetGauge(b.defaultMetrics.Business.UsersActive, count, nil)
}

// RecordUserSession records a user session
func (b *BusinessMetricsRecorder) RecordUserSession(sessionType string) {
	labels := map[string]string{
		"type": sessionType,
	}

	_ = b.metricsCollector.IncrementCounter(b.defaultMetrics.Business.UserSessions, labels)
}

// RecordFeatureUsage records feature usage
func (b *BusinessMetricsRecorder) RecordFeatureUsage(feature string) {
	labels := map[string]string{
		"feature": feature,
	}

	_ = b.metricsCollector.IncrementCounter(b.defaultMetrics.Business.FeatureUsage, labels)
}

// RecordError records an application error
func (b *BusinessMetricsRecorder) RecordError(code, severity string) {
	labels := map[string]string{
		"code":     code,
		"severity": severity,
	}

	_ = b.metricsCollector.IncrementCounter(b.defaultMetrics.Business.ErrorsTotal, labels)
}

// RecordFileUpload records a file upload
func (b *BusinessMetricsRecorder) RecordFileUpload(fileType string, size float64) {
	labels := map[string]string{
		"type": fileType,
	}

	_ = b.metricsCollector.IncrementCounter(b.defaultMetrics.Business.UploadedFiles, labels)

	// Also record file size if provided
	if size > 0 {
		_ = b.metricsCollector.ObserveHistogram("file_upload_size_bytes", size, labels)
	}
}

// SystemMetricsCollector collects system-level metrics
type SystemMetricsCollector struct {
	metricsCollector metrics.MetricsCollector
	defaultMetrics   *metrics.DefaultMetrics
	logger           *slog.Logger
}

// NewSystemMetricsCollector creates a new system metrics collector
func NewSystemMetricsCollector(metricsCollector metrics.MetricsCollector, logger *slog.Logger) *SystemMetricsCollector {
	return &SystemMetricsCollector{
		metricsCollector: metricsCollector,
		defaultMetrics:   metrics.GetDefaultMetrics(),
		logger:           logger,
	}
}

// StartSystemMetricsCollection starts collecting system metrics
func (s *SystemMetricsCollector) StartSystemMetricsCollection() {
	ticker := time.NewTicker(30 * time.Second) // Collect every 30 seconds
	go func() {
		defer ticker.Stop()
		for range ticker.C {
			s.collectSystemMetrics()
		}
	}()
}

// collectSystemMetrics collects various system metrics
func (s *SystemMetricsCollector) collectSystemMetrics() {
	// This would integrate with actual system monitoring
	// For now, we'll collect some basic Go runtime metrics

	// Memory stats
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	_ = s.metricsCollector.SetGauge(s.defaultMetrics.System.MemoryUsage, float64(m.Alloc), nil)
	_ = s.metricsCollector.SetGauge(s.defaultMetrics.System.GoroutinesCount, float64(runtime.NumGoroutine()), nil)

	// GC stats
	_ = s.metricsCollector.ObserveHistogram(s.defaultMetrics.System.GCDuration,
		float64(m.PauseTotalNs)/1e9, // Convert to seconds
		nil)
}

// GetAllRecorders returns all metrics recorders for easy access
func GetAllRecorders(
	metricsCollector metrics.MetricsCollector,
	logger *slog.Logger,
) (*EmailMetricsRecorder, *AuthMetricsRecorder, *BusinessMetricsRecorder, *SystemMetricsCollector) {
	return NewEmailMetricsRecorder(metricsCollector),
		NewAuthMetricsRecorder(metricsCollector),
		NewBusinessMetricsRecorder(metricsCollector),
		NewSystemMetricsCollector(metricsCollector, logger)
}

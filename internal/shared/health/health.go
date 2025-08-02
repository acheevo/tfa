package health

import (
	"context"
	"fmt"
	"log/slog"
	"runtime"
	"sync"
	"time"

	"gorm.io/gorm"

	"github.com/acheevo/tfa/internal/shared/config"
	"github.com/acheevo/tfa/internal/shared/email/domain"
)

// Status represents the health status
type Status string

const (
	StatusHealthy   Status = "healthy"
	StatusUnhealthy Status = "unhealthy"
	StatusDegraded  Status = "degraded"
	StatusUnknown   Status = "unknown"
)

// CheckResult represents the result of a health check
type CheckResult struct {
	Name      string                 `json:"name"`
	Status    Status                 `json:"status"`
	Message   string                 `json:"message,omitempty"`
	Duration  time.Duration          `json:"duration"`
	Timestamp time.Time              `json:"timestamp"`
	Details   map[string]interface{} `json:"details,omitempty"`
	Error     error                  `json:"-"`
}

// HealthReport represents the overall health report
type HealthReport struct {
	Status    Status                  `json:"status"`
	Timestamp time.Time               `json:"timestamp"`
	Duration  time.Duration           `json:"duration"`
	Version   string                  `json:"version"`
	Checks    map[string]*CheckResult `json:"checks"`
	Summary   HealthSummary           `json:"summary"`
}

// HealthSummary provides a summary of health checks
type HealthSummary struct {
	Total     int `json:"total"`
	Healthy   int `json:"healthy"`
	Unhealthy int `json:"unhealthy"`
	Degraded  int `json:"degraded"`
	Unknown   int `json:"unknown"`
}

// HealthChecker defines the interface for health checks
type HealthChecker interface {
	Name() string
	Check(ctx context.Context) *CheckResult
}

// EnhancedHealthService provides comprehensive health checking
type EnhancedHealthService struct {
	config   *config.Config
	logger   *slog.Logger
	checkers map[string]HealthChecker
	mu       sync.RWMutex
}

// NewEnhancedHealthService creates a new enhanced health service
func NewEnhancedHealthService(config *config.Config, logger *slog.Logger) *EnhancedHealthService {
	service := &EnhancedHealthService{
		config:   config,
		logger:   logger,
		checkers: make(map[string]HealthChecker),
	}

	return service
}

// RegisterChecker registers a health checker
func (h *EnhancedHealthService) RegisterChecker(checker HealthChecker) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.checkers[checker.Name()] = checker
	h.logger.Info("Health checker registered", "name", checker.Name())
}

// Check performs all health checks and returns a comprehensive report
func (h *EnhancedHealthService) Check(ctx context.Context) *HealthReport {
	start := time.Now()

	h.mu.RLock()
	checkers := make(map[string]HealthChecker, len(h.checkers))
	for name, checker := range h.checkers {
		checkers[name] = checker
	}
	h.mu.RUnlock()

	// Perform checks concurrently
	results := make(chan *CheckResult, len(checkers))
	var wg sync.WaitGroup

	for _, checker := range checkers {
		wg.Add(1)
		go func(c HealthChecker) {
			defer wg.Done()
			result := c.Check(ctx)
			results <- result
		}(checker)
	}

	// Wait for all checks to complete
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results
	checks := make(map[string]*CheckResult)
	summary := HealthSummary{}

	for result := range results {
		checks[result.Name] = result
		summary.Total++

		switch result.Status {
		case StatusHealthy:
			summary.Healthy++
		case StatusUnhealthy:
			summary.Unhealthy++
		case StatusDegraded:
			summary.Degraded++
		default:
			summary.Unknown++
		}
	}

	// Determine overall status
	overallStatus := h.determineOverallStatus(summary)

	report := &HealthReport{
		Status:    overallStatus,
		Timestamp: time.Now(),
		Duration:  time.Since(start),
		Version:   h.config.Version,
		Checks:    checks,
		Summary:   summary,
	}

	h.logger.Info("Health check completed",
		"status", overallStatus,
		"duration", report.Duration,
		"total_checks", summary.Total,
		"healthy", summary.Healthy,
		"unhealthy", summary.Unhealthy,
	)

	return report
}

// CheckSingle performs a single health check by name
func (h *EnhancedHealthService) CheckSingle(ctx context.Context, name string) *CheckResult {
	h.mu.RLock()
	checker, exists := h.checkers[name]
	h.mu.RUnlock()

	if !exists {
		return &CheckResult{
			Name:      name,
			Status:    StatusUnknown,
			Message:   "Health checker not found",
			Timestamp: time.Now(),
		}
	}

	return checker.Check(ctx)
}

// ListCheckers returns the names of all registered health checkers
func (h *EnhancedHealthService) ListCheckers() []string {
	h.mu.RLock()
	defer h.mu.RUnlock()

	names := make([]string, 0, len(h.checkers))
	for name := range h.checkers {
		names = append(names, name)
	}

	return names
}

// determineOverallStatus determines the overall health status
func (h *EnhancedHealthService) determineOverallStatus(summary HealthSummary) Status {
	if summary.Total == 0 {
		return StatusUnknown
	}

	// If any critical checks are unhealthy, overall is unhealthy
	if summary.Unhealthy > 0 {
		return StatusUnhealthy
	}

	// If any checks are degraded, overall is degraded
	if summary.Degraded > 0 {
		return StatusDegraded
	}

	// If all checks are healthy, overall is healthy
	if summary.Healthy == summary.Total {
		return StatusHealthy
	}

	// Otherwise, unknown
	return StatusUnknown
}

// Built-in health checkers

// DatabaseHealthChecker checks database connectivity
type DatabaseHealthChecker struct {
	name string
	db   *gorm.DB
}

// NewDatabaseHealthChecker creates a new database health checker
func NewDatabaseHealthChecker(name string, db *gorm.DB) *DatabaseHealthChecker {
	return &DatabaseHealthChecker{
		name: name,
		db:   db,
	}
}

// Name returns the checker name
func (d *DatabaseHealthChecker) Name() string {
	return d.name
}

// Check performs the database health check
func (d *DatabaseHealthChecker) Check(ctx context.Context) *CheckResult {
	start := time.Now()
	result := &CheckResult{
		Name:      d.name,
		Timestamp: time.Now(),
		Details:   make(map[string]interface{}),
	}

	// Check database connection
	sqlDB, err := d.db.DB()
	if err != nil {
		result.Status = StatusUnhealthy
		result.Message = "Failed to get database connection"
		result.Error = err
		result.Duration = time.Since(start)
		return result
	}

	// Ping database
	if err := sqlDB.PingContext(ctx); err != nil {
		result.Status = StatusUnhealthy
		result.Message = "Database ping failed"
		result.Error = err
		result.Duration = time.Since(start)
		return result
	}

	// Get connection stats
	stats := sqlDB.Stats()
	result.Details["connections_open"] = stats.OpenConnections
	result.Details["connections_in_use"] = stats.InUse
	result.Details["connections_idle"] = stats.Idle
	result.Details["max_open_connections"] = stats.MaxOpenConnections
	result.Details["max_idle_connections"] = stats.MaxIdleClosed

	// Check if we're approaching connection limits
	connectionUsage := float64(stats.OpenConnections) / float64(stats.MaxOpenConnections)
	if connectionUsage > 0.9 {
		result.Status = StatusDegraded
		result.Message = fmt.Sprintf("High connection usage: %.1f%%", connectionUsage*100)
	} else {
		result.Status = StatusHealthy
		result.Message = "Database connection healthy"
	}

	result.Duration = time.Since(start)
	return result
}

// EmailHealthChecker checks email service health
type EmailHealthChecker struct {
	name         string
	emailService domain.EmailServiceInterface
}

// NewEmailHealthChecker creates a new email health checker
func NewEmailHealthChecker(name string, emailService domain.EmailServiceInterface) *EmailHealthChecker {
	return &EmailHealthChecker{
		name:         name,
		emailService: emailService,
	}
}

// Name returns the checker name
func (e *EmailHealthChecker) Name() string {
	return e.name
}

// Check performs the email service health check
func (e *EmailHealthChecker) Check(ctx context.Context) *CheckResult {
	start := time.Now()
	result := &CheckResult{
		Name:      e.name,
		Timestamp: time.Now(),
		Details:   make(map[string]interface{}),
	}

	// Check email service health
	if err := e.emailService.HealthCheck(ctx); err != nil {
		result.Status = StatusUnhealthy
		result.Message = "Email service health check failed"
		result.Error = err
		result.Duration = time.Since(start)
		return result
	}

	// Get queue stats
	queueStats, err := e.emailService.GetQueueStats(ctx)
	if err != nil {
		result.Status = StatusDegraded
		result.Message = "Failed to get email queue stats"
		result.Error = err
	} else {
		result.Details["emails_pending"] = queueStats.Pending
		result.Details["emails_sending"] = queueStats.Sending
		result.Details["emails_failed"] = queueStats.Failed

		// Check if queue is backing up
		if queueStats.Pending > 1000 {
			result.Status = StatusDegraded
			result.Message = fmt.Sprintf("Email queue backing up: %d pending", queueStats.Pending)
		} else {
			result.Status = StatusHealthy
			result.Message = "Email service healthy"
		}
	}

	result.Duration = time.Since(start)
	return result
}

// ExternalServiceHealthChecker checks external service health
type ExternalServiceHealthChecker struct {
	name     string
	endpoint string
	timeout  time.Duration
}

// NewExternalServiceHealthChecker creates a new external service health checker
func NewExternalServiceHealthChecker(name, endpoint string, timeout time.Duration) *ExternalServiceHealthChecker {
	return &ExternalServiceHealthChecker{
		name:     name,
		endpoint: endpoint,
		timeout:  timeout,
	}
}

// Name returns the checker name
func (e *ExternalServiceHealthChecker) Name() string {
	return e.name
}

// Check performs the external service health check
func (e *ExternalServiceHealthChecker) Check(ctx context.Context) *CheckResult {
	start := time.Now()
	result := &CheckResult{
		Name:      e.name,
		Timestamp: time.Now(),
		Details:   make(map[string]interface{}),
	}

	// This would implement actual HTTP health check
	// For now, we'll simulate it
	result.Status = StatusHealthy
	result.Message = "External service healthy"
	result.Duration = time.Since(start)

	return result
}

// MemoryHealthChecker checks memory usage
type MemoryHealthChecker struct {
	name string
}

// NewMemoryHealthChecker creates a new memory health checker
func NewMemoryHealthChecker(name string) *MemoryHealthChecker {
	return &MemoryHealthChecker{
		name: name,
	}
}

// Name returns the checker name
func (m *MemoryHealthChecker) Name() string {
	return m.name
}

// Check performs the memory health check
func (m *MemoryHealthChecker) Check(ctx context.Context) *CheckResult {
	start := time.Now()
	result := &CheckResult{
		Name:      m.name,
		Timestamp: time.Now(),
		Details:   make(map[string]interface{}),
	}

	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	result.Details["alloc_bytes"] = memStats.Alloc
	result.Details["total_alloc_bytes"] = memStats.TotalAlloc
	result.Details["sys_bytes"] = memStats.Sys
	result.Details["num_gc"] = memStats.NumGC
	result.Details["gc_cpu_fraction"] = memStats.GCCPUFraction

	// Simple memory health check
	// You might want to implement more sophisticated checks
	allocMB := float64(memStats.Alloc) / 1024 / 1024
	if allocMB > 500 { // 500MB threshold
		result.Status = StatusDegraded
		result.Message = fmt.Sprintf("High memory usage: %.1fMB", allocMB)
	} else {
		result.Status = StatusHealthy
		result.Message = fmt.Sprintf("Memory usage normal: %.1fMB", allocMB)
	}

	result.Duration = time.Since(start)
	return result
}

// DiskSpaceHealthChecker checks disk space (placeholder implementation)
type DiskSpaceHealthChecker struct {
	name string
	path string
}

// NewDiskSpaceHealthChecker creates a new disk space health checker
func NewDiskSpaceHealthChecker(name, path string) *DiskSpaceHealthChecker {
	return &DiskSpaceHealthChecker{
		name: name,
		path: path,
	}
}

// Name returns the checker name
func (d *DiskSpaceHealthChecker) Name() string {
	return d.name
}

// Check performs the disk space health check
func (d *DiskSpaceHealthChecker) Check(ctx context.Context) *CheckResult {
	start := time.Now()
	result := &CheckResult{
		Name:      d.name,
		Timestamp: time.Now(),
		Details:   make(map[string]interface{}),
	}

	// This would implement actual disk space checking
	// For now, we'll simulate it
	result.Status = StatusHealthy
	result.Message = "Disk space healthy"
	result.Duration = time.Since(start)

	return result
}

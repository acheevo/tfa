package metrics

import (
	"context"
	"time"
)

// MetricType defines the type of metric
type MetricType string

const (
	MetricTypeCounter   MetricType = "counter"
	MetricTypeGauge     MetricType = "gauge"
	MetricTypeHistogram MetricType = "histogram"
	MetricTypeSummary   MetricType = "summary"
)

// Metric represents a metric measurement
type Metric struct {
	Name      string            `json:"name"`
	Type      MetricType        `json:"type"`
	Value     float64           `json:"value"`
	Labels    map[string]string `json:"labels,omitempty"`
	Timestamp time.Time         `json:"timestamp"`
	Help      string            `json:"help,omitempty"`
}

// MetricsCollector defines the interface for collecting metrics
type MetricsCollector interface {
	// Counter metrics
	IncrementCounter(name string, labels map[string]string) error
	IncrementCounterBy(name string, value float64, labels map[string]string) error

	// Gauge metrics
	SetGauge(name string, value float64, labels map[string]string) error
	IncrementGauge(name string, labels map[string]string) error
	DecrementGauge(name string, labels map[string]string) error

	// Histogram metrics
	ObserveHistogram(name string, value float64, labels map[string]string) error

	// Summary metrics
	ObserveSummary(name string, value float64, labels map[string]string) error

	// Timing utilities
	StartTimer(name string, labels map[string]string) Timer
	RecordDuration(name string, duration time.Duration, labels map[string]string) error

	// Registration
	RegisterMetric(metric *MetricDefinition) error

	// Collection
	Collect(ctx context.Context) ([]*Metric, error)
}

// Timer represents a timing measurement
type Timer interface {
	Stop() time.Duration
	StopAndRecord() error
}

// MetricDefinition defines a metric schema
type MetricDefinition struct {
	Name       string              `json:"name"`
	Type       MetricType          `json:"type"`
	Help       string              `json:"help"`
	Labels     []string            `json:"labels,omitempty"`
	Buckets    []float64           `json:"buckets,omitempty"`    // For histograms
	Objectives map[float64]float64 `json:"objectives,omitempty"` // For summaries
}

// HTTPMetrics represents HTTP-specific metrics
type HTTPMetrics struct {
	RequestsTotal    string
	RequestDuration  string
	RequestSize      string
	ResponseSize     string
	RequestsInFlight string
}

// DatabaseMetrics represents database-specific metrics
type DatabaseMetrics struct {
	ConnectionsOpen     string
	ConnectionsIdle     string
	ConnectionsInUse    string
	QueriesTotal        string
	QueryDuration       string
	TransactionsTotal   string
	TransactionDuration string
}

// EmailMetrics represents email-specific metrics
type EmailMetrics struct {
	EmailsSent         string
	EmailsFailed       string
	EmailsQueued       string
	EmailDeliveryTime  string
	EmailTemplatesUsed string
}

// AuthMetrics represents authentication-specific metrics
type AuthMetrics struct {
	LoginAttempts   string
	LoginSuccesses  string
	LoginFailures   string
	TokensIssued    string
	TokensValidated string
	PasswordResets  string
}

// SystemMetrics represents system-level metrics
type SystemMetrics struct {
	CPUUsage        string
	MemoryUsage     string
	DiskUsage       string
	NetworkBytesIn  string
	NetworkBytesOut string
	GoroutinesCount string
	GCDuration      string
}

// BusinessMetrics represents business-specific metrics
type BusinessMetrics struct {
	UsersRegistered string
	UsersActive     string
	UserSessions    string
	FeatureUsage    string
	ErrorsTotal     string
	UploadedFiles   string
}

// DefaultMetrics contains all standard metric names
type DefaultMetrics struct {
	HTTP     HTTPMetrics
	Database DatabaseMetrics
	Email    EmailMetrics
	Auth     AuthMetrics
	System   SystemMetrics
	Business BusinessMetrics
}

// GetDefaultMetrics returns the default metric definitions
func GetDefaultMetrics() *DefaultMetrics {
	return &DefaultMetrics{
		HTTP: HTTPMetrics{
			RequestsTotal:    "http_requests_total",
			RequestDuration:  "http_request_duration_seconds",
			RequestSize:      "http_request_size_bytes",
			ResponseSize:     "http_response_size_bytes",
			RequestsInFlight: "http_requests_in_flight",
		},
		Database: DatabaseMetrics{
			ConnectionsOpen:     "db_connections_open",
			ConnectionsIdle:     "db_connections_idle",
			ConnectionsInUse:    "db_connections_in_use",
			QueriesTotal:        "db_queries_total",
			QueryDuration:       "db_query_duration_seconds",
			TransactionsTotal:   "db_transactions_total",
			TransactionDuration: "db_transaction_duration_seconds",
		},
		Email: EmailMetrics{
			EmailsSent:         "emails_sent_total",
			EmailsFailed:       "emails_failed_total",
			EmailsQueued:       "emails_queued",
			EmailDeliveryTime:  "email_delivery_duration_seconds",
			EmailTemplatesUsed: "email_templates_used_total",
		},
		Auth: AuthMetrics{
			LoginAttempts:   "auth_login_attempts_total",
			LoginSuccesses:  "auth_login_successes_total",
			LoginFailures:   "auth_login_failures_total",
			TokensIssued:    "auth_tokens_issued_total",
			TokensValidated: "auth_tokens_validated_total",
			PasswordResets:  "auth_password_resets_total",
		},
		System: SystemMetrics{
			CPUUsage:        "system_cpu_usage_percent",
			MemoryUsage:     "system_memory_usage_bytes",
			DiskUsage:       "system_disk_usage_bytes",
			NetworkBytesIn:  "system_network_bytes_in_total",
			NetworkBytesOut: "system_network_bytes_out_total",
			GoroutinesCount: "system_goroutines_count",
			GCDuration:      "system_gc_duration_seconds",
		},
		Business: BusinessMetrics{
			UsersRegistered: "business_users_registered_total",
			UsersActive:     "business_users_active",
			UserSessions:    "business_user_sessions_total",
			FeatureUsage:    "business_feature_usage_total",
			ErrorsTotal:     "business_errors_total",
			UploadedFiles:   "business_uploaded_files_total",
		},
	}
}

// DefaultHistogramBuckets provides default histogram buckets for different use cases
var DefaultHistogramBuckets = map[string][]float64{
	"http_duration":  {0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
	"db_duration":    {0.0001, 0.0005, 0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1},
	"email_duration": {0.1, 0.5, 1, 2, 5, 10, 30, 60, 120, 300},
	"file_size":      {1024, 10240, 102400, 1048576, 10485760, 104857600, 1073741824}, // 1KB to 1GB
}

// DefaultSummaryObjectives provides default summary objectives
var DefaultSummaryObjectives = map[float64]float64{
	0.5:  0.05,
	0.9:  0.01,
	0.95: 0.005,
	0.99: 0.001,
}

// MetricsRegistry manages metric definitions and provides a central registry
type MetricsRegistry struct {
	definitions map[string]*MetricDefinition
	collector   MetricsCollector
}

// NewMetricsRegistry creates a new metrics registry
func NewMetricsRegistry(collector MetricsCollector) *MetricsRegistry {
	registry := &MetricsRegistry{
		definitions: make(map[string]*MetricDefinition),
		collector:   collector,
	}

	// Register default metrics
	registry.RegisterDefaultMetrics()

	return registry
}

// RegisterDefaultMetrics registers all default metric definitions
func (r *MetricsRegistry) RegisterDefaultMetrics() {
	metrics := GetDefaultMetrics()

	// HTTP metrics
	_ = r.RegisterMetric(&MetricDefinition{
		Name:   metrics.HTTP.RequestsTotal,
		Type:   MetricTypeCounter,
		Help:   "Total number of HTTP requests",
		Labels: []string{"method", "status", "endpoint"},
	})

	_ = r.RegisterMetric(&MetricDefinition{
		Name:    metrics.HTTP.RequestDuration,
		Type:    MetricTypeHistogram,
		Help:    "HTTP request duration in seconds",
		Labels:  []string{"method", "status", "endpoint"},
		Buckets: DefaultHistogramBuckets["http_duration"],
	})

	_ = r.RegisterMetric(&MetricDefinition{
		Name:   metrics.HTTP.RequestsInFlight,
		Type:   MetricTypeGauge,
		Help:   "Number of HTTP requests currently being processed",
		Labels: []string{"method", "endpoint"},
	})

	// Database metrics
	_ = r.RegisterMetric(&MetricDefinition{
		Name: metrics.Database.ConnectionsOpen,
		Type: MetricTypeGauge,
		Help: "Number of open database connections",
	})

	_ = r.RegisterMetric(&MetricDefinition{
		Name:    metrics.Database.QueryDuration,
		Type:    MetricTypeHistogram,
		Help:    "Database query duration in seconds",
		Labels:  []string{"operation", "table"},
		Buckets: DefaultHistogramBuckets["db_duration"],
	})

	// Email metrics
	_ = r.RegisterMetric(&MetricDefinition{
		Name:   metrics.Email.EmailsSent,
		Type:   MetricTypeCounter,
		Help:   "Total number of emails sent",
		Labels: []string{"provider", "template"},
	})

	_ = r.RegisterMetric(&MetricDefinition{
		Name:   metrics.Email.EmailsQueued,
		Type:   MetricTypeGauge,
		Help:   "Number of emails currently in queue",
		Labels: []string{"priority"},
	})

	// Auth metrics
	_ = r.RegisterMetric(&MetricDefinition{
		Name:   metrics.Auth.LoginAttempts,
		Type:   MetricTypeCounter,
		Help:   "Total number of login attempts",
		Labels: []string{"method", "result"},
	})

	// System metrics
	_ = r.RegisterMetric(&MetricDefinition{
		Name: metrics.System.CPUUsage,
		Type: MetricTypeGauge,
		Help: "CPU usage percentage",
	})

	_ = r.RegisterMetric(&MetricDefinition{
		Name: metrics.System.MemoryUsage,
		Type: MetricTypeGauge,
		Help: "Memory usage in bytes",
	})

	// Business metrics
	_ = r.RegisterMetric(&MetricDefinition{
		Name:   metrics.Business.UsersRegistered,
		Type:   MetricTypeCounter,
		Help:   "Total number of registered users",
		Labels: []string{"source"},
	})

	_ = r.RegisterMetric(&MetricDefinition{
		Name:   metrics.Business.ErrorsTotal,
		Type:   MetricTypeCounter,
		Help:   "Total number of application errors",
		Labels: []string{"code", "severity"},
	})
}

// RegisterMetric registers a metric definition
func (r *MetricsRegistry) RegisterMetric(definition *MetricDefinition) error {
	r.definitions[definition.Name] = definition
	return r.collector.RegisterMetric(definition)
}

// GetMetric returns a metric definition by name
func (r *MetricsRegistry) GetMetric(name string) (*MetricDefinition, bool) {
	def, exists := r.definitions[name]
	return def, exists
}

// ListMetrics returns all registered metric definitions
func (r *MetricsRegistry) ListMetrics() map[string]*MetricDefinition {
	return r.definitions
}

// GetCollector returns the underlying metrics collector
func (r *MetricsRegistry) GetCollector() MetricsCollector {
	return r.collector
}

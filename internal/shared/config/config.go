package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/kelseyhightower/envconfig"
)

const (
	MaskedValue = "***"
)

type Config struct {
	// Application Settings
	Environment string `envconfig:"ENVIRONMENT" default:"development" validate:"oneof=development staging production"`
	Port        string `envconfig:"PORT" default:"8080" validate:"numeric"`
	LogLevel    string `envconfig:"LOG_LEVEL" default:"info" validate:"oneof=debug info warn error"`
	AppName     string `envconfig:"APP_NAME" default:"Fullstack Template"`
	Version     string `envconfig:"APP_VERSION" default:"1.0.0"`

	// Database Configuration
	DatabaseHost     string `envconfig:"DATABASE_HOST" default:"localhost" validate:"required"`
	DatabasePort     string `envconfig:"DATABASE_PORT" default:"5432" validate:"numeric"`
	DatabaseUser     string `envconfig:"DATABASE_USER" default:"postgres" validate:"required"`
	DatabasePassword string `envconfig:"DATABASE_PASSWORD" default:"postgres"`
	DatabaseName     string `envconfig:"DATABASE_NAME" default:"fullstack_template" validate:"required"`
	DatabaseSSLMode  string `envconfig:"DATABASE_SSL_MODE" default:"disable" validate:"oneof=disable require verify-ca verify-full"`

	// Database Pool Configuration
	DBMaxIdleConns    int    `envconfig:"DB_MAX_IDLE_CONNS" default:"10" validate:"min=1,max=100"`
	DBMaxOpenConns    int    `envconfig:"DB_MAX_OPEN_CONNS" default:"100" validate:"min=1,max=1000"`
	DBConnMaxLifetime string `envconfig:"DB_CONN_MAX_LIFETIME" default:"1h" validate:"required"`
	DBConnMaxIdleTime string `envconfig:"DB_CONN_MAX_IDLE_TIME" default:"30m"`

	// JWT Configuration
	JWTSecret               string `envconfig:"JWT_SECRET" default:"your-super-secret-jwt-key-change-this-in-production-32chars-min" validate:"min=32"`
	JWTAccessTokenDuration  string `envconfig:"JWT_ACCESS_TOKEN_DURATION" default:"15m" validate:"required"`
	JWTRefreshTokenDuration string `envconfig:"JWT_REFRESH_TOKEN_DURATION" default:"7d" validate:"required"`
	JWTIssuer               string `envconfig:"JWT_ISSUER" default:"fullstack-template"`

	// Email Configuration
	EmailEnabled  bool   `envconfig:"EMAIL_ENABLED" default:"false"`
	EmailProvider string `envconfig:"EMAIL_PROVIDER" default:"smtp" validate:"oneof=smtp sendgrid postmark mailgun"`
	EmailFrom     string `envconfig:"EMAIL_FROM" default:"noreply@example.com"`
	EmailFromName string `envconfig:"EMAIL_FROM_NAME" default:"App"`

	// SMTP Configuration
	SMTPHost         string `envconfig:"SMTP_HOST" default:"localhost"`
	SMTPPort         int    `envconfig:"SMTP_PORT" default:"587" validate:"min=1,max=65535"`
	SMTPUsername     string `envconfig:"SMTP_USERNAME"`
	SMTPPassword     string `envconfig:"SMTP_PASSWORD"`
	SMTPUseTLS       bool   `envconfig:"SMTP_USE_TLS" default:"true"`
	SMTPSkipTLSCheck bool   `envconfig:"SMTP_SKIP_TLS_CHECK" default:"false"`

	// Email Service Provider Keys
	SendGridAPIKey string `envconfig:"SENDGRID_API_KEY"`
	PostmarkAPIKey string `envconfig:"POSTMARK_API_KEY"`
	MailgunAPIKey  string `envconfig:"MAILGUN_API_KEY"`
	MailgunDomain  string `envconfig:"MAILGUN_DOMAIN"`

	// Application URLs
	FrontendURL string `envconfig:"FRONTEND_URL" default:"http://localhost:3000" validate:"url"`
	BackendURL  string `envconfig:"BACKEND_URL" default:"http://localhost:8080" validate:"url"`

	// Security Configuration
	CSRFSecret       string `envconfig:"CSRF_SECRET" default:"your-super-secret-jwt-key-change-this-in-production-32chars-min" validate:"min=32"`
	CORSOrigins      string `envconfig:"CORS_ORIGINS" default:"http://localhost:3000,http://localhost:8080"`
	SecureHeaders    bool   `envconfig:"SECURE_HEADERS" default:"true"`
	RateLimitEnabled bool   `envconfig:"RATE_LIMIT_ENABLED" default:"true"`

	// Production Validation Settings
	StrictProductionValidation bool `envconfig:"STRICT_PRODUCTION_VALIDATION" default:"false"`
	AllowDevSecretsInProd      bool `envconfig:"ALLOW_DEV_SECRETS_IN_PROD" default:"false"`
	AllowInsecureDBInProd      bool `envconfig:"ALLOW_INSECURE_DB_IN_PROD" default:"false"`

	// Feature Flags
	FeatureFlags FeatureFlags `envconfig:"FEATURES"`

	// Monitoring Configuration
	MetricsEnabled  bool   `envconfig:"METRICS_ENABLED" default:"true"`
	MetricsPort     string `envconfig:"METRICS_PORT" default:"9090"`
	HealthCheckPath string `envconfig:"HEALTH_CHECK_PATH" default:"/api/health"`
	SentryDSN       string `envconfig:"SENTRY_DSN"`
	TracingEnabled  bool   `envconfig:"TRACING_ENABLED" default:"false"`

	// Cache Configuration
	RedisURL     string `envconfig:"REDIS_URL" default:"redis://localhost:6379/0"`
	CacheEnabled bool   `envconfig:"CACHE_ENABLED" default:"true"`
	CachePrefix  string `envconfig:"CACHE_PREFIX" default:"ft:"`
	CacheTTL     string `envconfig:"CACHE_TTL" default:"1h"`

	// File Storage Configuration
	StorageProvider  string `envconfig:"STORAGE_PROVIDER" default:"local" validate:"oneof=local s3 gcs"`
	S3Bucket         string `envconfig:"S3_BUCKET"`
	S3Region         string `envconfig:"S3_REGION" default:"us-east-1"`
	GCSBucket        string `envconfig:"GCS_BUCKET"`
	LocalStoragePath string `envconfig:"LOCAL_STORAGE_PATH" default:"./uploads"`

	// Bootstrap Configuration
	BootstrapEnabled bool   `envconfig:"BOOTSTRAP_ENABLED" default:"true"`
	AdminEmail       string `envconfig:"ADMIN_EMAIL" default:"admin@example.com"`
	AdminPassword    string `envconfig:"ADMIN_PASSWORD" default:"admin123"`
	DemoUserEmail    string `envconfig:"DEMO_USER_EMAIL" default:"user@example.com"`
	DemoUserPassword string `envconfig:"DEMO_USER_PASSWORD" default:"user1234"`
}

// FeatureFlags represents application feature flags
type FeatureFlags struct {
	EmailVerification bool `envconfig:"EMAIL_VERIFICATION" default:"true"`
	TwoFactorAuth     bool `envconfig:"TWO_FACTOR_AUTH" default:"false"`
	AdminAPI          bool `envconfig:"ADMIN_API" default:"true"`
	Metrics           bool `envconfig:"METRICS" default:"true"`
	FileUploads       bool `envconfig:"FILE_UPLOADS" default:"true"`
	SocialLogin       bool `envconfig:"SOCIAL_LOGIN" default:"false"`
	EmailTemplates    bool `envconfig:"EMAIL_TEMPLATES" default:"true"`
	RateLimiting      bool `envconfig:"RATE_LIMITING" default:"true"`
	CSRFProtection    bool `envconfig:"CSRF_PROTECTION" default:"true"`
	SecurityHeaders   bool `envconfig:"SECURITY_HEADERS" default:"true"`
}

// Load loads and validates the application configuration
func Load() (*Config, error) {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return nil, fmt.Errorf("failed to process environment config: %w", err)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	// Apply environment-specific defaults
	cfg.applyEnvironmentDefaults()

	// Validate critical production settings (only if strict validation is enabled)
	if cfg.IsProduction() && cfg.StrictProductionValidation {
		if err := cfg.validateProductionSettings(); err != nil {
			return nil, fmt.Errorf("production validation failed: %w", err)
		}
	}

	return &cfg, nil
}

// Validate validates the configuration using struct tags
func (c *Config) Validate() error {
	validate := validator.New()

	// Validate base struct
	if err := validate.Struct(c); err != nil {
		return err
	}

	// Conditional email validation
	if c.EmailEnabled {
		if err := validate.Var(c.EmailFrom, "required,email"); err != nil {
			return fmt.Errorf("EmailFrom validation failed: %w", err)
		}
	}

	return nil
}

// applyEnvironmentDefaults applies environment-specific configuration defaults
func (c *Config) applyEnvironmentDefaults() {
	switch c.Environment {
	case "production":
		// Production defaults
		if c.LogLevel == "debug" {
			c.LogLevel = "info"
		}
		c.SecureHeaders = true
		c.RateLimitEnabled = true
		c.FeatureFlags.CSRFProtection = true
		c.FeatureFlags.SecurityHeaders = true

	case "staging":
		// Staging defaults
		c.SecureHeaders = true
		c.RateLimitEnabled = true

	case "development":
		// Development defaults - already set in struct tags
		break
	}
}

// validateProductionSettings validates critical production configuration
func (c *Config) validateProductionSettings() error {
	var errors []string

	// Check for default secrets in production (unless explicitly allowed)
	if !c.AllowDevSecretsInProd {
		if strings.Contains(c.JWTSecret, "dev-") || strings.Contains(c.JWTSecret, "your-super-secret") {
			errors = append(errors, "JWT_SECRET must be changed from default value in production (set ALLOW_DEV_SECRETS_IN_PROD=true to override)")
		}

		if strings.Contains(c.CSRFSecret, "dev-") || strings.Contains(c.CSRFSecret, "your-super-secret") {
			errors = append(errors, "CSRF_SECRET must be changed from default value in production (set ALLOW_DEV_SECRETS_IN_PROD=true to override)")
		}
	}

	// Check database SSL mode (unless explicitly allowed)
	if !c.AllowInsecureDBInProd && c.DatabaseSSLMode == "disable" {
		errors = append(errors, "DATABASE_SSL_MODE should not be 'disable' in production (set ALLOW_INSECURE_DB_IN_PROD=true to override)")
	}

	// Check email configuration only if email is enabled
	if c.EmailEnabled {
		if c.EmailProvider == "smtp" && c.SMTPHost == "localhost" {
			errors = append(errors, "SMTP_HOST should not be 'localhost' in production")
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("production configuration errors: %s", strings.Join(errors, "; "))
	}

	return nil
}

func (c *Config) IsDevelopment() bool {
	return c.Environment == "development"
}

func (c *Config) DatabaseDSN() string {
	return "host=" + c.DatabaseHost +
		" port=" + c.DatabasePort +
		" user=" + c.DatabaseUser +
		" password=" + c.DatabasePassword +
		" dbname=" + c.DatabaseName +
		" sslmode=" + c.DatabaseSSLMode
}

func (c *Config) DBConnMaxLifetimeDuration() time.Duration {
	duration, err := time.ParseDuration(c.DBConnMaxLifetime)
	if err != nil {
		return time.Hour
	}
	return duration
}

func (c *Config) JWTAccessTokenDurationParsed() time.Duration {
	duration, err := time.ParseDuration(c.JWTAccessTokenDuration)
	if err != nil {
		return 15 * time.Minute
	}
	return duration
}

func (c *Config) JWTRefreshTokenDurationParsed() time.Duration {
	duration, err := time.ParseDuration(c.JWTRefreshTokenDuration)
	if err != nil {
		return 7 * 24 * time.Hour
	}
	return duration
}

// IsProduction returns true if the environment is production
func (c *Config) IsProduction() bool {
	return c.Environment == "production"
}

// IsStaging returns true if the environment is staging
func (c *Config) IsStaging() bool {
	return c.Environment == "staging"
}

// IsTest returns true if the environment is test
func (c *Config) IsTest() bool {
	return c.Environment == "test"
}

// DBConnMaxIdleTimeDuration parses the DB connection max idle time duration
func (c *Config) DBConnMaxIdleTimeDuration() time.Duration {
	duration, err := time.ParseDuration(c.DBConnMaxIdleTime)
	if err != nil {
		return 30 * time.Minute
	}
	return duration
}

// CacheTTLDuration parses the cache TTL duration
func (c *Config) CacheTTLDuration() time.Duration {
	duration, err := time.ParseDuration(c.CacheTTL)
	if err != nil {
		return time.Hour
	}
	return duration
}

// GetCORSOrigins returns the CORS origins as a slice
func (c *Config) GetCORSOrigins() []string {
	if c.CORSOrigins == "" {
		return []string{"*"}
	}
	return strings.Split(c.CORSOrigins, ",")
}

// GetEmailConfig returns email configuration based on provider
func (c *Config) GetEmailConfig() map[string]any {
	config := map[string]any{
		"provider":  c.EmailProvider,
		"from":      c.EmailFrom,
		"from_name": c.EmailFromName,
	}

	switch c.EmailProvider {
	case "smtp":
		config["host"] = c.SMTPHost
		config["port"] = c.SMTPPort
		config["username"] = c.SMTPUsername
		config["password"] = c.SMTPPassword
		config["use_tls"] = c.SMTPUseTLS
	case "sendgrid":
		config["api_key"] = c.SendGridAPIKey
	case "postmark":
		config["api_key"] = c.PostmarkAPIKey
	case "mailgun":
		config["api_key"] = c.MailgunAPIKey
		config["domain"] = c.MailgunDomain
	}

	return config
}

// IsFeatureEnabled checks if a specific feature is enabled
func (c *Config) IsFeatureEnabled(feature string) bool {
	switch feature {
	case "email_verification":
		return c.FeatureFlags.EmailVerification
	case "two_factor_auth":
		return c.FeatureFlags.TwoFactorAuth
	case "admin_api":
		return c.FeatureFlags.AdminAPI
	case "metrics":
		return c.FeatureFlags.Metrics
	case "file_uploads":
		return c.FeatureFlags.FileUploads
	case "social_login":
		return c.FeatureFlags.SocialLogin
	case "email_templates":
		return c.FeatureFlags.EmailTemplates
	case "rate_limiting":
		return c.FeatureFlags.RateLimiting
	case "csrf_protection":
		return c.FeatureFlags.CSRFProtection
	case "security_headers":
		return c.FeatureFlags.SecurityHeaders
	default:
		return false
	}
}

// GetDatabaseConfig returns database configuration map
func (c *Config) GetDatabaseConfig() map[string]any {
	return map[string]any{
		"host":               c.DatabaseHost,
		"port":               c.DatabasePort,
		"user":               c.DatabaseUser,
		"password":           c.DatabasePassword,
		"name":               c.DatabaseName,
		"ssl_mode":           c.DatabaseSSLMode,
		"max_idle_conns":     c.DBMaxIdleConns,
		"max_open_conns":     c.DBMaxOpenConns,
		"conn_max_lifetime":  c.DBConnMaxLifetime,
		"conn_max_idle_time": c.DBConnMaxIdleTime,
	}
}

// MaskSensitiveData returns a copy of the config with sensitive data masked
func (c *Config) MaskSensitiveData() *Config {
	masked := *c
	masked.DatabasePassword = MaskedValue
	masked.JWTSecret = MaskedValue
	masked.CSRFSecret = MaskedValue
	masked.SMTPPassword = MaskedValue
	masked.SendGridAPIKey = MaskedValue
	masked.PostmarkAPIKey = MaskedValue
	masked.MailgunAPIKey = MaskedValue
	return &masked
}

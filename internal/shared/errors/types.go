package errors

import (
	"fmt"
	"net/http"
	"time"
)

// ErrorCode represents a unique error code
type ErrorCode string

// Standard error codes
const (
	// Client errors (4xx)
	CodeBadRequest        ErrorCode = "BAD_REQUEST"
	CodeUnauthorized      ErrorCode = "UNAUTHORIZED"
	CodeForbidden         ErrorCode = "FORBIDDEN"
	CodeNotFound          ErrorCode = "NOT_FOUND"
	CodeConflict          ErrorCode = "CONFLICT"
	CodeValidationFailed  ErrorCode = "VALIDATION_FAILED"
	CodeRateLimitExceeded ErrorCode = "RATE_LIMIT_EXCEEDED"
	CodeRequestTooLarge   ErrorCode = "REQUEST_TOO_LARGE"
	CodeUnsupportedMedia  ErrorCode = "UNSUPPORTED_MEDIA_TYPE"

	// Authentication & Authorization
	// #nosec G101 -- This is an error code constant, not a credential
	CodeInvalidCredentials ErrorCode = "INVALID_CREDENTIALS"
	CodeTokenExpired       ErrorCode = "TOKEN_EXPIRED"
	CodeTokenInvalid       ErrorCode = "TOKEN_INVALID"
	CodeEmailNotVerified   ErrorCode = "EMAIL_NOT_VERIFIED"
	CodeAccountLocked      ErrorCode = "ACCOUNT_LOCKED"
	CodePermissionDenied   ErrorCode = "PERMISSION_DENIED"

	// Resource errors
	CodeUserNotFound       ErrorCode = "USER_NOT_FOUND"
	CodeUserAlreadyExists  ErrorCode = "USER_ALREADY_EXISTS"
	CodeEmailAlreadyExists ErrorCode = "EMAIL_ALREADY_EXISTS"
	CodeResourceNotFound   ErrorCode = "RESOURCE_NOT_FOUND"
	CodeResourceConflict   ErrorCode = "RESOURCE_CONFLICT"

	// Business logic errors
	CodeInvalidOperation      ErrorCode = "INVALID_OPERATION"
	CodeOperationFailed       ErrorCode = "OPERATION_FAILED"
	CodeDataInconsistent      ErrorCode = "DATA_INCONSISTENT"
	CodeBusinessRuleViolation ErrorCode = "BUSINESS_RULE_VIOLATION"

	// Server errors (5xx)
	CodeInternalError        ErrorCode = "INTERNAL_ERROR"
	CodeServiceUnavailable   ErrorCode = "SERVICE_UNAVAILABLE"
	CodeDatabaseError        ErrorCode = "DATABASE_ERROR"
	CodeExternalServiceError ErrorCode = "EXTERNAL_SERVICE_ERROR"
	CodeConfigurationError   ErrorCode = "CONFIGURATION_ERROR"
	CodeTimeoutError         ErrorCode = "TIMEOUT_ERROR"

	// Email errors
	CodeEmailSendFailed     ErrorCode = "EMAIL_SEND_FAILED"
	CodeEmailConfigError    ErrorCode = "EMAIL_CONFIG_ERROR"
	CodeTemplateNotFound    ErrorCode = "TEMPLATE_NOT_FOUND"
	CodeTemplateRenderError ErrorCode = "TEMPLATE_RENDER_ERROR"
)

// ErrorSeverity represents the severity level of an error
type ErrorSeverity string

const (
	SeverityLow      ErrorSeverity = "low"
	SeverityMedium   ErrorSeverity = "medium"
	SeverityHigh     ErrorSeverity = "high"
	SeverityCritical ErrorSeverity = "critical"
)

// AppError represents a structured application error
type AppError struct {
	Code         ErrorCode              `json:"code"`
	Message      string                 `json:"message"`
	Details      string                 `json:"details,omitempty"`
	Cause        error                  `json:"-"`
	Context      map[string]interface{} `json:"context,omitempty"`
	HTTPStatus   int                    `json:"-"`
	Severity     ErrorSeverity          `json:"-"`
	Timestamp    time.Time              `json:"timestamp"`
	TraceID      string                 `json:"trace_id,omitempty"`
	UserFriendly bool                   `json:"-"`
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("%s: %s - %s", e.Code, e.Message, e.Details)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap returns the underlying cause error
func (e *AppError) Unwrap() error {
	return e.Cause
}

// WithContext adds context information to the error
func (e *AppError) WithContext(key string, value interface{}) *AppError {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}
	e.Context[key] = value
	return e
}

// WithDetails adds additional details to the error
func (e *AppError) WithDetails(details string) *AppError {
	e.Details = details
	return e
}

// WithCause sets the underlying cause error
func (e *AppError) WithCause(cause error) *AppError {
	e.Cause = cause
	return e
}

// WithTraceID sets the trace ID for error tracking
func (e *AppError) WithTraceID(traceID string) *AppError {
	e.TraceID = traceID
	return e
}

// IsUserFriendly returns whether the error message is safe to show to users
func (e *AppError) IsUserFriendly() bool {
	return e.UserFriendly
}

// ErrorResponse represents the structure of error responses sent to clients
type ErrorResponse struct {
	Error     string                 `json:"error"`
	Code      ErrorCode              `json:"code"`
	Message   string                 `json:"message"`
	Details   string                 `json:"details,omitempty"`
	Context   map[string]interface{} `json:"context,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
	TraceID   string                 `json:"trace_id,omitempty"`
}

// ValidationError represents a validation error with field-specific details
type ValidationError struct {
	*AppError
	Fields map[string]string `json:"fields"`
}

// NewValidationError creates a new validation error
func NewValidationError(message string, fields map[string]string) *ValidationError {
	return &ValidationError{
		AppError: &AppError{
			Code:         CodeValidationFailed,
			Message:      message,
			HTTPStatus:   http.StatusBadRequest,
			Severity:     SeverityLow,
			Timestamp:    time.Now(),
			UserFriendly: true,
		},
		Fields: fields,
	}
}

// ErrorWithFields represents an error with additional field information
type ErrorWithFields struct {
	*AppError
	Fields map[string]interface{} `json:"fields,omitempty"`
}

// NewErrorWithFields creates a new error with field information
func NewErrorWithFields(code ErrorCode, message string, fields map[string]interface{}) *ErrorWithFields {
	return &ErrorWithFields{
		AppError: New(code, message),
		Fields:   fields,
	}
}

// ErrorMapper maps error codes to HTTP status codes and default messages
type ErrorMapper struct {
	mappings map[ErrorCode]ErrorMapping
}

// ErrorMapping defines the mapping for an error code
type ErrorMapping struct {
	HTTPStatus     int
	DefaultMessage string
	Severity       ErrorSeverity
	UserFriendly   bool
}

// NewErrorMapper creates a new error mapper with default mappings
func NewErrorMapper() *ErrorMapper {
	mapper := &ErrorMapper{
		mappings: make(map[ErrorCode]ErrorMapping),
	}

	// Register default mappings
	mapper.registerDefaultMappings()

	return mapper
}

// registerDefaultMappings registers the default error code mappings
func (em *ErrorMapper) registerDefaultMappings() {
	mappings := map[ErrorCode]ErrorMapping{
		// Client errors (4xx)
		CodeBadRequest:        {http.StatusBadRequest, "Bad request", SeverityLow, true},
		CodeUnauthorized:      {http.StatusUnauthorized, "Authentication required", SeverityMedium, true},
		CodeForbidden:         {http.StatusForbidden, "Access forbidden", SeverityMedium, true},
		CodeNotFound:          {http.StatusNotFound, "Resource not found", SeverityLow, true},
		CodeConflict:          {http.StatusConflict, "Resource conflict", SeverityLow, true},
		CodeValidationFailed:  {http.StatusBadRequest, "Validation failed", SeverityLow, true},
		CodeRateLimitExceeded: {http.StatusTooManyRequests, "Rate limit exceeded", SeverityMedium, true},
		CodeRequestTooLarge:   {http.StatusRequestEntityTooLarge, "Request too large", SeverityLow, true},
		CodeUnsupportedMedia:  {http.StatusUnsupportedMediaType, "Unsupported media type", SeverityLow, true},

		// Authentication & Authorization
		CodeInvalidCredentials: {http.StatusUnauthorized, "Invalid credentials", SeverityMedium, true},
		CodeTokenExpired:       {http.StatusUnauthorized, "Token expired", SeverityLow, true},
		CodeTokenInvalid:       {http.StatusUnauthorized, "Invalid token", SeverityMedium, true},
		CodeEmailNotVerified:   {http.StatusForbidden, "Email not verified", SeverityMedium, true},
		CodeAccountLocked:      {http.StatusForbidden, "Account locked", SeverityHigh, true},
		CodePermissionDenied:   {http.StatusForbidden, "Permission denied", SeverityMedium, true},

		// Resource errors
		CodeUserNotFound:       {http.StatusNotFound, "User not found", SeverityLow, true},
		CodeUserAlreadyExists:  {http.StatusConflict, "User already exists", SeverityLow, true},
		CodeEmailAlreadyExists: {http.StatusConflict, "Email already exists", SeverityLow, true},
		CodeResourceNotFound:   {http.StatusNotFound, "Resource not found", SeverityLow, true},
		CodeResourceConflict:   {http.StatusConflict, "Resource conflict", SeverityLow, true},

		// Business logic errors
		CodeInvalidOperation:      {http.StatusBadRequest, "Invalid operation", SeverityMedium, true},
		CodeOperationFailed:       {http.StatusUnprocessableEntity, "Operation failed", SeverityMedium, true},
		CodeDataInconsistent:      {http.StatusInternalServerError, "Data inconsistent", SeverityHigh, false},
		CodeBusinessRuleViolation: {http.StatusBadRequest, "Business rule violation", SeverityMedium, true},

		// Server errors (5xx)
		CodeInternalError:        {http.StatusInternalServerError, "Internal server error", SeverityCritical, false},
		CodeServiceUnavailable:   {http.StatusServiceUnavailable, "Service unavailable", SeverityHigh, true},
		CodeDatabaseError:        {http.StatusInternalServerError, "Database error", SeverityHigh, false},
		CodeExternalServiceError: {http.StatusBadGateway, "External service error", SeverityHigh, false},
		CodeConfigurationError:   {http.StatusInternalServerError, "Configuration error", SeverityCritical, false},
		CodeTimeoutError:         {http.StatusGatewayTimeout, "Request timeout", SeverityMedium, true},

		// Email errors
		CodeEmailSendFailed:     {http.StatusInternalServerError, "Email send failed", SeverityMedium, false},
		CodeEmailConfigError:    {http.StatusInternalServerError, "Email configuration error", SeverityHigh, false},
		CodeTemplateNotFound:    {http.StatusInternalServerError, "Email template not found", SeverityMedium, false},
		CodeTemplateRenderError: {http.StatusInternalServerError, "Template render error", SeverityMedium, false},
	}

	for code, mapping := range mappings {
		em.mappings[code] = mapping
	}
}

// GetMapping returns the mapping for an error code
func (em *ErrorMapper) GetMapping(code ErrorCode) (ErrorMapping, bool) {
	mapping, exists := em.mappings[code]
	return mapping, exists
}

// RegisterMapping registers a custom error code mapping
func (em *ErrorMapper) RegisterMapping(code ErrorCode, mapping ErrorMapping) {
	em.mappings[code] = mapping
}

// Global error mapper instance
var defaultErrorMapper = NewErrorMapper()

// Factory functions for creating common errors

// New creates a new AppError with the given code and message
func New(code ErrorCode, message string) *AppError {
	mapping, exists := defaultErrorMapper.GetMapping(code)
	if !exists {
		mapping = ErrorMapping{
			HTTPStatus:   http.StatusInternalServerError,
			Severity:     SeverityMedium,
			UserFriendly: false,
		}
	}

	if message == "" {
		message = mapping.DefaultMessage
	}

	return &AppError{
		Code:         code,
		Message:      message,
		HTTPStatus:   mapping.HTTPStatus,
		Severity:     mapping.Severity,
		Timestamp:    time.Now(),
		UserFriendly: mapping.UserFriendly,
	}
}

// Newf creates a new AppError with formatted message
func Newf(code ErrorCode, format string, args ...interface{}) *AppError {
	return New(code, fmt.Sprintf(format, args...))
}

// Wrap wraps an existing error with an AppError
func Wrap(err error, code ErrorCode, message string) *AppError {
	appErr := New(code, message)
	appErr.Cause = err
	return appErr
}

// Wrapf wraps an existing error with an AppError and formatted message
func Wrapf(err error, code ErrorCode, format string, args ...interface{}) *AppError {
	return Wrap(err, code, fmt.Sprintf(format, args...))
}

// Common error factory functions

// BadRequest creates a bad request error
func BadRequest(message string) *AppError {
	return New(CodeBadRequest, message)
}

// Unauthorized creates an unauthorized error
func Unauthorized(message string) *AppError {
	return New(CodeUnauthorized, message)
}

// Forbidden creates a forbidden error
func Forbidden(message string) *AppError {
	return New(CodeForbidden, message)
}

// NotFound creates a not found error
func NotFound(message string) *AppError {
	return New(CodeNotFound, message)
}

// Conflict creates a conflict error
func Conflict(message string) *AppError {
	return New(CodeConflict, message)
}

// InternalError creates an internal server error
func InternalError(message string) *AppError {
	return New(CodeInternalError, message)
}

// ValidationFailed creates a validation error
func ValidationFailed(message string) *AppError {
	return New(CodeValidationFailed, message)
}

// DatabaseError creates a database error
func DatabaseError(err error) *AppError {
	return Wrap(err, CodeDatabaseError, "Database operation failed")
}

// ExternalServiceError creates an external service error
func ExternalServiceError(service string, err error) *AppError {
	return Wrap(err, CodeExternalServiceError, fmt.Sprintf("External service '%s' error", service))
}

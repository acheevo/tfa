package errors

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"runtime"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ErrorHandler handles application errors and converts them to HTTP responses
type ErrorHandler struct {
	logger      *slog.Logger
	environment string
	mapper      *ErrorMapper
}

// NewErrorHandler creates a new error handler
func NewErrorHandler(logger *slog.Logger, environment string) *ErrorHandler {
	return &ErrorHandler{
		logger:      logger,
		environment: environment,
		mapper:      defaultErrorMapper,
	}
}

// HandleError processes an error and sends an appropriate HTTP response
func (h *ErrorHandler) HandleError(c *gin.Context, err error) {
	if err == nil {
		return
	}

	// Convert error to AppError
	appErr := h.toAppError(err)

	// Add trace ID if available
	if traceID := getTraceID(c); traceID != "" {
		appErr.TraceID = traceID
	}

	// Log the error with appropriate level
	h.logError(c, appErr)

	// Create response
	response := h.createErrorResponse(appErr)

	// Send HTTP response
	c.JSON(appErr.HTTPStatus, response)
}

// HandleErrorWithStatus handles an error with a specific HTTP status
func (h *ErrorHandler) HandleErrorWithStatus(c *gin.Context, err error, status int) {
	appErr := h.toAppError(err)
	appErr.HTTPStatus = status
	h.HandleError(c, appErr)
}

// toAppError converts any error to an AppError
func (h *ErrorHandler) toAppError(err error) *AppError {
	// If it's already an AppError, return it
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr
	}

	// Handle specific error types
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		return NotFound("Resource not found")
	case errors.Is(err, sql.ErrNoRows):
		return NotFound("Resource not found")
	case errors.Is(err, context.DeadlineExceeded):
		return New(CodeTimeoutError, "Request timeout")
	case errors.Is(err, context.Canceled):
		return New(CodeTimeoutError, "Request canceled")
	default:
		// Check for common error patterns
		errMsg := err.Error()

		// Database connection errors
		if strings.Contains(errMsg, "connection refused") ||
			strings.Contains(errMsg, "connection reset") ||
			strings.Contains(errMsg, "database is locked") {
			return Wrap(err, CodeDatabaseError, "Database connection error")
		}

		// Validation errors (you might want to integrate with a validation library)
		if strings.Contains(errMsg, "validation") ||
			strings.Contains(errMsg, "invalid") {
			return Wrap(err, CodeValidationFailed, "Validation failed")
		}

		// Default to internal error
		return Wrap(err, CodeInternalError, "An unexpected error occurred")
	}
}

// createErrorResponse creates an ErrorResponse from an AppError
func (h *ErrorHandler) createErrorResponse(appErr *AppError) *ErrorResponse {
	response := &ErrorResponse{
		Error:     appErr.Code.String(),
		Code:      appErr.Code,
		Timestamp: appErr.Timestamp,
		TraceID:   appErr.TraceID,
	}

	// Determine what message to show based on environment and user-friendliness
	if h.shouldShowDetails(appErr) {
		response.Message = appErr.Message
		response.Details = appErr.Details
		response.Context = appErr.Context
	} else {
		// Show generic message for internal errors
		mapping, exists := h.mapper.GetMapping(appErr.Code)
		if exists {
			response.Message = mapping.DefaultMessage
		} else {
			response.Message = "An error occurred"
		}
	}

	return response
}

// shouldShowDetails determines whether to show error details to the client
func (h *ErrorHandler) shouldShowDetails(appErr *AppError) bool {
	// Always show details in development
	if h.environment == "development" {
		return true
	}

	// Show details for user-friendly errors
	if appErr.UserFriendly {
		return true
	}

	// Hide details for internal errors in production
	return false
}

// logError logs the error with appropriate level and context
func (h *ErrorHandler) logError(c *gin.Context, appErr *AppError) {
	// Create log context
	logCtx := []interface{}{
		"error_code", appErr.Code,
		"http_status", appErr.HTTPStatus,
		"severity", appErr.Severity,
		"trace_id", appErr.TraceID,
		"method", c.Request.Method,
		"path", c.Request.URL.Path,
		"user_agent", c.Request.UserAgent(),
		"ip", c.ClientIP(),
	}

	// Add error context if available
	if appErr.Context != nil {
		contextJSON, _ := json.Marshal(appErr.Context)
		logCtx = append(logCtx, "error_context", string(contextJSON))
	}

	// Add request ID if available
	if requestID := c.GetString("request_id"); requestID != "" {
		logCtx = append(logCtx, "request_id", requestID)
	}

	// Add user ID if available
	if userID := c.GetString("user_id"); userID != "" {
		logCtx = append(logCtx, "user_id", userID)
	}

	// Add stack trace for high severity errors in development
	if (h.environment == "development" || appErr.Severity == SeverityCritical) && appErr.Cause != nil {
		logCtx = append(logCtx, "stack_trace", getStackTrace())
	}

	// Log with appropriate level based on severity
	message := fmt.Sprintf("Error handled: %s", appErr.Message)
	if appErr.Details != "" {
		message += fmt.Sprintf(" - %s", appErr.Details)
	}

	switch appErr.Severity {
	case SeverityLow:
		h.logger.Info(message, logCtx...)
	case SeverityMedium:
		h.logger.Warn(message, logCtx...)
	case SeverityHigh:
		h.logger.Error(message, logCtx...)
	case SeverityCritical:
		h.logger.Error(message, logCtx...)
		// Could also send to external error tracking service here
	default:
		h.logger.Error(message, logCtx...)
	}
}

// RecoveryMiddleware creates a middleware for panic recovery
func (h *ErrorHandler) RecoveryMiddleware() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		var err error

		switch r := recovered.(type) {
		case error:
			err = r
		case string:
			err = errors.New(r)
		default:
			err = fmt.Errorf("panic recovered: %v", r)
		}

		// Log the panic with stack trace
		h.logger.Error("Panic recovered",
			"error", err.Error(),
			"stack_trace", getStackTrace(),
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"trace_id", getTraceID(c),
		)

		// Handle as critical internal error
		appErr := New(CodeInternalError, "Internal server error")
		appErr.Severity = SeverityCritical
		appErr.Cause = err

		h.HandleError(c, appErr)
	})
}

// ErrorReportingMiddleware adds error context to requests
func (h *ErrorHandler) ErrorReportingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Add error handler to context so it can be used in handlers
		c.Set("error_handler", h)

		c.Next()

		// Check for errors in the context
		if len(c.Errors) > 0 {
			// Handle the last error
			lastError := c.Errors.Last()
			h.HandleError(c, lastError.Err)
		}
	}
}

// Helper functions

// getTraceID extracts trace ID from context
func getTraceID(c *gin.Context) string {
	if traceID := c.GetString("trace_id"); traceID != "" {
		return traceID
	}
	if traceID := c.GetHeader("X-Trace-ID"); traceID != "" {
		return traceID
	}
	return ""
}

// getStackTrace returns the current stack trace
func getStackTrace() string {
	buf := make([]byte, 1024*4)
	n := runtime.Stack(buf, false)
	return string(buf[:n])
}

// String method for ErrorCode
func (ec ErrorCode) String() string {
	return string(ec)
}

// Convenience functions for handlers

// HandleGinError is a convenience function for Gin handlers
func HandleGinError(c *gin.Context, err error) {
	if handler, exists := c.Get("error_handler"); exists {
		if h, ok := handler.(*ErrorHandler); ok {
			h.HandleError(c, err)
			return
		}
	}

	// Fallback to basic error handling
	c.JSON(http.StatusInternalServerError, gin.H{
		"error": "Internal server error",
		"code":  CodeInternalError,
	})
}

// MustGetErrorHandler gets the error handler from Gin context or panics
func MustGetErrorHandler(c *gin.Context) *ErrorHandler {
	handler, exists := c.Get("error_handler")
	if !exists {
		panic("error handler not found in context")
	}

	h, ok := handler.(*ErrorHandler)
	if !ok {
		panic("invalid error handler type in context")
	}

	return h
}

// AbortWithError aborts the request with an error
func AbortWithError(c *gin.Context, err error) {
	HandleGinError(c, err)
	c.Abort()
}

// AbortWithAppError aborts the request with an AppError
func AbortWithAppError(c *gin.Context, appErr *AppError) {
	HandleGinError(c, appErr)
	c.Abort()
}

// ErrorMiddleware creates a comprehensive error handling middleware
func ErrorMiddleware(logger *slog.Logger, environment string) gin.HandlerFunc {
	handler := NewErrorHandler(logger, environment)

	return gin.HandlerFunc(func(c *gin.Context) {
		// Add error handler to context
		c.Set("error_handler", handler)

		// Continue with request
		c.Next()

		// Handle any errors that occurred during request processing
		if len(c.Errors) > 0 {
			// Get the last error (most recent)
			lastError := c.Errors.Last()

			// Only handle if response hasn't been written yet
			if !c.Writer.Written() {
				handler.HandleError(c, lastError.Err)
			}
		}
	})
}

package middleware

import (
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/acheevo/tfa/internal/shared/config"
	"github.com/acheevo/tfa/internal/shared/errors"
)

// SecurityHeaders adds security headers to responses
func SecurityHeaders(config *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Prevent MIME type sniffing
		c.Header("X-Content-Type-Options", "nosniff")

		// Prevent clickjacking
		c.Header("X-Frame-Options", "DENY")

		// Enable XSS protection
		c.Header("X-XSS-Protection", "1; mode=block")

		// Prevent downloading of files with dangerous extensions
		c.Header("X-Download-Options", "noopen")

		// Prevent content from being embedded in frames from other origins
		c.Header("Content-Security-Policy", generateCSP(config))

		// Force HTTPS in production
		if config.IsProduction() {
			// HTTP Strict Transport Security (HSTS)
			c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")

			// Prevent HTTP access in production
			if c.Request.Header.Get("X-Forwarded-Proto") != "https" && c.Request.TLS == nil {
				redirectURL := "https://" + c.Request.Host + c.Request.RequestURI
				c.Redirect(http.StatusMovedPermanently, redirectURL)
				c.Abort()
				return
			}
		}

		// Referrer policy
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		// Feature policy / Permissions policy
		c.Header("Permissions-Policy", "camera=(), microphone=(), geolocation=(), interest-cohort=()")

		c.Next()
	}
}

// generateCSP generates a Content Security Policy header
func generateCSP(config *config.Config) string {
	policies := []string{
		"default-src 'self'",
		// Note: 'unsafe-inline' and 'unsafe-eval' should be removed in production with proper nonce/hash
		"script-src 'self' 'unsafe-inline' 'unsafe-eval'",
		"style-src 'self' 'unsafe-inline' https://fonts.googleapis.com",
		"font-src 'self' https://fonts.gstatic.com",
		"img-src 'self' data: https:",
		"connect-src 'self'",
		"frame-ancestors 'none'",
		"base-uri 'self'",
		"form-action 'self'",
	}

	// Add frontend URL to connect-src for API calls
	if config.FrontendURL != "" {
		policies = append(policies, fmt.Sprintf("connect-src 'self' %s", config.FrontendURL))
	}

	return strings.Join(policies, "; ")
}

// CSRFProtection provides CSRF protection using double-submit cookie pattern
func CSRFProtection(config *config.Config, logger *slog.Logger) gin.HandlerFunc {
	if !config.IsFeatureEnabled("csrf_protection") {
		logger.Info("CSRF protection disabled by feature flag")
		return func(c *gin.Context) { c.Next() }
	}

	return func(c *gin.Context) {
		// Skip CSRF protection for safe methods
		if isSafeMethod(c.Request.Method) {
			c.Next()
			return
		}

		// Skip CSRF protection for API endpoints with proper authentication
		if isAPIEndpoint(c.Request.URL.Path) && hasValidAPIAuth(c) {
			c.Next()
			return
		}

		token := getCSRFToken(c)
		if token == "" {
			logger.Warn("CSRF token missing",
				"method", c.Request.Method,
				"path", c.Request.URL.Path,
				"ip", c.ClientIP(),
			)
			errors.AbortWithError(c, errors.New(errors.CodeForbidden, "CSRF token required"))
			return
		}

		if !validateCSRFToken(c, token) {
			logger.Warn("CSRF token validation failed",
				"method", c.Request.Method,
				"path", c.Request.URL.Path,
				"ip", c.ClientIP(),
				"user_agent", c.Request.UserAgent(),
			)
			errors.AbortWithError(c, errors.New(errors.CodeForbidden, "CSRF token invalid"))
			return
		}

		c.Next()
	}
}

// getCSRFToken extracts CSRF token from request
func getCSRFToken(c *gin.Context) string {
	// Try header first
	token := c.GetHeader("X-CSRF-Token")
	if token != "" {
		return token
	}

	// Try form field
	token = c.PostForm("_csrf_token")
	if token != "" {
		return token
	}

	// Try query parameter (less secure, only for specific cases)
	return c.Query("_csrf_token")
}

// validateCSRFToken validates a CSRF token
func validateCSRFToken(c *gin.Context, token string) bool {
	// Get the expected token from cookie
	cookie, err := c.Request.Cookie("_csrf_token")
	if err != nil {
		return false
	}

	// Use constant-time comparison to prevent timing attacks
	return subtle.ConstantTimeCompare([]byte(token), []byte(cookie.Value)) == 1
}

// GenerateCSRFToken generates a new CSRF token
func GenerateCSRFToken(c *gin.Context, config *config.Config) string {
	// Generate a random token
	token := generateSecureToken()

	// Set cookie with the token
	c.SetCookie(
		"_csrf_token",
		token,
		3600, // 1 hour
		"/",
		"",                    // domain
		config.IsProduction(), // secure
		true,                  // httpOnly
	)

	return token
}

// isSafeMethod checks if HTTP method is safe (doesn't modify state)
func isSafeMethod(method string) bool {
	safeMethods := []string{"GET", "HEAD", "OPTIONS", "TRACE"}
	for _, safe := range safeMethods {
		if method == safe {
			return true
		}
	}
	return false
}

// isAPIEndpoint checks if the path is an API endpoint
func isAPIEndpoint(path string) bool {
	return strings.HasPrefix(path, "/api/")
}

// hasValidAPIAuth checks if request has valid API authentication
func hasValidAPIAuth(c *gin.Context) bool {
	// Check for Bearer token
	auth := c.GetHeader("Authorization")
	if strings.HasPrefix(auth, "Bearer ") {
		return true
	}

	// Check for API key
	apiKey := c.GetHeader("X-API-Key")
	return apiKey != ""
}

// generateSecureToken generates a cryptographically secure random token
func generateSecureToken() string {
	return base64.URLEncoding.EncodeToString([]byte(uuid.New().String()))
}

// RequestID middleware adds a unique request ID to each request
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if request ID is already provided
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}

		// Set in context and response header
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)

		c.Next()
	}
}

// TraceID middleware adds a trace ID for distributed tracing
func TraceID() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if trace ID is already provided
		traceID := c.GetHeader("X-Trace-ID")
		if traceID == "" {
			traceID = uuid.New().String()
		}

		// Set in context and response header
		c.Set("trace_id", traceID)
		c.Header("X-Trace-ID", traceID)

		c.Next()
	}
}

// TrustedProxies middleware validates trusted proxy headers
func TrustedProxies(config *config.Config, logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// In production, validate that proxy headers come from trusted sources
		if config.IsProduction() {
			// Check X-Forwarded-For header for suspicious values
			if xff := c.GetHeader("X-Forwarded-For"); xff != "" {
				// Log for monitoring
				logger.Debug("Request with X-Forwarded-For header",
					"xff", xff,
					"remote_addr", c.Request.RemoteAddr,
					"path", c.Request.URL.Path,
				)
			}

			// Check X-Real-IP header
			if xri := c.GetHeader("X-Real-IP"); xri != "" {
				logger.Debug("Request with X-Real-IP header",
					"real_ip", xri,
					"remote_addr", c.Request.RemoteAddr,
					"path", c.Request.URL.Path,
				)
			}
		}

		c.Next()
	}
}

// InputSanitization middleware provides basic input sanitization
func InputSanitization(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check for potentially malicious patterns in query parameters
		for key, values := range c.Request.URL.Query() {
			for _, value := range values {
				if containsSuspiciousContent(value) {
					logger.Warn("Suspicious query parameter detected",
						"key", key,
						"value", value,
						"ip", c.ClientIP(),
						"user_agent", c.Request.UserAgent(),
						"path", c.Request.URL.Path,
					)

					errors.AbortWithError(c, errors.BadRequest("Invalid request parameters"))
					return
				}
			}
		}

		// Check User-Agent for suspicious patterns
		userAgent := c.Request.UserAgent()
		if containsSuspiciousUserAgent(userAgent) {
			logger.Warn("Suspicious User-Agent detected",
				"user_agent", userAgent,
				"ip", c.ClientIP(),
				"path", c.Request.URL.Path,
			)
		}

		c.Next()
	}
}

// containsSuspiciousContent checks for common attack patterns
func containsSuspiciousContent(content string) bool {
	suspiciousPatterns := []string{
		"<script",
		"javascript:",
		"onload=",
		"onerror=",
		"eval(",
		"alert(",
		"document.cookie",
		"document.write",
		"../",
		"..\\",
		"<iframe",
		"<object",
		"<embed",
		"data:text/html",
		"vbscript:",
	}

	lowerContent := strings.ToLower(content)
	for _, pattern := range suspiciousPatterns {
		if strings.Contains(lowerContent, pattern) {
			return true
		}
	}

	return false
}

// containsSuspiciousUserAgent checks for suspicious user agents
func containsSuspiciousUserAgent(userAgent string) bool {
	suspiciousPatterns := []string{
		"nikto",
		"sqlmap",
		"nmap",
		"masscan",
		"zap",
		"burp",
		"acunetix",
		"nessus",
		"openvas",
		"w3af",
		"havij",
		"grabber",
	}

	lowerUA := strings.ToLower(userAgent)
	for _, pattern := range suspiciousPatterns {
		if strings.Contains(lowerUA, pattern) {
			return true
		}
	}

	return false
}

// CORS middleware with security considerations
func SecureCORS(config *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// Check if origin is allowed
		if isAllowedOrigin(origin, config.GetCORSOrigins()) {
			c.Header("Access-Control-Allow-Origin", origin)
		} else if config.IsDevelopment() {
			// In development, be more permissive
			c.Header("Access-Control-Allow-Origin", "*")
		}

		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Headers",
			"Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, "+
				"accept, origin, Cache-Control, X-Requested-With, X-API-Key")
		c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, PATCH")
		c.Header("Access-Control-Expose-Headers", "X-Request-ID, X-Trace-ID")
		c.Header("Access-Control-Max-Age", "86400") // 24 hours

		// Handle preflight requests
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// isAllowedOrigin checks if an origin is in the allowed list
func isAllowedOrigin(origin string, allowedOrigins []string) bool {
	for _, allowed := range allowedOrigins {
		if allowed == "*" || allowed == origin {
			return true
		}
	}
	return false
}

// ContentLengthLimit middleware limits request body size
func ContentLengthLimit(maxBytes int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.ContentLength > maxBytes {
			errors.AbortWithError(c, errors.New(errors.CodeRequestTooLarge,
				fmt.Sprintf("Request body too large (max: %d bytes)", maxBytes)))
			return
		}

		// Also set a limit on the request body reader
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBytes)

		c.Next()
	}
}

// SecurityHeaders middleware that combines multiple security measures
func ComprehensiveSecurity(config *config.Config, logger *slog.Logger) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// Apply security headers
		SecurityHeaders(config)(c)

		// Apply request/trace IDs
		RequestID()(c)
		TraceID()(c)

		// Apply input sanitization
		InputSanitization(logger)(c)

		// Apply trusted proxies validation
		TrustedProxies(config, logger)(c)

		// Apply content length limit (10MB default)
		ContentLengthLimit(10 * 1024 * 1024)(c)

		c.Next()
	})
}

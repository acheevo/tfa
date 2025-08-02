package middleware

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/acheevo/tfa/internal/auth/domain"
	"github.com/acheevo/tfa/internal/auth/service"
)

// AuthMiddleware provides authentication middleware
type AuthMiddleware struct {
	logger      *slog.Logger
	authService *service.AuthService
}

// NewAuthMiddleware creates a new authentication middleware
func NewAuthMiddleware(logger *slog.Logger, authService *service.AuthService) *AuthMiddleware {
	return &AuthMiddleware{
		logger:      logger,
		authService: authService,
	}
}

// RequireAuth middleware that requires valid authentication
func (m *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := m.extractToken(c)
		if token == "" {
			c.JSON(http.StatusUnauthorized, domain.ErrorResponse{
				Error: "authentication required",
			})
			c.Abort()
			return
		}

		claims, err := m.authService.ValidateAccessToken(token)
		if err != nil {
			m.logger.Warn("invalid access token", "error", err)
			c.JSON(http.StatusUnauthorized, domain.ErrorResponse{
				Error: "invalid or expired token",
			})
			c.Abort()
			return
		}

		// Set user information in context
		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)
		c.Set("user_role", claims.Role)
		c.Set("token_type", claims.TokenType)
		c.Set("jwt_claims", claims)

		c.Next()
	}
}

// OptionalAuth middleware that optionally authenticates if token is present
func (m *AuthMiddleware) OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := m.extractToken(c)
		if token == "" {
			c.Next()
			return
		}

		claims, err := m.authService.ValidateAccessToken(token)
		if err != nil {
			// Don't abort for optional auth, just continue without user context
			m.logger.Debug("invalid access token in optional auth", "error", err)
			c.Next()
			return
		}

		// Set user information in context
		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)
		c.Set("user_role", claims.Role)
		c.Set("token_type", claims.TokenType)
		c.Set("jwt_claims", claims)

		c.Next()
	}
}

// RequireEmailVerified middleware that requires email to be verified
func (m *AuthMiddleware) RequireEmailVerified() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, domain.ErrorResponse{
				Error: "authentication required",
			})
			c.Abort()
			return
		}

		uid, ok := userID.(uint)
		if !ok {
			c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
				Error: "invalid user ID",
			})
			c.Abort()
			return
		}

		// Get user profile to check email verification status
		profile, err := m.authService.GetUserProfile(uid)
		if err != nil {
			m.logger.Error("failed to get user profile for email verification check", "user_id", uid, "error", err)
			c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
				Error: "failed to verify user status",
			})
			c.Abort()
			return
		}

		if !profile.EmailVerified {
			c.JSON(http.StatusForbidden, domain.ErrorResponse{
				Error: "email verification required",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireActiveUser middleware that requires user to be active
func (m *AuthMiddleware) RequireActiveUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, domain.ErrorResponse{
				Error: "authentication required",
			})
			c.Abort()
			return
		}

		uid, ok := userID.(uint)
		if !ok {
			c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
				Error: "invalid user ID",
			})
			c.Abort()
			return
		}

		// Get user profile to check active status
		profile, err := m.authService.GetUserProfile(uid)
		if err != nil {
			m.logger.Error("failed to get user profile for active check", "user_id", uid, "error", err)
			c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
				Error: "failed to verify user status",
			})
			c.Abort()
			return
		}

		if profile.Status != domain.StatusActive {
			c.JSON(http.StatusForbidden, domain.ErrorResponse{
				Error: "user account is inactive",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// extractToken extracts the token from the request
// Checks in order: Authorization header, access_token cookie
func (m *AuthMiddleware) extractToken(c *gin.Context) string {
	// Check Authorization header first
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" {
		// Bearer token format: "Bearer <token>"
		if strings.HasPrefix(authHeader, "Bearer ") {
			return strings.TrimPrefix(authHeader, "Bearer ")
		}
	}

	// Check access_token cookie
	token, err := c.Cookie("access_token")
	if err == nil && token != "" {
		return token
	}

	return ""
}

// RequireRole middleware that requires a specific role
func (m *AuthMiddleware) RequireRole(role domain.UserRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, domain.ErrorResponse{
				Error: "authentication required",
			})
			c.Abort()
			return
		}

		uid, ok := userID.(uint)
		if !ok {
			c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
				Error: "invalid user ID",
			})
			c.Abort()
			return
		}

		// Get user profile to check role
		profile, err := m.authService.GetUserProfile(uid)
		if err != nil {
			m.logger.Error("failed to get user profile for role check", "user_id", uid, "error", err)
			c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
				Error: "failed to verify user role",
			})
			c.Abort()
			return
		}

		if profile.Role != role {
			c.JSON(http.StatusForbidden, domain.ErrorResponse{
				Error: "insufficient permissions",
			})
			c.Abort()
			return
		}

		// Set user profile in context for convenience
		c.Set("user_profile", profile)

		c.Next()
	}
}

// RequireAdmin middleware that requires admin role
func (m *AuthMiddleware) RequireAdmin() gin.HandlerFunc {
	return m.RequireRole(domain.RoleAdmin)
}

// RequireUserRole middleware that requires user role or higher
func (m *AuthMiddleware) RequireUserRole() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, domain.ErrorResponse{
				Error: "authentication required",
			})
			c.Abort()
			return
		}

		uid, ok := userID.(uint)
		if !ok {
			c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
				Error: "invalid user ID",
			})
			c.Abort()
			return
		}

		// Get user profile to check role
		profile, err := m.authService.GetUserProfile(uid)
		if err != nil {
			m.logger.Error("failed to get user profile for role check", "user_id", uid, "error", err)
			c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
				Error: "failed to verify user role",
			})
			c.Abort()
			return
		}

		// Check if user has valid role (user or admin)
		if profile.Role != domain.RoleUser && profile.Role != domain.RoleAdmin {
			c.JSON(http.StatusForbidden, domain.ErrorResponse{
				Error: "insufficient permissions",
			})
			c.Abort()
			return
		}

		// Set user profile in context for convenience
		c.Set("user_profile", profile)

		c.Next()
	}
}

// RequireActiveUserWithRole combines active user check with role validation
func (m *AuthMiddleware) RequireActiveUserWithRole(role domain.UserRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, domain.ErrorResponse{
				Error: "authentication required",
			})
			c.Abort()
			return
		}

		uid, ok := userID.(uint)
		if !ok {
			c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
				Error: "invalid user ID",
			})
			c.Abort()
			return
		}

		// Get user profile to check status and role
		profile, err := m.authService.GetUserProfile(uid)
		if err != nil {
			m.logger.Error("failed to get user profile for active role check", "user_id", uid, "error", err)
			c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
				Error: "failed to verify user status",
			})
			c.Abort()
			return
		}

		// Check if user is active
		if profile.Status != domain.StatusActive {
			c.JSON(http.StatusForbidden, domain.ErrorResponse{
				Error: "user account is inactive",
			})
			c.Abort()
			return
		}

		// Check role
		if profile.Role != role {
			c.JSON(http.StatusForbidden, domain.ErrorResponse{
				Error: "insufficient permissions",
			})
			c.Abort()
			return
		}

		// Set user profile in context for convenience
		c.Set("user_profile", profile)

		c.Next()
	}
}

// RequireActiveAdmin combines active user and admin role checks
func (m *AuthMiddleware) RequireActiveAdmin() gin.HandlerFunc {
	return m.RequireActiveUserWithRole(domain.RoleAdmin)
}

// Helper functions

// GetCurrentUserID is a helper function to get the current user ID from context
func GetCurrentUserID(c *gin.Context) (uint, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		return 0, false
	}

	uid, ok := userID.(uint)
	return uid, ok
}

// GetCurrentUserEmail is a helper function to get the current user email from context
func GetCurrentUserEmail(c *gin.Context) (string, bool) {
	email, exists := c.Get("user_email")
	if !exists {
		return "", false
	}

	userEmail, ok := email.(string)
	return userEmail, ok
}

// GetCurrentUserProfile is a helper function to get the current user profile from context
func GetCurrentUserProfile(c *gin.Context) (*domain.UserResponse, bool) {
	profile, exists := c.Get("user_profile")
	if !exists {
		return nil, false
	}

	userProfile, ok := profile.(*domain.UserResponse)
	return userProfile, ok
}

// IsAuthenticated checks if the current request is authenticated
func IsAuthenticated(c *gin.Context) bool {
	_, exists := c.Get("user_id")
	return exists
}

// IsAdmin checks if the current user has admin role
func IsAdmin(c *gin.Context) bool {
	profile, exists := GetCurrentUserProfile(c)
	if !exists {
		return false
	}
	return profile.Role == domain.RoleAdmin
}

// IsActiveUser checks if the current user is active
func IsActiveUser(c *gin.Context) bool {
	profile, exists := GetCurrentUserProfile(c)
	if !exists {
		return false
	}
	return profile.Status == domain.StatusActive
}

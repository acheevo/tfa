package middleware

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/acheevo/tfa/internal/auth/domain"
	"github.com/acheevo/tfa/internal/auth/service"
)

// RBACMiddleware provides role-based access control middleware
type RBACMiddleware struct {
	logger      *slog.Logger
	authService *service.AuthService
}

// NewRBACMiddleware creates a new RBAC middleware
func NewRBACMiddleware(logger *slog.Logger, authService *service.AuthService) *RBACMiddleware {
	return &RBACMiddleware{
		logger:      logger,
		authService: authService,
	}
}

// RequirePermission middleware that requires a specific permission
func (m *RBACMiddleware) RequirePermission(permission domain.Permission) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user role from context (set by auth middleware)
		userRole, exists := m.getUserRole(c)
		if !exists {
			m.logger.Warn("permission check failed: no user role in context", "permission", permission)
			c.JSON(http.StatusUnauthorized, domain.ErrorResponse{
				Error: "authentication required",
			})
			c.Abort()
			return
		}

		// Check permission
		if !domain.HasPermission(userRole, permission) {
			userID, _ := c.Get("user_id")
			m.logger.Warn("permission denied",
				"user_id", userID,
				"user_role", userRole,
				"required_permission", permission,
				"path", c.Request.URL.Path,
			)

			c.JSON(http.StatusForbidden, domain.ErrorResponse{
				Error: "insufficient permissions",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireAnyPermission middleware that requires any of the specified permissions
func (m *RBACMiddleware) RequireAnyPermission(permissions []domain.Permission) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := m.getUserRole(c)
		if !exists {
			m.logger.Warn("permission check failed: no user role in context", "permissions", permissions)
			c.JSON(http.StatusUnauthorized, domain.ErrorResponse{
				Error: "authentication required",
			})
			c.Abort()
			return
		}

		if !domain.HasAnyPermission(userRole, permissions) {
			userID, _ := c.Get("user_id")
			m.logger.Warn("permission denied",
				"user_id", userID,
				"user_role", userRole,
				"required_permissions", permissions,
				"path", c.Request.URL.Path,
			)

			c.JSON(http.StatusForbidden, domain.ErrorResponse{
				Error: "insufficient permissions",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireAllPermissions middleware that requires all of the specified permissions
func (m *RBACMiddleware) RequireAllPermissions(permissions []domain.Permission) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := m.getUserRole(c)
		if !exists {
			m.logger.Warn("permission check failed: no user role in context", "permissions", permissions)
			c.JSON(http.StatusUnauthorized, domain.ErrorResponse{
				Error: "authentication required",
			})
			c.Abort()
			return
		}

		if !domain.HasAllPermissions(userRole, permissions) {
			userID, _ := c.Get("user_id")
			m.logger.Warn("permission denied",
				"user_id", userID,
				"user_role", userRole,
				"required_permissions", permissions,
				"path", c.Request.URL.Path,
			)

			c.JSON(http.StatusForbidden, domain.ErrorResponse{
				Error: "insufficient permissions",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireResourceAccess middleware that requires access to a specific resource with an action
func (m *RBACMiddleware) RequireResourceAccess(resource domain.Resource, action domain.Action) gin.HandlerFunc {
	permission := domain.BuildPermission(resource, action)
	return m.RequirePermission(permission)
}

// RequireUserManagement middleware for user management operations
func (m *RBACMiddleware) RequireUserManagement() gin.HandlerFunc {
	return m.RequirePermission(domain.PermissionUserManage)
}

// RequireAdminAccess middleware for admin-only operations
func (m *RBACMiddleware) RequireAdminAccess() gin.HandlerFunc {
	return m.RequireRole(domain.RoleAdmin)
}

// RequireRole middleware that requires a specific role (enhanced version)
func (m *RBACMiddleware) RequireRole(role domain.UserRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := m.getUserRole(c)
		if !exists {
			m.logger.Warn("role check failed: no user role in context", "required_role", role)
			c.JSON(http.StatusUnauthorized, domain.ErrorResponse{
				Error: "authentication required",
			})
			c.Abort()
			return
		}

		if userRole != role {
			userID, _ := c.Get("user_id")
			m.logger.Warn("role check failed",
				"user_id", userID,
				"user_role", userRole,
				"required_role", role,
				"path", c.Request.URL.Path,
			)

			c.JSON(http.StatusForbidden, domain.ErrorResponse{
				Error: "insufficient role permissions",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireMinimumRole middleware that requires at least the specified role level
func (m *RBACMiddleware) RequireMinimumRole(minRole domain.UserRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := m.getUserRole(c)
		if !exists {
			m.logger.Warn("role check failed: no user role in context", "minimum_role", minRole)
			c.JSON(http.StatusUnauthorized, domain.ErrorResponse{
				Error: "authentication required",
			})
			c.Abort()
			return
		}

		// Check if user role is equal to or higher than minimum role
		if userRole != minRole && !domain.IsRoleHigherThan(userRole, minRole) {
			userID, _ := c.Get("user_id")
			m.logger.Warn("minimum role check failed",
				"user_id", userID,
				"user_role", userRole,
				"minimum_role", minRole,
				"path", c.Request.URL.Path,
			)

			c.JSON(http.StatusForbidden, domain.ErrorResponse{
				Error: "insufficient role level",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireOwnResourceOrPermission middleware for operations on own resources or with specific permission
func (m *RBACMiddleware) RequireOwnResourceOrPermission(permission domain.Permission) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := m.getCurrentUserID(c)
		if !exists {
			c.JSON(http.StatusUnauthorized, domain.ErrorResponse{
				Error: "authentication required",
			})
			c.Abort()
			return
		}

		userRole, _ := m.getUserRole(c)

		// Get target user ID from URL parameter
		targetIDStr := c.Param("id")
		if targetIDStr == "" {
			c.JSON(http.StatusBadRequest, domain.ErrorResponse{
				Error: "missing user ID parameter",
			})
			c.Abort()
			return
		}

		targetID, err := strconv.ParseUint(targetIDStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, domain.ErrorResponse{
				Error: "invalid user ID parameter",
			})
			c.Abort()
			return
		}

		// Allow if accessing own resource
		if userID == uint(targetID) {
			c.Next()
			return
		}

		// Otherwise, check for required permission
		if !domain.HasPermission(userRole, permission) {
			m.logger.Warn("resource access denied",
				"user_id", userID,
				"target_id", targetID,
				"user_role", userRole,
				"required_permission", permission,
				"path", c.Request.URL.Path,
			)

			c.JSON(http.StatusForbidden, domain.ErrorResponse{
				Error: "insufficient permissions",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// WithPermissionLogging middleware that logs permission checks for auditing
func (m *RBACMiddleware) WithPermissionLogging() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _ := m.getCurrentUserID(c)
		userRole, _ := m.getUserRole(c)

		m.logger.Info("API access",
			"user_id", userID,
			"user_role", userRole,
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"ip", c.ClientIP(),
			"user_agent", c.Request.UserAgent(),
		)

		c.Next()
	}
}

// Helper functions

// getUserRole gets the user role from the context
func (m *RBACMiddleware) getUserRole(c *gin.Context) (domain.UserRole, bool) {
	// First try to get from JWT claims if available
	if claims, exists := c.Get("jwt_claims"); exists {
		if jwtClaims, ok := claims.(*domain.JWTClaims); ok {
			return jwtClaims.Role, true
		}
	}

	// Fallback to getting from user profile (set by auth middleware)
	if profile, exists := c.Get("user_profile"); exists {
		if userProfile, ok := profile.(*domain.UserResponse); ok {
			return userProfile.Role, true
		}
	}

	// If no user profile in context, try to fetch it
	userID, exists := m.getCurrentUserID(c)
	if !exists {
		return "", false
	}

	profile, err := m.authService.GetUserProfile(userID)
	if err != nil {
		m.logger.Error("failed to get user profile for role check", "user_id", userID, "error", err)
		return "", false
	}

	// Set in context for future use
	c.Set("user_profile", profile)
	return profile.Role, true
}

// getCurrentUserID gets the current user ID from context
func (m *RBACMiddleware) getCurrentUserID(c *gin.Context) (uint, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		return 0, false
	}

	uid, ok := userID.(uint)
	if !ok {
		return 0, false
	}

	return uid, true
}

// SetRoleInContext sets the user role in context (used by auth middleware)
func SetRoleInContext(c *gin.Context, role domain.UserRole) {
	c.Set("user_role", role)
}

// GetRoleFromContext gets the user role from context
func GetRoleFromContext(c *gin.Context) (domain.UserRole, bool) {
	role, exists := c.Get("user_role")
	if !exists {
		return "", false
	}

	userRole, ok := role.(domain.UserRole)
	if !ok {
		return "", false
	}

	return userRole, true
}

// HasCurrentUserPermission checks if the current user has a specific permission
func HasCurrentUserPermission(c *gin.Context, permission domain.Permission) bool {
	role, exists := GetRoleFromContext(c)
	if !exists {
		return false
	}
	return domain.HasPermission(role, permission)
}

// IsCurrentUserAdmin checks if the current user is an admin
func IsCurrentUserAdmin(c *gin.Context) bool {
	role, exists := GetRoleFromContext(c)
	if !exists {
		return false
	}
	return role == domain.RoleAdmin
}

// Convenience middleware functions

// RequireUserRead requires user read permission
func (m *RBACMiddleware) RequireUserRead() gin.HandlerFunc {
	return m.RequirePermission(domain.PermissionUserRead)
}

// RequireUserWrite requires user write permission
func (m *RBACMiddleware) RequireUserWrite() gin.HandlerFunc {
	return m.RequirePermission(domain.PermissionUserWrite)
}

// RequireProfileAccess requires profile access (own profile or admin permission)
func (m *RBACMiddleware) RequireProfileAccess() gin.HandlerFunc {
	return m.RequireOwnResourceOrPermission(domain.PermissionUserRead)
}

// RequireAuditAccess requires audit log access
func (m *RBACMiddleware) RequireAuditAccess() gin.HandlerFunc {
	return m.RequirePermission(domain.PermissionAuditRead)
}

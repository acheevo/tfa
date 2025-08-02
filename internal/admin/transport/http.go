package transport

import (
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/acheevo/tfa/internal/admin/domain"
	"github.com/acheevo/tfa/internal/admin/service"
	authdomain "github.com/acheevo/tfa/internal/auth/domain"
	"github.com/acheevo/tfa/internal/shared/config"
	userdomain "github.com/acheevo/tfa/internal/user/domain"
)

// AdminHandler handles HTTP requests for admin user management
type AdminHandler struct {
	config       *config.Config
	logger       *slog.Logger
	adminService *service.AdminService
}

// NewAdminHandler creates a new admin handler
func NewAdminHandler(config *config.Config, logger *slog.Logger, adminService *service.AdminService) *AdminHandler {
	return &AdminHandler{
		config:       config,
		logger:       logger,
		adminService: adminService,
	}
}

// ListUsers handles GET /api/admin/users
func (h *AdminHandler) ListUsers(c *gin.Context) {
	adminID := h.getUserID(c)
	if adminID == 0 {
		c.JSON(http.StatusUnauthorized, authdomain.ErrorResponse{Error: "unauthorized"})
		return
	}

	var req userdomain.UserListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		h.handleValidationError(c, err)
		return
	}

	response, err := h.adminService.ListUsers(adminID, &req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetUserDetails handles GET /api/admin/users/:id
func (h *AdminHandler) GetUserDetails(c *gin.Context) {
	adminID := h.getUserID(c)
	if adminID == 0 {
		c.JSON(http.StatusUnauthorized, authdomain.ErrorResponse{Error: "unauthorized"})
		return
	}

	targetUserID, err := h.getTargetUserID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, authdomain.ErrorResponse{Error: "invalid user ID"})
		return
	}

	response, err := h.adminService.GetUserDetails(adminID, targetUserID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, response)
}

// UpdateUserRole handles PUT /api/admin/users/:id/role
func (h *AdminHandler) UpdateUserRole(c *gin.Context) {
	adminID := h.getUserID(c)
	if adminID == 0 {
		c.JSON(http.StatusUnauthorized, authdomain.ErrorResponse{Error: "unauthorized"})
		return
	}

	targetUserID, err := h.getTargetUserID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, authdomain.ErrorResponse{Error: "invalid user ID"})
		return
	}

	var req domain.UpdateUserRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleValidationError(c, err)
		return
	}

	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	err = h.adminService.UpdateUserRole(adminID, targetUserID, &req, ipAddress, userAgent)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, authdomain.MessageResponse{Message: "user role updated successfully"})
}

// UpdateUserStatus handles PUT /api/admin/users/:id/status
func (h *AdminHandler) UpdateUserStatus(c *gin.Context) {
	adminID := h.getUserID(c)
	if adminID == 0 {
		c.JSON(http.StatusUnauthorized, authdomain.ErrorResponse{Error: "unauthorized"})
		return
	}

	targetUserID, err := h.getTargetUserID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, authdomain.ErrorResponse{Error: "invalid user ID"})
		return
	}

	var req domain.UpdateUserStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleValidationError(c, err)
		return
	}

	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	err = h.adminService.UpdateUserStatus(adminID, targetUserID, &req, ipAddress, userAgent)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, authdomain.MessageResponse{Message: "user status updated successfully"})
}

// UpdateUser handles PUT /api/admin/users/:id
func (h *AdminHandler) UpdateUser(c *gin.Context) {
	adminID := h.getUserID(c)
	if adminID == 0 {
		c.JSON(http.StatusUnauthorized, authdomain.ErrorResponse{Error: "unauthorized"})
		return
	}

	targetUserID, err := h.getTargetUserID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, authdomain.ErrorResponse{Error: "invalid user ID"})
		return
	}

	var req domain.AdminUpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleValidationError(c, err)
		return
	}

	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	err = h.adminService.UpdateUser(adminID, targetUserID, &req, ipAddress, userAgent)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, authdomain.MessageResponse{Message: "user updated successfully"})
}

// DeleteUsers handles DELETE /api/admin/users
func (h *AdminHandler) DeleteUsers(c *gin.Context) {
	adminID := h.getUserID(c)
	if adminID == 0 {
		c.JSON(http.StatusUnauthorized, authdomain.ErrorResponse{Error: "unauthorized"})
		return
	}

	var deleteReq domain.DeleteUserRequest
	if err := c.ShouldBindJSON(&deleteReq); err != nil {
		h.handleValidationError(c, err)
		return
	}

	// Get user IDs from query parameter
	userIDsStr := c.Query("ids")
	if userIDsStr == "" {
		c.JSON(http.StatusBadRequest, authdomain.ErrorResponse{Error: "user IDs required"})
		return
	}

	userIDs, err := h.parseUserIDs(userIDsStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, authdomain.ErrorResponse{Error: "invalid user IDs"})
		return
	}

	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	err = h.adminService.DeleteUsers(adminID, &deleteReq, userIDs, ipAddress, userAgent)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, authdomain.MessageResponse{Message: "users deleted successfully"})
}

// BulkUpdateUsers handles POST /api/admin/users/bulk
func (h *AdminHandler) BulkUpdateUsers(c *gin.Context) {
	adminID := h.getUserID(c)
	if adminID == 0 {
		c.JSON(http.StatusUnauthorized, authdomain.ErrorResponse{Error: "unauthorized"})
		return
	}

	var req domain.BulkUserActionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleValidationError(c, err)
		return
	}

	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	result, err := h.adminService.BulkUpdateUsers(adminID, &req, ipAddress, userAgent)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetStats handles GET /api/admin/stats
func (h *AdminHandler) GetStats(c *gin.Context) {
	adminID := h.getUserID(c)
	if adminID == 0 {
		c.JSON(http.StatusUnauthorized, authdomain.ErrorResponse{Error: "unauthorized"})
		return
	}

	stats, err := h.adminService.GetAdminStats(adminID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, stats)
}

// GetAuditLogs handles GET /api/admin/audit-logs
func (h *AdminHandler) GetAuditLogs(c *gin.Context) {
	adminID := h.getUserID(c)
	if adminID == 0 {
		c.JSON(http.StatusUnauthorized, authdomain.ErrorResponse{Error: "unauthorized"})
		return
	}

	var req domain.AdminAuditLogRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		h.handleValidationError(c, err)
		return
	}

	response, err := h.adminService.GetAuditLogs(adminID, &req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, response)
}

// RegisterRoutes registers all admin routes
func (h *AdminHandler) RegisterRoutes(router *gin.RouterGroup) {
	admin := router.Group("/admin")
	{
		// User management
		admin.GET("/users", h.ListUsers)
		admin.GET("/users/:id", h.GetUserDetails)
		admin.PUT("/users/:id", h.UpdateUser)
		admin.PUT("/users/:id/role", h.UpdateUserRole)
		admin.PUT("/users/:id/status", h.UpdateUserStatus)
		admin.DELETE("/users", h.DeleteUsers)
		admin.POST("/users/bulk", h.BulkUpdateUsers)

		// Admin dashboard
		admin.GET("/stats", h.GetStats)
		admin.GET("/audit-logs", h.GetAuditLogs)
	}
}

// Helper methods

// getUserID extracts user ID from Gin context
func (h *AdminHandler) getUserID(c *gin.Context) uint {
	if userID, exists := c.Get("user_id"); exists {
		if id, ok := userID.(uint); ok {
			return id
		}
	}
	return 0
}

// getTargetUserID extracts target user ID from URL parameter
func (h *AdminHandler) getTargetUserID(c *gin.Context) (uint, error) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return 0, err
	}
	return uint(id), nil
}

// parseUserIDs parses comma-separated user IDs
func (h *AdminHandler) parseUserIDs(idsStr string) ([]uint, error) {
	idStrs := strings.Split(idsStr, ",")
	userIDs := make([]uint, 0, len(idStrs))

	for _, idStr := range idStrs {
		idStr = strings.TrimSpace(idStr)
		if idStr == "" {
			continue
		}

		id, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			return nil, err
		}
		userIDs = append(userIDs, uint(id))
	}

	return userIDs, nil
}

// handleError handles service errors and returns appropriate HTTP responses
func (h *AdminHandler) handleError(c *gin.Context, err error) {
	switch err {
	case domain.ErrNotAuthorized:
		c.JSON(http.StatusForbidden, authdomain.ErrorResponse{Error: "not authorized for admin operations"})
	case domain.ErrCannotManageSelf:
		c.JSON(http.StatusForbidden, authdomain.ErrorResponse{Error: "cannot manage own account through admin interface"})
	case domain.ErrBulkActionFailed:
		c.JSON(http.StatusInternalServerError, authdomain.ErrorResponse{Error: "bulk action failed"})
	case domain.ErrAuditLogNotFound:
		c.JSON(http.StatusNotFound, authdomain.ErrorResponse{Error: "audit log not found"})
	case domain.ErrInvalidDateRange:
		c.JSON(http.StatusBadRequest, authdomain.ErrorResponse{Error: "invalid date range"})
	case domain.ErrTooManyUsers:
		c.JSON(http.StatusBadRequest, authdomain.ErrorResponse{Error: "too many users selected for bulk action"})
	case userdomain.ErrUserNotFound:
		c.JSON(http.StatusNotFound, authdomain.ErrorResponse{Error: "user not found"})
	case userdomain.ErrEmailAlreadyExists:
		c.JSON(http.StatusConflict, authdomain.ErrorResponse{Error: "email already exists"})
	default:
		h.logger.Error("unhandled admin service error", "error", err)
		c.JSON(http.StatusInternalServerError, authdomain.ErrorResponse{Error: "internal server error"})
	}
}

// handleValidationError handles validation errors from request binding
func (h *AdminHandler) handleValidationError(c *gin.Context, err error) {
	h.logger.Error("validation error", "error", err)
	c.JSON(http.StatusBadRequest, authdomain.ErrorResponse{
		Error:   "validation failed",
		Details: extractValidationErrors(err),
	})
}

// extractValidationErrors extracts field-specific validation errors
func extractValidationErrors(err error) map[string]string {
	// This is a simplified version - you might want to use a more sophisticated
	// validation error extraction based on your validation library
	return map[string]string{
		"general": err.Error(),
	}
}

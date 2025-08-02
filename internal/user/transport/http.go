package transport

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	authdomain "github.com/acheevo/tfa/internal/auth/domain"
	"github.com/acheevo/tfa/internal/shared/config"
	"github.com/acheevo/tfa/internal/user/domain"
	"github.com/acheevo/tfa/internal/user/service"
)

// UserHandler handles HTTP requests for user management
type UserHandler struct {
	config      *config.Config
	logger      *slog.Logger
	userService *service.UserService
}

// NewUserHandler creates a new user handler
func NewUserHandler(config *config.Config, logger *slog.Logger, userService *service.UserService) *UserHandler {
	return &UserHandler{
		config:      config,
		logger:      logger,
		userService: userService,
	}
}

// GetProfile handles GET /api/user/profile
func (h *UserHandler) GetProfile(c *gin.Context) {
	userID := h.getUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, authdomain.ErrorResponse{Error: "unauthorized"})
		return
	}

	profile, err := h.userService.GetProfile(userID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, profile)
}

// UpdateProfile handles PUT /api/user/profile
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	userID := h.getUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, authdomain.ErrorResponse{Error: "unauthorized"})
		return
	}

	var req domain.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleValidationError(c, err)
		return
	}

	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	profile, err := h.userService.UpdateProfile(userID, &req, ipAddress, userAgent)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, profile)
}

// GetPreferences handles GET /api/user/preferences
func (h *UserHandler) GetPreferences(c *gin.Context) {
	userID := h.getUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, authdomain.ErrorResponse{Error: "unauthorized"})
		return
	}

	preferences, err := h.userService.GetPreferences(userID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, preferences)
}

// UpdatePreferences handles PUT /api/user/preferences
func (h *UserHandler) UpdatePreferences(c *gin.Context) {
	userID := h.getUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, authdomain.ErrorResponse{Error: "unauthorized"})
		return
	}

	var req domain.UpdatePreferencesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleValidationError(c, err)
		return
	}

	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	preferences, err := h.userService.UpdatePreferences(userID, &req, ipAddress, userAgent)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, preferences)
}

// ChangeEmail handles POST /api/user/change-email
func (h *UserHandler) ChangeEmail(c *gin.Context) {
	userID := h.getUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, authdomain.ErrorResponse{Error: "unauthorized"})
		return
	}

	var req domain.ChangeEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleValidationError(c, err)
		return
	}

	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	err := h.userService.ChangeEmail(userID, &req, ipAddress, userAgent)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, authdomain.MessageResponse{
		Message: "Email change requested. Please verify your new email address.",
	})
}

// GetDashboard handles GET /api/user/dashboard
func (h *UserHandler) GetDashboard(c *gin.Context) {
	userID := h.getUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, authdomain.ErrorResponse{Error: "unauthorized"})
		return
	}

	dashboard, err := h.userService.GetDashboard(userID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, dashboard)
}

// RegisterRoutes registers all user management routes
func (h *UserHandler) RegisterRoutes(router *gin.RouterGroup) {
	user := router.Group("/user")
	{
		user.GET("/profile", h.GetProfile)
		user.PUT("/profile", h.UpdateProfile)
		user.GET("/preferences", h.GetPreferences)
		user.PUT("/preferences", h.UpdatePreferences)
		user.POST("/change-email", h.ChangeEmail)
		user.GET("/dashboard", h.GetDashboard)
	}
}

// Helper methods

// getUserID extracts user ID from Gin context
func (h *UserHandler) getUserID(c *gin.Context) uint {
	if userID, exists := c.Get("user_id"); exists {
		if id, ok := userID.(uint); ok {
			return id
		}
	}
	return 0
}

// handleError handles service errors and returns appropriate HTTP responses
func (h *UserHandler) handleError(c *gin.Context, err error) {
	switch err {
	case domain.ErrUserNotFound:
		c.JSON(http.StatusNotFound, authdomain.ErrorResponse{Error: "user not found"})
	case domain.ErrUnauthorized:
		c.JSON(http.StatusUnauthorized, authdomain.ErrorResponse{Error: "unauthorized"})
	case domain.ErrForbidden:
		c.JSON(http.StatusForbidden, authdomain.ErrorResponse{Error: "forbidden"})
	case domain.ErrEmailAlreadyExists:
		c.JSON(http.StatusConflict, authdomain.ErrorResponse{Error: "email already exists"})
	case domain.ErrInvalidPreferences:
		c.JSON(http.StatusBadRequest, authdomain.ErrorResponse{Error: "invalid preferences"})
	case domain.ErrProfileUpdateFailed:
		c.JSON(http.StatusInternalServerError, authdomain.ErrorResponse{Error: "profile update failed"})
	case authdomain.ErrInvalidCredentials:
		c.JSON(http.StatusUnauthorized, authdomain.ErrorResponse{Error: "invalid credentials"})
	default:
		h.logger.Error("unhandled user service error", "error", err)
		c.JSON(http.StatusInternalServerError, authdomain.ErrorResponse{Error: "internal server error"})
	}
}

// handleValidationError handles validation errors from request binding
func (h *UserHandler) handleValidationError(c *gin.Context, err error) {
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

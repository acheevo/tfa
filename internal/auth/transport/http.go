package transport

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/acheevo/tfa/internal/auth/domain"
	"github.com/acheevo/tfa/internal/auth/service"
	"github.com/acheevo/tfa/internal/shared/config"
)

// AuthHandler handles HTTP requests for authentication
type AuthHandler struct {
	config      *config.Config
	logger      *slog.Logger
	authService *service.AuthService
}

// NewAuthHandler creates a new authentication handler
func NewAuthHandler(config *config.Config, logger *slog.Logger, authService *service.AuthService) *AuthHandler {
	return &AuthHandler{
		config:      config,
		logger:      logger,
		authService: authService,
	}
}

// Register handles user registration
func (h *AuthHandler) Register(c *gin.Context) {
	var req domain.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleValidationError(c, err)
		return
	}

	response, err := h.authService.Register(&req)
	if err != nil {
		h.handleAuthError(c, err)
		return
	}

	// Set HTTP-only cookies for tokens
	h.setAuthCookies(c, response.AccessToken, response.RefreshToken)

	c.JSON(http.StatusCreated, response)
}

// Login handles user login
func (h *AuthHandler) Login(c *gin.Context) {
	var req domain.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleValidationError(c, err)
		return
	}

	response, err := h.authService.Login(&req)
	if err != nil {
		h.handleAuthError(c, err)
		return
	}

	// Set HTTP-only cookies for tokens
	h.setAuthCookies(c, response.AccessToken, response.RefreshToken)

	c.JSON(http.StatusOK, response)
}

// RefreshToken handles token refresh
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	// Try to get refresh token from cookie first, then from request body
	refreshToken, err := c.Cookie("refresh_token")
	if err != nil || refreshToken == "" {
		var req domain.RefreshTokenRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			h.handleValidationError(c, err)
			return
		}
		refreshToken = req.RefreshToken
	}

	if refreshToken == "" {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Error: "refresh token is required",
		})
		return
	}

	req := &domain.RefreshTokenRequest{RefreshToken: refreshToken}
	response, err := h.authService.RefreshToken(req)
	if err != nil {
		h.handleAuthError(c, err)
		return
	}

	// Set HTTP-only cookies for tokens
	h.setAuthCookies(c, response.AccessToken, response.RefreshToken)

	c.JSON(http.StatusOK, response)
}

// Logout handles user logout
func (h *AuthHandler) Logout(c *gin.Context) {
	// Get refresh token from cookie
	refreshToken, err := c.Cookie("refresh_token")
	if err != nil || refreshToken == "" {
		// Clear cookies anyway
		h.clearAuthCookies(c)
		c.JSON(http.StatusOK, domain.MessageResponse{Message: "logged out successfully"})
		return
	}

	if err := h.authService.Logout(refreshToken); err != nil {
		h.logger.Error("failed to logout", "error", err)
		// Still clear cookies even if logout fails
	}

	h.clearAuthCookies(c)
	c.JSON(http.StatusOK, domain.MessageResponse{Message: "logged out successfully"})
}

// LogoutAll handles logout from all devices
func (h *AuthHandler) LogoutAll(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, domain.ErrorResponse{Error: "unauthorized"})
		return
	}

	uid, ok := userID.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{Error: "invalid user ID"})
		return
	}

	if err := h.authService.LogoutAll(uid); err != nil {
		h.logger.Error("failed to logout from all devices", "user_id", uid, "error", err)
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{Error: "failed to logout from all devices"})
		return
	}

	h.clearAuthCookies(c)
	c.JSON(http.StatusOK, domain.MessageResponse{Message: "logged out from all devices successfully"})
}

// VerifyEmail handles email verification
func (h *AuthHandler) VerifyEmail(c *gin.Context) {
	var req domain.EmailVerificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleValidationError(c, err)
		return
	}

	if err := h.authService.VerifyEmail(&req); err != nil {
		h.handleAuthError(c, err)
		return
	}

	c.JSON(http.StatusOK, domain.MessageResponse{Message: "email verified successfully"})
}

// ForgotPassword handles forgot password requests
func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var req domain.ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleValidationError(c, err)
		return
	}

	if err := h.authService.ForgotPassword(&req); err != nil {
		h.logger.Error("forgot password error", "error", err)
		// Don't reveal specific errors for security
		c.JSON(http.StatusOK, domain.MessageResponse{
			Message: "if the email exists, a password reset link has been sent",
		})
		return
	}

	c.JSON(http.StatusOK, domain.MessageResponse{
		Message: "if the email exists, a password reset link has been sent",
	})
}

// ResetPassword handles password reset
func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req domain.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleValidationError(c, err)
		return
	}

	if err := h.authService.ResetPassword(&req); err != nil {
		h.handleAuthError(c, err)
		return
	}

	c.JSON(http.StatusOK, domain.MessageResponse{Message: "password reset successfully"})
}

// ChangePassword handles password change
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, domain.ErrorResponse{Error: "unauthorized"})
		return
	}

	uid, ok := userID.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{Error: "invalid user ID"})
		return
	}

	var req domain.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleValidationError(c, err)
		return
	}

	if err := h.authService.ChangePassword(uid, &req); err != nil {
		h.handleAuthError(c, err)
		return
	}

	c.JSON(http.StatusOK, domain.MessageResponse{Message: "password changed successfully"})
}

// GetProfile handles getting user profile
func (h *AuthHandler) GetProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, domain.ErrorResponse{Error: "unauthorized"})
		return
	}

	uid, ok := userID.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{Error: "invalid user ID"})
		return
	}

	profile, err := h.authService.GetUserProfile(uid)
	if err != nil {
		h.logger.Error("failed to get user profile", "user_id", uid, "error", err)
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{Error: "failed to get profile"})
		return
	}

	c.JSON(http.StatusOK, profile)
}

// ResendEmailVerification handles resending email verification
func (h *AuthHandler) ResendEmailVerification(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, domain.ErrorResponse{Error: "unauthorized"})
		return
	}

	uid, ok := userID.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{Error: "invalid user ID"})
		return
	}

	if err := h.authService.ResendEmailVerification(uid); err != nil {
		h.logger.Error("failed to resend email verification", "user_id", uid, "error", err)
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, domain.MessageResponse{Message: "verification email sent"})
}

// CheckAuth handles checking authentication status
func (h *AuthHandler) CheckAuth(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, domain.ErrorResponse{Error: "unauthorized"})
		return
	}

	uid, ok := userID.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{Error: "invalid user ID"})
		return
	}

	profile, err := h.authService.GetUserProfile(uid)
	if err != nil {
		h.logger.Error("failed to get user profile", "user_id", uid, "error", err)
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{Error: "failed to get profile"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"authenticated": true,
		"user":          profile,
	})
}

// Helper methods

func (h *AuthHandler) setAuthCookies(c *gin.Context, accessToken, refreshToken string) {
	// Set access token cookie (shorter expiry)
	c.SetCookie(
		"access_token",
		accessToken,
		int(h.config.JWTAccessTokenDurationParsed().Seconds()),
		"/",
		"",
		!h.config.IsDevelopment(), // secure in production
		true,                      // httpOnly
	)

	// Set refresh token cookie (longer expiry)
	c.SetCookie(
		"refresh_token",
		refreshToken,
		int(h.config.JWTRefreshTokenDurationParsed().Seconds()),
		"/",
		"",
		!h.config.IsDevelopment(), // secure in production
		true,                      // httpOnly
	)
}

func (h *AuthHandler) clearAuthCookies(c *gin.Context) {
	c.SetCookie("access_token", "", -1, "/", "", !h.config.IsDevelopment(), true)
	c.SetCookie("refresh_token", "", -1, "/", "", !h.config.IsDevelopment(), true)
}

func (h *AuthHandler) handleValidationError(c *gin.Context, err error) {
	h.logger.Warn("validation error", "error", err)
	c.JSON(http.StatusBadRequest, domain.ErrorResponse{
		Error: "validation failed",
		Details: map[string]string{
			"message": err.Error(),
		},
	})
}

func (h *AuthHandler) handleAuthError(c *gin.Context, err error) {
	switch err {
	case domain.ErrInvalidCredentials:
		c.JSON(http.StatusUnauthorized, domain.ErrorResponse{Error: "invalid credentials"})
	case domain.ErrUserAlreadyExists:
		c.JSON(http.StatusConflict, domain.ErrorResponse{Error: "user already exists"})
	case domain.ErrEmailNotVerified:
		c.JSON(http.StatusForbidden, domain.ErrorResponse{Error: "email not verified"})
	case domain.ErrUserInactive:
		c.JSON(http.StatusForbidden, domain.ErrorResponse{Error: "user account is inactive"})
	case domain.ErrInvalidToken, domain.ErrTokenNotFound:
		c.JSON(http.StatusUnauthorized, domain.ErrorResponse{Error: "invalid token"})
	case domain.ErrTokenExpired:
		c.JSON(http.StatusUnauthorized, domain.ErrorResponse{Error: "token expired"})
	case domain.ErrTokenAlreadyUsed:
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{Error: "token already used"})
	case domain.ErrPasswordsDoNotMatch:
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{Error: "passwords do not match"})
	case domain.ErrWeakPassword:
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{Error: "password is too weak"})
	case domain.ErrUnauthorized:
		c.JSON(http.StatusUnauthorized, domain.ErrorResponse{Error: "unauthorized"})
	case domain.ErrForbidden:
		c.JSON(http.StatusForbidden, domain.ErrorResponse{Error: "forbidden"})
	default:
		if strings.Contains(err.Error(), "too many") {
			c.JSON(http.StatusTooManyRequests, domain.ErrorResponse{Error: err.Error()})
		} else {
			h.logger.Error("auth service error", "error", err)
			c.JSON(http.StatusInternalServerError, domain.ErrorResponse{Error: "internal server error"})
		}
	}
}

// RegisterRoutes registers all authentication routes
func (h *AuthHandler) RegisterRoutes(router *gin.RouterGroup) {
	auth := router.Group("/auth")
	{
		auth.POST("/register", h.Register)
		auth.POST("/login", h.Login)
		auth.POST("/refresh", h.RefreshToken)
		auth.POST("/logout", h.Logout)
		auth.POST("/verify-email", h.VerifyEmail)
		auth.POST("/forgot-password", h.ForgotPassword)
		auth.POST("/reset-password", h.ResetPassword)
		auth.GET("/check", h.CheckAuth) // This will require auth middleware
	}

	// Protected routes (require authentication middleware)
	protected := router.Group("/auth")
	{
		protected.POST("/logout-all", h.LogoutAll)
		protected.POST("/change-password", h.ChangePassword)
		protected.GET("/profile", h.GetProfile)
		protected.POST("/resend-verification", h.ResendEmailVerification)
	}
}

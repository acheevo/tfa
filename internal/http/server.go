package http

import (
	"context"
	"log/slog"
	"net/http"
	"path/filepath"
	"time"

	admintransport "github.com/acheevo/tfa/internal/admin/transport"
	authtransport "github.com/acheevo/tfa/internal/auth/transport"
	healthtransport "github.com/acheevo/tfa/internal/health/transport"
	infotransport "github.com/acheevo/tfa/internal/info/transport"
	"github.com/acheevo/tfa/internal/middleware"
	"github.com/acheevo/tfa/internal/shared/config"
	usertransport "github.com/acheevo/tfa/internal/user/transport"
	"github.com/gin-gonic/gin"
)

type Server struct {
	config         *config.Config
	logger         *slog.Logger
	healthHandler  *healthtransport.HealthHandler
	infoHandler    *infotransport.InfoHandler
	authHandler    *authtransport.AuthHandler
	userHandler    *usertransport.UserHandler
	adminHandler   *admintransport.AdminHandler
	authMiddleware *middleware.AuthMiddleware
	rbacMiddleware *middleware.RBACMiddleware
	rateLimiter    *middleware.RateLimiter
	router         *gin.Engine
	server         *http.Server
}

func NewServer(
	config *config.Config,
	logger *slog.Logger,
	healthHandler *healthtransport.HealthHandler,
	infoHandler *infotransport.InfoHandler,
	authHandler *authtransport.AuthHandler,
	userHandler *usertransport.UserHandler,
	adminHandler *admintransport.AdminHandler,
	authMiddleware *middleware.AuthMiddleware,
	rbacMiddleware *middleware.RBACMiddleware,
	rateLimiter *middleware.RateLimiter,
) *Server {
	if !config.IsDevelopment() {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	s := &Server{
		config:         config,
		logger:         logger,
		healthHandler:  healthHandler,
		infoHandler:    infoHandler,
		authHandler:    authHandler,
		userHandler:    userHandler,
		adminHandler:   adminHandler,
		authMiddleware: authMiddleware,
		rbacMiddleware: rbacMiddleware,
		rateLimiter:    rateLimiter,
		router:         router,
	}

	s.setupMiddleware()
	s.setupRoutes()

	s.server = &http.Server{
		Addr:              ":" + config.Port,
		Handler:           router,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
	}

	return s
}

func (s *Server) setupMiddleware() {
	s.router.Use(middleware.Logger(s.logger))
	s.router.Use(middleware.Recovery(s.logger))
	s.router.Use(middleware.CORS())
}

func (s *Server) setupRoutes() {
	api := s.router.Group("/api")
	{
		// Health and info endpoints
		api.GET("/health", s.healthHandler.GetHealth)
		api.GET("/info", s.infoHandler.GetInfo)

		// Authentication routes with rate limiting
		authGroup := api.Group("/auth")
		authGroup.Use(s.rateLimiter.AuthRateLimit())

		// Login with specific rate limiting
		authGroup.POST("/login", s.authHandler.Login)

		// Other auth routes
		authGroup.POST("/register", s.authHandler.Register)
		authGroup.POST("/refresh", s.authHandler.RefreshToken)
		authGroup.POST("/logout", s.authHandler.Logout)
		authGroup.POST("/verify-email", s.authHandler.VerifyEmail)
		authGroup.POST("/forgot-password", s.authHandler.ForgotPassword)
		authGroup.POST("/reset-password", s.authHandler.ResetPassword)

		// Protected auth routes
		protectedAuth := authGroup.Group("/")
		protectedAuth.Use(s.authMiddleware.RequireAuth())
		{
			protectedAuth.GET("/check", s.authHandler.CheckAuth)
			protectedAuth.POST("/logout-all", s.authHandler.LogoutAll)
			protectedAuth.POST("/change-password", s.authHandler.ChangePassword)
			protectedAuth.GET("/profile", s.authHandler.GetProfile)
			protectedAuth.POST("/resend-verification", s.authHandler.ResendEmailVerification)
		}

		// User management routes (require authentication, active user, and profile permissions)
		userGroup := api.Group("/user")
		userGroup.Use(s.authMiddleware.RequireAuth(), s.authMiddleware.RequireActiveUser())
		{
			userGroup.GET("/profile", s.rbacMiddleware.RequirePermission("profile:read"), s.userHandler.GetProfile)
			userGroup.PUT("/profile", s.rbacMiddleware.RequirePermission("profile:update"), s.userHandler.UpdateProfile)
			userGroup.GET("/preferences", s.rbacMiddleware.RequirePermission("profile:read"), s.userHandler.GetPreferences)
			userGroup.PUT("/preferences", s.rbacMiddleware.RequirePermission("profile:update"), s.userHandler.UpdatePreferences)
			userGroup.POST("/change-email", s.rbacMiddleware.RequirePermission("profile:update"), s.userHandler.ChangeEmail)
			userGroup.GET("/dashboard", s.rbacMiddleware.RequirePermission("profile:read"), s.userHandler.GetDashboard)
		}

		// Admin routes (require authentication, active status, and specific permissions)
		adminGroup := api.Group("/admin")
		adminGroup.Use(
			s.authMiddleware.RequireAuth(),
			s.authMiddleware.RequireActiveUser(),
			s.rbacMiddleware.RequireAdminAccess(),
		)
		{
			// User management (require user management permissions)
			adminGroup.GET("/users", s.rbacMiddleware.RequireUserRead(), s.adminHandler.ListUsers)
			adminGroup.GET("/users/:id", s.rbacMiddleware.RequireUserRead(), s.adminHandler.GetUserDetails)
			adminGroup.PUT("/users/:id", s.rbacMiddleware.RequireUserManagement(), s.adminHandler.UpdateUser)
			adminGroup.PUT("/users/:id/role", s.rbacMiddleware.RequireUserManagement(), s.adminHandler.UpdateUserRole)
			adminGroup.PUT("/users/:id/status", s.rbacMiddleware.RequireUserManagement(), s.adminHandler.UpdateUserStatus)
			adminGroup.DELETE("/users", s.rbacMiddleware.RequirePermission("user:delete"), s.adminHandler.DeleteUsers)
			adminGroup.POST("/users/bulk", s.rbacMiddleware.RequireUserManagement(), s.adminHandler.BulkUpdateUsers)

			// Admin dashboard and monitoring
			adminGroup.GET("/stats", s.rbacMiddleware.RequirePermission("admin:read"), s.adminHandler.GetStats)
			adminGroup.GET("/audit-logs", s.rbacMiddleware.RequireAuditAccess(), s.adminHandler.GetAuditLogs)
		}
	}

	s.setupStaticRoutes()
}

func (s *Server) setupStaticRoutes() {
	frontendPath := "./frontend/dist"

	s.router.Static("/assets", filepath.Join(frontendPath, "assets"))

	s.router.StaticFile("/favicon.ico", filepath.Join(frontendPath, "favicon.ico"))

	s.router.NoRoute(func(c *gin.Context) {
		indexPath := filepath.Join(frontendPath, "index.html")
		c.File(indexPath)
	})
}

func (s *Server) Start() error {
	s.logger.Info("starting server", "port", s.config.Port)
	return s.server.ListenAndServe()
}

func (s *Server) Stop(ctx context.Context) error {
	s.logger.Info("stopping server")
	return s.server.Shutdown(ctx)
}

package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	adminservice "github.com/acheevo/tfa/internal/admin/service"
	admintransport "github.com/acheevo/tfa/internal/admin/transport"
	"github.com/acheevo/tfa/internal/auth/repository"
	authservice "github.com/acheevo/tfa/internal/auth/service"
	authtransport "github.com/acheevo/tfa/internal/auth/transport"
	"github.com/acheevo/tfa/internal/health/service"
	"github.com/acheevo/tfa/internal/health/transport"
	"github.com/acheevo/tfa/internal/http"
	infoservice "github.com/acheevo/tfa/internal/info/service"
	infotransport "github.com/acheevo/tfa/internal/info/transport"
	"github.com/acheevo/tfa/internal/middleware"
	"github.com/acheevo/tfa/internal/shared/bootstrap"
	"github.com/acheevo/tfa/internal/shared/config"
	"github.com/acheevo/tfa/internal/shared/database"
	"github.com/acheevo/tfa/internal/shared/logger"
	userrepository "github.com/acheevo/tfa/internal/user/repository"
	userservice "github.com/acheevo/tfa/internal/user/service"
	usertransport "github.com/acheevo/tfa/internal/user/transport"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	appLogger := logger.New(cfg.LogLevel, cfg.IsDevelopment())

	db, err := database.New(cfg.DatabaseDSN(), cfg.IsDevelopment(), appLogger, cfg.Environment)
	if err != nil {
		appLogger.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := db.Close(); err != nil {
			appLogger.Error("failed to close database connection", "error", err)
		}
	}()

	if err := db.SetConnectionPool(
		cfg.DBMaxIdleConns,
		cfg.DBMaxOpenConns,
		cfg.DBConnMaxLifetimeDuration(),
	); err != nil {
		appLogger.Error("failed to configure database connection pool", "error", err)
		return
	}

	// Bootstrap demo users and initial data
	bootstrapService := bootstrap.NewService(cfg, db.DB, appLogger)
	if err := bootstrapService.Bootstrap(); err != nil {
		appLogger.Error("bootstrap failed", "error", err)
		return
	}

	// Initialize repositories
	authUserRepo := repository.NewUserRepository(db.DB)
	refreshTokenRepo := repository.NewRefreshTokenRepository(db.DB)
	passwordResetRepo := repository.NewPasswordResetRepository(db.DB)
	userRepo := userrepository.NewUserRepository(db.DB)
	auditRepo := userrepository.NewAuditRepository(db.DB)

	// Initialize services
	jwtService := authservice.NewJWTService(cfg)
	emailService := authservice.NewEmailService(cfg, appLogger)
	authService := authservice.NewAuthService(
		cfg,
		appLogger,
		authUserRepo,
		refreshTokenRepo,
		passwordResetRepo,
		jwtService,
		emailService,
	)

	userSvc := userservice.NewUserService(
		cfg,
		appLogger,
		userRepo,
		auditRepo,
		authUserRepo,
	)

	adminSvc := adminservice.NewAdminService(
		cfg,
		appLogger,
		userRepo,
		auditRepo,
	)

	healthService := service.NewHealthService(cfg, db, appLogger)
	infoSvc := infoservice.NewInfoService(cfg, db, appLogger)

	// Initialize middleware
	authMiddleware := middleware.NewAuthMiddleware(appLogger, authService)
	rbacMiddleware := middleware.NewRBACMiddleware(appLogger, authService)
	rateLimiter := middleware.NewRateLimiter(appLogger, 10, time.Minute) // 10 requests per minute

	// Initialize handlers
	authHandler := authtransport.NewAuthHandler(cfg, appLogger, authService)
	userHandler := usertransport.NewUserHandler(cfg, appLogger, userSvc)
	adminHandler := admintransport.NewAdminHandler(cfg, appLogger, adminSvc)
	healthHandler := transport.NewHealthHandler(healthService)
	infoHandler := infotransport.NewInfoHandler(infoSvc)

	server := http.NewServer(
		cfg,
		appLogger,
		healthHandler,
		infoHandler,
		authHandler,
		userHandler,
		adminHandler,
		authMiddleware,
		rbacMiddleware,
		rateLimiter,
	)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := server.Start(); err != nil {
			appLogger.Error("server failed to start", "error", err)
			os.Exit(1)
		}
	}()

	appLogger.Info("server started successfully")

	<-quit
	appLogger.Info("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Stop(ctx); err != nil {
		appLogger.Error("server forced to shutdown", "error", err)
	} else {
		appLogger.Info("server exited gracefully")
	}
}

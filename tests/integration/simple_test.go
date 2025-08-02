//go:build integration

package integration

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"golang.org/x/crypto/bcrypt"

	authDomain "github.com/acheevo/tfa/internal/auth/domain"
	authRepo "github.com/acheevo/tfa/internal/auth/repository"
	authService "github.com/acheevo/tfa/internal/auth/service"
	authTransport "github.com/acheevo/tfa/internal/auth/transport"
	"github.com/acheevo/tfa/internal/shared/config"
	"github.com/acheevo/tfa/internal/shared/database"
)

func TestIntegration_SimpleAuth(t *testing.T) {
	ctx := context.Background()

	// Create PostgreSQL container
	postgresContainer, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("fullstack_template_test"),
		postgres.WithUsername("test"),
		postgres.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second)),
	)
	if err != nil {
		t.Fatalf("Failed to start postgres container: %v", err)
	}
	defer postgresContainer.Terminate(ctx)

	// Get connection details
	host, err := postgresContainer.Host(ctx)
	if err != nil {
		t.Fatalf("Failed to get container host: %v", err)
	}

	port, err := postgresContainer.MappedPort(ctx, "5432")
	if err != nil {
		t.Fatalf("Failed to get container port: %v", err)
	}

	// Build DSN
	dsn := fmt.Sprintf("postgres://test:test@%s:%s/fullstack_template_test?sslmode=disable", host, port.Port())

	// Create logger for testing
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelError,
	}))

	// Initialize database
	db, err := database.New(dsn, false, logger, "test")
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}
	defer db.Close()

	// Run migrations
	if err := db.AutoMigrate(); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	// Get underlying sql.DB
	sqlDB, err := db.DB.DB()
	if err != nil {
		t.Fatalf("Failed to get sql.DB: %v", err)
	}

	// Seed test data
	if err := seedSimpleTestData(sqlDB); err != nil {
		t.Fatalf("Failed to seed test data: %v", err)
	}

	// Create test config
	cfg := &config.Config{
		JWTSecret: "test-jwt-secret-key-for-testing-only-and-this-is-long-enough",
		SMTPHost:  "localhost",
		SMTPPort:  587,
		EmailFrom: "test@example.com",
	}

	// Initialize services
	userRepo := authRepo.NewUserRepository(db.DB)
	refreshTokenRepo := authRepo.NewRefreshTokenRepository(db.DB)
	passwordResetRepo := authRepo.NewPasswordResetRepository(db.DB)
	jwtSvc := authService.NewJWTService(cfg)
	emailSvc := authService.NewEmailService(cfg, logger)
	authSvc := authService.NewAuthService(cfg, logger, userRepo, refreshTokenRepo, passwordResetRepo, jwtSvc, emailSvc)

	// Initialize handler
	authHandler := authTransport.NewAuthHandler(cfg, logger, authSvc)

	// Set Gin mode for testing
	gin.SetMode(gin.TestMode)
	router := gin.Default()

	// Setup auth routes
	api := router.Group("/api")
	auth := api.Group("/auth")
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
	}

	t.Run("Login_Success", func(t *testing.T) {
		loginReq := authDomain.LoginRequest{
			Email:    "admin@fullstack.dev",
			Password: "password",
		}

		body, _ := json.Marshal(loginReq)
		req := httptest.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d. Body: %s", http.StatusOK, w.Code, w.Body.String())
			return
		}

		var response authDomain.AuthResponse
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if response.AccessToken == "" {
			t.Error("Expected access token in response")
		}

		if response.User.Email != "admin@fullstack.dev" {
			t.Errorf("Expected user email admin@fullstack.dev, got %s", response.User.Email)
		}
	})

	t.Run("Register_Success", func(t *testing.T) {
		registerReq := authDomain.RegisterRequest{
			Email:     "newuser@fullstack.dev",
			Password:  "newpassword123",
			FirstName: "New",
			LastName:  "User",
		}

		body, _ := json.Marshal(registerReq)
		req := httptest.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("expected status %d, got %d. Body: %s", http.StatusCreated, w.Code, w.Body.String())
			return
		}

		var response authDomain.AuthResponse
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if response.AccessToken == "" {
			t.Error("Expected access token in response")
		}

		if response.User.Email != "newuser@fullstack.dev" {
			t.Errorf("Expected user email newuser@fullstack.dev, got %s", response.User.Email)
		}
	})
}

func seedSimpleTestData(db *sql.DB) error {
	// Generate proper bcrypt hash for "password"
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Create test admin user
	insertAdminUser := `
	INSERT INTO users (email, password_hash, first_name, last_name, role, status, email_verified, created_at, updated_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`

	_, err = db.Exec(insertAdminUser,
		"admin@fullstack.dev",
		string(hashedPassword),
		"Admin",
		"User",
		string(authDomain.RoleAdmin),
		string(authDomain.StatusActive),
		true)
	if err != nil {
		return fmt.Errorf("failed to seed admin user: %w", err)
	}

	return nil
}

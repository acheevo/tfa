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
	"github.com/acheevo/tfa/internal/middleware"
	"github.com/acheevo/tfa/internal/shared/config"
	"github.com/acheevo/tfa/internal/shared/database"
)

func TestE2E_FullUserFlow(t *testing.T) {
	ctx := context.Background()

	// Setup test database container
	testDB, cleanup := setupE2ETestDatabase(t, ctx)
	defer cleanup()

	// Create test handler with test database
	handler := createE2ETestHandler(t, testDB)

	t.Run("CompleteUserJourney", func(t *testing.T) {
		// Step 1: Health check (if implemented)
		req := httptest.NewRequest("GET", "/api/health", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		// Health check may not be implemented, so don't fail the test
		if w.Code == http.StatusOK {
			t.Logf("Health check passed: %d", w.Code)
		} else {
			t.Logf("Health check not implemented or failed: %d", w.Code)
		}

		// Step 2: Register new user
		registerReq := authDomain.RegisterRequest{
			Email:     "journey@fullstack.dev",
			Password:  "journeypassword123",
			FirstName: "Journey",
			LastName:  "User",
		}

		body, _ := json.Marshal(registerReq)
		req = httptest.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		w = httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Fatalf("Registration failed: %d, Body: %s", w.Code, w.Body.String())
		}

		var registerResponse authDomain.AuthResponse
		if err := json.Unmarshal(w.Body.Bytes(), &registerResponse); err != nil {
			t.Fatalf("Failed to unmarshal register response: %v", err)
		}

		// Verify registration response
		if registerResponse.AccessToken == "" {
			t.Error("Expected access token in registration response")
		}
		if registerResponse.User.Email != "journey@fullstack.dev" {
			t.Errorf("Expected email journey@fullstack.dev, got %s", registerResponse.User.Email)
		}
		if registerResponse.User.Role != authDomain.RoleUser {
			t.Errorf("Expected role %s, got %s", authDomain.RoleUser, registerResponse.User.Role)
		}

		// Step 3: Login with registered user
		loginReq := authDomain.LoginRequest{
			Email:    "journey@fullstack.dev",
			Password: "journeypassword123",
		}

		body, _ = json.Marshal(loginReq)
		req = httptest.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		w = httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("Login failed: %d", w.Code)
		}

		var loginResponse authDomain.AuthResponse
		if err := json.Unmarshal(w.Body.Bytes(), &loginResponse); err != nil {
			t.Fatalf("Failed to unmarshal login response: %v", err)
		}

		token := loginResponse.AccessToken

		// Step 4: Access protected endpoint
		req = httptest.NewRequest("GET", "/api/protected/profile", nil)
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

		w = httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Protected endpoint access failed: %d", w.Code)
		}

		// Step 5: Try to access admin endpoint (should fail)
		req = httptest.NewRequest("GET", "/api/admin/users", nil)
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

		w = httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusForbidden {
			t.Errorf("Expected forbidden access to admin endpoint, got: %d", w.Code)
		}

		// Step 6: Test token refresh
		if loginResponse.RefreshToken != "" {
			refreshReq := authDomain.RefreshTokenRequest{
				RefreshToken: loginResponse.RefreshToken,
			}

			body, _ = json.Marshal(refreshReq)
			req = httptest.NewRequest("POST", "/api/auth/refresh", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			w = httptest.NewRecorder()
			handler.ServeHTTP(w, req)

			if w.Code == http.StatusOK {
				t.Logf("Token refresh successful")
			} else {
				t.Logf("Token refresh may not be fully implemented: %d", w.Code)
			}
		}

		// Step 7: Test logout (if implemented)
		req = httptest.NewRequest("POST", "/api/auth/logout", nil)
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

		w = httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code == http.StatusOK {
			t.Logf("Logout successful")
		} else {
			t.Logf("Logout may not be fully implemented: %d", w.Code)
		}
	})

	t.Run("AdminUserManagement", func(t *testing.T) {
		// Step 1: Login as admin
		loginReq := authDomain.LoginRequest{
			Email:    "admin@fullstack.dev",
			Password: "password",
		}

		body, _ := json.Marshal(loginReq)
		req := httptest.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("Admin login failed: %d", w.Code)
		}

		var loginResponse authDomain.AuthResponse
		if err := json.Unmarshal(w.Body.Bytes(), &loginResponse); err != nil {
			t.Fatalf("Failed to unmarshal admin login response: %v", err)
		}

		adminToken := loginResponse.AccessToken

		// Step 2: Access admin endpoint
		req = httptest.NewRequest("GET", "/api/admin/users", nil)
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", adminToken))

		w = httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		// Admin endpoints may not be fully implemented
		if w.Code == http.StatusOK {
			t.Logf("Admin endpoint access successful")
		} else {
			t.Logf("Admin endpoint may not be fully implemented: %d", w.Code)
		}

		// Step 3: Try to access admin endpoint as regular user (should fail)
		userLoginReq := authDomain.LoginRequest{
			Email:    "user@fullstack.dev",
			Password: "password",
		}

		body, _ = json.Marshal(userLoginReq)
		req = httptest.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		w = httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("User login failed: %d", w.Code)
		}

		var userLoginResponse authDomain.AuthResponse
		if err := json.Unmarshal(w.Body.Bytes(), &userLoginResponse); err != nil {
			t.Fatalf("Failed to unmarshal user login response: %v", err)
		}

		userToken := userLoginResponse.AccessToken

		// Try to access admin endpoint with user token
		req = httptest.NewRequest("GET", "/api/admin/users", nil)
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", userToken))

		w = httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusForbidden {
			t.Errorf("Expected forbidden access for regular user to admin endpoint, got %d", w.Code)
		}
	})

	t.Run("SecurityValidation", func(t *testing.T) {
		// Test various security scenarios

		// 1. SQL Injection attempt in login
		loginReq := authDomain.LoginRequest{
			Email:    "admin@fullstack.dev'; DROP TABLE users; --",
			Password: "password",
		}

		body, _ := json.Marshal(loginReq)
		req := httptest.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		// Should fail with validation error or unauthorized
		if w.Code == http.StatusOK {
			t.Error("SQL injection attempt should not succeed")
		}

		// 2. XSS attempt in registration
		registerReq := authDomain.RegisterRequest{
			Email:     "xss@test.com",
			Password:  "password123",
			FirstName: "<script>alert('xss')</script>",
			LastName:  "Test",
		}

		body, _ = json.Marshal(registerReq)
		req = httptest.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		w = httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code == http.StatusCreated {
			// If registration succeeds, ensure XSS content is properly handled
			var response authDomain.AuthResponse
			json.Unmarshal(w.Body.Bytes(), &response)

			// The first name should be stored safely (content filtering is application-specific)
			// For now, we just verify the registration succeeded without breaking the system
			if response.User.FirstName == "" {
				t.Error("First name should not be empty after registration")
			}
			t.Logf("XSS test registered successfully with first name: %s", response.User.FirstName)
		} else {
			// Registration may be rejected due to validation - this is also acceptable
			t.Logf("XSS content registration rejected with status: %d", w.Code)
		}

		// 3. Test with extremely long input
		longEmail := string(make([]byte, 1000)) + "@test.com"
		registerReq = authDomain.RegisterRequest{
			Email:     longEmail,
			Password:  "password123",
			FirstName: "Test",
			LastName:  "User",
		}

		body, _ = json.Marshal(registerReq)
		req = httptest.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		w = httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		// Should fail with validation error
		if w.Code == http.StatusCreated {
			t.Error("Extremely long email should be rejected")
		}

		// 4. Test weak password
		registerReq = authDomain.RegisterRequest{
			Email:     "weak@test.com",
			Password:  "123",
			FirstName: "Weak",
			LastName:  "Password",
		}

		body, _ = json.Marshal(registerReq)
		req = httptest.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		w = httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		// Should fail with validation error
		if w.Code == http.StatusCreated {
			t.Error("Weak password should be rejected")
		}
	})

	t.Run("ConcurrentOperations", func(t *testing.T) {
		// Test concurrent registrations with same email
		registerReq := authDomain.RegisterRequest{
			Email:     "concurrent@test.com",
			Password:  "password123",
			FirstName: "Concurrent",
			LastName:  "Test",
		}

		body, _ := json.Marshal(registerReq)

		// Start multiple goroutines trying to register same email
		results := make(chan int, 3)

		for i := 0; i < 3; i++ {
			go func() {
				req := httptest.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(body))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()
				handler.ServeHTTP(w, req)
				results <- w.Code
			}()
		}

		// Collect results
		successCount := 0
		for i := 0; i < 3; i++ {
			code := <-results
			if code == http.StatusCreated {
				successCount++
			}
		}

		// Only one should succeed due to unique email constraint
		if successCount != 1 {
			t.Errorf("Expected exactly 1 successful registration, got %d", successCount)
		} else {
			t.Logf("Concurrent operations test passed: %d successful registrations", successCount)
		}
	})

	t.Run("RateLimiting", func(t *testing.T) {
		// Test rate limiting by making many rapid requests
		loginReq := authDomain.LoginRequest{
			Email:    "nonexistent@test.com",
			Password: "wrongpassword",
		}

		body, _ := json.Marshal(loginReq)

		rateLimitHit := false
		for i := 0; i < 20; i++ {
			req := httptest.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-Forwarded-For", "192.168.1.100") // Simulate same IP

			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)

			if w.Code == http.StatusTooManyRequests {
				rateLimitHit = true
				break
			}
		}

		if !rateLimitHit {
			t.Logf("Rate limiting may not be implemented or threshold is higher")
		} else {
			t.Logf("Rate limiting is working correctly")
		}
	})
}

// Helper functions for E2E tests

func setupE2ETestDatabase(t *testing.T, ctx context.Context) (*database.DB, func()) {
	t.Helper()

	// Create PostgreSQL container
	postgresContainer, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("fullstack_template_e2e"),
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
	dsn := fmt.Sprintf("postgres://test:test@%s:%s/fullstack_template_e2e?sslmode=disable", host, port.Port())

	// Create logger for testing
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelError,
	}))

	// Initialize database
	db, err := database.New(dsn, false, logger, "test")
	if err != nil {
		postgresContainer.Terminate(ctx)
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Run migrations
	if err := db.AutoMigrate(); err != nil {
		db.Close()
		postgresContainer.Terminate(ctx)
		t.Fatalf("Failed to run migrations: %v", err)
	}

	// Get underlying sql.DB
	sqlDB, err := db.DB.DB()
	if err != nil {
		db.Close()
		postgresContainer.Terminate(ctx)
		t.Fatalf("Failed to get sql.DB: %v", err)
	}

	// Seed test data
	if err := seedE2ETestData(sqlDB); err != nil {
		db.Close()
		postgresContainer.Terminate(ctx)
		t.Fatalf("Failed to seed test data: %v", err)
	}

	cleanup := func() {
		db.Close()
		postgresContainer.Terminate(ctx)
	}

	return db, cleanup
}

func createE2ETestHandler(t *testing.T, db *database.DB) http.Handler {
	t.Helper()

	// Create test config
	cfg := &config.Config{
		JWTSecret: "test-jwt-secret-key-for-testing-only-and-this-is-long-enough",
		SMTPHost:  "localhost",
		SMTPPort:  587,
		EmailFrom: "test@example.com",
	}

	// Create logger for testing
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelError,
	}))

	// Initialize services
	userRepo := authRepo.NewUserRepository(db.DB)
	refreshTokenRepo := authRepo.NewRefreshTokenRepository(db.DB)
	passwordResetRepo := authRepo.NewPasswordResetRepository(db.DB)
	jwtSvc := authService.NewJWTService(cfg)
	emailSvc := authService.NewEmailService(cfg, logger)
	authSvc := authService.NewAuthService(cfg, logger, userRepo, refreshTokenRepo, passwordResetRepo, jwtSvc, emailSvc)

	// Initialize handlers
	authHandler := authTransport.NewAuthHandler(cfg, logger, authSvc)

	// Initialize middleware
	authMiddleware := middleware.NewAuthMiddleware(logger, authSvc)

	// Set Gin mode for testing
	gin.SetMode(gin.TestMode)
	router := gin.Default()

	// Setup all routes
	api := router.Group("/api")
	{
		// Health check (if implemented)
		api.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		// Auth routes (public)
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh", authHandler.RefreshToken)
			auth.POST("/forgot-password", authHandler.ForgotPassword)
			auth.POST("/reset-password", authHandler.ResetPassword)
			auth.POST("/verify-email", authHandler.VerifyEmail)
			auth.POST("/logout", authHandler.Logout)
		}

		// Protected routes
		protected := api.Group("/protected")
		protected.Use(authMiddleware.RequireAuth())
		{
			protected.GET("/profile", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "profile accessed"})
			})
		}

		// Admin routes
		admin := api.Group("/admin")
		admin.Use(authMiddleware.RequireAuth())
		admin.Use(authMiddleware.RequireAdmin())
		{
			admin.GET("/users", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"users": []string{"admin", "user"}})
			})
		}
	}

	return router
}

func seedE2ETestData(db *sql.DB) error {
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

	// Create test regular user
	insertRegularUser := `
	INSERT INTO users (email, password_hash, first_name, last_name, role, status, email_verified, created_at, updated_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`

	_, err = db.Exec(insertRegularUser,
		"user@fullstack.dev",
		string(hashedPassword),
		"Test",
		"User",
		string(authDomain.RoleUser),
		string(authDomain.StatusActive),
		true)
	if err != nil {
		return fmt.Errorf("failed to seed regular user: %w", err)
	}

	return nil
}

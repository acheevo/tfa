package service

import (
	"fmt"
	"log/slog"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/acheevo/tfa/internal/auth/domain"
	"github.com/acheevo/tfa/internal/auth/repository"
	"github.com/acheevo/tfa/internal/shared/config"
)

// AuthService handles authentication operations
type AuthService struct {
	config            *config.Config
	logger            *slog.Logger
	userRepo          *repository.UserRepository
	refreshTokenRepo  *repository.RefreshTokenRepository
	passwordResetRepo *repository.PasswordResetRepository
	jwtService        *JWTService
	emailService      *EmailService
}

// NewAuthService creates a new authentication service
func NewAuthService(
	config *config.Config,
	logger *slog.Logger,
	userRepo *repository.UserRepository,
	refreshTokenRepo *repository.RefreshTokenRepository,
	passwordResetRepo *repository.PasswordResetRepository,
	jwtService *JWTService,
	emailService *EmailService,
) *AuthService {
	return &AuthService{
		config:            config,
		logger:            logger,
		userRepo:          userRepo,
		refreshTokenRepo:  refreshTokenRepo,
		passwordResetRepo: passwordResetRepo,
		jwtService:        jwtService,
		emailService:      emailService,
	}
}

// Register registers a new user
func (s *AuthService) Register(req *domain.RegisterRequest) (*domain.AuthResponse, error) {
	// Check if user already exists
	exists, err := s.userRepo.ExistsByEmail(req.Email)
	if err != nil {
		s.logger.Error("failed to check if user exists", "email", req.Email, "error", err)
		return nil, fmt.Errorf("failed to check user existence: %w", err)
	}
	if exists {
		return nil, domain.ErrUserAlreadyExists
	}

	// Validate password strength
	if err := s.validatePassword(req.Password); err != nil {
		return nil, err
	}

	// Hash password
	passwordHash, err := s.hashPassword(req.Password)
	if err != nil {
		s.logger.Error("failed to hash password", "error", err)
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Generate email verification token
	emailVerifyToken, err := s.jwtService.GenerateRandomToken()
	if err != nil {
		s.logger.Error("failed to generate email verification token", "error", err)
		return nil, fmt.Errorf("failed to generate email verification token: %w", err)
	}

	// Create user
	user := &domain.User{
		Email:            strings.ToLower(strings.TrimSpace(req.Email)),
		PasswordHash:     passwordHash,
		FirstName:        strings.TrimSpace(req.FirstName),
		LastName:         strings.TrimSpace(req.LastName),
		EmailVerified:    false,
		EmailVerifyToken: emailVerifyToken,
		Status:           domain.StatusActive,
	}

	if err := s.userRepo.Create(user); err != nil {
		s.logger.Error("failed to create user", "email", req.Email, "error", err)
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Send email verification email
	if err := s.emailService.SendEmailVerification(user.Email, emailVerifyToken, user.FirstName); err != nil {
		s.logger.Error("failed to send email verification", "email", user.Email, "error", err)
		// Don't fail registration if email fails to send
	}

	// Generate tokens
	accessToken, err := s.jwtService.GenerateAccessToken(user)
	if err != nil {
		s.logger.Error("failed to generate access token", "user_id", user.ID, "error", err)
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := s.createRefreshToken(user.ID)
	if err != nil {
		s.logger.Error("failed to create refresh token", "user_id", user.ID, "error", err)
		return nil, fmt.Errorf("failed to create refresh token: %w", err)
	}

	s.logger.Info("user registered successfully", "user_id", user.ID, "email", user.Email)

	return &domain.AuthResponse{
		User:         user.ToResponse(),
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(s.jwtService.GetAccessTokenDuration().Seconds()),
	}, nil
}

// Login authenticates a user and returns tokens
func (s *AuthService) Login(req *domain.LoginRequest) (*domain.AuthResponse, error) {
	// Get user by email
	user, err := s.userRepo.GetByEmail(strings.ToLower(strings.TrimSpace(req.Email)))
	if err != nil {
		if err == domain.ErrUserNotFound {
			return nil, domain.ErrInvalidCredentials
		}
		s.logger.Error("failed to get user by email", "email", req.Email, "error", err)
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Check if user is active
	if !user.IsActive() {
		return nil, domain.ErrUserInactive
	}

	// Verify password
	if err := s.verifyPassword(req.Password, user.PasswordHash); err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	// Update last login time
	if err := s.userRepo.UpdateLastLogin(user.ID); err != nil {
		s.logger.Error("failed to update last login", "user_id", user.ID, "error", err)
		// Don't fail login if this fails
	}

	// Generate tokens
	accessToken, err := s.jwtService.GenerateAccessToken(user)
	if err != nil {
		s.logger.Error("failed to generate access token", "user_id", user.ID, "error", err)
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := s.createRefreshToken(user.ID)
	if err != nil {
		s.logger.Error("failed to create refresh token", "user_id", user.ID, "error", err)
		return nil, fmt.Errorf("failed to create refresh token: %w", err)
	}

	s.logger.Info("user logged in successfully", "user_id", user.ID, "email", user.Email)

	return &domain.AuthResponse{
		User:         user.ToResponse(),
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(s.jwtService.GetAccessTokenDuration().Seconds()),
	}, nil
}

// RefreshToken refreshes an access token using a refresh token
func (s *AuthService) RefreshToken(req *domain.RefreshTokenRequest) (*domain.AuthResponse, error) {
	// Get refresh token from database
	refreshToken, err := s.refreshTokenRepo.GetByToken(req.RefreshToken)
	if err != nil {
		return nil, domain.ErrInvalidToken
	}

	// Check if token is expired
	if refreshToken.IsExpired() {
		// Clean up expired token
		_ = s.refreshTokenRepo.Delete(refreshToken.Token)
		return nil, domain.ErrTokenExpired
	}

	// Get user
	user, err := s.userRepo.GetByID(refreshToken.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Check if user is active
	if !user.IsActive() {
		return nil, domain.ErrUserInactive
	}

	// Generate new access token
	accessToken, err := s.jwtService.GenerateAccessToken(user)
	if err != nil {
		s.logger.Error("failed to generate access token", "user_id", user.ID, "error", err)
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	s.logger.Info("token refreshed successfully", "user_id", user.ID)

	return &domain.AuthResponse{
		User:         user.ToResponse(),
		AccessToken:  accessToken,
		RefreshToken: refreshToken.Token, // Return the same refresh token
		ExpiresIn:    int64(s.jwtService.GetAccessTokenDuration().Seconds()),
	}, nil
}

// Logout invalidates a refresh token
func (s *AuthService) Logout(refreshToken string) error {
	if err := s.refreshTokenRepo.Delete(refreshToken); err != nil {
		s.logger.Error("failed to delete refresh token", "error", err)
		return fmt.Errorf("failed to logout: %w", err)
	}

	s.logger.Info("user logged out successfully")
	return nil
}

// LogoutAll invalidates all refresh tokens for a user
func (s *AuthService) LogoutAll(userID uint) error {
	if err := s.refreshTokenRepo.DeleteByUserID(userID); err != nil {
		s.logger.Error("failed to delete all refresh tokens", "user_id", userID, "error", err)
		return fmt.Errorf("failed to logout from all devices: %w", err)
	}

	s.logger.Info("user logged out from all devices", "user_id", userID)
	return nil
}

// VerifyEmail verifies a user's email address
func (s *AuthService) VerifyEmail(req *domain.EmailVerificationRequest) error {
	// Get user by email verification token
	user, err := s.userRepo.GetByEmailVerifyToken(req.Token)
	if err != nil {
		return domain.ErrInvalidToken
	}

	// Mark email as verified and clear token
	user.EmailVerified = true
	user.EmailVerifyToken = ""

	if err := s.userRepo.Update(user); err != nil {
		s.logger.Error("failed to update user email verification", "user_id", user.ID, "error", err)
		return fmt.Errorf("failed to verify email: %w", err)
	}

	// Send welcome email
	if err := s.emailService.SendWelcomeEmail(user.Email, user.FirstName); err != nil {
		s.logger.Error("failed to send welcome email", "email", user.Email, "error", err)
		// Don't fail verification if welcome email fails
	}

	s.logger.Info("email verified successfully", "user_id", user.ID, "email", user.Email)
	return nil
}

// ForgotPassword initiates password reset process
func (s *AuthService) ForgotPassword(req *domain.ForgotPasswordRequest) error {
	email := strings.ToLower(strings.TrimSpace(req.Email))

	// Check if user exists
	user, err := s.userRepo.GetByEmail(email)
	if err != nil {
		if err == domain.ErrUserNotFound {
			// Don't reveal if user exists or not for security
			s.logger.Info("password reset requested for non-existent email", "email", email)
			return nil
		}
		s.logger.Error("failed to get user by email", "email", email, "error", err)
		return fmt.Errorf("failed to process password reset request: %w", err)
	}

	// Check rate limiting - don't allow too many reset requests
	count, err := s.passwordResetRepo.GetValidTokensCount(email)
	if err != nil {
		s.logger.Error("failed to get valid tokens count", "email", email, "error", err)
		return fmt.Errorf("failed to process password reset request: %w", err)
	}
	if count >= 3 {
		s.logger.Warn("too many password reset requests", "email", email, "count", count)
		return fmt.Errorf("too many password reset requests, please try again later")
	}

	// Generate reset token
	token, err := s.jwtService.GenerateRandomToken()
	if err != nil {
		s.logger.Error("failed to generate reset token", "error", err)
		return fmt.Errorf("failed to generate reset token: %w", err)
	}

	// Create password reset record
	reset := &domain.PasswordReset{
		Email:     email,
		Token:     token,
		ExpiresAt: time.Now().Add(24 * time.Hour), // 24 hours expiry
		Used:      false,
	}

	if err := s.passwordResetRepo.Create(reset); err != nil {
		s.logger.Error("failed to create password reset", "email", email, "error", err)
		return fmt.Errorf("failed to create password reset: %w", err)
	}

	// Send password reset email
	if err := s.emailService.SendPasswordReset(email, token, user.FirstName); err != nil {
		s.logger.Error("failed to send password reset email", "email", email, "error", err)
		return fmt.Errorf("failed to send password reset email: %w", err)
	}

	s.logger.Info("password reset requested", "email", email)
	return nil
}

// ResetPassword resets a user's password using a reset token
func (s *AuthService) ResetPassword(req *domain.ResetPasswordRequest) error {
	// Validate passwords match
	if req.Password != req.ConfirmPassword {
		return domain.ErrPasswordsDoNotMatch
	}

	// Validate password strength
	if err := s.validatePassword(req.Password); err != nil {
		return err
	}

	// Get password reset token
	reset, err := s.passwordResetRepo.GetByToken(req.Token)
	if err != nil {
		return domain.ErrInvalidToken
	}

	// Check if token is expired or used
	if reset.IsExpired() {
		return domain.ErrTokenExpired
	}
	if reset.Used {
		return domain.ErrTokenAlreadyUsed
	}

	// Get user
	user, err := s.userRepo.GetByEmail(reset.Email)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	// Hash new password
	passwordHash, err := s.hashPassword(req.Password)
	if err != nil {
		s.logger.Error("failed to hash password", "error", err)
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Update user password
	user.PasswordHash = passwordHash
	if err := s.userRepo.Update(user); err != nil {
		s.logger.Error("failed to update user password", "user_id", user.ID, "error", err)
		return fmt.Errorf("failed to update password: %w", err)
	}

	// Mark token as used
	if err := s.passwordResetRepo.MarkAsUsed(req.Token); err != nil {
		s.logger.Error("failed to mark reset token as used", "token", req.Token, "error", err)
		// Don't fail if this fails
	}

	// Invalidate all refresh tokens to force re-login
	if err := s.refreshTokenRepo.DeleteByUserID(user.ID); err != nil {
		s.logger.Error("failed to invalidate refresh tokens", "user_id", user.ID, "error", err)
		// Don't fail if this fails
	}

	s.logger.Info("password reset successfully", "user_id", user.ID, "email", user.Email)
	return nil
}

// ChangePassword changes a user's password
func (s *AuthService) ChangePassword(userID uint, req *domain.ChangePasswordRequest) error {
	// Validate passwords match
	if req.NewPassword != req.ConfirmPassword {
		return domain.ErrPasswordsDoNotMatch
	}

	// Validate password strength
	if err := s.validatePassword(req.NewPassword); err != nil {
		return err
	}

	// Get user
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	// Verify current password
	if err := s.verifyPassword(req.CurrentPassword, user.PasswordHash); err != nil {
		return domain.ErrInvalidCredentials
	}

	// Hash new password
	passwordHash, err := s.hashPassword(req.NewPassword)
	if err != nil {
		s.logger.Error("failed to hash password", "error", err)
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Update user password
	user.PasswordHash = passwordHash
	if err := s.userRepo.Update(user); err != nil {
		s.logger.Error("failed to update user password", "user_id", user.ID, "error", err)
		return fmt.Errorf("failed to update password: %w", err)
	}

	s.logger.Info("password changed successfully", "user_id", user.ID)
	return nil
}

// GetUserProfile gets a user's profile
func (s *AuthService) GetUserProfile(userID uint) (*domain.UserResponse, error) {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user.ToResponse(), nil
}

// ValidateAccessToken validates an access token and returns user claims
func (s *AuthService) ValidateAccessToken(tokenString string) (*domain.JWTClaims, error) {
	return s.jwtService.ValidateAccessToken(tokenString)
}

// ResendEmailVerification resends email verification email
func (s *AuthService) ResendEmailVerification(userID uint) error {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	if user.EmailVerified {
		return fmt.Errorf("email already verified")
	}

	// Generate new verification token if empty
	if user.EmailVerifyToken == "" {
		token, err := s.jwtService.GenerateRandomToken()
		if err != nil {
			s.logger.Error("failed to generate email verification token", "error", err)
			return fmt.Errorf("failed to generate email verification token: %w", err)
		}
		user.EmailVerifyToken = token
		if err := s.userRepo.Update(user); err != nil {
			s.logger.Error("failed to update user email verification token", "user_id", user.ID, "error", err)
			return fmt.Errorf("failed to update user: %w", err)
		}
	}

	// Send email verification
	if err := s.emailService.SendEmailVerification(user.Email, user.EmailVerifyToken, user.FirstName); err != nil {
		s.logger.Error("failed to send email verification", "email", user.Email, "error", err)
		return fmt.Errorf("failed to send email verification: %w", err)
	}

	s.logger.Info("email verification resent", "user_id", user.ID, "email", user.Email)
	return nil
}

// Helper methods

func (s *AuthService) createRefreshToken(userID uint) (string, error) {
	// Generate refresh token
	tokenStr, err := s.jwtService.GenerateRefreshToken()
	if err != nil {
		return "", err
	}

	// Create refresh token record
	refreshToken := &domain.RefreshToken{
		UserID:    userID,
		Token:     tokenStr,
		ExpiresAt: time.Now().Add(s.jwtService.GetRefreshTokenDuration()),
	}

	if err := s.refreshTokenRepo.Create(refreshToken); err != nil {
		return "", err
	}

	// Clean up old tokens (keep max 5 per user)
	if err := s.refreshTokenRepo.DeleteOldestTokensForUser(userID, 5); err != nil {
		s.logger.Error("failed to clean up old refresh tokens", "user_id", userID, "error", err)
		// Don't fail if cleanup fails
	}

	return tokenStr, nil
}

func (s *AuthService) hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func (s *AuthService) verifyPassword(password, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

func (s *AuthService) validatePassword(password string) error {
	if len(password) < 8 {
		return domain.ErrWeakPassword
	}
	// Add more password strength validation as needed
	return nil
}

// CleanupExpiredTokens removes expired tokens from the database
func (s *AuthService) CleanupExpiredTokens() error {
	if err := s.refreshTokenRepo.DeleteExpired(); err != nil {
		s.logger.Error("failed to cleanup expired refresh tokens", "error", err)
		return err
	}

	if err := s.passwordResetRepo.DeleteExpired(); err != nil {
		s.logger.Error("failed to cleanup expired password reset tokens", "error", err)
		return err
	}

	if err := s.passwordResetRepo.DeleteUsed(); err != nil {
		s.logger.Error("failed to cleanup used password reset tokens", "error", err)
		return err
	}

	s.logger.Info("expired tokens cleaned up successfully")
	return nil
}

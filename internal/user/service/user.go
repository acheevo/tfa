package service

import (
	"fmt"
	"log/slog"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	authdomain "github.com/acheevo/tfa/internal/auth/domain"
	authrepo "github.com/acheevo/tfa/internal/auth/repository"
	"github.com/acheevo/tfa/internal/shared/config"
	"github.com/acheevo/tfa/internal/user/domain"
	"github.com/acheevo/tfa/internal/user/repository"
)

// UserService handles user management operations
type UserService struct {
	config       *config.Config
	logger       *slog.Logger
	userRepo     *repository.UserRepository
	auditRepo    *repository.AuditRepository
	authUserRepo *authrepo.UserRepository
}

// NewUserService creates a new user service
func NewUserService(
	config *config.Config,
	logger *slog.Logger,
	userRepo *repository.UserRepository,
	auditRepo *repository.AuditRepository,
	authUserRepo *authrepo.UserRepository,
) *UserService {
	return &UserService{
		config:       config,
		logger:       logger,
		userRepo:     userRepo,
		auditRepo:    auditRepo,
		authUserRepo: authUserRepo,
	}
}

// GetProfile retrieves a user's profile
func (s *UserService) GetProfile(userID uint) (*authdomain.UserResponse, error) {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		s.logger.Error("failed to get user profile", "user_id", userID, "error", err)
		return nil, err
	}

	return user.ToResponse(), nil
}

// UpdateProfile updates a user's profile information
func (s *UserService) UpdateProfile(
	userID uint,
	req *domain.UpdateProfileRequest,
	ipAddress, userAgent string,
) (*authdomain.UserResponse, error) {
	// Get current user to compare changes
	currentUser, err := s.userRepo.GetByID(userID)
	if err != nil {
		s.logger.Error("failed to get user for profile update", "user_id", userID, "error", err)
		return nil, err
	}

	// Update profile
	err = s.userRepo.UpdateProfile(userID, req)
	if err != nil {
		s.logger.Error("failed to update user profile", "user_id", userID, "error", err)
		return nil, domain.ErrProfileUpdateFailed
	}

	// Create audit log
	changes := s.buildProfileChanges(currentUser, req)
	if err := s.auditRepo.CreateAuditEntry(
		&userID,
		&userID,
		authdomain.AuditActionUserUpdated,
		authdomain.AuditLevelInfo,
		"user",
		fmt.Sprintf("Profile updated: %s", changes),
		ipAddress,
		userAgent,
		map[string]interface{}{
			"changes": changes,
		},
	); err != nil {
		s.logger.Error("failed to create audit log for profile update", "user_id", userID, "error", err)
	}

	// Return updated profile
	return s.GetProfile(userID)
}

// UpdatePreferences updates a user's preferences
func (s *UserService) UpdatePreferences(
	userID uint,
	req *domain.UpdatePreferencesRequest,
	ipAddress, userAgent string,
) (*authdomain.UserPreferences, error) {
	// Get current preferences for audit
	currentPrefs, err := s.userRepo.GetPreferences(userID)
	if err != nil && err != domain.ErrUserNotFound {
		s.logger.Error("failed to get current preferences", "user_id", userID, "error", err)
		return nil, err
	}

	// Build new preferences
	newPrefs := authdomain.UserPreferences{
		Theme:         req.Theme,
		Language:      req.Language,
		Timezone:      req.Timezone,
		Notifications: req.Notifications,
		Privacy:       req.Privacy,
		Custom:        req.Custom,
	}

	// Validate timezone if provided
	if newPrefs.Timezone != "" {
		if _, err := time.LoadLocation(newPrefs.Timezone); err != nil {
			return nil, domain.ErrInvalidPreferences
		}
	}

	// Update preferences
	err = s.userRepo.UpdatePreferences(userID, newPrefs)
	if err != nil {
		s.logger.Error("failed to update user preferences", "user_id", userID, "error", err)
		return nil, domain.ErrInvalidPreferences
	}

	// Create audit log
	changes := s.buildPreferencesChanges(currentPrefs, &newPrefs)
	if err := s.auditRepo.CreateAuditEntry(
		&userID,
		&userID,
		authdomain.AuditActionPreferencesUpdated,
		authdomain.AuditLevelInfo,
		"user",
		fmt.Sprintf("Preferences updated: %s", changes),
		ipAddress,
		userAgent,
		map[string]interface{}{
			"changes": changes,
		},
	); err != nil {
		s.logger.Error("failed to create audit log for preferences update", "user_id", userID, "error", err)
	}

	return &newPrefs, nil
}

// GetPreferences retrieves a user's preferences
func (s *UserService) GetPreferences(userID uint) (*authdomain.UserPreferences, error) {
	return s.userRepo.GetPreferences(userID)
}

// ChangeEmail initiates an email change process
func (s *UserService) ChangeEmail(userID uint, req *domain.ChangeEmailRequest, ipAddress, userAgent string) error {
	// Get current user
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return err
	}

	// Verify current password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return authdomain.ErrInvalidCredentials
	}

	// Check if new email already exists
	exists, err := s.userRepo.CheckEmailExists(req.NewEmail, userID)
	if err != nil {
		s.logger.Error("failed to check email exists", "email", req.NewEmail, "error", err)
		return err
	}
	if exists {
		return domain.ErrEmailAlreadyExists
	}

	// Update email
	oldEmail := user.Email
	err = s.userRepo.UpdateEmail(userID, req.NewEmail)
	if err != nil {
		s.logger.Error("failed to update user email", "user_id", userID, "error", err)
		return err
	}

	// Create audit log
	if err := s.auditRepo.CreateAuditEntry(
		&userID,
		&userID,
		authdomain.AuditActionUserUpdated,
		authdomain.AuditLevelInfo,
		"user",
		fmt.Sprintf("Email changed from %s to %s", oldEmail, req.NewEmail),
		ipAddress,
		userAgent,
		map[string]interface{}{
			"old_email": oldEmail,
			"new_email": req.NewEmail,
		},
	); err != nil {
		s.logger.Error("failed to create audit log for email change", "user_id", userID, "error", err)
	}

	return nil
}

// GetDashboard retrieves dashboard data for a user
func (s *UserService) GetDashboard(userID uint) (*domain.DashboardResponse, error) {
	// Get user profile
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, err
	}

	// Get user stats
	stats, err := s.userRepo.GetUserStats(userID)
	if err != nil {
		s.logger.Error("failed to get user stats", "user_id", userID, "error", err)
		// Continue with empty stats rather than failing
		stats = &domain.UserStats{}
	}

	// Get recent login history (simplified for now)
	recentLogins := []domain.LoginHistoryEntry{}
	if user.LastLoginAt != nil {
		recentLogins = append(recentLogins, domain.LoginHistoryEntry{
			ID:        1,
			IPAddress: "Unknown",
			UserAgent: "Unknown",
			Success:   true,
			CreatedAt: *user.LastLoginAt,
		})
	}

	// Get notifications (placeholder)
	notifications := []domain.NotificationItem{
		{
			ID:        1,
			Type:      domain.NotificationTypeInfo,
			Title:     "Welcome!",
			Message:   "Welcome to your dashboard. Complete your profile to get started.",
			Read:      false,
			Priority:  domain.NotificationPriorityMedium,
			CreatedAt: time.Now(),
		},
	}

	return &domain.DashboardResponse{
		User:          user.ToResponse(),
		Stats:         stats,
		RecentLogins:  recentLogins,
		Notifications: notifications,
	}, nil
}

// buildProfileChanges builds a human-readable string of profile changes
func (s *UserService) buildProfileChanges(current *authdomain.User, req *domain.UpdateProfileRequest) string {
	var changes []string

	if current.FirstName != strings.TrimSpace(req.FirstName) {
		changes = append(changes, fmt.Sprintf("first name: '%s' -> '%s'", current.FirstName, req.FirstName))
	}

	if current.LastName != strings.TrimSpace(req.LastName) {
		changes = append(changes, fmt.Sprintf("last name: '%s' -> '%s'", current.LastName, req.LastName))
	}

	if req.Avatar != "" && current.Avatar != req.Avatar {
		changes = append(changes, "avatar updated")
	}

	if len(changes) == 0 {
		return "no changes"
	}

	return strings.Join(changes, ", ")
}

// buildPreferencesChanges builds a human-readable string of preferences changes
func (s *UserService) buildPreferencesChanges(current, new *authdomain.UserPreferences) string {
	var changes []string

	if current == nil {
		return "preferences initialized"
	}

	if current.Theme != new.Theme {
		changes = append(changes, fmt.Sprintf("theme: '%s' -> '%s'", current.Theme, new.Theme))
	}

	if current.Language != new.Language {
		changes = append(changes, fmt.Sprintf("language: '%s' -> '%s'", current.Language, new.Language))
	}

	if current.Timezone != new.Timezone {
		changes = append(changes, fmt.Sprintf("timezone: '%s' -> '%s'", current.Timezone, new.Timezone))
	}

	// Check notification preferences
	if current.Notifications.Email != new.Notifications.Email ||
		current.Notifications.SMS != new.Notifications.SMS ||
		current.Notifications.Push != new.Notifications.Push {
		changes = append(changes, "notification preferences updated")
	}

	// Check privacy preferences
	if current.Privacy.ProfileVisible != new.Privacy.ProfileVisible ||
		current.Privacy.ShowEmail != new.Privacy.ShowEmail {
		changes = append(changes, "privacy preferences updated")
	}

	if len(changes) == 0 {
		return "no changes"
	}

	return strings.Join(changes, ", ")
}

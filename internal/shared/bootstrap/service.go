package bootstrap

import (
	"errors"
	"log/slog"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/acheevo/tfa/internal/auth/domain"
	"github.com/acheevo/tfa/internal/shared/config"
)

// Service handles bootstrap operations for the application
type Service struct {
	config *config.Config
	db     *gorm.DB
	logger *slog.Logger
}

// NewService creates a new bootstrap service
func NewService(cfg *config.Config, db *gorm.DB, logger *slog.Logger) *Service {
	return &Service{
		config: cfg,
		db:     db,
		logger: logger,
	}
}

// Bootstrap runs all bootstrap operations
func (s *Service) Bootstrap() error {
	if !s.config.BootstrapEnabled {
		s.logger.Info("bootstrap disabled, skipping")
		return nil
	}

	s.logger.Info("starting bootstrap process")

	if err := s.createDemoUsers(); err != nil {
		s.logger.Error("failed to create demo users", "error", err)
		return err
	}

	s.logger.Info("bootstrap process completed successfully")
	return nil
}

// createDemoUsers creates the demo admin and user accounts
func (s *Service) createDemoUsers() error {
	// Create admin user
	if err := s.createUserIfNotExists(
		s.config.AdminEmail,
		s.config.AdminPassword,
		"Admin",
		"User",
		domain.RoleAdmin,
	); err != nil {
		return err
	}

	// Create demo user
	if err := s.createUserIfNotExists(
		s.config.DemoUserEmail,
		s.config.DemoUserPassword,
		"Demo",
		"User",
		domain.RoleUser,
	); err != nil {
		return err
	}

	return nil
}

// createUserIfNotExists creates a user if they don't already exist
func (s *Service) createUserIfNotExists(email, password, firstName, lastName string, role domain.UserRole) error {
	// Check if user already exists
	var existingUser domain.User
	err := s.db.Where("email = ?", email).First(&existingUser).Error
	if err == nil {
		// User exists, check if role matches
		if existingUser.Role != role {
			s.logger.Info("updating user role", "email", email, "old_role", existingUser.Role, "new_role", role)
			if err := s.db.Model(&existingUser).Update("role", role).Error; err != nil {
				return err
			}
		} else {
			s.logger.Debug("user already exists", "email", email, "role", role)
		}
		return nil
	}

	// If error is not "record not found", return the error
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// Create new user
	user := &domain.User{
		Email:         email,
		PasswordHash:  string(hashedPassword),
		FirstName:     firstName,
		LastName:      lastName,
		Role:          role,
		Status:        domain.StatusActive,
		EmailVerified: true, // Bootstrap users are auto-verified
		Preferences: domain.UserPreferences{
			Theme:    "light",
			Language: "en",
			Timezone: "UTC",
			Notifications: domain.NotificationPrefs{
				Email: true,
				SMS:   false,
				Push:  true,
			},
			Privacy: domain.PrivacyPrefs{
				ProfileVisible: true,
				ShowEmail:      false,
			},
			Custom: make(map[string]any),
		},
	}

	if err := s.db.Create(user).Error; err != nil {
		return err
	}

	s.logger.Info("created bootstrap user", "email", email, "role", role, "id", user.ID)
	return nil
}

// DropDemoUsers removes demo users (useful for testing)
func (s *Service) DropDemoUsers() error {
	emails := []string{s.config.AdminEmail, s.config.DemoUserEmail}

	for _, email := range emails {
		if err := s.db.Where("email = ?", email).Delete(&domain.User{}).Error; err != nil {
			s.logger.Error("failed to delete demo user", "email", email, "error", err)
			return err
		}
		s.logger.Info("deleted demo user", "email", email)
	}

	return nil
}

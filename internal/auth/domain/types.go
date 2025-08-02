package domain

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

// UserRole represents the role of a user
type UserRole string

const (
	RoleUser  UserRole = "user"
	RoleAdmin UserRole = "admin"
)

// UserStatus represents the status of a user
type UserStatus string

const (
	StatusActive    UserStatus = "active"
	StatusInactive  UserStatus = "inactive"
	StatusSuspended UserStatus = "suspended"
)

// UserPreferences represents user preferences stored as JSONB
type UserPreferences struct {
	Theme         string            `json:"theme,omitempty"`         // "light", "dark", "system"
	Language      string            `json:"language,omitempty"`      // "en", "es", etc.
	Timezone      string            `json:"timezone,omitempty"`      // "UTC", "America/New_York", etc.
	Notifications NotificationPrefs `json:"notifications,omitempty"` // notification preferences
	Privacy       PrivacyPrefs      `json:"privacy,omitempty"`       // privacy preferences
	Custom        map[string]any    `json:"custom,omitempty"`        // custom application-specific preferences
}

// Value implements the driver.Valuer interface for database storage
func (p UserPreferences) Value() (driver.Value, error) {
	return json.Marshal(p)
}

// Scan implements the sql.Scanner interface for database retrieval
func (p *UserPreferences) Scan(value interface{}) error {
	if value == nil {
		*p = UserPreferences{}
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return errors.New("cannot scan UserPreferences from non-string/[]byte type")
	}

	if len(bytes) == 0 {
		*p = UserPreferences{}
		return nil
	}

	return json.Unmarshal(bytes, p)
}

// NotificationPrefs represents notification preferences
type NotificationPrefs struct {
	Email bool `json:"email"`
	SMS   bool `json:"sms"`
	Push  bool `json:"push"`
}

// PrivacyPrefs represents privacy preferences
type PrivacyPrefs struct {
	ProfileVisible bool `json:"profile_visible"`
	ShowEmail      bool `json:"show_email"`
}

// User represents a user in the system
type User struct {
	ID               uint            `json:"id" gorm:"primarykey"`
	Email            string          `json:"email" gorm:"uniqueIndex;not null"`
	PasswordHash     string          `json:"-" gorm:"not null"`
	FirstName        string          `json:"first_name" gorm:"not null"`
	LastName         string          `json:"last_name" gorm:"not null"`
	EmailVerified    bool            `json:"email_verified" gorm:"default:false"`
	EmailVerifyToken string          `json:"-" gorm:"index"`
	Role             UserRole        `json:"role" gorm:"default:'user';not null"`
	Status           UserStatus      `json:"status" gorm:"default:'active';not null"`
	Preferences      UserPreferences `json:"preferences" gorm:"type:jsonb;default:'{}'"`
	Avatar           string          `json:"avatar"` // URL to avatar image
	LastLoginAt      *time.Time      `json:"last_login_at"`
	CreatedAt        time.Time       `json:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at"`
	DeletedAt        gorm.DeletedAt  `json:"-" gorm:"index"`

	// Relationships
	RefreshTokens []RefreshToken `json:"-" gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
}

// IsActive checks if the user is active
func (u *User) IsActive() bool {
	return u.Status == StatusActive
}

// IsAdmin checks if the user has admin role
func (u *User) IsAdmin() bool {
	return u.Role == RoleAdmin
}

// UserResponse represents the user data returned to the client
type UserResponse struct {
	ID            uint            `json:"id"`
	Email         string          `json:"email"`
	FirstName     string          `json:"first_name"`
	LastName      string          `json:"last_name"`
	EmailVerified bool            `json:"email_verified"`
	Role          UserRole        `json:"role"`
	Status        UserStatus      `json:"status"`
	Preferences   UserPreferences `json:"preferences"`
	Avatar        string          `json:"avatar,omitempty"`
	LastLoginAt   *time.Time      `json:"last_login_at"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
}

// ToResponse converts User to UserResponse
func (u *User) ToResponse() *UserResponse {
	return &UserResponse{
		ID:            u.ID,
		Email:         u.Email,
		FirstName:     u.FirstName,
		LastName:      u.LastName,
		EmailVerified: u.EmailVerified,
		Role:          u.Role,
		Status:        u.Status,
		Preferences:   u.Preferences,
		Avatar:        u.Avatar,
		LastLoginAt:   u.LastLoginAt,
		CreatedAt:     u.CreatedAt,
		UpdatedAt:     u.UpdatedAt,
	}
}

// RefreshToken represents a refresh token for JWT authentication
type RefreshToken struct {
	ID        uint           `json:"id" gorm:"primarykey"`
	UserID    uint           `json:"user_id" gorm:"not null;index"`
	Token     string         `json:"-" gorm:"uniqueIndex;not null"`
	ExpiresAt time.Time      `json:"expires_at" gorm:"not null"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Relationships
	User User `json:"-" gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
}

// IsExpired checks if the refresh token is expired
func (rt *RefreshToken) IsExpired() bool {
	return time.Now().After(rt.ExpiresAt)
}

// PasswordReset represents a password reset request
type PasswordReset struct {
	ID        uint           `json:"id" gorm:"primarykey"`
	Email     string         `json:"email" gorm:"not null;index"`
	Token     string         `json:"token" gorm:"uniqueIndex;not null"`
	ExpiresAt time.Time      `json:"expires_at" gorm:"not null"`
	Used      bool           `json:"used" gorm:"default:false"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

// IsExpired checks if the password reset token is expired
func (pr *PasswordReset) IsExpired() bool {
	return time.Now().After(pr.ExpiresAt)
}

// AuditAction represents the type of audit action
type AuditAction string

const (
	AuditActionUserCreated        AuditAction = "user_created"
	AuditActionUserUpdated        AuditAction = "user_updated"
	AuditActionUserDeleted        AuditAction = "user_deleted"
	AuditActionUserStatusChanged  AuditAction = "user_status_changed"
	AuditActionUserRoleChanged    AuditAction = "user_role_changed"
	AuditActionPasswordChanged    AuditAction = "password_changed"
	AuditActionEmailVerified      AuditAction = "email_verified"
	AuditActionLoginSuccess       AuditAction = "login_success"
	AuditActionLoginFailed        AuditAction = "login_failed"
	AuditActionLogout             AuditAction = "logout"
	AuditActionPasswordResetReq   AuditAction = "password_reset_requested"
	AuditActionPasswordResetUsed  AuditAction = "password_reset_used"
	AuditActionPreferencesUpdated AuditAction = "preferences_updated"
)

// AuditLevel represents the severity level of the audit event
type AuditLevel string

const (
	AuditLevelInfo    AuditLevel = "info"
	AuditLevelWarning AuditLevel = "warning"
	AuditLevelError   AuditLevel = "error"
)

// AuditLog represents an audit log entry for tracking system events
type AuditLog struct {
	ID          uint                   `json:"id" gorm:"primarykey"`
	UserID      *uint                  `json:"user_id" gorm:"index"`
	TargetID    *uint                  `json:"target_id" gorm:"index"`
	Action      AuditAction            `json:"action" gorm:"not null;index"`
	Level       AuditLevel             `json:"level" gorm:"default:'info';not null"`
	Resource    string                 `json:"resource" gorm:"not null"` // e.g., "user", "admin", "auth"
	Description string                 `json:"description" gorm:"not null"`
	IPAddress   string                 `json:"ip_address"`
	UserAgent   string                 `json:"user_agent"`
	Metadata    map[string]interface{} `json:"metadata" gorm:"type:jsonb;default:'{}'"` // Additional structured data
	CreatedAt   time.Time              `json:"created_at"`

	// Relationships
	User   *User `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Target *User `json:"target,omitempty" gorm:"foreignKey:TargetID"`
}

// Authentication DTOs

// RegisterRequest represents a user registration request
type RegisterRequest struct {
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required,min=8"`
	FirstName string `json:"first_name" binding:"required,min=1"`
	LastName  string `json:"last_name" binding:"required,min=1"`
}

// LoginRequest represents a user login request
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// RefreshTokenRequest represents a token refresh request
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// ForgotPasswordRequest represents a forgot password request
type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// ResetPasswordRequest represents a password reset request
type ResetPasswordRequest struct {
	Token           string `json:"token" binding:"required"`
	Password        string `json:"password" binding:"required,min=8"`
	ConfirmPassword string `json:"confirm_password" binding:"required"`
}

// ChangePasswordRequest represents a password change request
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=8"`
	ConfirmPassword string `json:"confirm_password" binding:"required"`
}

// EmailVerificationRequest represents an email verification request
type EmailVerificationRequest struct {
	Token string `json:"token" binding:"required"`
}

// AuthResponse represents the response after successful authentication
type AuthResponse struct {
	User         *UserResponse `json:"user"`
	AccessToken  string        `json:"access_token"`
	RefreshToken string        `json:"refresh_token"`
	ExpiresIn    int64         `json:"expires_in"` // seconds
}

// MessageResponse represents a simple message response
type MessageResponse struct {
	Message string `json:"message"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string            `json:"error"`
	Details map[string]string `json:"details,omitempty"`
}

// JWT Claims

// JWTClaims represents the claims in a JWT token
// Implements jwt.Claims interface
type JWTClaims struct {
	UserID    uint     `json:"user_id"`
	Email     string   `json:"email"`
	Role      UserRole `json:"role"`       // User role for authorization
	TokenType string   `json:"token_type"` // "access" or "refresh"
	jwt.RegisteredClaims
}

// Valid validates the JWT claims
func (c *JWTClaims) Valid() error {
	// Check expiration using the new jwt library
	now := time.Now()
	if c.ExpiresAt != nil && now.After(c.ExpiresAt.Time) {
		return ErrTokenExpired
	}
	return nil
}

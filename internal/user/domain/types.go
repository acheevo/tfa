package domain

import (
	"time"

	authdomain "github.com/acheevo/tfa/internal/auth/domain"
)

// User management DTOs

// UpdateProfileRequest represents a user profile update request
type UpdateProfileRequest struct {
	FirstName string `json:"first_name" binding:"required,min=1,max=50"`
	LastName  string `json:"last_name" binding:"required,min=1,max=50"`
	Avatar    string `json:"avatar" binding:"omitempty,url"`
}

// UpdatePreferencesRequest represents a user preferences update request
type UpdatePreferencesRequest struct {
	Theme         string                       `json:"theme" binding:"omitempty,oneof=light dark system"`
	Language      string                       `json:"language" binding:"omitempty,len=2"`
	Timezone      string                       `json:"timezone" binding:"omitempty"`
	Notifications authdomain.NotificationPrefs `json:"notifications"`
	Privacy       authdomain.PrivacyPrefs      `json:"privacy"`
	Custom        map[string]interface{}       `json:"custom"`
}

// ChangeEmailRequest represents an email change request
type ChangeEmailRequest struct {
	NewEmail string `json:"new_email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// UserListRequest represents a request to list users with filtering and pagination
type UserListRequest struct {
	Page      int                   `form:"page,default=1" binding:"min=1"`
	PageSize  int                   `form:"page_size,default=20" binding:"min=1,max=100"`
	Search    string                `form:"search"`
	Role      authdomain.UserRole   `form:"role" binding:"omitempty,oneof=user admin"`
	Status    authdomain.UserStatus `form:"status" binding:"omitempty,oneof=active inactive suspended"`
	SortBy    string                `form:"sort_by,default=created_at" binding:"omitempty"`
	SortOrder string                `form:"sort_order,default=desc" binding:"omitempty,oneof=asc desc"`
}

// UserListResponse represents the response for user list requests
type UserListResponse struct {
	Users      []*UserSummary `json:"users"`
	Pagination Pagination     `json:"pagination"`
}

// UserSummary represents a summary of user information for list views
type UserSummary struct {
	ID            uint                  `json:"id"`
	Email         string                `json:"email"`
	FirstName     string                `json:"first_name"`
	LastName      string                `json:"last_name"`
	Role          authdomain.UserRole   `json:"role"`
	Status        authdomain.UserStatus `json:"status"`
	EmailVerified bool                  `json:"email_verified"`
	Avatar        string                `json:"avatar,omitempty"`
	LastLoginAt   *time.Time            `json:"last_login_at"`
	CreatedAt     time.Time             `json:"created_at"`
	UpdatedAt     time.Time             `json:"updated_at"`
}

// UserDetailResponse represents detailed user information
type UserDetailResponse struct {
	*authdomain.UserResponse
	LoginHistory []LoginHistoryEntry `json:"login_history,omitempty"`
	AuditTrail   []AuditLogEntry     `json:"audit_trail,omitempty"`
}

// LoginHistoryEntry represents a login history entry
type LoginHistoryEntry struct {
	ID        uint      `json:"id"`
	IPAddress string    `json:"ip_address"`
	UserAgent string    `json:"user_agent"`
	Success   bool      `json:"success"`
	CreatedAt time.Time `json:"created_at"`
}

// AuditLogEntry represents an audit log entry
type AuditLogEntry struct {
	ID          uint                   `json:"id"`
	Action      authdomain.AuditAction `json:"action"`
	Level       authdomain.AuditLevel  `json:"level"`
	Resource    string                 `json:"resource"`
	Description string                 `json:"description"`
	IPAddress   string                 `json:"ip_address"`
	UserAgent   string                 `json:"user_agent"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
}

// Pagination represents pagination information
type Pagination struct {
	Page       int  `json:"page"`
	PageSize   int  `json:"page_size"`
	Total      int  `json:"total"`
	TotalPages int  `json:"total_pages"`
	HasNext    bool `json:"has_next"`
	HasPrev    bool `json:"has_prev"`
}

// Dashboard response types

// DashboardResponse represents the user dashboard data
type DashboardResponse struct {
	User          *authdomain.UserResponse `json:"user"`
	Stats         *UserStats               `json:"stats"`
	RecentLogins  []LoginHistoryEntry      `json:"recent_logins"`
	Notifications []NotificationItem       `json:"notifications"`
}

// UserStats represents user statistics
type UserStats struct {
	TotalLogins     int        `json:"total_logins"`
	LastLoginAt     *time.Time `json:"last_login_at"`
	AccountAge      int        `json:"account_age_days"`
	ProfileComplete bool       `json:"profile_complete"`
}

// NotificationItem represents a notification item
type NotificationItem struct {
	ID        uint                 `json:"id"`
	Type      NotificationType     `json:"type"`
	Title     string               `json:"title"`
	Message   string               `json:"message"`
	Read      bool                 `json:"read"`
	Priority  NotificationPriority `json:"priority"`
	CreatedAt time.Time            `json:"created_at"`
}

// NotificationType represents the type of notification
type NotificationType string

const (
	NotificationTypeInfo    NotificationType = "info"
	NotificationTypeWarning NotificationType = "warning"
	NotificationTypeError   NotificationType = "error"
	NotificationTypeSuccess NotificationType = "success"
)

// NotificationPriority represents the priority of a notification
type NotificationPriority string

const (
	NotificationPriorityLow    NotificationPriority = "low"
	NotificationPriorityMedium NotificationPriority = "medium"
	NotificationPriorityHigh   NotificationPriority = "high"
)

// Helper functions

// ToUserSummary converts a User to UserSummary
func ToUserSummary(u *authdomain.User) *UserSummary {
	return &UserSummary{
		ID:            u.ID,
		Email:         u.Email,
		FirstName:     u.FirstName,
		LastName:      u.LastName,
		Role:          u.Role,
		Status:        u.Status,
		EmailVerified: u.EmailVerified,
		Avatar:        u.Avatar,
		LastLoginAt:   u.LastLoginAt,
		CreatedAt:     u.CreatedAt,
		UpdatedAt:     u.UpdatedAt,
	}
}

// ToUserDetailResponse converts a User to UserDetailResponse
func ToUserDetailResponse(u *authdomain.User) *UserDetailResponse {
	return &UserDetailResponse{
		UserResponse: u.ToResponse(),
	}
}

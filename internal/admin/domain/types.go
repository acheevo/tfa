package domain

import (
	"time"

	authdomain "github.com/acheevo/tfa/internal/auth/domain"
	userdomain "github.com/acheevo/tfa/internal/user/domain"
)

// Admin user management DTOs

// UpdateUserRoleRequest represents a request to update a user's role
type UpdateUserRoleRequest struct {
	Role   authdomain.UserRole `json:"role" binding:"required,oneof=user admin"`
	Reason string              `json:"reason" binding:"required,min=1,max=255"`
}

// UpdateUserStatusRequest represents a request to update a user's status
type UpdateUserStatusRequest struct {
	Status authdomain.UserStatus `json:"status" binding:"required,oneof=active inactive suspended"`
	Reason string                `json:"reason" binding:"required,min=1,max=255"`
}

// AdminUpdateUserRequest represents an admin request to update user information
type AdminUpdateUserRequest struct {
	FirstName     string                `json:"first_name" binding:"omitempty,min=1,max=50"`
	LastName      string                `json:"last_name" binding:"omitempty,min=1,max=50"`
	Email         string                `json:"email" binding:"omitempty,email"`
	EmailVerified *bool                 `json:"email_verified"`
	Role          authdomain.UserRole   `json:"role" binding:"omitempty,oneof=user admin"`
	Status        authdomain.UserStatus `json:"status" binding:"omitempty,oneof=active inactive suspended"`
	Avatar        string                `json:"avatar" binding:"omitempty,url"`
	Reason        string                `json:"reason" binding:"required,min=1,max=255"`
}

// DeleteUserRequest represents a request to delete a user
type DeleteUserRequest struct {
	Reason string `json:"reason" binding:"required,min=1,max=255"`
	Force  bool   `json:"force"` // Force delete (hard delete) vs soft delete
}

// BulkUserActionRequest represents a request to perform bulk actions on users
type BulkUserActionRequest struct {
	UserIDs []uint               `json:"user_ids" binding:"required,min=1"`
	Action  BulkActionType       `json:"action" binding:"required,oneof=activate deactivate suspend delete role_change"`
	Role    *authdomain.UserRole `json:"role" binding:"required_if=Action role_change"`
	Reason  string               `json:"reason" binding:"required,min=1,max=255"`
}

// BulkActionType represents the type of bulk action
type BulkActionType string

const (
	BulkActionActivate   BulkActionType = "activate"
	BulkActionDeactivate BulkActionType = "deactivate"
	BulkActionSuspend    BulkActionType = "suspend"
	BulkActionDelete     BulkActionType = "delete"
	BulkActionRoleChange BulkActionType = "role_change"
)

// BulkActionResult represents the result of a bulk action
type BulkActionResult struct {
	TotalRequested int                    `json:"total_requested"`
	Successful     int                    `json:"successful"`
	Failed         int                    `json:"failed"`
	Results        []BulkActionItemResult `json:"results"`
}

// BulkActionItemResult represents the result of a single item in a bulk action
type BulkActionItemResult struct {
	UserID  uint   `json:"user_id"`
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

// AdminStatsResponse represents admin dashboard statistics
type AdminStatsResponse struct {
	TotalUsers       int              `json:"total_users"`
	ActiveUsers      int              `json:"active_users"`
	InactiveUsers    int              `json:"inactive_users"`
	SuspendedUsers   int              `json:"suspended_users"`
	AdminUsers       int              `json:"admin_users"`
	NewUsersToday    int              `json:"new_users_today"`
	NewUsersThisWeek int              `json:"new_users_this_week"`
	UserGrowth       []UserGrowthData `json:"user_growth"`
	TopCountries     []CountryData    `json:"top_countries,omitempty"`
}

// UserGrowthData represents user growth data for charts
type UserGrowthData struct {
	Date  string `json:"date"`
	Count int    `json:"count"`
}

// CountryData represents user count by country
type CountryData struct {
	Country string `json:"country"`
	Count   int    `json:"count"`
}

// AdminAuditLogRequest represents a request to fetch audit logs
type AdminAuditLogRequest struct {
	Page      int                    `form:"page,default=1" binding:"min=1"`
	PageSize  int                    `form:"page_size,default=50" binding:"min=1,max=100"`
	UserID    *uint                  `form:"user_id"`
	TargetID  *uint                  `form:"target_id"`
	Action    authdomain.AuditAction `form:"action"`
	Level     authdomain.AuditLevel  `form:"level" binding:"omitempty,oneof=info warning error"`
	Resource  string                 `form:"resource"`
	DateFrom  *time.Time             `form:"date_from" time_format:"2006-01-02"`
	DateTo    *time.Time             `form:"date_to" time_format:"2006-01-02"`
	IPAddress string                 `form:"ip_address"`
}

// AdminAuditLogResponse represents the response for audit log requests
type AdminAuditLogResponse struct {
	Logs       []*EnhancedAuditLogEntry `json:"logs"`
	Pagination userdomain.Pagination    `json:"pagination"`
}

// EnhancedAuditLogEntry represents an enhanced audit log entry with user details
type EnhancedAuditLogEntry struct {
	userdomain.AuditLogEntry
	User   *userdomain.UserSummary `json:"user,omitempty"`
	Target *userdomain.UserSummary `json:"target,omitempty"`
}

// SystemHealthResponse represents system health information
type SystemHealthResponse struct {
	Status    string        `json:"status"`
	Timestamp time.Time     `json:"timestamp"`
	Version   string        `json:"version"`
	Uptime    string        `json:"uptime"`
	Database  HealthStatus  `json:"database"`
	Redis     HealthStatus  `json:"redis,omitempty"`
	Email     HealthStatus  `json:"email"`
	Metrics   SystemMetrics `json:"metrics"`
}

// HealthStatus represents the health status of a component
type HealthStatus struct {
	Status    string    `json:"status"`
	Latency   string    `json:"latency,omitempty"`
	Error     string    `json:"error,omitempty"`
	LastCheck time.Time `json:"last_check"`
}

// SystemMetrics represents system performance metrics
type SystemMetrics struct {
	ActiveSessions  int     `json:"active_sessions"`
	RequestsPerMin  int     `json:"requests_per_minute"`
	AvgResponseTime string  `json:"avg_response_time"`
	ErrorRate       float64 `json:"error_rate"`
	MemoryUsage     string  `json:"memory_usage"`
	DiskUsage       string  `json:"disk_usage"`
}

// Permission and role management

// PermissionCheck represents a permission check
type PermissionCheck struct {
	UserID     uint   `json:"user_id"`
	Resource   string `json:"resource"`
	Action     string `json:"action"`
	Permission string `json:"permission"`
}

// PermissionResponse represents a permission check response
type PermissionResponse struct {
	Allowed bool   `json:"allowed"`
	Reason  string `json:"reason,omitempty"`
}

// Helper methods

// ToEnhancedAuditLogEntry converts an AuditLog to EnhancedAuditLogEntry
func ToEnhancedAuditLogEntry(log *authdomain.AuditLog) *EnhancedAuditLogEntry {
	entry := &EnhancedAuditLogEntry{
		AuditLogEntry: userdomain.AuditLogEntry{
			ID:          log.ID,
			Action:      log.Action,
			Level:       log.Level,
			Resource:    log.Resource,
			Description: log.Description,
			IPAddress:   log.IPAddress,
			UserAgent:   log.UserAgent,
			Metadata:    log.Metadata,
			CreatedAt:   log.CreatedAt,
		},
	}

	if log.User != nil {
		entry.User = userdomain.ToUserSummary(log.User)
	}

	if log.Target != nil {
		entry.Target = userdomain.ToUserSummary(log.Target)
	}

	return entry
}

// Admin permissions

// IsAuthorizedForUserManagement checks if a user can manage other users
func IsAuthorizedForUserManagement(user *authdomain.User) bool {
	return user.IsAdmin() && user.IsActive()
}

// CanManageUser checks if an admin can manage a specific user
func CanManageUser(admin, target *authdomain.User) bool {
	if !IsAuthorizedForUserManagement(admin) {
		return false
	}

	// Admins cannot manage themselves through admin endpoints
	if admin.ID == target.ID {
		return false
	}

	// Super admin logic could be added here if needed
	return true
}

// ValidateBulkAction validates bulk action requests
func (r *BulkUserActionRequest) Validate() error {
	if len(r.UserIDs) == 0 {
		return userdomain.ErrInvalidRequest
	}

	if r.Action == BulkActionRoleChange && r.Role == nil {
		return userdomain.ErrInvalidRequest
	}

	return nil
}

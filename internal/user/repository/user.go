package repository

import (
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"

	authdomain "github.com/acheevo/tfa/internal/auth/domain"
	"github.com/acheevo/tfa/internal/user/domain"
)

// UserRepository handles user-related database operations
type UserRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{
		db: db,
	}
}

// GetByID retrieves a user by ID
func (r *UserRepository) GetByID(id uint) (*authdomain.User, error) {
	var user authdomain.User
	err := r.db.First(&user, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

// Update updates a user's information
func (r *UserRepository) Update(user *authdomain.User) error {
	return r.db.Save(user).Error
}

// UpdateProfile updates a user's profile information
func (r *UserRepository) UpdateProfile(userID uint, req *domain.UpdateProfileRequest) error {
	updates := map[string]interface{}{
		"first_name": strings.TrimSpace(req.FirstName),
		"last_name":  strings.TrimSpace(req.LastName),
		"updated_at": time.Now(),
	}

	if req.Avatar != "" {
		updates["avatar"] = req.Avatar
	}

	return r.db.Model(&authdomain.User{}).Where("id = ?", userID).Updates(updates).Error
}

// UpdatePreferences updates a user's preferences
func (r *UserRepository) UpdatePreferences(userID uint, preferences authdomain.UserPreferences) error {
	return r.db.Model(&authdomain.User{}).
		Where("id = ?", userID).
		Update("preferences", preferences).Error
}

// GetPreferences retrieves a user's preferences
func (r *UserRepository) GetPreferences(userID uint) (*authdomain.UserPreferences, error) {
	var user authdomain.User
	err := r.db.Select("preferences").First(&user, userID).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}
	return &user.Preferences, nil
}

// List retrieves users with filtering and pagination
func (r *UserRepository) List(req *domain.UserListRequest) ([]*authdomain.User, int, error) {
	var users []*authdomain.User
	var total int64

	query := r.db.Model(&authdomain.User{})

	// Apply filters
	if req.Search != "" {
		searchTerm := "%" + strings.ToLower(req.Search) + "%"
		query = query.Where(
			"LOWER(email) LIKE ? OR LOWER(first_name) LIKE ? OR LOWER(last_name) LIKE ?",
			searchTerm, searchTerm, searchTerm,
		)
	}

	if req.Role != "" {
		query = query.Where("role = ?", req.Role)
	}

	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}

	// Count total records
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply sorting
	orderClause := fmt.Sprintf("%s %s", req.SortBy, strings.ToUpper(req.SortOrder))
	query = query.Order(orderClause)

	// Apply pagination
	offset := (req.Page - 1) * req.PageSize
	if err := query.Offset(offset).Limit(req.PageSize).Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, int(total), nil
}

// GetUserStats retrieves user statistics for dashboard
func (r *UserRepository) GetUserStats(userID uint) (*domain.UserStats, error) {
	var user authdomain.User
	err := r.db.Select("created_at, last_login_at").First(&user, userID).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}

	// Calculate account age
	accountAge := int(time.Since(user.CreatedAt).Hours() / 24)

	// Get total login count (would need a login history table for accurate count)
	// For now, we'll use a placeholder
	totalLogins := 1
	if user.LastLoginAt != nil {
		totalLogins = 5 // Placeholder
	}

	// Check if profile is complete
	profileComplete := user.FirstName != "" && user.LastName != "" && user.EmailVerified

	return &domain.UserStats{
		TotalLogins:     totalLogins,
		LastLoginAt:     user.LastLoginAt,
		AccountAge:      accountAge,
		ProfileComplete: profileComplete,
	}, nil
}

// UpdateEmail updates a user's email address
func (r *UserRepository) UpdateEmail(userID uint, newEmail string) error {
	updates := map[string]interface{}{
		"email":          strings.ToLower(strings.TrimSpace(newEmail)),
		"email_verified": false, // Reset email verification when email changes
		"updated_at":     time.Now(),
	}

	return r.db.Model(&authdomain.User{}).Where("id = ?", userID).Updates(updates).Error
}

// CheckEmailExists checks if an email already exists (excluding a specific user ID)
func (r *UserRepository) CheckEmailExists(email string, excludeUserID uint) (bool, error) {
	var count int64
	query := r.db.Model(&authdomain.User{}).Where("email = ?", strings.ToLower(strings.TrimSpace(email)))
	if excludeUserID > 0 {
		query = query.Where("id != ?", excludeUserID)
	}

	err := query.Count(&count).Error
	return count > 0, err
}

// GetUsersByIDs retrieves multiple users by their IDs
func (r *UserRepository) GetUsersByIDs(ids []uint) ([]*authdomain.User, error) {
	var users []*authdomain.User
	err := r.db.Where("id IN ?", ids).Find(&users).Error
	return users, err
}

// UpdateUserRole updates a user's role
func (r *UserRepository) UpdateUserRole(userID uint, role authdomain.UserRole) error {
	return r.db.Model(&authdomain.User{}).
		Where("id = ?", userID).
		Update("role", role).Error
}

// UpdateUserStatus updates a user's status
func (r *UserRepository) UpdateUserStatus(userID uint, status authdomain.UserStatus) error {
	return r.db.Model(&authdomain.User{}).
		Where("id = ?", userID).
		Update("status", status).Error
}

// BulkUpdateStatus updates status for multiple users
func (r *UserRepository) BulkUpdateStatus(userIDs []uint, status authdomain.UserStatus) error {
	return r.db.Model(&authdomain.User{}).
		Where("id IN ?", userIDs).
		Update("status", status).Error
}

// BulkUpdateRole updates role for multiple users
func (r *UserRepository) BulkUpdateRole(userIDs []uint, role authdomain.UserRole) error {
	return r.db.Model(&authdomain.User{}).
		Where("id IN ?", userIDs).
		Update("role", role).Error
}

// SoftDelete soft deletes users
func (r *UserRepository) SoftDelete(userIDs []uint) error {
	return r.db.Delete(&authdomain.User{}, userIDs).Error
}

// HardDelete permanently deletes users
func (r *UserRepository) HardDelete(userIDs []uint) error {
	return r.db.Unscoped().Delete(&authdomain.User{}, userIDs).Error
}

// GetAdminStats retrieves admin dashboard statistics
func (r *UserRepository) GetAdminStats() (*AdminStats, error) {
	stats := &AdminStats{}

	// Total users
	r.db.Model(&authdomain.User{}).Count(&stats.TotalUsers)

	// Users by status
	r.db.Model(&authdomain.User{}).Where("status = ?", authdomain.StatusActive).Count(&stats.ActiveUsers)
	r.db.Model(&authdomain.User{}).Where("status = ?", authdomain.StatusInactive).Count(&stats.InactiveUsers)
	r.db.Model(&authdomain.User{}).Where("status = ?", authdomain.StatusSuspended).Count(&stats.SuspendedUsers)

	// Admin users
	r.db.Model(&authdomain.User{}).Where("role = ?", authdomain.RoleAdmin).Count(&stats.AdminUsers)

	// New users today
	today := time.Now().Truncate(24 * time.Hour)
	r.db.Model(&authdomain.User{}).Where("created_at >= ?", today).Count(&stats.NewUsersToday)

	// New users this week
	weekStart := time.Now().AddDate(0, 0, -7).Truncate(24 * time.Hour)
	r.db.Model(&authdomain.User{}).Where("created_at >= ?", weekStart).Count(&stats.NewUsersThisWeek)

	return stats, nil
}

// GetUserGrowthData retrieves user growth data for the last 30 days
func (r *UserRepository) GetUserGrowthData(days int) ([]UserGrowthDataPoint, error) {
	var results []UserGrowthDataPoint

	// Query to get user registration counts per day
	query := `
		SELECT 
			DATE(created_at) as date,
			COUNT(*) as count
		FROM users 
		WHERE created_at >= ? AND deleted_at IS NULL
		GROUP BY DATE(created_at)
		ORDER BY date DESC
		LIMIT ?`

	startDate := time.Now().AddDate(0, 0, -days)
	err := r.db.Raw(query, startDate, days).Scan(&results).Error

	return results, err
}

// AdminStats represents statistics for admin dashboard
type AdminStats struct {
	TotalUsers       int64 `json:"total_users"`
	ActiveUsers      int64 `json:"active_users"`
	InactiveUsers    int64 `json:"inactive_users"`
	SuspendedUsers   int64 `json:"suspended_users"`
	AdminUsers       int64 `json:"admin_users"`
	NewUsersToday    int64 `json:"new_users_today"`
	NewUsersThisWeek int64 `json:"new_users_this_week"`
}

// UserGrowthDataPoint represents a data point for user growth charts
type UserGrowthDataPoint struct {
	Date  string `json:"date"`
	Count int    `json:"count"`
}

package repository

import (
	"strings"
	"time"

	"gorm.io/gorm"

	admindomain "github.com/acheevo/tfa/internal/admin/domain"
	authdomain "github.com/acheevo/tfa/internal/auth/domain"
)

// AuditRepository handles audit log database operations
type AuditRepository struct {
	db *gorm.DB
}

// NewAuditRepository creates a new audit repository
func NewAuditRepository(db *gorm.DB) *AuditRepository {
	return &AuditRepository{
		db: db,
	}
}

// Create creates a new audit log entry
func (r *AuditRepository) Create(log *authdomain.AuditLog) error {
	return r.db.Create(log).Error
}

// CreateAuditEntry creates an audit log entry with minimal parameters
func (r *AuditRepository) CreateAuditEntry(
	userID *uint,
	targetID *uint,
	action authdomain.AuditAction,
	level authdomain.AuditLevel,
	resource string,
	description string,
	ipAddress string,
	userAgent string,
	metadata map[string]interface{},
) error {
	log := &authdomain.AuditLog{
		UserID:      userID,
		TargetID:    targetID,
		Action:      action,
		Level:       level,
		Resource:    resource,
		Description: description,
		IPAddress:   ipAddress,
		UserAgent:   userAgent,
		Metadata:    metadata,
		CreatedAt:   time.Now(),
	}

	return r.Create(log)
}

// List retrieves audit logs with filtering and pagination
func (r *AuditRepository) List(req *admindomain.AdminAuditLogRequest) ([]*authdomain.AuditLog, int, error) {
	var logs []*authdomain.AuditLog
	var total int64

	query := r.db.Model(&authdomain.AuditLog{}).Preload("User").Preload("Target")

	// Apply filters
	if req.UserID != nil {
		query = query.Where("user_id = ?", *req.UserID)
	}

	if req.TargetID != nil {
		query = query.Where("target_id = ?", *req.TargetID)
	}

	if req.Action != "" {
		query = query.Where("action = ?", req.Action)
	}

	if req.Level != "" {
		query = query.Where("level = ?", req.Level)
	}

	if req.Resource != "" {
		query = query.Where("resource = ?", req.Resource)
	}

	if req.IPAddress != "" {
		query = query.Where("ip_address = ?", req.IPAddress)
	}

	if req.DateFrom != nil {
		query = query.Where("created_at >= ?", *req.DateFrom)
	}

	if req.DateTo != nil {
		// Add time to end of day
		endOfDay := req.DateTo.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
		query = query.Where("created_at <= ?", endOfDay)
	}

	// Count total records
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination and sorting
	offset := (req.Page - 1) * req.PageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(req.PageSize).Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, int(total), nil
}

// GetByID retrieves an audit log by ID
func (r *AuditRepository) GetByID(id uint) (*authdomain.AuditLog, error) {
	var log authdomain.AuditLog
	err := r.db.Preload("User").Preload("Target").First(&log, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, admindomain.ErrAuditLogNotFound
		}
		return nil, err
	}
	return &log, nil
}

// GetUserAuditHistory retrieves audit history for a specific user
func (r *AuditRepository) GetUserAuditHistory(userID uint, limit int) ([]*authdomain.AuditLog, error) {
	var logs []*authdomain.AuditLog
	query := r.db.Where("user_id = ? OR target_id = ?", userID, userID).
		Preload("User").
		Preload("Target").
		Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Find(&logs).Error
	return logs, err
}

// GetRecentLogs retrieves recent audit logs
func (r *AuditRepository) GetRecentLogs(limit int) ([]*authdomain.AuditLog, error) {
	var logs []*authdomain.AuditLog
	err := r.db.Preload("User").Preload("Target").
		Order("created_at DESC").
		Limit(limit).
		Find(&logs).Error
	return logs, err
}

// DeleteOldLogs deletes audit logs older than the specified duration
func (r *AuditRepository) DeleteOldLogs(olderThan time.Duration) (int64, error) {
	cutoffDate := time.Now().Add(-olderThan)
	result := r.db.Where("created_at < ?", cutoffDate).Delete(&authdomain.AuditLog{})
	return result.RowsAffected, result.Error
}

// GetLogsByAction retrieves logs by specific action
func (r *AuditRepository) GetLogsByAction(action authdomain.AuditAction, limit int) ([]*authdomain.AuditLog, error) {
	var logs []*authdomain.AuditLog
	query := r.db.Where("action = ?", action).
		Preload("User").
		Preload("Target").
		Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Find(&logs).Error
	return logs, err
}

// GetLogsByLevel retrieves logs by severity level
func (r *AuditRepository) GetLogsByLevel(level authdomain.AuditLevel, limit int) ([]*authdomain.AuditLog, error) {
	var logs []*authdomain.AuditLog
	query := r.db.Where("level = ?", level).
		Preload("User").
		Preload("Target").
		Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Find(&logs).Error
	return logs, err
}

// SearchLogs searches audit logs by description
func (r *AuditRepository) SearchLogs(searchTerm string, limit int) ([]*authdomain.AuditLog, error) {
	var logs []*authdomain.AuditLog
	searchPattern := "%" + strings.ToLower(searchTerm) + "%"

	query := r.db.Where("LOWER(description) LIKE ?", searchPattern).
		Preload("User").
		Preload("Target").
		Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Find(&logs).Error
	return logs, err
}

// GetAuditStatistics retrieves audit log statistics
func (r *AuditRepository) GetAuditStatistics(days int) (*AuditStatistics, error) {
	stats := &AuditStatistics{}
	startDate := time.Now().AddDate(0, 0, -days)

	// Total logs
	r.db.Model(&authdomain.AuditLog{}).Where("created_at >= ?", startDate).Count(&stats.TotalLogs)

	// Logs by level
	r.db.Model(&authdomain.AuditLog{}).
		Where("level = ? AND created_at >= ?", authdomain.AuditLevelInfo, startDate).
		Count(&stats.InfoLogs)
	r.db.Model(&authdomain.AuditLog{}).
		Where("level = ? AND created_at >= ?", authdomain.AuditLevelWarning, startDate).
		Count(&stats.WarningLogs)
	r.db.Model(&authdomain.AuditLog{}).
		Where("level = ? AND created_at >= ?", authdomain.AuditLevelError, startDate).
		Count(&stats.ErrorLogs)

	// Most active users (top 10)
	var activeUsers []ActiveUserStat
	r.db.Model(&authdomain.AuditLog{}).
		Select("user_id, COUNT(*) as log_count").
		Where("user_id IS NOT NULL AND created_at >= ?", startDate).
		Group("user_id").
		Order("log_count DESC").
		Limit(10).
		Find(&activeUsers)
	stats.MostActiveUsers = activeUsers

	return stats, nil
}

// AuditStatistics represents audit log statistics
type AuditStatistics struct {
	TotalLogs       int64            `json:"total_logs"`
	InfoLogs        int64            `json:"info_logs"`
	WarningLogs     int64            `json:"warning_logs"`
	ErrorLogs       int64            `json:"error_logs"`
	MostActiveUsers []ActiveUserStat `json:"most_active_users"`
}

// ActiveUserStat represents active user statistics
type ActiveUserStat struct {
	UserID   uint  `json:"user_id"`
	LogCount int64 `json:"log_count"`
}

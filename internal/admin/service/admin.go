package service

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/acheevo/tfa/internal/admin/domain"
	authdomain "github.com/acheevo/tfa/internal/auth/domain"
	"github.com/acheevo/tfa/internal/shared/config"
	userdomain "github.com/acheevo/tfa/internal/user/domain"
	"github.com/acheevo/tfa/internal/user/repository"
)

// AdminService handles admin user management operations
type AdminService struct {
	config    *config.Config
	logger    *slog.Logger
	userRepo  *repository.UserRepository
	auditRepo *repository.AuditRepository
}

// NewAdminService creates a new admin service
func NewAdminService(
	config *config.Config,
	logger *slog.Logger,
	userRepo *repository.UserRepository,
	auditRepo *repository.AuditRepository,
) *AdminService {
	return &AdminService{
		config:    config,
		logger:    logger,
		userRepo:  userRepo,
		auditRepo: auditRepo,
	}
}

// ListUsers retrieves a paginated list of users with filtering
func (s *AdminService) ListUsers(adminID uint, req *userdomain.UserListRequest) (*userdomain.UserListResponse, error) {
	// Check admin authorization
	admin, err := s.userRepo.GetByID(adminID)
	if err != nil {
		return nil, err
	}

	if !domain.IsAuthorizedForUserManagement(admin) {
		return nil, domain.ErrNotAuthorized
	}

	// Get users
	users, total, err := s.userRepo.List(req)
	if err != nil {
		s.logger.Error("failed to list users", "admin_id", adminID, "error", err)
		return nil, err
	}

	// Convert to summary format
	userSummaries := make([]*userdomain.UserSummary, len(users))
	for i, user := range users {
		userSummaries[i] = userdomain.ToUserSummary(user)
	}

	// Build pagination
	totalPages := (total + req.PageSize - 1) / req.PageSize
	pagination := userdomain.Pagination{
		Page:       req.Page,
		PageSize:   req.PageSize,
		Total:      total,
		TotalPages: totalPages,
		HasNext:    req.Page < totalPages,
		HasPrev:    req.Page > 1,
	}

	return &userdomain.UserListResponse{
		Users:      userSummaries,
		Pagination: pagination,
	}, nil
}

// GetUserDetails retrieves detailed information about a user
func (s *AdminService) GetUserDetails(adminID, targetUserID uint) (*userdomain.UserDetailResponse, error) {
	// Check admin authorization
	admin, err := s.userRepo.GetByID(adminID)
	if err != nil {
		return nil, err
	}

	if !domain.IsAuthorizedForUserManagement(admin) {
		return nil, domain.ErrNotAuthorized
	}

	// Get target user
	targetUser, err := s.userRepo.GetByID(targetUserID)
	if err != nil {
		return nil, err
	}

	// Build response
	response := userdomain.ToUserDetailResponse(targetUser)

	// Get audit trail for this user
	auditLogs, err := s.auditRepo.GetUserAuditHistory(targetUserID, 50)
	if err != nil {
		s.logger.Error("failed to get user audit history", "user_id", targetUserID, "error", err)
		// Continue without audit trail rather than failing
	} else {
		response.AuditTrail = make([]userdomain.AuditLogEntry, len(auditLogs))
		for i, log := range auditLogs {
			response.AuditTrail[i] = userdomain.AuditLogEntry{
				ID:          log.ID,
				Action:      log.Action,
				Level:       log.Level,
				Resource:    log.Resource,
				Description: log.Description,
				IPAddress:   log.IPAddress,
				UserAgent:   log.UserAgent,
				Metadata:    log.Metadata,
				CreatedAt:   log.CreatedAt,
			}
		}
	}

	return response, nil
}

// UpdateUserRole updates a user's role with comprehensive security validation
func (s *AdminService) UpdateUserRole(
	adminID, targetUserID uint,
	req *domain.UpdateUserRoleRequest,
	ipAddress, userAgent string,
) error {
	// Check admin authorization
	admin, err := s.userRepo.GetByID(adminID)
	if err != nil {
		return err
	}

	if !domain.IsAuthorizedForUserManagement(admin) {
		return domain.ErrNotAuthorized
	}

	// Get target user
	targetUser, err := s.userRepo.GetByID(targetUserID)
	if err != nil {
		return err
	}

	// Check if admin can manage this user
	if !domain.CanManageUser(admin, targetUser) {
		return domain.ErrCannotManageSelf
	}

	// Perform comprehensive security validation
	securityCheck := &authdomain.RoleChangeSecurityCheck{
		AdminID:       adminID,
		AdminRole:     admin.Role,
		TargetID:      targetUserID,
		TargetRole:    targetUser.Role,
		NewRole:       req.Role,
		Reason:        req.Reason,
		IPAddress:     ipAddress,
		UserAgent:     userAgent,
		RequestSource: "web",
	}

	validationResult := authdomain.ValidateRoleChange(securityCheck)
	if !validationResult.Valid {
		s.logger.Warn("role change validation failed",
			"admin_id", adminID,
			"target_user_id", targetUserID,
			"errors", validationResult.Errors,
			"risk_level", validationResult.RiskLevel,
		)
		return fmt.Errorf("role change validation failed: %s", strings.Join(validationResult.Errors, "; "))
	}

	// Log security warnings
	if len(validationResult.Warnings) > 0 {
		s.logger.Warn("role change security warnings",
			"admin_id", adminID,
			"target_user_id", targetUserID,
			"warnings", validationResult.Warnings,
			"risk_level", validationResult.RiskLevel,
			"audit_flags", validationResult.AuditFlags,
		)
	}

	// Store the old role for audit
	oldRole := targetUser.Role

	// Create comprehensive audit entry before making changes
	auditEntry := authdomain.CreateRoleChangeAuditEntry(
		admin,
		targetUser,
		req.Role,
		req.Reason,
		ipAddress,
		userAgent,
		"web",
		validationResult,
	)

	// Log the audit entry details
	s.logger.Info("role change initiated",
		"admin_id", adminID,
		"admin_email", admin.Email,
		"target_user_id", targetUserID,
		"target_email", targetUser.Email,
		"old_role", oldRole,
		"new_role", req.Role,
		"risk_level", validationResult.RiskLevel,
		"requires_secondary_auth", validationResult.RequiresSecondaryAuth,
	)

	// TODO: Implement secondary authentication if required
	if validationResult.RequiresSecondaryAuth {
		s.logger.Info("secondary authentication required for role change",
			"admin_id", adminID,
			"target_user_id", targetUserID,
			"new_role", req.Role,
		)
		// For now, we'll proceed, but in production you might want to:
		// 1. Send email to security team
		// 2. Require MFA confirmation
		// 3. Implement approval workflow
	}

	// Update role
	err = s.userRepo.UpdateUserRole(targetUserID, req.Role)
	if err != nil {
		s.logger.Error("failed to update user role",
			"admin_id", adminID,
			"target_user_id", targetUserID,
			"error", err,
		)
		return err
	}

	// Create enhanced audit log with security validation details
	auditDetails := map[string]interface{}{
		"old_role":          oldRole,
		"new_role":          req.Role,
		"reason":            req.Reason,
		"validation_result": validationResult,
		"audit_entry":       auditEntry,
		"security_flags":    validationResult.AuditFlags,
		"risk_level":        validationResult.RiskLevel,
	}

	if err := s.auditRepo.CreateAuditEntry(
		&adminID,
		&targetUserID,
		authdomain.AuditActionUserRoleChanged,
		authdomain.AuditLevelInfo,
		"admin",
		fmt.Sprintf("Role changed from %s to %s: %s [Risk: %s]", oldRole, req.Role, req.Reason, validationResult.RiskLevel),
		ipAddress,
		userAgent,
		auditDetails,
	); err != nil {
		s.logger.Error("failed to create audit log for role change",
			"admin_id", adminID,
			"target_user_id", targetUserID,
			"error", err,
		)
	}

	// Generate security alerts for high-risk changes
	if validationResult.RiskLevel == "high" || validationResult.RiskLevel == "critical" {
		alertData := map[string]interface{}{
			"admin_id":     adminID,
			"admin_email":  admin.Email,
			"target_id":    targetUserID,
			"target_email": targetUser.Email,
			"old_role":     oldRole,
			"new_role":     req.Role,
			"reason":       req.Reason,
			"risk_level":   validationResult.RiskLevel,
			"audit_flags":  validationResult.AuditFlags,
			"ip_address":   ipAddress,
		}

		alert := authdomain.GenerateSecurityAlert(
			"role_change",
			validationResult.RiskLevel,
			fmt.Sprintf("High-risk role change: %s → %s", oldRole, req.Role),
			fmt.Sprintf("Admin %s changed role of %s from %s to %s", admin.Email, targetUser.Email, oldRole, req.Role),
			admin,
			alertData,
		)

		s.logger.Warn("security alert generated for role change",
			"alert_id", alert.ID,
			"alert_type", alert.Type,
			"severity", alert.Severity,
			"admin_id", adminID,
			"target_user_id", targetUserID,
		)

		// TODO: Send alert to security monitoring system
	}

	s.logger.Info("role change completed successfully",
		"admin_id", adminID,
		"target_user_id", targetUserID,
		"old_role", oldRole,
		"new_role", req.Role,
		"risk_level", validationResult.RiskLevel,
	)

	return nil
}

// UpdateUserStatus updates a user's status
func (s *AdminService) UpdateUserStatus(
	adminID, targetUserID uint,
	req *domain.UpdateUserStatusRequest,
	ipAddress, userAgent string,
) error {
	// Check admin authorization
	admin, err := s.userRepo.GetByID(adminID)
	if err != nil {
		return err
	}

	if !domain.IsAuthorizedForUserManagement(admin) {
		return domain.ErrNotAuthorized
	}

	// Get target user
	targetUser, err := s.userRepo.GetByID(targetUserID)
	if err != nil {
		return err
	}

	// Check if admin can manage this user
	if !domain.CanManageUser(admin, targetUser) {
		return domain.ErrCannotManageSelf
	}

	// Update status
	oldStatus := targetUser.Status
	err = s.userRepo.UpdateUserStatus(targetUserID, req.Status)
	if err != nil {
		s.logger.Error("failed to update user status", "admin_id", adminID, "target_user_id", targetUserID, "error", err)
		return err
	}

	// Create audit log
	if err := s.auditRepo.CreateAuditEntry(
		&adminID,
		&targetUserID,
		authdomain.AuditActionUserStatusChanged,
		authdomain.AuditLevelInfo,
		"admin",
		fmt.Sprintf("Status changed from %s to %s: %s", oldStatus, req.Status, req.Reason),
		ipAddress,
		userAgent,
		map[string]interface{}{
			"old_status": oldStatus,
			"new_status": req.Status,
			"reason":     req.Reason,
		},
	); err != nil {
		s.logger.Error("failed to create audit log for status change",
			"admin_id", adminID,
			"target_user_id", targetUserID,
			"error", err)
	}

	return nil
}

// UpdateUser updates user information (admin version)
func (s *AdminService) UpdateUser(
	adminID, targetUserID uint,
	req *domain.AdminUpdateUserRequest,
	ipAddress, userAgent string,
) error {
	// Check admin authorization
	admin, err := s.userRepo.GetByID(adminID)
	if err != nil {
		return err
	}

	if !domain.IsAuthorizedForUserManagement(admin) {
		return domain.ErrNotAuthorized
	}

	// Get target user
	targetUser, err := s.userRepo.GetByID(targetUserID)
	if err != nil {
		return err
	}

	// Check if admin can manage this user
	if !domain.CanManageUser(admin, targetUser) {
		return domain.ErrCannotManageSelf
	}

	// Check if email change is requested and if it already exists
	if req.Email != "" && req.Email != targetUser.Email {
		exists, err := s.userRepo.CheckEmailExists(req.Email, targetUserID)
		if err != nil {
			return err
		}
		if exists {
			return userdomain.ErrEmailAlreadyExists
		}
	}

	// Build changes for audit
	changes := s.buildUserChanges(targetUser, req)

	// Apply updates
	if req.FirstName != "" {
		targetUser.FirstName = req.FirstName
	}
	if req.LastName != "" {
		targetUser.LastName = req.LastName
	}
	if req.Email != "" {
		targetUser.Email = req.Email
		if req.EmailVerified != nil {
			targetUser.EmailVerified = *req.EmailVerified
		}
	}
	if req.Role != "" {
		targetUser.Role = req.Role
	}
	if req.Status != "" {
		targetUser.Status = req.Status
	}
	if req.Avatar != "" {
		targetUser.Avatar = req.Avatar
	}

	// Save changes
	err = s.userRepo.Update(targetUser)
	if err != nil {
		s.logger.Error("failed to update user", "admin_id", adminID, "target_user_id", targetUserID, "error", err)
		return err
	}

	// Create audit log
	if err := s.auditRepo.CreateAuditEntry(
		&adminID,
		&targetUserID,
		authdomain.AuditActionUserUpdated,
		authdomain.AuditLevelInfo,
		"admin",
		fmt.Sprintf("User updated by admin: %s. Reason: %s", changes, req.Reason),
		ipAddress,
		userAgent,
		map[string]interface{}{
			"changes": changes,
			"reason":  req.Reason,
		},
	); err != nil {
		s.logger.Error("failed to create audit log for admin user update",
			"admin_id", adminID,
			"target_user_id", targetUserID,
			"error", err)
	}

	return nil
}

// DeleteUsers deletes multiple users (soft or hard delete)
func (s *AdminService) DeleteUsers(
	adminID uint,
	req *domain.DeleteUserRequest,
	userIDs []uint,
	ipAddress, userAgent string,
) error {
	// Check admin authorization
	admin, err := s.userRepo.GetByID(adminID)
	if err != nil {
		return err
	}

	if !domain.IsAuthorizedForUserManagement(admin) {
		return domain.ErrNotAuthorized
	}

	// Get target users to check permissions and for audit
	targetUsers, err := s.userRepo.GetUsersByIDs(userIDs)
	if err != nil {
		return err
	}

	// Check permissions for each user
	for _, targetUser := range targetUsers {
		if !domain.CanManageUser(admin, targetUser) {
			return domain.ErrCannotManageSelf
		}
	}

	// Perform deletion
	var deleteErr error
	if req.Force {
		deleteErr = s.userRepo.HardDelete(userIDs)
	} else {
		deleteErr = s.userRepo.SoftDelete(userIDs)
	}

	if deleteErr != nil {
		s.logger.Error("failed to delete users",
			"admin_id", adminID,
			"user_ids", userIDs,
			"force", req.Force,
			"error", deleteErr)
		return deleteErr
	}

	// Create audit logs for each deleted user
	deleteType := "soft"
	if req.Force {
		deleteType = "hard"
	}

	for _, targetUser := range targetUsers {
		if err := s.auditRepo.CreateAuditEntry(
			&adminID,
			&targetUser.ID,
			authdomain.AuditActionUserDeleted,
			authdomain.AuditLevelWarning,
			"admin",
			fmt.Sprintf("User %s deleted (%s delete): %s", targetUser.Email, deleteType, req.Reason),
			ipAddress,
			userAgent,
			map[string]interface{}{
				"delete_type": deleteType,
				"reason":      req.Reason,
				"user_email":  targetUser.Email,
			},
		); err != nil {
			s.logger.Error("failed to create audit log for user deletion",
				"admin_id", adminID,
				"target_user_id", targetUser.ID,
				"error", err)
		}
	}

	return nil
}

// BulkUpdateUsers performs bulk operations on multiple users
func (s *AdminService) BulkUpdateUsers(
	adminID uint,
	req *domain.BulkUserActionRequest,
	ipAddress, userAgent string,
) (*domain.BulkActionResult, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// Check admin authorization
	admin, err := s.userRepo.GetByID(adminID)
	if err != nil {
		return nil, err
	}

	if !domain.IsAuthorizedForUserManagement(admin) {
		return nil, domain.ErrNotAuthorized
	}

	// Limit bulk operations
	if len(req.UserIDs) > 100 {
		return nil, domain.ErrTooManyUsers
	}

	// Get target users
	targetUsers, err := s.userRepo.GetUsersByIDs(req.UserIDs)
	if err != nil {
		return nil, err
	}

	result := &domain.BulkActionResult{
		TotalRequested: len(req.UserIDs),
		Results:        make([]domain.BulkActionItemResult, 0, len(req.UserIDs)),
	}

	// Process each user
	for _, userID := range req.UserIDs {
		itemResult := domain.BulkActionItemResult{
			UserID:  userID,
			Success: false,
		}

		// Find user in fetched users
		var targetUser *authdomain.User
		for _, user := range targetUsers {
			if user.ID == userID {
				targetUser = user
				break
			}
		}

		if targetUser == nil {
			itemResult.Error = "user not found"
			result.Results = append(result.Results, itemResult)
			result.Failed++
			continue
		}

		// Check if admin can manage this user
		if !domain.CanManageUser(admin, targetUser) {
			itemResult.Error = "cannot manage this user"
			result.Results = append(result.Results, itemResult)
			result.Failed++
			continue
		}

		// Perform action
		var actionErr error
		var actionDescription string

		switch req.Action {
		case domain.BulkActionActivate:
			actionErr = s.userRepo.UpdateUserStatus(userID, authdomain.StatusActive)
			actionDescription = "User activated"

		case domain.BulkActionDeactivate:
			actionErr = s.userRepo.UpdateUserStatus(userID, authdomain.StatusInactive)
			actionDescription = "User deactivated"

		case domain.BulkActionSuspend:
			actionErr = s.userRepo.UpdateUserStatus(userID, authdomain.StatusSuspended)
			actionDescription = "User suspended"

		case domain.BulkActionDelete:
			actionErr = s.userRepo.SoftDelete([]uint{userID})
			actionDescription = "User deleted"

		case domain.BulkActionRoleChange:
			if req.Role != nil {
				actionErr = s.userRepo.UpdateUserRole(userID, *req.Role)
				actionDescription = fmt.Sprintf("Role changed to %s", *req.Role)
			} else {
				actionErr = fmt.Errorf("role not specified")
			}
		}

		if actionErr != nil {
			itemResult.Error = actionErr.Error()
			result.Failed++
		} else {
			itemResult.Success = true
			result.Successful++

			// Create audit log
			if err := s.auditRepo.CreateAuditEntry(
				&adminID,
				&userID,
				s.getAuditActionForBulkAction(req.Action),
				authdomain.AuditLevelInfo,
				"admin",
				fmt.Sprintf("Bulk operation: %s. Reason: %s", actionDescription, req.Reason),
				ipAddress,
				userAgent,
				map[string]interface{}{
					"bulk_action":    req.Action,
					"reason":         req.Reason,
					"target_email":   targetUser.Email,
					"target_user_id": userID,
				},
			); err != nil {
				s.logger.Error("failed to create audit log for bulk operation",
					"admin_id", adminID,
					"target_user_id", userID,
					"error", err)
			}
		}

		result.Results = append(result.Results, itemResult)
	}

	return result, nil
}

// GetAdminStats retrieves admin dashboard statistics
func (s *AdminService) GetAdminStats(adminID uint) (*domain.AdminStatsResponse, error) {
	// Check admin authorization
	admin, err := s.userRepo.GetByID(adminID)
	if err != nil {
		return nil, err
	}

	if !domain.IsAuthorizedForUserManagement(admin) {
		return nil, domain.ErrNotAuthorized
	}

	// Get basic stats
	stats, err := s.userRepo.GetAdminStats()
	if err != nil {
		s.logger.Error("failed to get admin stats", "admin_id", adminID, "error", err)
		return nil, err
	}

	// Get user growth data
	growthData, err := s.userRepo.GetUserGrowthData(30)
	if err != nil {
		s.logger.Error("failed to get user growth data", "admin_id", adminID, "error", err)
		// Continue with empty growth data rather than failing
		growthData = []repository.UserGrowthDataPoint{}
	}

	// Convert growth data
	userGrowth := make([]domain.UserGrowthData, len(growthData))
	for i, data := range growthData {
		userGrowth[i] = domain.UserGrowthData{
			Date:  data.Date,
			Count: data.Count,
		}
	}

	return &domain.AdminStatsResponse{
		TotalUsers:       int(stats.TotalUsers),
		ActiveUsers:      int(stats.ActiveUsers),
		InactiveUsers:    int(stats.InactiveUsers),
		SuspendedUsers:   int(stats.SuspendedUsers),
		AdminUsers:       int(stats.AdminUsers),
		NewUsersToday:    int(stats.NewUsersToday),
		NewUsersThisWeek: int(stats.NewUsersThisWeek),
		UserGrowth:       userGrowth,
	}, nil
}

// GetAuditLogs retrieves audit logs with filtering
func (s *AdminService) GetAuditLogs(
	adminID uint,
	req *domain.AdminAuditLogRequest,
) (*domain.AdminAuditLogResponse, error) {
	// Check admin authorization
	admin, err := s.userRepo.GetByID(adminID)
	if err != nil {
		return nil, err
	}

	if !domain.IsAuthorizedForUserManagement(admin) {
		return nil, domain.ErrNotAuthorized
	}

	// Validate date range
	if req.DateFrom != nil && req.DateTo != nil && req.DateFrom.After(*req.DateTo) {
		return nil, domain.ErrInvalidDateRange
	}

	// Get audit logs
	logs, total, err := s.auditRepo.List(req)
	if err != nil {
		s.logger.Error("failed to get audit logs", "admin_id", adminID, "error", err)
		return nil, err
	}

	// Convert to enhanced format
	enhancedLogs := make([]*domain.EnhancedAuditLogEntry, len(logs))
	for i, log := range logs {
		enhancedLogs[i] = domain.ToEnhancedAuditLogEntry(log)
	}

	// Build pagination
	totalPages := (total + req.PageSize - 1) / req.PageSize
	pagination := userdomain.Pagination{
		Page:       req.Page,
		PageSize:   req.PageSize,
		Total:      total,
		TotalPages: totalPages,
		HasNext:    req.Page < totalPages,
		HasPrev:    req.Page > 1,
	}

	return &domain.AdminAuditLogResponse{
		Logs:       enhancedLogs,
		Pagination: pagination,
	}, nil
}

// Helper methods

// buildUserChanges builds a human-readable string of user changes
func (s *AdminService) buildUserChanges(current *authdomain.User, req *domain.AdminUpdateUserRequest) string {
	var changes []string

	if req.FirstName != "" && current.FirstName != req.FirstName {
		changes = append(changes, fmt.Sprintf("first name: '%s' -> '%s'", current.FirstName, req.FirstName))
	}

	if req.LastName != "" && current.LastName != req.LastName {
		changes = append(changes, fmt.Sprintf("last name: '%s' -> '%s'", current.LastName, req.LastName))
	}

	if req.Email != "" && current.Email != req.Email {
		changes = append(changes, fmt.Sprintf("email: '%s' -> '%s'", current.Email, req.Email))
	}

	if req.Role != "" && current.Role != req.Role {
		changes = append(changes, fmt.Sprintf("role: '%s' -> '%s'", current.Role, req.Role))
	}

	if req.Status != "" && current.Status != req.Status {
		changes = append(changes, fmt.Sprintf("status: '%s' -> '%s'", current.Status, req.Status))
	}

	if req.EmailVerified != nil && current.EmailVerified != *req.EmailVerified {
		changes = append(changes, fmt.Sprintf("email verified: %t -> %t", current.EmailVerified, *req.EmailVerified))
	}

	if req.Avatar != "" && current.Avatar != req.Avatar {
		changes = append(changes, "avatar updated")
	}

	if len(changes) == 0 {
		return "no changes"
	}

	return fmt.Sprintf("[%s]", strings.Join(changes, ", "))
}

// getAuditActionForBulkAction maps bulk actions to audit actions
func (s *AdminService) getAuditActionForBulkAction(action domain.BulkActionType) authdomain.AuditAction {
	switch action {
	case domain.BulkActionActivate, domain.BulkActionDeactivate, domain.BulkActionSuspend:
		return authdomain.AuditActionUserStatusChanged
	case domain.BulkActionDelete:
		return authdomain.AuditActionUserDeleted
	case domain.BulkActionRoleChange:
		return authdomain.AuditActionUserRoleChanged
	default:
		return authdomain.AuditActionUserUpdated
	}
}

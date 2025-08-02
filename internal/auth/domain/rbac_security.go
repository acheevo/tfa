package domain

import (
	"fmt"
	"strings"
	"time"
)

// Security validation and escalation prevention for RBAC

// Risk levels
const (
	RiskLevelLow      = "low"
	RiskLevelMedium   = "medium"
	RiskLevelHigh     = "high"
	RiskLevelCritical = "critical"
)

// RoleChangeSecurityCheck performs security validation for role changes
type RoleChangeSecurityCheck struct {
	AdminID       uint     `json:"admin_id"`
	AdminRole     UserRole `json:"admin_role"`
	TargetID      uint     `json:"target_id"`
	TargetRole    UserRole `json:"target_role"`
	NewRole       UserRole `json:"new_role"`
	Reason        string   `json:"reason"`
	IPAddress     string   `json:"ip_address"`
	UserAgent     string   `json:"user_agent"`
	SessionID     string   `json:"session_id,omitempty"`
	RequestSource string   `json:"request_source"` // "web", "api", "cli", etc.
}

// SecurityValidationResult represents the result of security validation
type SecurityValidationResult struct {
	Valid                 bool     `json:"valid"`
	Errors                []string `json:"errors,omitempty"`
	Warnings              []string `json:"warnings,omitempty"`
	RiskLevel             string   `json:"risk_level"` // "low", "medium", "high", "critical"
	RequiresSecondaryAuth bool     `json:"requires_secondary_auth"`
	AuditFlags            []string `json:"audit_flags,omitempty"`
}

// ValidateRoleChange performs comprehensive security validation for role changes
func ValidateRoleChange(check *RoleChangeSecurityCheck) *SecurityValidationResult {
	result := &SecurityValidationResult{
		Valid:      true,
		Errors:     []string{},
		Warnings:   []string{},
		RiskLevel:  RiskLevelLow,
		AuditFlags: []string{},
	}

	// 1. Prevent self-role escalation
	if check.AdminID == check.TargetID {
		result.Valid = false
		result.Errors = append(result.Errors, "administrators cannot modify their own role")
		result.RiskLevel = RiskLevelCritical
		result.AuditFlags = append(result.AuditFlags, "self_role_modification_attempt")
	}

	// 2. Validate admin has permission to change roles
	if !HasPermission(check.AdminRole, PermissionUserManage) {
		result.Valid = false
		result.Errors = append(result.Errors, "insufficient permissions to modify user roles")
		result.RiskLevel = RiskLevelHigh
		result.AuditFlags = append(result.AuditFlags, "unauthorized_role_change_attempt")
	}

	// 3. Validate role transition is allowed
	if !isValidRoleTransition(check.TargetRole, check.NewRole) {
		result.Valid = false
		result.Errors = append(result.Errors,
			fmt.Sprintf("invalid role transition from %s to %s", check.TargetRole, check.NewRole))
		result.RiskLevel = RiskLevelHigh
		result.AuditFlags = append(result.AuditFlags, "invalid_role_transition")
	}

	// 4. Check for privilege escalation patterns
	if isPrivilegeEscalation(check.TargetRole, check.NewRole) {
		result.RiskLevel = RiskLevelHigh
		result.RequiresSecondaryAuth = true
		result.Warnings = append(result.Warnings, "role change involves privilege escalation")
		result.AuditFlags = append(result.AuditFlags, "privilege_escalation")
	}

	// 5. Validate reason is provided and meaningful
	if strings.TrimSpace(check.Reason) == "" {
		result.Valid = false
		result.Errors = append(result.Errors, "reason for role change is required")
	} else if len(strings.TrimSpace(check.Reason)) < 10 {
		result.Warnings = append(result.Warnings, "reason for role change is very brief")
		result.AuditFlags = append(result.AuditFlags, "brief_reason")
	}

	// 6. Check for suspicious patterns in reason
	suspiciousPatterns := []string{"test", "temp", "temporary", "quick", "urgent"}
	reasonLower := strings.ToLower(check.Reason)
	for _, pattern := range suspiciousPatterns {
		if strings.Contains(reasonLower, pattern) {
			result.Warnings = append(result.Warnings, fmt.Sprintf("reason contains potentially suspicious term: %s", pattern))
			result.AuditFlags = append(result.AuditFlags, "suspicious_reason")
			if result.RiskLevel == RiskLevelLow {
				result.RiskLevel = RiskLevelMedium
			}
			break
		}
	}

	// 7. Validate IP address and user agent
	if check.IPAddress == "" {
		result.Warnings = append(result.Warnings, "IP address not recorded")
		result.AuditFlags = append(result.AuditFlags, "missing_ip")
	}

	if check.UserAgent == "" {
		result.Warnings = append(result.Warnings, "user agent not recorded")
		result.AuditFlags = append(result.AuditFlags, "missing_user_agent")
	}

	// 8. Additional security checks for admin role assignments
	if check.NewRole == RoleAdmin {
		result.RiskLevel = RiskLevelHigh
		result.RequiresSecondaryAuth = true
		result.AuditFlags = append(result.AuditFlags, "admin_role_assignment")

		// Extra validation for admin assignments
		if !strings.Contains(strings.ToLower(check.Reason), "admin") &&
			!strings.Contains(strings.ToLower(check.Reason), "administrator") {
			result.Warnings = append(result.Warnings, "admin role assignment without explicit admin-related reason")
		}
	}

	return result
}

// isValidRoleTransition checks if a role transition is allowed
func isValidRoleTransition(from, to UserRole) bool {
	// Define allowed transitions
	allowedTransitions := map[UserRole][]UserRole{
		RoleUser:  {RoleAdmin},
		RoleAdmin: {RoleUser},
	}

	validTransitions, exists := allowedTransitions[from]
	if !exists {
		return false
	}

	for _, validTo := range validTransitions {
		if to == validTo {
			return true
		}
	}

	return false
}

// isPrivilegeEscalation checks if the role change is a privilege escalation
func isPrivilegeEscalation(from, to UserRole) bool {
	// Define role hierarchy (higher number = more privileges)
	roleHierarchy := map[UserRole]int{
		RoleUser:  1,
		RoleAdmin: 2,
	}

	fromLevel, fromExists := roleHierarchy[from]
	toLevel, toExists := roleHierarchy[to]

	if !fromExists || !toExists {
		return true // Unknown role is considered escalation
	}

	return toLevel > fromLevel
}

// RoleChangeAuditEntry represents a comprehensive audit entry for role changes
type RoleChangeAuditEntry struct {
	ID                    uint                      `json:"id"`
	AdminID               uint                      `json:"admin_id"`
	AdminEmail            string                    `json:"admin_email"`
	AdminRole             UserRole                  `json:"admin_role"`
	TargetID              uint                      `json:"target_id"`
	TargetEmail           string                    `json:"target_email"`
	PreviousRole          UserRole                  `json:"previous_role"`
	NewRole               UserRole                  `json:"new_role"`
	Reason                string                    `json:"reason"`
	ValidationResult      *SecurityValidationResult `json:"validation_result"`
	IPAddress             string                    `json:"ip_address"`
	UserAgent             string                    `json:"user_agent"`
	RequestSource         string                    `json:"request_source"`
	SessionID             string                    `json:"session_id,omitempty"`
	SecondaryAuthRequired bool                      `json:"secondary_auth_required"`
	SecondaryAuthPassed   bool                      `json:"secondary_auth_passed,omitempty"`
	Status                string                    `json:"status"` // "pending", "approved", "rejected", "completed"
	CreatedAt             time.Time                 `json:"created_at"`
	CompletedAt           *time.Time                `json:"completed_at,omitempty"`
	Notes                 string                    `json:"notes,omitempty"`
}

// CreateRoleChangeAuditEntry creates a comprehensive audit entry
func CreateRoleChangeAuditEntry(
	adminUser *User,
	targetUser *User,
	newRole UserRole,
	reason string,
	ipAddress, userAgent, requestSource string,
	validationResult *SecurityValidationResult,
) *RoleChangeAuditEntry {
	return &RoleChangeAuditEntry{
		AdminID:               adminUser.ID,
		AdminEmail:            adminUser.Email,
		AdminRole:             adminUser.Role,
		TargetID:              targetUser.ID,
		TargetEmail:           targetUser.Email,
		PreviousRole:          targetUser.Role,
		NewRole:               newRole,
		Reason:                reason,
		ValidationResult:      validationResult,
		IPAddress:             ipAddress,
		UserAgent:             userAgent,
		RequestSource:         requestSource,
		SecondaryAuthRequired: validationResult.RequiresSecondaryAuth,
		Status:                determineInitialStatus(validationResult),
		CreatedAt:             time.Now(),
	}
}

// determineInitialStatus determines the initial status based on validation
func determineInitialStatus(result *SecurityValidationResult) string {
	if !result.Valid {
		return "rejected"
	}
	if result.RequiresSecondaryAuth {
		return "pending"
	}
	return "approved"
}

// Role change monitoring and alerting

// RoleChangeMonitor handles monitoring and alerting for role changes
type RoleChangeMonitor struct {
	AlertThresholds AlertThresholds `json:"alert_thresholds"`
}

// AlertThresholds defines thresholds for triggering alerts
type AlertThresholds struct {
	AdminRoleAssignmentsPerHour int           `json:"admin_role_assignments_per_hour"`
	RoleChangesPerAdmin         int           `json:"role_changes_per_admin"`
	TimeWindow                  time.Duration `json:"time_window"`
	HighRiskActionsPerDay       int           `json:"high_risk_actions_per_day"`
}

// DefaultAlertThresholds returns default alert thresholds
func DefaultAlertThresholds() AlertThresholds {
	return AlertThresholds{
		AdminRoleAssignmentsPerHour: 5,
		RoleChangesPerAdmin:         10,
		TimeWindow:                  24 * time.Hour,
		HighRiskActionsPerDay:       3,
	}
}

// SecurityAlert represents a security alert
type SecurityAlert struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Severity    string                 `json:"severity"` // "low", "medium", "high", "critical"
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	AdminID     uint                   `json:"admin_id"`
	AdminEmail  string                 `json:"admin_email"`
	Data        map[string]interface{} `json:"data"`
	CreatedAt   time.Time              `json:"created_at"`
	Resolved    bool                   `json:"resolved"`
	ResolvedAt  *time.Time             `json:"resolved_at,omitempty"`
	Notes       string                 `json:"notes,omitempty"`
}

// GenerateSecurityAlert creates a security alert for suspicious activity
func GenerateSecurityAlert(
	alertType, severity, title, description string,
	adminUser *User,
	data map[string]interface{},
) *SecurityAlert {
	return &SecurityAlert{
		ID:          fmt.Sprintf("%d_%s_%d", adminUser.ID, alertType, time.Now().Unix()),
		Type:        alertType,
		Severity:    severity,
		Title:       title,
		Description: description,
		AdminID:     adminUser.ID,
		AdminEmail:  adminUser.Email,
		Data:        data,
		CreatedAt:   time.Now(),
		Resolved:    false,
	}
}

// Security compliance helpers

// ComplianceRequirement represents a compliance requirement for role changes
type ComplianceRequirement struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Required    bool   `json:"required"`
	Met         bool   `json:"met"`
	Details     string `json:"details,omitempty"`
}

// CheckComplianceRequirements checks compliance requirements for role changes
func CheckComplianceRequirements(entry *RoleChangeAuditEntry) []ComplianceRequirement {
	requirements := []ComplianceRequirement{
		{
			Name:        "Audit Trail",
			Description: "Complete audit trail of role change",
			Required:    true,
			Met:         true, // Always met if entry exists
		},
		{
			Name:        "Justification",
			Description: "Valid business justification provided",
			Required:    true,
			Met:         len(strings.TrimSpace(entry.Reason)) >= 10,
			Details:     "Reason must be at least 10 characters",
		},
		{
			Name:        "Administrator Authentication",
			Description: "Administrator properly authenticated",
			Required:    true,
			Met:         entry.AdminID > 0,
		},
		{
			Name:        "IP Address Logging",
			Description: "Source IP address recorded",
			Required:    true,
			Met:         entry.IPAddress != "",
		},
		{
			Name:        "Secondary Authentication",
			Description: "Secondary authentication for high-risk changes",
			Required:    entry.SecondaryAuthRequired,
			Met:         !entry.SecondaryAuthRequired || entry.SecondaryAuthPassed,
			Details:     "Required for privilege escalation",
		},
	}

	return requirements
}

// GenerateComplianceReport generates a compliance report for role changes
func GenerateComplianceReport(entries []*RoleChangeAuditEntry) map[string]interface{} {
	report := map[string]interface{}{
		"total_role_changes":     len(entries),
		"compliance_summary":     map[string]int{},
		"failed_requirements":    []string{},
		"high_risk_changes":      0,
		"privileged_escalations": 0,
		"generated_at":           time.Now(),
	}

	for _, entry := range entries {
		requirements := CheckComplianceRequirements(entry)

		if entry.ValidationResult != nil && entry.ValidationResult.RiskLevel == RiskLevelHigh {
			if count, ok := report["high_risk_changes"].(int); ok {
				report["high_risk_changes"] = count + 1
			}
		}

		if isPrivilegeEscalation(entry.PreviousRole, entry.NewRole) {
			if count, ok := report["privileged_escalations"].(int); ok {
				report["privileged_escalations"] = count + 1
			}
		}

		for _, req := range requirements {
			if req.Required && !req.Met {
				if failures, ok := report["failed_requirements"].([]string); ok {
					report["failed_requirements"] = append(
						failures,
						fmt.Sprintf("%s: %s", entry.TargetEmail, req.Name),
					)
				}
			}
		}
	}

	return report
}

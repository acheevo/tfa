package domain

import "errors"

// Admin management errors
var (
	ErrNotAuthorized     = errors.New("not authorized for admin operations")
	ErrCannotManageSelf  = errors.New("cannot manage own account through admin interface")
	ErrBulkActionFailed  = errors.New("bulk action failed")
	ErrAuditLogNotFound  = errors.New("audit log not found")
	ErrSystemHealthCheck = errors.New("system health check failed")
	ErrInvalidDateRange  = errors.New("invalid date range")
	ErrTooManyUsers      = errors.New("too many users selected for bulk action")
)

// IsAdminError checks if the error is an admin management error
func IsAdminError(err error) bool {
	return err == ErrNotAuthorized ||
		err == ErrCannotManageSelf ||
		err == ErrBulkActionFailed ||
		err == ErrAuditLogNotFound ||
		err == ErrSystemHealthCheck ||
		err == ErrInvalidDateRange ||
		err == ErrTooManyUsers
}

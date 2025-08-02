package domain

import "errors"

// User management errors
var (
	ErrUserNotFound          = errors.New("user not found")
	ErrInvalidRequest        = errors.New("invalid request")
	ErrUnauthorized          = errors.New("unauthorized")
	ErrForbidden             = errors.New("forbidden")
	ErrEmailAlreadyExists    = errors.New("email already exists")
	ErrCannotUpdateOwnRole   = errors.New("cannot update own role")
	ErrCannotUpdateOwnStatus = errors.New("cannot update own status")
	ErrInvalidPreferences    = errors.New("invalid preferences")
	ErrPreferencesNotFound   = errors.New("preferences not found")
	ErrProfileUpdateFailed   = errors.New("profile update failed")
)

// IsUserError checks if the error is a user management error
func IsUserError(err error) bool {
	return err == ErrUserNotFound ||
		err == ErrInvalidRequest ||
		err == ErrUnauthorized ||
		err == ErrForbidden ||
		err == ErrEmailAlreadyExists ||
		err == ErrCannotUpdateOwnRole ||
		err == ErrCannotUpdateOwnStatus ||
		err == ErrInvalidPreferences ||
		err == ErrPreferencesNotFound ||
		err == ErrProfileUpdateFailed
}

package domain

import "errors"

// Authentication errors
var (
	ErrInvalidCredentials      = errors.New("invalid email or password")
	ErrUserNotFound            = errors.New("user not found")
	ErrUserAlreadyExists       = errors.New("user already exists")
	ErrEmailNotVerified        = errors.New("email not verified")
	ErrUserInactive            = errors.New("user account is inactive")
	ErrInvalidToken            = errors.New("invalid token")
	ErrTokenExpired            = errors.New("token expired")
	ErrTokenNotFound           = errors.New("token not found")
	ErrTokenAlreadyUsed        = errors.New("token already used")
	ErrPasswordsDoNotMatch     = errors.New("passwords do not match")
	ErrWeakPassword            = errors.New("password is too weak")
	ErrInvalidEmail            = errors.New("invalid email address")
	ErrEmailVerificationFailed = errors.New("email verification failed")
	ErrPasswordResetFailed     = errors.New("password reset failed")
	ErrUnauthorized            = errors.New("unauthorized")
	ErrForbidden               = errors.New("forbidden")
)

// IsValidationError checks if the error is a validation error
func IsValidationError(err error) bool {
	return err == ErrInvalidCredentials ||
		err == ErrUserAlreadyExists ||
		err == ErrPasswordsDoNotMatch ||
		err == ErrWeakPassword ||
		err == ErrInvalidEmail
}

// IsAuthError checks if the error is an authentication error
func IsAuthError(err error) bool {
	return err == ErrInvalidCredentials ||
		err == ErrEmailNotVerified ||
		err == ErrUserInactive ||
		err == ErrUnauthorized ||
		err == ErrForbidden
}

// IsTokenError checks if the error is a token-related error
func IsTokenError(err error) bool {
	return err == ErrInvalidToken ||
		err == ErrTokenExpired ||
		err == ErrTokenNotFound ||
		err == ErrTokenAlreadyUsed
}

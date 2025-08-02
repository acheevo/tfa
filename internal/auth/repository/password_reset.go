package repository

import (
	"time"

	"gorm.io/gorm"

	"github.com/acheevo/tfa/internal/auth/domain"
)

// PasswordResetRepository handles database operations for password reset tokens
type PasswordResetRepository struct {
	db *gorm.DB
}

// NewPasswordResetRepository creates a new password reset repository
func NewPasswordResetRepository(db *gorm.DB) *PasswordResetRepository {
	return &PasswordResetRepository{
		db: db,
	}
}

// Create creates a new password reset token
func (r *PasswordResetRepository) Create(reset *domain.PasswordReset) error {
	return r.db.Create(reset).Error
}

// GetByToken gets a password reset by token
func (r *PasswordResetRepository) GetByToken(token string) (*domain.PasswordReset, error) {
	var reset domain.PasswordReset
	err := r.db.Where("token = ? AND used = false AND expires_at > ?", token, time.Now()).First(&reset).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, domain.ErrTokenNotFound
		}
		return nil, err
	}
	return &reset, nil
}

// GetByEmail gets all password reset tokens for an email
func (r *PasswordResetRepository) GetByEmail(email string) ([]*domain.PasswordReset, error) {
	var resets []*domain.PasswordReset
	err := r.db.Where("email = ?", email).Find(&resets).Error
	return resets, err
}

// MarkAsUsed marks a password reset token as used
func (r *PasswordResetRepository) MarkAsUsed(token string) error {
	return r.db.Model(&domain.PasswordReset{}).
		Where("token = ?", token).
		Update("used", true).Error
}

// Delete deletes a password reset token
func (r *PasswordResetRepository) Delete(token string) error {
	return r.db.Where("token = ?", token).Delete(&domain.PasswordReset{}).Error
}

// DeleteByEmail deletes all password reset tokens for an email
func (r *PasswordResetRepository) DeleteByEmail(email string) error {
	return r.db.Where("email = ?", email).Delete(&domain.PasswordReset{}).Error
}

// DeleteExpired deletes all expired password reset tokens
func (r *PasswordResetRepository) DeleteExpired() error {
	return r.db.Where("expires_at < ?", time.Now()).Delete(&domain.PasswordReset{}).Error
}

// DeleteUsed deletes all used password reset tokens
func (r *PasswordResetRepository) DeleteUsed() error {
	return r.db.Where("used = true").Delete(&domain.PasswordReset{}).Error
}

// Update updates a password reset token
func (r *PasswordResetRepository) Update(reset *domain.PasswordReset) error {
	return r.db.Save(reset).Error
}

// GetValidTokensCount returns the count of valid (unused and not expired) tokens for an email
func (r *PasswordResetRepository) GetValidTokensCount(email string) (int64, error) {
	var count int64
	err := r.db.Model(&domain.PasswordReset{}).
		Where("email = ? AND used = false AND expires_at > ?", email, time.Now()).
		Count(&count).Error
	return count, err
}

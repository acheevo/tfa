package repository

import (
	"time"

	"gorm.io/gorm"

	"github.com/acheevo/tfa/internal/auth/domain"
)

// RefreshTokenRepository handles database operations for refresh tokens
type RefreshTokenRepository struct {
	db *gorm.DB
}

// NewRefreshTokenRepository creates a new refresh token repository
func NewRefreshTokenRepository(db *gorm.DB) *RefreshTokenRepository {
	return &RefreshTokenRepository{
		db: db,
	}
}

// Create creates a new refresh token
func (r *RefreshTokenRepository) Create(token *domain.RefreshToken) error {
	return r.db.Create(token).Error
}

// GetByToken gets a refresh token by token string
func (r *RefreshTokenRepository) GetByToken(token string) (*domain.RefreshToken, error) {
	var refreshToken domain.RefreshToken
	err := r.db.Where("token = ?", token).First(&refreshToken).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, domain.ErrTokenNotFound
		}
		return nil, err
	}
	return &refreshToken, nil
}

// GetByUserID gets all refresh tokens for a user
func (r *RefreshTokenRepository) GetByUserID(userID uint) ([]*domain.RefreshToken, error) {
	var tokens []*domain.RefreshToken
	err := r.db.Where("user_id = ?", userID).Find(&tokens).Error
	return tokens, err
}

// Delete deletes a refresh token
func (r *RefreshTokenRepository) Delete(token string) error {
	return r.db.Where("token = ?", token).Delete(&domain.RefreshToken{}).Error
}

// DeleteByUserID deletes all refresh tokens for a user
func (r *RefreshTokenRepository) DeleteByUserID(userID uint) error {
	return r.db.Where("user_id = ?", userID).Delete(&domain.RefreshToken{}).Error
}

// DeleteExpired deletes all expired refresh tokens
func (r *RefreshTokenRepository) DeleteExpired() error {
	return r.db.Where("expires_at < ?", time.Now()).Delete(&domain.RefreshToken{}).Error
}

// Update updates a refresh token
func (r *RefreshTokenRepository) Update(token *domain.RefreshToken) error {
	return r.db.Save(token).Error
}

// GetActiveTokensCount returns the count of active tokens for a user
func (r *RefreshTokenRepository) GetActiveTokensCount(userID uint) (int64, error) {
	var count int64
	err := r.db.Model(&domain.RefreshToken{}).
		Where("user_id = ? AND expires_at > ?", userID, time.Now()).
		Count(&count).Error
	return count, err
}

// DeleteOldestTokensForUser deletes the oldest tokens for a user, keeping only the specified limit
func (r *RefreshTokenRepository) DeleteOldestTokensForUser(userID uint, keepCount int) error {
	// Get tokens ordered by creation date (oldest first)
	var tokens []*domain.RefreshToken
	err := r.db.Where("user_id = ?", userID).
		Order("created_at ASC").
		Find(&tokens).Error
	if err != nil {
		return err
	}

	// If we have more tokens than the limit, delete the oldest ones
	if len(tokens) > keepCount {
		tokensToDelete := tokens[:len(tokens)-keepCount]
		for _, token := range tokensToDelete {
			if err := r.db.Delete(token).Error; err != nil {
				return err
			}
		}
	}

	return nil
}

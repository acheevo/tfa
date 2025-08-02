package repository

import (
	"time"

	"gorm.io/gorm"

	"github.com/acheevo/tfa/internal/auth/domain"
)

// UserRepository handles database operations for users
type UserRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{
		db: db,
	}
}

// Create creates a new user
func (r *UserRepository) Create(user *domain.User) error {
	return r.db.Create(user).Error
}

// GetByID gets a user by ID
func (r *UserRepository) GetByID(id uint) (*domain.User, error) {
	var user domain.User
	err := r.db.First(&user, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

// GetByEmail gets a user by email
func (r *UserRepository) GetByEmail(email string) (*domain.User, error) {
	var user domain.User
	err := r.db.Where("email = ?", email).First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

// GetByEmailVerifyToken gets a user by email verification token
func (r *UserRepository) GetByEmailVerifyToken(token string) (*domain.User, error) {
	var user domain.User
	err := r.db.Where("email_verify_token = ?", token).First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

// Update updates a user
func (r *UserRepository) Update(user *domain.User) error {
	return r.db.Save(user).Error
}

// UpdateLastLogin updates the last login time for a user
func (r *UserRepository) UpdateLastLogin(userID uint) error {
	now := time.Now()
	return r.db.Model(&domain.User{}).Where("id = ?", userID).Update("last_login_at", &now).Error
}

// Delete soft deletes a user
func (r *UserRepository) Delete(id uint) error {
	return r.db.Delete(&domain.User{}, id).Error
}

// ExistsByEmail checks if a user exists by email
func (r *UserRepository) ExistsByEmail(email string) (bool, error) {
	var count int64
	err := r.db.Model(&domain.User{}).Where("email = ?", email).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// CountUsers returns the total number of users
func (r *UserRepository) CountUsers() (int64, error) {
	var count int64
	err := r.db.Model(&domain.User{}).Count(&count).Error
	return count, err
}

// GetUsers returns a paginated list of users
func (r *UserRepository) GetUsers(limit, offset int) ([]*domain.User, error) {
	var users []*domain.User
	err := r.db.Limit(limit).Offset(offset).Find(&users).Error
	return users, err
}

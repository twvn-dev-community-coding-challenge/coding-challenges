package repository

import (
	"context"
	"errors"

	"github.com/dotdak/sms-otp/models"
	"gorm.io/gorm"
)

// UserRepository defines the interface for user data access operations
type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	Delete(ctx context.Context, id uint) error
	GetByID(ctx context.Context, id uint) (*models.User, error)
	GetByPhoneNumber(ctx context.Context, phone string) (*models.User, error)
	GetByUsername(ctx context.Context, username string) (*models.User, error)
	Update(ctx context.Context, user *models.User) error
}

type postgresUserRepository struct {
	db *gorm.DB
}

// NewUserRepository returns a new instance of UserRepository backed by PostgreSQL
func NewUserRepository(db *gorm.DB) UserRepository {
	return &postgresUserRepository{db: db}
}

func (r *postgresUserRepository) Create(ctx context.Context, user *models.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *postgresUserRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&models.User{}, id).Error
}

func (r *postgresUserRepository) GetByID(ctx context.Context, id uint) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).First(&user, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // Or return a specific ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (r *postgresUserRepository) GetByPhoneNumber(ctx context.Context, phone string) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).Where(&models.User{PhoneNumber: phone}).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // Or return a specific ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (r *postgresUserRepository) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).Where(&models.User{Username: username}).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // Or return a specific ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (r *postgresUserRepository) Update(ctx context.Context, user *models.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

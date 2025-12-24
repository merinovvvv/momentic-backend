package repository

import (
	"context"

	"github.com/merinovvvv/momentic-backend/models"
	"gorm.io/gorm"
)

type UserRepository interface {
	UpdateAvatarPath(ctx context.Context, userID uint64, avatarPath string) error
}

type userRepositoryImpl struct {
	DB *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepositoryImpl{DB: db}
}

func (r *userRepositoryImpl) UpdateAvatarPath(ctx context.Context, userID uint64, avatarPath string) error {
	return r.DB.WithContext(ctx).Model(&models.User{}).
		Where("user_id = ?", userID).
		Update("avatar_filepath", avatarPath).Error
}

package repository

import (
	"context"

	"github.com/merinovvvv/momentic-backend/models"

	"gorm.io/gorm"
)

type CommentRepository interface {
	Create(ctx context.Context, comment *models.Comment) error
}

type commentRepositoryImpl struct {
	DB *gorm.DB
}

func NewCommentRepository(db *gorm.DB) CommentRepository {
	return &commentRepositoryImpl{DB: db}
}

func (r *commentRepositoryImpl) Create(ctx context.Context, comment *models.Comment) error {
	// Сохраняем комментарий и загружаем связанные данные пользователя (Nickname)
	if err := r.DB.WithContext(ctx).Create(comment).Error; err != nil {
		return err
	}
	// Загружаем Nickname автора для отображения в чате
	return r.DB.WithContext(ctx).Preload("User").First(comment, comment.CommentID).Error
}

package repository

import (
	"context"

	"github.com/merinovvvv/momentic-backend/models"

	"gorm.io/gorm"
)

type CommentRepository interface {
	Create(ctx context.Context, comment *models.Comment) error
	GetByVideoID(ctx context.Context, videoID int64) ([]*models.Comment, error)
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
	if err := r.DB.WithContext(ctx).Preload("User").First(comment, comment.CommentID).Error; err != nil {
		return err
	}

	// 3. Заполняем поле Nickname (так как оно помечено gorm:"-")
	comment.Nickname = comment.User.Nickname

	return nil
}

func (r *commentRepositoryImpl) GetByVideoID(ctx context.Context, videoID int64) ([]*models.Comment, error) {
	var comments []*models.Comment

	// Добавляем Preload сюда тоже, чтобы при получении списка были никнеймы
	err := r.DB.WithContext(ctx).
		Preload("User").
		Where("video_id = ?", videoID).
		Find(&comments).Error

	if err != nil {
		return nil, err
	}

	// Заполняем виртуальные поля Nickname для всего списка
	for _, c := range comments {
		c.Nickname = c.User.Nickname
	}

	return comments, nil
}

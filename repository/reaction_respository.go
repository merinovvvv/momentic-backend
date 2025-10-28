package repository

import (
	"context"

	"github.com/merinovvvv/momentic-backend/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ReactionRepository определяет интерфейс для работы с реакциями в БД
type ReactionRepository interface {
	SetReaction(ctx context.Context, userID int64, videoID int64, kind models.ReactionKind) error
	DeleteReaction(ctx context.Context, userID int64, videoID int64) (int64, error)
	GetReactingUsers(ctx context.Context, videoID int64) ([]models.ReactingUserResponse, error)
}

// reactionRepositoryImpl - реализация ReactionRepository
type reactionRepositoryImpl struct {
	DB *gorm.DB
}

// NewReactionRepository создает новый экземпляр репозитория реакций
func NewReactionRepository(db *gorm.DB) ReactionRepository {
	return &reactionRepositoryImpl{DB: db}
}

// SetReaction создает новую реакцию или обновляет существующую
func (r *reactionRepositoryImpl) SetReaction(ctx context.Context, userID int64, videoID int64, kind models.ReactionKind) error {
	reaction := models.Reaction{
		UserID:   userID,
		VideoID:  videoID,
		Reaction: kind,
	}

	result := r.DB.WithContext(ctx).
		Clauses(
			clause.OnConflict{
				Columns: []clause.Column{{Name: "user_id"}, {Name: "video_id"}},
				DoUpdates: clause.Assignments(map[string]interface{}{
					"reaction": reaction.Reaction,
				}),
			}).
		Create(&reaction)

	if result.Error != nil {
		return result.Error
	}
	return nil
}

// DeleteReaction удаляет реакцию пользователя на видео
func (r *reactionRepositoryImpl) DeleteReaction(ctx context.Context, userID int64, videoID int64) (int64, error) {
	result := r.DB.WithContext(ctx).
		Where("user_id = ?", userID).
		Where("video_id = ?", videoID).
		Delete(&models.Reaction{})

	if result.Error != nil {
		return 0, result.Error
	}
	return result.RowsAffected, nil
}

// Получить список пользователей, поставивших реакцию
func (r *reactionRepositoryImpl) GetReactingUsers(ctx context.Context, videoID int64) ([]models.ReactingUserResponse, error) {
	var reactions []models.Reaction

	err := r.DB.WithContext(ctx).
		Preload("User").
		Where("video_id = ?", videoID).
		Find(&reactions).Error

	if err != nil {
		return nil, err
	}

	response := make([]models.ReactingUserResponse, len(reactions))
	for i, r := range reactions {
		response[i] = models.ReactingUserResponse{
			UserID:   r.UserID,
			Nickname: r.User.Nickname,
			Reaction: r.Reaction,
		}
	}
	return response, nil
}

package repository

import (
	"context"
	"errors"
	"time"

	"github.com/merinovvvv/momentic-backend/models"
	"gorm.io/gorm"
)

var ErrRecordNotFound = gorm.ErrRecordNotFound

// VideoRepository определяет методы для работы с БД
type VideoRepository interface {
	CreateVideo(ctx context.Context, video *models.Video) error
	GetVideoByID(ctx context.Context, videoID int64) (*models.Video, error)
	GetFriendsIDs(ctx context.Context, userID int64) ([]int64, error)
	GetTodayVideosByAuthors(ctx context.Context, authorIDs []int64) ([]models.Video, error)
	DeleteVideo(ctx context.Context, videoID int64) (*models.Video, error)
	UpdateDescription(ctx context.Context, videoID int64, description string) (rowsAffected int64, err error)
}

// videoRepositoryImpl - реализация интерфейса
type videoRepositoryImpl struct {
	db *gorm.DB
}

func NewVideoRepository(db *gorm.DB) VideoRepository {
	return &videoRepositoryImpl{db: db}
}

// --- Реализации CRUD для Video ---
func (r *videoRepositoryImpl) CreateVideo(ctx context.Context, video *models.Video) error {
	return r.db.WithContext(ctx).Create(video).Error
}

func (r *videoRepositoryImpl) GetVideoByID(ctx context.Context, videoID int64) (*models.Video, error) {
	var video models.Video
	err := r.db.WithContext(ctx).First(&video, videoID).Error
	if err != nil {
		return nil, err
	}
	return &video, nil
}

func (r *videoRepositoryImpl) DeleteVideo(ctx context.Context, videoID int64) (*models.Video, error) {
	video, err := r.GetVideoByID(ctx, videoID)
	if err != nil {
		return nil, err
	}

	deleteResult := r.db.WithContext(ctx).Delete(&video)
	if deleteResult.Error != nil {
		return nil, deleteResult.Error
	}

	return video, nil // Возвращаем модель, чтобы сервис мог удалить файл
}

func (r *videoRepositoryImpl) UpdateDescription(ctx context.Context, videoID int64, description string) (int64, error) {
	result := r.db.WithContext(ctx).Model(&models.Video{}).
		Where("video_id = ?", videoID).
		Update("description", description)

	return result.RowsAffected, result.Error
}

// --- Реализация логики для Friends ---
func (r *videoRepositoryImpl) GetFriendsIDs(ctx context.Context, userID int64) ([]int64, error) {
	if userID == 0 {
		return nil, errors.New("invalid user ID")
	}

	var friendships []models.Friendship
	result := r.db.WithContext(ctx).
		Where("status = ?", models.StatusFriends).
		Where("user_id1 = ? OR user_id2 = ?", userID, userID).
		Find(&friendships)

	if result.Error != nil {
		return nil, result.Error
	}

	friendIDs := make([]int64, 0, len(friendships))
	for _, f := range friendships {
		if f.UserID1 == userID {
			friendIDs = append(friendIDs, f.UserID2)
		} else {
			friendIDs = append(friendIDs, f.UserID1)
		}
	}
	return friendIDs, nil
}

func (r *videoRepositoryImpl) GetTodayVideosByAuthors(ctx context.Context, authorIDs []int64) ([]models.Video, error) {
	startOfDay := time.Now().Truncate(24 * time.Hour)
	endOfToday := startOfDay.Add(24 * time.Hour)
	var videos []models.Video

	result := r.db.WithContext(ctx).
		Where("author_id IN (?)", authorIDs).
		Where("created_at >= ? AND created_at < ?", startOfDay, endOfToday).
		Order("created_at DESC").
		Find(&videos)

	return videos, result.Error
}

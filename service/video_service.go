package service

import (
	"context"
	"errors"
	"log"
	"os"

	"github.com/merinovvvv/momentic-backend/models"
	"github.com/merinovvvv/momentic-backend/repository"
)

// Ошибки, специфичные для сервисного слоя
var ErrVideoNotFound = errors.New("video not found")
var ErrDescriptionTooLong = errors.New("description is too long (max 70 chars)")
var ErrNoFriends = errors.New("user has no friends")
var ErrAuthorIDRequired = errors.New("author_id is required")

// VideoService определяет все методы
type VideoService interface {
	UploadVideo(ctx context.Context, filepath string, authorID int64, description string) (*models.Video, error)
	GetTodayFeed(ctx context.Context, userID int64) ([]models.Video, error)
	DeleteVideo(ctx context.Context, videoID int64) error
	UpdateDescription(ctx context.Context, videoID int64, description string) error
}

type videoServiceImpl struct {
	Repo repository.VideoRepository
}

func NewVideoService(repo repository.VideoRepository) VideoService {
	return &videoServiceImpl{Repo: repo}
}

// --- UploadVideo (Создание) ---
func (s *videoServiceImpl) UploadVideo(ctx context.Context, filepath string, authorID int64, description string) (*models.Video, error) {
	if authorID == 0 {
		return nil, ErrAuthorIDRequired
	}

	newVideo := models.Video{
		Filepath:    filepath,
		AuthorID:    authorID,
		Description: description,
	}

	err := s.Repo.CreateVideo(ctx, &newVideo)
	if err != nil {
		log.Printf("ERROR: Failed to create video in DB for author %d: %v", authorID, err)
	}

	log.Printf("INFO: Video uploaded successfully. ID: %d, AuthorID: %d", newVideo.VideoID, authorID)

	return &newVideo, err
}

// --- GetTodayFeed (Чтение) ---
func (s *videoServiceImpl) GetTodayFeed(ctx context.Context, userID int64) ([]models.Video, error) {
	friendIDs, err := s.Repo.GetFriendsIDs(ctx, userID)
	if err != nil {
		return nil, err
	}

	if len(friendIDs) == 0 {
		log.Printf("INFO: User %d has no friends or no new videos today (returned empty feed).", userID)
		return []models.Video{}, ErrNoFriends
	}

	videos, err := s.Repo.GetTodayVideosByAuthors(ctx, friendIDs)

	if err != nil {
		log.Printf("ERROR: Failed to fetch today's videos for user %d: %v", userID, err)
		return nil, err
	}

	log.Printf("INFO: Successfully retrieved %d videos for user %d.", len(videos), userID)
	return videos, err
}

// --- DeleteVideo (Удаление) ---
func (s *videoServiceImpl) DeleteVideo(ctx context.Context, videoID int64) error {

	video, err := s.Repo.DeleteVideo(ctx, videoID)
	if errors.Is(err, repository.ErrRecordNotFound) {
		log.Printf("ERROR: Video not found. Video id %d:", videoID)
		return ErrVideoNotFound
	}
	if err != nil {
		log.Printf("ERROR: DB delete failed for videoID %d: %v", videoID, err)
		return err
	}

	if err := os.Remove(video.Filepath); err != nil {
		log.Printf("WARNING: Could not delete file %s from disk after DB success: %v", video.Filepath, err)
	}

	log.Printf("INFO: Video deleted successfully. ID: %d", videoID)
	return nil
}

// --- UpdateDescription (Обновление) ---
func (s *videoServiceImpl) UpdateDescription(ctx context.Context, videoID int64, description string) error {
	if len(description) > 70 {
		return ErrDescriptionTooLong
	}

	rowsAffected, err := s.Repo.UpdateDescription(ctx, videoID, description)
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		_, err = s.Repo.GetVideoByID(ctx, videoID)
		if errors.Is(err, repository.ErrRecordNotFound) {
			return ErrVideoNotFound
		}
		if err != nil {
			return err
		}
		log.Printf("INFO: Video description update skipped. ID: %d", videoID)
	}

	log.Printf("INFO: Video description updated successfully. ID: %d", videoID)
	return nil
}

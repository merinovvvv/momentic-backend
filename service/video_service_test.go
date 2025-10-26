package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/merinovvvv/momentic-backend/models"
	"github.com/merinovvvv/momentic-backend/repository"
)

var errTestDB = errors.New("DB test error")

// MockVideoRepository - структура, имитирующая репозиторий
type MockVideoRepository struct {
	// Поля для имитации поведения
	CreateVideoFn             func(ctx context.Context, video *models.Video) error
	DeleteVideoFn             func(ctx context.Context, videoID int64) (*models.Video, error)
	UpdateDescriptionFn       func(ctx context.Context, videoID int64, description string) (int64, error)
	GetFriendsIDsFn           func(ctx context.Context, userID int64) ([]int64, error)
	GetTodayVideosByAuthorsFn func(ctx context.Context, authorIDs []int64) ([]models.Video, error)
	GetVideoByIDFn            func(ctx context.Context, videoID int64) (*models.Video, error)
}

// Реализация методов интерфейса Repository
func (m *MockVideoRepository) CreateVideo(ctx context.Context, video *models.Video) error {
	return m.CreateVideoFn(ctx, video)
}
func (m *MockVideoRepository) DeleteVideo(ctx context.Context, videoID int64) (*models.Video, error) {
	return m.DeleteVideoFn(ctx, videoID)
}
func (m *MockVideoRepository) UpdateDescription(ctx context.Context, videoID int64, description string) (int64, error) {
	return m.UpdateDescriptionFn(ctx, videoID, description)
}
func (m *MockVideoRepository) GetFriendsIDs(ctx context.Context, userID int64) ([]int64, error) {
	return m.GetFriendsIDsFn(ctx, userID)
}
func (m *MockVideoRepository) GetTodayVideosByAuthors(ctx context.Context, authorIDs []int64) ([]models.Video, error) {
	return m.GetTodayVideosByAuthorsFn(ctx, authorIDs)
}
func (m *MockVideoRepository) GetVideoByID(ctx context.Context, videoID int64) (*models.Video, error) {
	return m.GetVideoByIDFn(ctx, videoID)
}

// --- UploadVideo (Создание) ---

func TestVideoService_UploadVideo(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name        string
		authorID    int64
		filepath    string
		description string
		mockRepoFn  func(t *testing.T, video *models.Video) error
		wantErr     error
	}{
		{
			name:        "Success",
			authorID:    10,
			filepath:    "uploads/10_test.mp4",
			description: "Test video",
			mockRepoFn: func(t *testing.T, video *models.Video) error {
				if video.AuthorID != 10 {
					t.Errorf("Expected AuthorID 10, got %d", video.AuthorID)
				}
				return nil
			},
			wantErr: nil,
		},
		{
			name:        "Error_MissingAuthorID",
			authorID:    0,
			filepath:    "uploads/0_test.mp4",
			description: "Test video",
			mockRepoFn: func(t *testing.T, video *models.Video) error {
				t.Fatalf("Repository should not be called")
				return nil
			},
			wantErr: ErrAuthorIDRequired,
		},
		{
			name:        "Error_DBFailure",
			authorID:    10,
			filepath:    "uploads/10_test.mp4",
			description: "Test video",
			mockRepoFn: func(t *testing.T, video *models.Video) error {
				return errTestDB
			},
			wantErr: errTestDB,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockVideoRepository{
				CreateVideoFn: func(ctx context.Context, video *models.Video) error {
					return tt.mockRepoFn(t, video)
				},
			}
			s := NewVideoService(mockRepo)

			_, err := s.UploadVideo(ctx, tt.filepath, tt.authorID, tt.description)

			if !errors.Is(err, tt.wantErr) && err != tt.wantErr {
				t.Errorf("UploadVideo() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// --- UpdateDescription (Обновление) ---

func TestVideoService_UpdateDescription(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name          string
		videoID       int64
		description   string
		mockUpdateFn  func() (int64, error)
		mockGetByIDFn func() (*models.Video, error)
		wantErr       error
	}{
		{
			name:        "Success_Update",
			videoID:     101,
			description: "New description",
			mockUpdateFn: func() (int64, error) {
				return 1, nil
			},
			wantErr: nil,
		},
		{
			name:        "Error_DescriptionTooLong",
			videoID:     101,
			description: "This description is definitely way too long, exceeding the maximum limit of 70 characters.",
			mockUpdateFn: func() (int64, error) {
				t.Fatalf("Repository should not be called due to validation error")
				return 0, nil
			},
			wantErr: ErrDescriptionTooLong,
		},
		{
			name:        "Error_VideoNotFound",
			videoID:     999,
			description: "Short description",
			mockUpdateFn: func() (int64, error) {
				return 0, nil
			},
			mockGetByIDFn: func() (*models.Video, error) {
				return nil, repository.ErrRecordNotFound
			},
			wantErr: ErrVideoNotFound,
		},
		{
			name:        "Error_DBFailure",
			videoID:     101,
			description: "Short description",
			mockUpdateFn: func() (int64, error) {
				return 0, errTestDB
			},
			mockGetByIDFn: func() (*models.Video, error) { return nil, nil },
			wantErr:       errTestDB,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockVideoRepository{
				UpdateDescriptionFn: func(ctx context.Context, videoID int64, desc string) (int64, error) {
					return tt.mockUpdateFn()
				},
				GetVideoByIDFn: func(ctx context.Context, videoID int64) (*models.Video, error) {
					if tt.mockGetByIDFn != nil {
						return tt.mockGetByIDFn()
					}
					return nil, nil
				},
			}
			s := NewVideoService(mockRepo)

			err := s.UpdateDescription(ctx, tt.videoID, tt.description)

			if !errors.Is(err, tt.wantErr) && (err == nil || tt.wantErr == nil || err.Error() != tt.wantErr.Error()) {
				t.Errorf("UpdateDescription() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// --- GetTodayFeed (Получение ленты) ---

func TestVideoService_GetTodayFeed(t *testing.T) {
	ctx := context.Background()

	sampleVideos := []models.Video{
		{VideoID: 1, AuthorID: 15, CreatedAt: time.Now()},
		{VideoID: 2, AuthorID: 20, CreatedAt: time.Now()},
	}

	tests := []struct {
		name             string
		userID           int64
		mockFriendsIDsFn func() ([]int64, error)
		mockGetVideosFn  func() ([]models.Video, error)
		wantLen          int
		wantErr          error
	}{
		{
			name:   "Success_WithVideos",
			userID: 10,
			mockFriendsIDsFn: func() ([]int64, error) {
				return []int64{15, 20}, nil
			},
			mockGetVideosFn: func() ([]models.Video, error) {
				return sampleVideos, nil
			},
			wantLen: 2,
			wantErr: nil,
		},
		{
			name:   "Success_NoVideosPublishedToday",
			userID: 10,
			mockFriendsIDsFn: func() ([]int64, error) {
				return []int64{15, 20}, nil
			},
			mockGetVideosFn: func() ([]models.Video, error) {
				return []models.Video{}, nil
			},
			wantLen: 0,
			wantErr: nil,
		},
		{
			name:   "Success_NoFriends",
			userID: 10,
			mockFriendsIDsFn: func() ([]int64, error) {
				return []int64{}, nil
			},
			mockGetVideosFn: func() ([]models.Video, error) {
				t.Fatalf("GetTodayVideosByAuthors should not be called")
				return nil, nil
			},
			wantLen: 0,
			wantErr: ErrNoFriends,
		},
		{
			name:   "Error_FriendsDBFailure",
			userID: 10,
			mockFriendsIDsFn: func() ([]int64, error) {
				return nil, errTestDB
			},
			mockGetVideosFn: func() ([]models.Video, error) {
				return nil, nil
			},
			wantLen: 0,
			wantErr: errTestDB,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockVideoRepository{
				GetFriendsIDsFn: func(ctx context.Context, userID int64) ([]int64, error) {
					return tt.mockFriendsIDsFn()
				},
				GetTodayVideosByAuthorsFn: func(ctx context.Context, authorIDs []int64) ([]models.Video, error) {
					return tt.mockGetVideosFn()
				},
			}
			s := NewVideoService(mockRepo)

			videos, err := s.GetTodayFeed(ctx, tt.userID)

			if !errors.Is(err, tt.wantErr) && err != tt.wantErr {
				t.Errorf("GetTodayFeed() error = %v, wantErr %v", err, tt.wantErr)
			}
			if len(videos) != tt.wantLen {
				t.Errorf("GetTodayFeed() got %d videos, want %d", len(videos), tt.wantLen)
			}
		})
	}
}

// --- DeleteVideo (Удаление) ---

func TestVideoService_DeleteVideo(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name         string
		videoID      int64
		mockDeleteFn func() (*models.Video, error)
		wantErr      error
	}{
		{
			name:    "Success_Deletion",
			videoID: 101,
			mockDeleteFn: func() (*models.Video, error) {
				return &models.Video{VideoID: 101, Filepath: "uploads/101_exists.mp4"}, nil
			},
			wantErr: nil,
		},
		{
			name:    "Error_VideoNotFound",
			videoID: 999,
			mockDeleteFn: func() (*models.Video, error) {
				return nil, repository.ErrRecordNotFound
			},
			wantErr: ErrVideoNotFound,
		},
		{
			name:    "Error_DBFailure",
			videoID: 101,
			mockDeleteFn: func() (*models.Video, error) {
				return nil, errors.New("DB delete failed")
			},
			wantErr: errors.New("DB delete failed"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockVideoRepository{
				DeleteVideoFn: func(ctx context.Context, videoID int64) (*models.Video, error) {
					return tt.mockDeleteFn()
				},
			}
			s := NewVideoService(mockRepo)

			err := s.DeleteVideo(ctx, tt.videoID)

			if !errors.Is(err, tt.wantErr) && (err == nil || tt.wantErr == nil || err.Error() != tt.wantErr.Error()) {
				t.Errorf("DeleteVideo() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

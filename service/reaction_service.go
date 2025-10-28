package service

import (
	"context"
	"errors"
	"log"

	"github.com/merinovvvv/momentic-backend/models"
	"github.com/merinovvvv/momentic-backend/repository"
)

var (
	ErrInvalidReactionKind = errors.New("invalid reaction kind")
	ErrReactionNotFound    = errors.New("reaction not found")
)

// ReactionService определяет методы бизнес-логики для реакций.
type ReactionService interface {
	HandleReaction(ctx context.Context, userID int64, videoID int64, kind models.ReactionKind) error
	RemoveReaction(ctx context.Context, userID int64, videoID int64) error
	GetVideoReactions(ctx context.Context, videoID int64) ([]models.ReactingUserResponse, error)
}

type reactionServiceImpl struct {
	Repo repository.ReactionRepository
}

func NewReactionService(repo repository.ReactionRepository) ReactionService {
	return &reactionServiceImpl{Repo: repo}
}

func isValidReactionKind(kind models.ReactionKind) bool {
	switch kind {
	case models.ReactionHeart, models.ReactionFlame, models.ReactionFunny, models.ReactionAngry:
		return true
	default:
		return false
	}
}

// HandleReaction устанавливает или обновляет реакцию пользователя на видео.
func (s *reactionServiceImpl) HandleReaction(ctx context.Context, userID int64, videoID int64, kind models.ReactionKind) error {
	if !isValidReactionKind(kind) {
		log.Printf("ERROR: Invalid reaction kind received: %s by UserID %d", kind, userID)
		return ErrInvalidReactionKind
	}

	err := s.Repo.SetReaction(ctx, userID, videoID, kind)
	if err != nil {
		log.Printf("ERROR: Failed to set reaction %s for VideoID %d by UserID %d: %v", kind, videoID, userID, err)
		return err
	}

	log.Printf("INFO: Reaction '%s' set/updated for VideoID %d by UserID %d", kind, videoID, userID)
	return nil
}

func (s *reactionServiceImpl) RemoveReaction(ctx context.Context, userID int64, videoID int64) error {
	rowsAffected, err := s.Repo.DeleteReaction(ctx, userID, videoID)
	if err != nil {
		log.Printf("ERROR: Failed to delete reaction for VideoID %d by UserID %d: %v", videoID, userID, err)
		return err
	}

	if rowsAffected == 0 {
		return ErrReactionNotFound
	}

	log.Printf("INFO: Reaction removed for VideoID %d by UserID %d", videoID, userID)
	return nil
}

func (s *reactionServiceImpl) GetVideoReactions(ctx context.Context, videoID int64) ([]models.ReactingUserResponse, error) {
	users, err := s.Repo.GetReactingUsers(ctx, videoID)
	if err != nil {
		log.Printf("ERROR: Failed to fetch reactions for VideoID %d: %v", videoID, err)
		return nil, err
	}

	log.Printf("INFO: Successfully fetched %d reactions for VideoID %d", len(users), videoID)
	return users, nil
}

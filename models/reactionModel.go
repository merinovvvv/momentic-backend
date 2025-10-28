package models

import "time"

type ReactionKind string

const (
	ReactionHeart ReactionKind = "heart"
	ReactionFlame ReactionKind = "flame"
	ReactionFunny ReactionKind = "funny"
	ReactionAngry ReactionKind = "angry"
)

type Reaction struct {
	UserID    int64        `gorm:"column:user_id;primaryKey"`
	VideoID   int64        `gorm:"column:video_id;primaryKey"`
	Reaction  ReactionKind `gorm:"column:reaction;type:reaction_kind"`
	CreatedAt time.Time    `gorm:"column:created_at"`

	User User
}

type ReactionRequest struct {
	Kind ReactionKind `json:"reaction_kind" binding:"required"`
}

type ReactingUserResponse struct {
	UserID   int64        `json:"user_id"`
	Nickname string       `json:"nickname"`
	Reaction ReactionKind `json:"reaction_kind"`
}

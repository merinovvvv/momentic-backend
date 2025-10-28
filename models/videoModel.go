package models

import (
	"time"
)

type Video struct {
	// video_id BIGSERIAL PRIMARY KEY
	VideoID int64 `gorm:"primaryKey;column:video_id;primaryKey;autoIncrement"`

	// filepath TEXT NOT NULL
	Filepath string `gorm:"column:filepath;type:TEXT;not null"`

	// author_id BIGINT REFERENCES users(user_id) ON DELETE CASCADE
	AuthorID int64 `gorm:"column:author_id;type:BIGINT;not null"`

	// description VARCHAR(70) NOT NULL DEFAULT ''
	Description string `gorm:"column:description;type:VARCHAR(70);not null;default:''"`

	// created_at TIMESTAMPTZ NOT NULL DEFAULT now()
	CreatedAt time.Time `gorm:"column:created_at;type:TIMESTAMPTZ;not null;default:now()"`
}

func (Video) TableName() string {
	return "videos"
}

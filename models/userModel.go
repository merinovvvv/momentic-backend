package models

import "time"

type User struct {
    ID         uint64    `gorm:"primaryKey;column:user_id"`
    Nickname       string    `gorm:"size:50;not null;unique"`
    Email          string    `gorm:"size:255;not null;unique"`
	Password       string    `gorm:"size:255;not null"`
    Rating         int       `gorm:"not null;default:0;check:rating >= 0"`
    MaxStreak      int       `gorm:"not null;default:0;check:max_streak >= 0"`
    CurrentStreak  int       `gorm:"not null;default:0;check:current_streak >= 0"`
    MaxReactions   int       `gorm:"not null;default:0;check:max_reactions >= 0"`
    AvatarFilepath *string   `gorm:"column:avatar_filepath"` // nullable
    Bio            string    `gorm:"size:100;not null;default:''"`
    CreatedAt      time.Time `gorm:"column:created_at;not null;default:now()"`
}

func (User) TableName() string {
    return "users"
}

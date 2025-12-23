package models

import "time"

// Comment представляет структуру комментария в БД и для передачи через WS
type Comment struct {
	CommentID int64     `gorm:"primaryKey;column:comment_id" json:"comment_id"`
	VideoID   int64     `gorm:"column:video_id;index" json:"video_id"`
	UserID    int64     `gorm:"column:user_id" json:"user_id"`
	Nickname  string    `gorm:"-" json:"nickname"` // Заполняется из связи с User
	Content   string    `gorm:"column:content" json:"content" binding:"required"`
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
}

// CommentWSMessage — структура сообщения, которая гуляет по сокетам
type CommentWSMessage struct {
	Type    string  `json:"type"` // "new_comment"
	Payload Comment `json:"payload"`
}

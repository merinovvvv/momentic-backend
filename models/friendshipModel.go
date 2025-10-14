package models

import "time"

// FriendshipStatus определяет возможные статусы дружбы.
type FriendshipStatus string

const (
	StatusFriends   FriendshipStatus = "friends"
	StatusPending1  FriendshipStatus = "pending1" // Ожидание (от user1 к user2)
	StatusPending2  FriendshipStatus = "pending2"
	StatusBlocked1  FriendshipStatus = "blocked1" // user1 заблокировал user2
	StatusBlocked2  FriendshipStatus = "blocked2"
	StatusBlocked12 FriendshipStatus = "blocked12"
)

type Friendship struct {
	// user_id1 BIGINT PRIMARY KEY REFERENCES users(user_id) ON DELETE CASCADE
	UserID1 int64 `gorm:"column:user_id1;primaryKey;autoIncrement:false;type:BIGINT;not null"`

	// user_id2 BIGINT PRIMARY KEY REFERENCES users(user_id) ON DELETE CASCADE
	UserID2 int64 `gorm:"column:user_id2;primaryKey;autoIncrement:false;type:BIGINT;not null"`

	Status FriendshipStatus `gorm:"column:status;type:friendship_status;not null"`

	// created_at TIMESTAMPTZ NOT NULL DEFAULT now()
	CreatedAt time.Time `gorm:"column:created_at;type:TIMESTAMPTZ;not null;default:now()"`
}

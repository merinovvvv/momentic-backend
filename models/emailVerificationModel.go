package models

import "time"

type EmailVerification struct {
	// ID  	   uint64	 `gorm:"primaryKey;column:id"`
	Email      string    `gorm:"primaryKey;size:255;unique"`
	Code       string    `gorm:"size:64;not null"` //hashed
	ExpiresAt time.Time `gorm:"not null;column:expires_at"`
	// Used bool
}

func (e EmailVerification) TableName() string {
  return "email_verifications"
}

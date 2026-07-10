package models

import (
	"time"

	"gorm.io/gorm"
)

type RevokedToken struct {
	gorm.Model
	JTI       string    `json:"jti" gorm:"type:VARCHAR(64);uniqueIndex;not null"`
	ExpiresAt time.Time `json:"expires_at" gorm:"not null"`
}

package models

import (
	"gorm.io/gorm"
)

type Reward struct {
	gorm.Model
	Name        string `json:"name" gorm:"type:VARCHAR(50);not null"`
	Description string `json:"description" gorm:"type:VARCHAR(255)"`
	RequiredXp  int    `json:"required_xp" gorm:"not null"`
}

// RewardClaim records that a user has claimed a reward. The unique index
// on (UserID, RewardID) makes a reward claimable only once per user, even
// under concurrent requests.
type RewardClaim struct {
	gorm.Model
	UserID   uint `json:"user_id" gorm:"uniqueIndex:idx_user_reward;not null"`
	RewardID uint `json:"reward_id" gorm:"uniqueIndex:idx_user_reward;not null"`
}

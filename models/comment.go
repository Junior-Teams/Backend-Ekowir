package models

import (
	"gorm.io/gorm"
)

type Comment struct {
	gorm.Model
	Content   string `json:"content" form:"content" gorm:"type:TEXT;not null;default:null"`
	CreatedBy string `json:"created_by" form:"createdBy" gorm:"type:VARCHAR(30);not null;default:null"`
	Image     string `json:"image" form:"image" gorm:"type:VARCHAR(255);null;default:null"`
	// UserID is the authoritative owner, set server-side from the
	// authenticated session - never trust a client-supplied CreatedBy.
	UserID  uint   `json:"user_id" gorm:"not null;index"`
	User    *User  `json:"user,omitempty" gorm:"foreignKey:UserID;references:ID"`
	IDForum uint   `json:"id_forum" form:"idForum" gorm:"type:INT;not null;default:null"`
	Forum   Forum  `gorm:"foreignKey:IDForum;references:ID"`
}

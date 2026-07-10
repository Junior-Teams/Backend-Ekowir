package models

import (
	"gorm.io/gorm"
)

type Forum struct {
	gorm.Model
	Title       string `json:"title" form:"title" gorm:"type:VARCHAR(150);not null;default:null"`
	Description string `json:"description" form:"description" gorm:"type:TEXT;not null;default:null"`
	CreatedBy   string `json:"created_by" form:"createdBy" gorm:"type:VARCHAR(30);not null;default:null"`
	// UserID is the authoritative owner, set server-side from the
	// authenticated session - never trust a client-supplied CreatedBy.
	UserID   uint   `json:"user_id" gorm:"not null;index"`
	User     *User  `json:"user,omitempty" gorm:"foreignKey:UserID;references:ID"`
	IDModule uint   `json:"id_module" form:"idModule" gorm:"type:INT;not null;default:null"`
	Module   Module `gorm:"foreignKey:IDModule;references:ID"`
}

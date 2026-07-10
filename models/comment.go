package models

import (
	"gorm.io/gorm"
)

type Comment struct {
	gorm.Model
	Content   string `json:"content" form:"content" gorm:"type:VARCHAR(30);not null;default:null"`
	CreatedBy string `json:"created_by" form:"createdBy" gorm:"type:VARCHAR(30);not null;default:null"`
	Image     string `json:"image" form:"image" gorm:"type:VARCHAR(30);null;default:null"`
	IDForum   uint   `json:"id_forum" form:"idForum" gorm:"type:INT;not null;default:null"`
	Forum     Forum  `gorm:"foreignKey:IDForum;references:ID"`
}

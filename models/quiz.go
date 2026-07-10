package models

import (
	"gorm.io/gorm"
)

type Quiz struct {
	gorm.Model
	Title       string `json:"title" form:"title" gorm:"type:VARCHAR(30);not null;default:null"`
	Description string `json:"description" form:"description" gorm:"type:VARCHAR(30);not null;default:null"`
	CreatedBy   string `json:"created_by" form:"createdBy" gorm:"type:VARCHAR(30);not null;default:null"`
	IDModule    uint   `json:"id_module" form:"idModule" gorm:"type:INT;not null;default:null"`
	Module      Module `gorm:"foreignKey:IDModule;references:ID"`
}

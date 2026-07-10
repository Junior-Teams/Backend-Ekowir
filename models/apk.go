package models

import (
	"gorm.io/gorm"
)

type Apk struct {
	gorm.Model
	Name        string `json:"name" form:"name" gorm:"type:VARCHAR(30);not null;default:null"`
	Cover       string `json:"cover"`
	Title       string `json:"title" form:"title"`
	Description string `json:"description" form:"description"`
	Game        string `json:"game" form:"game"`
	Footage     string `json:"footage"`
	Creator     string `json:"creator" form:"creator"`
}

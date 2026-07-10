package models

import (
	"gorm.io/gorm"
)

type Tier struct {
	gorm.Model
	Name  string `json:"name" form:"name" gorm:"type:VARCHAR(30);not null;default:null"`
	Point int    `json:"point" form:"point" gorm:"type:INT;not null;default:null"`
}

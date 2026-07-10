package models

import (
	"gorm.io/gorm"
)

type Materi struct {
	gorm.Model
	Name         string `json:"name" form:"name" gorm:"type:VARCHAR(30);not null;default:null"`
	ArrayElement string `json:"array_element" form:"arrayElement" gorm:"type:VARCHAR(30);not null;default:null"`
	IDModule     uint   `json:"id_module" form:"idModule" gorm:"type:INT;not null;default:null"`
	Module       Module `gorm:"foreignKey:IDModule;references:ID"`
}

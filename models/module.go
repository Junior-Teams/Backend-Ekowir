package models

import (
	"gorm.io/gorm"
)

type Module struct {
	gorm.Model
	CodeModule        string `json:"code_module" form:"codeModule" gorm:"type:VARCHAR(30);not null;default:null"`
	NameModule        string `json:"name_module" form:"nameModule" gorm:"type:VARCHAR(30);not null;default:null"`
	DescriptionModule string `json:"description_module" form:"descriptionModule" gorm:"type:VARCHAR(30);not null;default:null"`
	Image             string `json:"image" form:"image" gorm:"type:VARCHAR(30);not null;default:null"`
}

package models

import (
	"gorm.io/gorm"
)

type AnswerOption struct {
	gorm.Model
	OptionText string `json:"option_text" gorm:"type:VARCHAR(255);not null"`
	IsCorrect  bool   `json:"is_correct" gorm:"not null;default:false"`
	IDQuestion uint   `json:"id_question" gorm:"not null"`
}

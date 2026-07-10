package models

import (
	"gorm.io/gorm"
)

type Question struct {
	gorm.Model
	QuestionText string `json:"question_text" form:"questionText" gorm:"type:VARCHAR(30);not null;default:null"`
	Answer       string `json:"answer" form:"answer" gorm:"type:VARCHAR(30);not null;default:null"`
	Point        int    `json:"point" form:"point" gorm:"type:INT;not null;default:null"`
	IDQuiz       uint   `json:"id_quiz" form:"idQuiz" gorm:"type:INT;not null;default:null"`
	Quiz		 Quiz   `gorm:"foreignKey:IDQuiz;references:ID"`
}

package models

import (
	"gorm.io/gorm"
)

type Quiz struct {
	gorm.Model
	Title        string `json:"title" form:"title" gorm:"type:VARCHAR(30);not null;default:null"`
	Description  string `json:"description" form:"description" gorm:"type:VARCHAR(30);not null;default:null"`
	CreatedBy    string `json:"created_by" form:"createdBy" gorm:"type:VARCHAR(30);not null;default:null"`
	IDModule     uint   `json:"id_module" form:"idModule" gorm:"type:INT;not null;default:null"`
	Module       Module `json:"-" gorm:"foreignKey:IDModule;references:ID"`
	PassingScore int    `json:"passing_score" gorm:"not null;default:0"`
	BonusXp      int    `json:"bonus_xp" gorm:"not null;default:0"`
}

// QuizCompletion marks that a user has already passed & been paid out for a
// quiz, so XP/bonus is only ever credited once per user per quiz (retries
// after passing are for practice only, not more XP).
type QuizCompletion struct {
	gorm.Model
	UserID   uint `json:"user_id" gorm:"uniqueIndex:idx_user_quiz;not null"`
	IDQuiz   uint `json:"id_quiz" gorm:"uniqueIndex:idx_user_quiz;not null"`
	Score    int  `json:"score" gorm:"not null"`
	XpEarned int  `json:"xp_earned" gorm:"not null"`
}

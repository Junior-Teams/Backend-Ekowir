package models

import "gorm.io/gorm"

// MateriCompletion marks that a user has viewed/completed a materi. Mirrors
// QuizCompletion's shape: row existence is the completion signal, unique per
// (user, materi) so revisiting is a no-op, not a new row.
type MateriCompletion struct {
	gorm.Model
	UserID   uint `json:"user_id" gorm:"uniqueIndex:idx_user_materi;not null"`
	IDMateri uint `json:"id_materi" gorm:"uniqueIndex:idx_user_materi;not null"`
}

package models

import (
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Name     string  `json:"name"`
	Username string  `json:"username" gorm:"unique"`
	Email    string  `json:"email" gorm:"unique"`
	Password string  `json:"-"`
	Picture  string  `json:"picture"`
	Xp       int     `json:"xp" gorm:"not null;default:0"`
	Role     string  `json:"role" gorm:"type:VARCHAR(20);not null;default:'user'"`
	IDTier   *uint   `json:"id_tier" gorm:"index"`
	Tier     *Tier   `json:"tier,omitempty" gorm:"foreignKey:IDTier;references:ID"`
	GoogleID *string `json:"-" gorm:"uniqueIndex"`
}

func (user *User) HashPassword(password string) error {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		return err
	}
	user.Password = string(bytes)
	return nil
}

func (user *User) CheckPassword(providedPassword string) error {
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(providedPassword))
	if err != nil {
		return err
	}
	return nil
}

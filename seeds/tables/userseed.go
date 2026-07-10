package tables

import (
	"fmt"

	"github.com/ALZEE23/ApiGo/database"
	"github.com/ALZEE23/ApiGo/models"
	"golang.org/x/crypto/bcrypt"
	"github.com/bxcodec/faker/v3"
)

func SeedUsers() {
	fmt.Println("Seeding users...")
	var user models.User
	password := "password123"
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		fmt.Println("Error hashing password:", err)
		return
	}
	user.Name = faker.Name()
	user.Email = faker.Email()
	user.Username = faker.Username()
	user.Password = string(hashedPassword)

	if err := database.DB.Db.Create(&user).Error; err != nil {
		fmt.Println("Error seeding user:", err)
		return
	}
}
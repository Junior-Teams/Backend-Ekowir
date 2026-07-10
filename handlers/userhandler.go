package handlers

import (
	"net/http"

	"github.com/ALZEE23/ApiGo/database"
	"github.com/ALZEE23/ApiGo/models"
	"github.com/gin-gonic/gin"
)

func RegisterUser(context *gin.Context) {
	var input struct {
		Name     string `json:"name"`
		Username string `json:"username" binding:"required"`
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}
	if err := context.ShouldBindJSON(&input); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		context.Abort()
		return
	}

	user := models.User{
		Name:     input.Name,
		Username: input.Username,
		Email:    input.Email,
	}
	if err := user.HashPassword(input.Password); err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		context.Abort()
		return
	}
	record := database.DB.Db.Create(&user)
	if record.Error != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": record.Error.Error()})
		context.Abort()
		return
	}
	context.JSON(http.StatusCreated, gin.H{"userId": user.ID, "email": user.Email, "username": user.Username})
}

func GetUsers(context *gin.Context) {
	var users []models.User
	if err := database.DB.Db.Find(&users).Error; err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		context.Abort()
		return
	}
	context.JSON(http.StatusOK, users)
}

func GetUserByID(context *gin.Context) {
	id := context.Param("id")
	var user models.User
	if err := database.DB.Db.First(&user, id).Error; err != nil {
		context.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		context.Abort()
		return
	}
	context.JSON(http.StatusOK, user)
}

func UpdateUser(context *gin.Context) {
	id := context.Param("id")
	var user models.User
	if err := database.DB.Db.First(&user, id).Error; err != nil {
		context.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		context.Abort()
		return
	}

	var input struct {
		Name     string `json:"name"`
		Username string `json:"username"`
		Email    string `json:"email"`
		OldPassword string `json:"old_password"`
		Password string `json:"password"`
	}

	if err := context.ShouldBindJSON(&input); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		context.Abort()
		return
	}

	if input.OldPassword != "" {
		if err := user.CheckPassword(input.OldPassword); err != nil {
			context.JSON(http.StatusUnauthorized, gin.H{"error": "Old password is incorrect"})
			context.Abort()
			return
		}
	}

	if input.Password != "" {
		if err := user.HashPassword(input.Password); err != nil {
			context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			context.Abort()
			return
		}
	}

	if input.Name != "" {
		user.Name = input.Name
	}
	if input.Username != "" {
		user.Username = input.Username
	}
	if input.Email != "" {
		user.Email = input.Email
	}

	if err := database.DB.Db.Save(&user).Error; err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		context.Abort()
		return
	}

	context.JSON(http.StatusOK, user)
}

func DeleteUser(context *gin.Context) {
	id := context.Param("id")
	var user models.User
	if err := database.DB.Db.First(&user, id).Error; err != nil {
		context.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		context.Abort()
		return
	}

	if err := database.DB.Db.Delete(&user).Error; err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		context.Abort()
		return
	}

	context.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}
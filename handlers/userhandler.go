package handlers

import (
	"net/http"

	"github.com/ALZEE23/ApiGo/database"
	"github.com/ALZEE23/ApiGo/models"
	"github.com/ALZEE23/ApiGo/utils"
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
		utils.RespondValidationError(context, "Data yang Anda masukkan tidak valid, mohon periksa kembali")
		return
	}

	user := models.User{
		Name:     input.Name,
		Username: input.Username,
		Email:    input.Email,
	}
	if err := user.HashPassword(input.Password); err != nil {
		utils.RespondServerError(context, err)
		return
	}
	record := database.DB.Db.Create(&user)
	if record.Error != nil {
		utils.RespondDBError(context, record.Error, "Data tidak ditemukan")
		return
	}
	context.JSON(http.StatusCreated, gin.H{"userId": user.ID, "email": user.Email, "username": user.Username})
}

func CreateUser(context *gin.Context) {
	var input struct {
		Name     string `json:"name"`
		Username string `json:"username" binding:"required"`
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
		Role     string `json:"role"`
	}
	if err := context.ShouldBindJSON(&input); err != nil {
		utils.RespondValidationError(context, "Data yang Anda masukkan tidak valid, mohon periksa kembali")
		return
	}

	role := input.Role
	if role == "" {
		role = "user"
	}

	user := models.User{
		Name:     input.Name,
		Username: input.Username,
		Email:    input.Email,
		Role:     role,
	}
	if err := user.HashPassword(input.Password); err != nil {
		utils.RespondServerError(context, err)
		return
	}
	record := database.DB.Db.Create(&user)
	if record.Error != nil {
		utils.RespondDBError(context, record.Error, "Data tidak ditemukan")
		return
	}
	context.JSON(http.StatusCreated, gin.H{"userId": user.ID, "email": user.Email, "username": user.Username, "role": user.Role})
}

func GetMe(context *gin.Context) {
	email := context.GetString("email")
	var user models.User
	if err := database.DB.Db.Where("email = ?", email).First(&user).Error; err != nil {
		context.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		context.Abort()
		return
	}
	context.JSON(http.StatusOK, gin.H{
		"id":       user.ID,
		"name":     user.Name,
		"username": user.Username,
		"email":    user.Email,
		"picture":  user.Picture,
		"role":     user.Role,
	})
}

func GetUsers(context *gin.Context) {
	var users []models.User
	if err := database.DB.Db.Find(&users).Error; err != nil {
		utils.RespondServerError(context, err)
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
		Name        string `json:"name"`
		Username    string `json:"username"`
		Email       string `json:"email"`
		Role        string `json:"role"`
		OldPassword string `json:"old_password"`
		Password    string `json:"password"`
	}

	if err := context.ShouldBindJSON(&input); err != nil {
		utils.RespondValidationError(context, "Data yang Anda masukkan tidak valid, mohon periksa kembali")
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
			utils.RespondServerError(context, err)
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
	if input.Role != "" {
		user.Role = input.Role
	}

	if err := database.DB.Db.Save(&user).Error; err != nil {
		utils.RespondDBError(context, err, "Pengguna tidak ditemukan")
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
		utils.RespondServerError(context, err)
		return
	}

	context.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}

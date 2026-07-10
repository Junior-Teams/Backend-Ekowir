package handlers

import (
	"errors"
	"net/http"
	"time"

	"github.com/ALZEE23/ApiGo/auth"
	"github.com/ALZEE23/ApiGo/database"
	"github.com/ALZEE23/ApiGo/models"
	"github.com/ALZEE23/ApiGo/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type TokenRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func GenerateToken(context *gin.Context) {
	var request TokenRequest
	var user models.User
	if err := context.ShouldBindJSON(&request); err != nil {
		utils.RespondValidationError(context, "Email dan password wajib diisi")
		return
	}

	record := database.DB.Db.Where("email = ?", request.Email).First(&user)
	if record.Error != nil {
		if errors.Is(record.Error, gorm.ErrRecordNotFound) {
			context.JSON(http.StatusUnauthorized, gin.H{"error": "Email atau password salah"})
			context.Abort()
			return
		}
		utils.RespondServerError(context, record.Error)
		return
	}
	credentialError := user.CheckPassword(request.Password)
	if credentialError != nil {
		context.JSON(http.StatusUnauthorized, gin.H{"error": "Email atau password salah"})
		context.Abort()
		return
	}
	tokenString, err := auth.GenerateJWT(user.Email, user.Username, user.Role)
	if err != nil {
		utils.RespondServerError(context, err)
		return
	}
	context.JSON(http.StatusOK, gin.H{"token": tokenString})
}

func Logout(context *gin.Context) {
	jti := context.GetString("jti")
	expiresAt, _ := context.Get("exp")

	exp, ok := expiresAt.(time.Time)
	if jti == "" || !ok {
		context.JSON(http.StatusBadRequest, gin.H{"error": "no active session to revoke"})
		context.Abort()
		return
	}

	revoked := models.RevokedToken{
		JTI:       jti,
		ExpiresAt: exp,
	}
	if err := database.DB.Db.Create(&revoked).Error; err != nil {
		utils.RespondServerError(context, err)
		return
	}

	context.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}

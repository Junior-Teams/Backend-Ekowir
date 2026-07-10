package handlers

import (
	"net/http"

	"github.com/ALZEE23/ApiGo/database"
	"github.com/ALZEE23/ApiGo/models"
	"github.com/ALZEE23/ApiGo/utils"
	"github.com/gin-gonic/gin"
)

func Quiz(context *gin.Context) {
	var input struct {
		Title       string `json:"title" binding:"required"`
		Description string `json:"description" binding:"required"`
		CreatedBy   string `json:"created_by" binding:"required"`
		IDModule    uint   `json:"id_module" binding:"required"`
	}
	if err := context.ShouldBindJSON(&input); err != nil {
		utils.RespondValidationError(context, "Data yang Anda masukkan tidak valid, mohon periksa kembali")
		return
	}

	var module models.Module
	if err := database.DB.Db.First(&module, input.IDModule).Error; err != nil {
		context.JSON(http.StatusNotFound, gin.H{"error": "Module not found"})
		context.Abort()
		return
	}

	quiz := models.Quiz{
		Title:       input.Title,
		Description: input.Description,
		CreatedBy:   input.CreatedBy,
		IDModule:    input.IDModule,
	}

	record := database.DB.Db.Create(&quiz)
	if record.Error != nil {
		utils.RespondDBError(context, record.Error, "Data tidak ditemukan")
		return
	}
	context.JSON(http.StatusCreated, gin.H{"quizId": quiz.ID, "title": quiz.Title, "description": quiz.Description, "created_by": quiz.CreatedBy, "id_module": quiz.IDModule})
}

func GetQuizzes(context *gin.Context) {
	var quizzes []models.Quiz
	query := database.DB.Db.Preload("Module")

	if idModule := context.Query("idModule"); idModule != "" {
		query = query.Where("id_module = ?", idModule)
	}

	if err := query.Find(&quizzes).Error; err != nil {
		utils.RespondServerError(context, err)
		return
	}
	context.JSON(http.StatusOK, gin.H{"data": quizzes})
}

func GetQuizByID(context *gin.Context) {
	id := context.Param("id")
	var quiz models.Quiz
	if err := database.DB.Db.Preload("Module").First(&quiz, id).Error; err != nil {
		context.JSON(http.StatusNotFound, gin.H{"error": "Quiz not found"})
		context.Abort()
		return
	}
	context.JSON(http.StatusOK, quiz)
}

func UpdateQuiz(context *gin.Context) {
	id := context.Param("id")
	var quiz models.Quiz
	if err := database.DB.Db.First(&quiz, id).Error; err != nil {
		context.JSON(http.StatusNotFound, gin.H{"error": "Quiz not found"})
		context.Abort()
		return
	}

	var input struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		CreatedBy   string `json:"created_by"`
		IDModule    uint   `json:"id_module"`
	}
	if err := context.ShouldBindJSON(&input); err != nil {
		utils.RespondValidationError(context, "Data yang Anda masukkan tidak valid, mohon periksa kembali")
		return
	}

	if input.IDModule != 0 {
		var module models.Module
		if err := database.DB.Db.First(&module, input.IDModule).Error; err != nil {
			context.JSON(http.StatusNotFound, gin.H{"error": "Module not found"})
			context.Abort()
			return
		}
		quiz.IDModule = input.IDModule
	}
	if input.Title != "" {
		quiz.Title = input.Title
	}
	if input.Description != "" {
		quiz.Description = input.Description
	}
	if input.CreatedBy != "" {
		quiz.CreatedBy = input.CreatedBy
	}

	if err := database.DB.Db.Save(&quiz).Error; err != nil {
		utils.RespondDBError(context, err, "Quiz tidak ditemukan")
		return
	}
	context.JSON(http.StatusOK, quiz)
}

func DeleteQuiz(context *gin.Context) {
	id := context.Param("id")
	var quiz models.Quiz
	if err := database.DB.Db.First(&quiz, id).Error; err != nil {
		context.JSON(http.StatusNotFound, gin.H{"error": "Quiz not found"})
		context.Abort()
		return
	}

	if err := database.DB.Db.Delete(&quiz).Error; err != nil {
		utils.RespondServerError(context, err)
		return
	}
	context.JSON(http.StatusOK, gin.H{"message": "Quiz deleted successfully"})
}

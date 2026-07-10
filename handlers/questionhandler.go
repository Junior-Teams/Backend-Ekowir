package handlers

import (
	"net/http"

	"github.com/ALZEE23/ApiGo/database"
	"github.com/ALZEE23/ApiGo/models"
	"github.com/ALZEE23/ApiGo/utils"
	"github.com/gin-gonic/gin"
)

func Question(context *gin.Context) {
	var input struct {
		QuestionText string `json:"question_text" binding:"required"`
		Answer       string `json:"answer" binding:"required"`
		Point        int    `json:"point" binding:"required"`
		IDQuiz       uint   `json:"id_quiz" binding:"required"`
	}
	if err := context.ShouldBindJSON(&input); err != nil {
		utils.RespondValidationError(context, "Data yang Anda masukkan tidak valid, mohon periksa kembali")
		return
	}

	var quiz models.Quiz
	if err := database.DB.Db.First(&quiz, input.IDQuiz).Error; err != nil {
		context.JSON(http.StatusNotFound, gin.H{"error": "Quiz not found"})
		context.Abort()
		return
	}

	question := models.Question{
		QuestionText: input.QuestionText,
		Answer:       input.Answer,
		Point:        input.Point,
		IDQuiz:       input.IDQuiz,
	}

	record := database.DB.Db.Create(&question)
	if record.Error != nil {
		utils.RespondDBError(context, record.Error, "Data tidak ditemukan")
		return
	}
	context.JSON(http.StatusCreated, gin.H{"questionId": question.ID, "question_text": question.QuestionText, "answer": question.Answer, "point": question.Point, "id_quiz": question.IDQuiz})
}

func GetQuestions(context *gin.Context) {
	var questions []models.Question
	query := database.DB.Db.Preload("Quiz")

	if idQuiz := context.Query("idQuiz"); idQuiz != "" {
		query = query.Where("id_quiz = ?", idQuiz)
	}

	if err := query.Find(&questions).Error; err != nil {
		utils.RespondServerError(context, err)
		return
	}
	context.JSON(http.StatusOK, gin.H{"data": questions})
}

func GetQuestionByID(context *gin.Context) {
	id := context.Param("id")
	var question models.Question
	if err := database.DB.Db.Preload("Quiz").First(&question, id).Error; err != nil {
		context.JSON(http.StatusNotFound, gin.H{"error": "Question not found"})
		context.Abort()
		return
	}
	context.JSON(http.StatusOK, question)
}

func UpdateQuestion(context *gin.Context) {
	id := context.Param("id")
	var question models.Question
	if err := database.DB.Db.First(&question, id).Error; err != nil {
		context.JSON(http.StatusNotFound, gin.H{"error": "Question not found"})
		context.Abort()
		return
	}

	var input struct {
		QuestionText string `json:"question_text"`
		Answer       string `json:"answer"`
		Point        int    `json:"point"`
		IDQuiz       uint   `json:"id_quiz"`
	}
	if err := context.ShouldBindJSON(&input); err != nil {
		utils.RespondValidationError(context, "Data yang Anda masukkan tidak valid, mohon periksa kembali")
		return
	}

	if input.IDQuiz != 0 {
		var quiz models.Quiz
		if err := database.DB.Db.First(&quiz, input.IDQuiz).Error; err != nil {
			context.JSON(http.StatusNotFound, gin.H{"error": "Quiz not found"})
			context.Abort()
			return
		}
		question.IDQuiz = input.IDQuiz
	}
	if input.QuestionText != "" {
		question.QuestionText = input.QuestionText
	}
	if input.Answer != "" {
		question.Answer = input.Answer
	}
	if input.Point != 0 {
		question.Point = input.Point
	}

	if err := database.DB.Db.Save(&question).Error; err != nil {
		utils.RespondDBError(context, err, "Question tidak ditemukan")
		return
	}
	context.JSON(http.StatusOK, question)
}

func DeleteQuestion(context *gin.Context) {
	id := context.Param("id")
	var question models.Question
	if err := database.DB.Db.First(&question, id).Error; err != nil {
		context.JSON(http.StatusNotFound, gin.H{"error": "Question not found"})
		context.Abort()
		return
	}

	if err := database.DB.Db.Delete(&question).Error; err != nil {
		utils.RespondServerError(context, err)
		return
	}
	context.JSON(http.StatusOK, gin.H{"message": "Question deleted successfully"})
}

package handlers

import (
	"net/http"

	"github.com/ALZEE23/ApiGo/database"
	"github.com/ALZEE23/ApiGo/models"
	"github.com/ALZEE23/ApiGo/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type optionInput struct {
	OptionText string `json:"option_text" binding:"required"`
	IsCorrect  bool   `json:"is_correct"`
}

func validateOptions(options []optionInput) string {
	if len(options) < 2 {
		return "a question needs at least 2 options"
	}
	correctCount := 0
	for _, o := range options {
		if o.IsCorrect {
			correctCount++
		}
	}
	if correctCount != 1 {
		return "a question must have exactly one correct option"
	}
	return ""
}

func Question(context *gin.Context) {
	var input struct {
		QuestionText string        `json:"question_text" binding:"required"`
		Point        int           `json:"point" binding:"required"`
		IDQuiz       uint          `json:"id_quiz" binding:"required"`
		Options      []optionInput `json:"options" binding:"required"`
	}
	if err := context.ShouldBindJSON(&input); err != nil {
		utils.RespondValidationError(context, "Data yang Anda masukkan tidak valid, mohon periksa kembali")
		return
	}

	if msg := validateOptions(input.Options); msg != "" {
		utils.RespondValidationError(context, msg)
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
		Point:        input.Point,
		IDQuiz:       input.IDQuiz,
	}

	err := database.DB.Db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&question).Error; err != nil {
			return err
		}
		for _, o := range input.Options {
			option := models.AnswerOption{
				OptionText: o.OptionText,
				IsCorrect:  o.IsCorrect,
				IDQuestion: question.ID,
			}
			if err := tx.Create(&option).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		utils.RespondDBError(context, err, "Data tidak ditemukan")
		return
	}

	database.DB.Db.Preload("Options").First(&question, question.ID)
	context.JSON(http.StatusCreated, question)
}

// publicOption/publicQuestion hide IsCorrect so quiz-takers can't read the
// correct answer straight from the API; correctness is only checked server-side.
type publicOption struct {
	ID         uint   `json:"id"`
	OptionText string `json:"option_text"`
}

type publicQuestion struct {
	ID           uint           `json:"id"`
	QuestionText string         `json:"question_text"`
	Point        int            `json:"point"`
	IDQuiz       uint           `json:"id_quiz"`
	Options      []publicOption `json:"options"`
}

func toPublicQuestion(q models.Question) publicQuestion {
	options := make([]publicOption, 0, len(q.Options))
	for _, o := range q.Options {
		options = append(options, publicOption{ID: o.ID, OptionText: o.OptionText})
	}
	return publicQuestion{
		ID:           q.ID,
		QuestionText: q.QuestionText,
		Point:        q.Point,
		IDQuiz:       q.IDQuiz,
		Options:      options,
	}
}

func GetQuestions(context *gin.Context) {
	var questions []models.Question
	query := database.DB.Db.Preload("Options")

	if idQuiz := context.Query("idQuiz"); idQuiz != "" {
		query = query.Where("id_quiz = ?", idQuiz)
	}

	if err := query.Find(&questions).Error; err != nil {
		utils.RespondServerError(context, err)
		return
	}

	result := make([]publicQuestion, 0, len(questions))
	for _, q := range questions {
		result = append(result, toPublicQuestion(q))
	}
	context.JSON(http.StatusOK, gin.H{"data": result})
}

func GetQuestionByID(context *gin.Context) {
	id := context.Param("id")
	var question models.Question
	if err := database.DB.Db.Preload("Options").First(&question, id).Error; err != nil {
		context.JSON(http.StatusNotFound, gin.H{"error": "Question not found"})
		context.Abort()
		return
	}
	context.JSON(http.StatusOK, toPublicQuestion(question))
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
		QuestionText string        `json:"question_text"`
		Point        int           `json:"point"`
		IDQuiz       uint          `json:"id_quiz"`
		Options      []optionInput `json:"options"`
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
	if input.Point != 0 {
		question.Point = input.Point
	}

	if input.Options != nil {
		if msg := validateOptions(input.Options); msg != "" {
			utils.RespondValidationError(context, msg)
			return
		}
	}

	err := database.DB.Db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(&question).Error; err != nil {
			return err
		}
		if input.Options != nil {
			if err := tx.Where("id_question = ?", question.ID).Delete(&models.AnswerOption{}).Error; err != nil {
				return err
			}
			for _, o := range input.Options {
				option := models.AnswerOption{
					OptionText: o.OptionText,
					IsCorrect:  o.IsCorrect,
					IDQuestion: question.ID,
				}
				if err := tx.Create(&option).Error; err != nil {
					return err
				}
			}
		}
		return nil
	})
	if err != nil {
		utils.RespondDBError(context, err, "Question tidak ditemukan")
		return
	}

	database.DB.Db.Preload("Options").First(&question, question.ID)
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

	err := database.DB.Db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("id_question = ?", question.ID).Delete(&models.AnswerOption{}).Error; err != nil {
			return err
		}
		return tx.Delete(&question).Error
	})
	if err != nil {
		utils.RespondServerError(context, err)
		return
	}
	context.JSON(http.StatusOK, gin.H{"message": "Question deleted successfully"})
}

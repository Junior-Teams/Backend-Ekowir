package handlers

import (
	"net/http"

	"github.com/ALZEE23/ApiGo/database"
	"github.com/ALZEE23/ApiGo/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SubmitQuiz(context *gin.Context) {
	quizID := context.Param("id")
	var quiz models.Quiz
	if err := database.DB.Db.First(&quiz, quizID).Error; err != nil {
		context.JSON(http.StatusNotFound, gin.H{"error": "Quiz not found"})
		context.Abort()
		return
	}

	var input struct {
		Answers []struct {
			QuestionID uint `json:"question_id" binding:"required"`
			OptionID   uint `json:"option_id" binding:"required"`
		} `json:"answers" binding:"required"`
	}
	if err := context.ShouldBindJSON(&input); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		context.Abort()
		return
	}

	var questions []models.Question
	if err := database.DB.Db.Where("id_quiz = ?", quiz.ID).Preload("Options").Find(&questions).Error; err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		context.Abort()
		return
	}
	if len(questions) == 0 {
		context.JSON(http.StatusBadRequest, gin.H{"error": "quiz has no questions"})
		context.Abort()
		return
	}

	questionByID := make(map[uint]models.Question, len(questions))
	for _, q := range questions {
		questionByID[q.ID] = q
	}

	score := 0
	xpFromQuestions := 0
	for _, a := range input.Answers {
		question, ok := questionByID[a.QuestionID]
		if !ok {
			continue
		}
		for _, opt := range question.Options {
			if opt.ID == a.OptionID && opt.IsCorrect {
				score++
				xpFromQuestions += question.Point
			}
		}
	}

	passed := score >= quiz.PassingScore
	response := gin.H{
		"score":           score,
		"total_questions": len(questions),
		"passed":          passed,
		"xp_earned":       0,
	}

	if !passed {
		context.JSON(http.StatusOK, response)
		return
	}

	email := context.GetString("email")
	var user models.User
	if err := database.DB.Db.Where("email = ?", email).First(&user).Error; err != nil {
		context.JSON(http.StatusOK, response)
		return
	}

	var existing models.QuizCompletion
	err := database.DB.Db.Where("user_id = ? AND id_quiz = ?", user.ID, quiz.ID).First(&existing).Error
	if err == nil {
		response["already_completed"] = true
		context.JSON(http.StatusOK, response)
		return
	}

	xpEarned := xpFromQuestions + quiz.BonusXp
	completion := models.QuizCompletion{UserID: user.ID, IDQuiz: quiz.ID, Score: score, XpEarned: xpEarned}

	txErr := database.DB.Db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&completion).Error; err != nil {
			return err
		}
		return tx.Model(&user).UpdateColumn("xp", gorm.Expr("xp + ?", xpEarned)).Error
	})
	if txErr != nil {
		// most likely a duplicate completion from a concurrent request; the
		// unique index on (user_id, id_quiz) is what actually guarantees
		// this can't be claimed twice, this is just the friendly response
		response["already_completed"] = true
		context.JSON(http.StatusOK, response)
		return
	}

	response["xp_earned"] = xpEarned
	response["bonus_xp"] = quiz.BonusXp

	if tier, changed, err := syncUserTier(user.ID); err == nil && tier != nil {
		response["tier"] = tier
		response["tier_changed"] = changed
	}

	context.JSON(http.StatusOK, response)
}

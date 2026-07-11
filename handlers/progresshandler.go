package handlers

import (
	"net/http"

	"github.com/ALZEE23/ApiGo/database"
	"github.com/ALZEE23/ApiGo/models"
	"github.com/ALZEE23/ApiGo/utils"
	"github.com/gin-gonic/gin"
)

// CompleteMateri idempotently marks a materi as completed by the current
// user. Visiting the same materi twice is a no-op thanks to FirstOrCreate +
// the unique (user_id, id_materi) index.
func CompleteMateri(context *gin.Context) {
	id := context.Param("id")
	var materi models.Materi
	if err := database.DB.Db.First(&materi, id).Error; err != nil {
		context.JSON(http.StatusNotFound, gin.H{"error": "Materi not found"})
		context.Abort()
		return
	}

	user, ok := currentUser(context)
	if !ok {
		return
	}

	var completion models.MateriCompletion
	err := database.DB.Db.
		Where(models.MateriCompletion{UserID: user.ID, IDMateri: materi.ID}).
		FirstOrCreate(&completion).Error
	if err != nil {
		utils.RespondDBError(context, err, "Materi tidak ditemukan")
		return
	}

	context.JSON(http.StatusOK, gin.H{"completed": true, "id_materi": materi.ID})
}

// moduleMateriProgress returns the materi ids of a module together with which
// of them the given user has completed. Shared by GetModuleProgress and the
// quiz-gating guard in SubmitQuiz.
func moduleMateriProgress(moduleID uint, userID uint) (materiIDs []uint, completedIDs []uint, err error) {
	if err = database.DB.Db.Model(&models.Materi{}).
		Where("id_module = ?", moduleID).
		Order("created_at asc, id asc").
		Pluck("id", &materiIDs).Error; err != nil {
		return nil, nil, err
	}

	completedIDs = []uint{}
	if len(materiIDs) > 0 {
		if err = database.DB.Db.Model(&models.MateriCompletion{}).
			Where("user_id = ? AND id_materi IN ?", userID, materiIDs).
			Pluck("id_materi", &completedIDs).Error; err != nil {
			return nil, nil, err
		}
	}
	return materiIDs, completedIDs, nil
}

// GetModuleProgress reports the current user's completion state for one
// module: which materi are done, whether all of them are, and per-quiz
// completion — everything the course-viewer sidebar needs in one call.
func GetModuleProgress(context *gin.Context) {
	id := context.Param("id")
	var module models.Module
	if err := database.DB.Db.First(&module, id).Error; err != nil {
		context.JSON(http.StatusNotFound, gin.H{"error": "Module not found"})
		context.Abort()
		return
	}

	user, ok := currentUser(context)
	if !ok {
		return
	}

	materiIDs, completedIDs, err := moduleMateriProgress(module.ID, user.ID)
	if err != nil {
		utils.RespondServerError(context, err)
		return
	}

	var quizzes []models.Quiz
	if err := database.DB.Db.Where("id_module = ?", module.ID).
		Order("created_at asc, id asc").Find(&quizzes).Error; err != nil {
		utils.RespondServerError(context, err)
		return
	}

	completedQuizIDs := []uint{}
	if len(quizzes) > 0 {
		quizIDs := make([]uint, 0, len(quizzes))
		for _, quiz := range quizzes {
			quizIDs = append(quizIDs, quiz.ID)
		}
		if err := database.DB.Db.Model(&models.QuizCompletion{}).
			Where("user_id = ? AND id_quiz IN ?", user.ID, quizIDs).
			Pluck("id_quiz", &completedQuizIDs).Error; err != nil {
			utils.RespondServerError(context, err)
			return
		}
	}
	completedQuizSet := make(map[uint]bool, len(completedQuizIDs))
	for _, quizID := range completedQuizIDs {
		completedQuizSet[quizID] = true
	}

	quizStates := make([]gin.H, 0, len(quizzes))
	for _, quiz := range quizzes {
		quizStates = append(quizStates, gin.H{
			"id_quiz":   quiz.ID,
			"completed": completedQuizSet[quiz.ID],
		})
	}

	context.JSON(http.StatusOK, gin.H{
		"id_module":             module.ID,
		"total_materi":          len(materiIDs),
		"completed_materi_ids":  completedIDs,
		"all_materi_completed":  len(completedIDs) == len(materiIDs),
		"quizzes":               quizStates,
	})
}

// GetMyCourseHistory lists every module the current user has touched (any
// materi or quiz completion), with enough aggregate info for the profile
// page to show an in-progress bar or a completed badge per course.
func GetMyCourseHistory(context *gin.Context) {
	user, ok := currentUser(context)
	if !ok {
		return
	}

	// Modules touched via materi completions.
	var materiModuleIDs []uint
	if err := database.DB.Db.Model(&models.MateriCompletion{}).
		Joins("JOIN materis ON materis.id = materi_completions.id_materi").
		Where("materi_completions.user_id = ?", user.ID).
		Distinct().Pluck("materis.id_module", &materiModuleIDs).Error; err != nil {
		utils.RespondServerError(context, err)
		return
	}

	// Modules touched via quiz completions.
	var quizModuleIDs []uint
	if err := database.DB.Db.Model(&models.QuizCompletion{}).
		Joins("JOIN quizzes ON quizzes.id = quiz_completions.id_quiz").
		Where("quiz_completions.user_id = ?", user.ID).
		Distinct().Pluck("quizzes.id_module", &quizModuleIDs).Error; err != nil {
		utils.RespondServerError(context, err)
		return
	}

	moduleIDSet := make(map[uint]bool, len(materiModuleIDs)+len(quizModuleIDs))
	moduleIDs := []uint{}
	for _, moduleID := range append(materiModuleIDs, quizModuleIDs...) {
		if !moduleIDSet[moduleID] {
			moduleIDSet[moduleID] = true
			moduleIDs = append(moduleIDs, moduleID)
		}
	}

	courses := []gin.H{}
	if len(moduleIDs) > 0 {
		var modules []models.Module
		if err := database.DB.Db.Where("id IN ?", moduleIDs).
			Order("created_at asc, id asc").Find(&modules).Error; err != nil {
			utils.RespondServerError(context, err)
			return
		}

		for _, module := range modules {
			materiIDs, completedIDs, err := moduleMateriProgress(module.ID, user.ID)
			if err != nil {
				utils.RespondServerError(context, err)
				return
			}

			var quizzes []models.Quiz
			if err := database.DB.Db.Where("id_module = ?", module.ID).
				Order("created_at asc, id asc").Find(&quizzes).Error; err != nil {
				utils.RespondServerError(context, err)
				return
			}

			completedQuizCount := int64(0)
			if len(quizzes) > 0 {
				quizIDs := make([]uint, 0, len(quizzes))
				for _, quiz := range quizzes {
					quizIDs = append(quizIDs, quiz.ID)
				}
				if err := database.DB.Db.Model(&models.QuizCompletion{}).
					Where("user_id = ? AND id_quiz IN ?", user.ID, quizIDs).
					Count(&completedQuizCount).Error; err != nil {
					utils.RespondServerError(context, err)
					return
				}
			}

			allMateriDone := len(completedIDs) == len(materiIDs)
			allQuizzesDone := completedQuizCount == int64(len(quizzes))
			status := "in_progress"
			if allMateriDone && allQuizzesDone {
				status = "completed"
			}

			courses = append(courses, gin.H{
				"id_module":         module.ID,
				"code_module":       module.CodeModule,
				"name_module":       module.NameModule,
				"image":             module.Image,
				"total_materi":      len(materiIDs),
				"completed_materi":  len(completedIDs),
				"total_quizzes":     len(quizzes),
				"completed_quizzes": completedQuizCount,
				"status":            status,
			})
		}
	}

	context.JSON(http.StatusOK, gin.H{"data": courses})
}

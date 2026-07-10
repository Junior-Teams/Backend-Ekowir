package handlers

import (
	"net/http"

	"github.com/ALZEE23/ApiGo/database"
	"github.com/ALZEE23/ApiGo/models"
	"github.com/ALZEE23/ApiGo/utils"
	"github.com/gin-gonic/gin"
)

func Forum(context *gin.Context) {
	var input struct {
		Title       string `json:"title" binding:"required"`
		Description string `json:"description" binding:"required"`
		IDModule    uint   `json:"id_module" binding:"required"`
	}
	if err := context.ShouldBindJSON(&input); err != nil {
		utils.RespondValidationError(context, "Data yang Anda masukkan tidak valid, mohon periksa kembali")
		return
	}

	user, ok := currentUser(context)
	if !ok {
		return
	}

	var module models.Module
	if err := database.DB.Db.First(&module, input.IDModule).Error; err != nil {
		context.JSON(http.StatusNotFound, gin.H{"error": "Module not found"})
		context.Abort()
		return
	}

	forum := models.Forum{
		Title:       input.Title,
		Description: input.Description,
		CreatedBy:   user.Username,
		UserID:      user.ID,
		IDModule:    input.IDModule,
	}

	if err := database.DB.Db.Create(&forum).Error; err != nil {
		utils.RespondDBError(context, err, "Data tidak ditemukan")
		return
	}

	xp := ForumPostXP
	capReached := forumPostsToday(user.ID) > ForumPostDailyCap
	if capReached {
		xp = 0
	}

	gamify, err := awardXp(user.ID, xp)
	if err != nil {
		utils.RespondServerError(context, err)
		return
	}
	gamify.DailyCapReached = capReached

	context.JSON(http.StatusCreated, gin.H{
		"forumId":      forum.ID,
		"title":        forum.Title,
		"description":  forum.Description,
		"created_by":   forum.CreatedBy,
		"id_module":    forum.IDModule,
		"gamification": gamify,
	})
}

func GetForums(context *gin.Context) {
	forums := []models.Forum{}
	query := database.DB.Db.Preload("Module").Preload("User").Preload("User.Tier")

	if idModule := context.Query("idModule"); idModule != "" {
		query = query.Where("id_module = ?", idModule)
	}

	if err := query.Order("created_at desc").Find(&forums).Error; err != nil {
		utils.RespondServerError(context, err)
		return
	}
	context.JSON(http.StatusOK, gin.H{"data": forums, "count": len(forums)})
}

func GetForumByID(context *gin.Context) {
	id := context.Param("id")
	var forum models.Forum
	if err := database.DB.Db.Preload("Module").Preload("User").Preload("User.Tier").First(&forum, id).Error; err != nil {
		context.JSON(http.StatusNotFound, gin.H{"error": "Forum not found"})
		context.Abort()
		return
	}

	// Comments are the forum's "turunan" (derivative) data - always fetched
	// as a real, possibly-empty slice so the client can render a proper
	// "belum ada komentar" state instead of choking on null.
	comments := []models.Comment{}
	database.DB.Db.Preload("User").Preload("User.Tier").
		Where("id_forum = ?", forum.ID).
		Order("created_at asc").
		Find(&comments)

	type forumDetail struct {
		models.Forum
		Comments     []models.Comment `json:"comments"`
		CommentCount int              `json:"comment_count"`
	}

	context.JSON(http.StatusOK, forumDetail{
		Forum:        forum,
		Comments:     comments,
		CommentCount: len(comments),
	})
}

func UpdateForum(context *gin.Context) {
	id := context.Param("id")
	var forum models.Forum
	if err := database.DB.Db.First(&forum, id).Error; err != nil {
		context.JSON(http.StatusNotFound, gin.H{"error": "Forum not found"})
		context.Abort()
		return
	}

	user, ok := currentUser(context)
	if !ok {
		return
	}
	if !isOwnerOrAdmin(user, forum.UserID) {
		context.JSON(http.StatusForbidden, gin.H{"error": "Anda tidak berhak mengubah forum ini"})
		context.Abort()
		return
	}

	var input struct {
		Title       string `json:"title"`
		Description string `json:"description"`
	}
	if err := context.ShouldBindJSON(&input); err != nil {
		utils.RespondValidationError(context, "Data yang Anda masukkan tidak valid, mohon periksa kembali")
		return
	}

	if input.Title != "" {
		forum.Title = input.Title
	}
	if input.Description != "" {
		forum.Description = input.Description
	}

	if err := database.DB.Db.Save(&forum).Error; err != nil {
		utils.RespondDBError(context, err, "Forum tidak ditemukan")
		return
	}
	context.JSON(http.StatusOK, forum)
}

func DeleteForum(context *gin.Context) {
	id := context.Param("id")
	var forum models.Forum
	if err := database.DB.Db.First(&forum, id).Error; err != nil {
		context.JSON(http.StatusNotFound, gin.H{"error": "Forum not found"})
		context.Abort()
		return
	}

	user, ok := currentUser(context)
	if !ok {
		return
	}
	if !isOwnerOrAdmin(user, forum.UserID) {
		context.JSON(http.StatusForbidden, gin.H{"error": "Anda tidak berhak menghapus forum ini"})
		context.Abort()
		return
	}

	if err := database.DB.Db.Delete(&forum).Error; err != nil {
		utils.RespondServerError(context, err)
		return
	}
	context.JSON(http.StatusOK, gin.H{"message": "Forum deleted successfully"})
}

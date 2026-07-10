package handlers

import (
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/ALZEE23/ApiGo/database"
	"github.com/ALZEE23/ApiGo/models"
	"github.com/ALZEE23/ApiGo/utils"
	"github.com/gin-gonic/gin"
)

func Comment(context *gin.Context) {
	content := context.PostForm("content")
	idForumRaw := context.PostForm("idForum")

	if content == "" || idForumRaw == "" {
		utils.RespondValidationError(context, "content dan idForum wajib diisi")
		return
	}

	idForum, err := strconv.ParseUint(idForumRaw, 10, 64)
	if err != nil {
		utils.RespondValidationError(context, "idForum harus berupa angka yang valid")
		return
	}

	user, ok := currentUser(context)
	if !ok {
		return
	}

	var forum models.Forum
	if err := database.DB.Db.First(&forum, idForum).Error; err != nil {
		context.JSON(http.StatusNotFound, gin.H{"error": "Forum not found"})
		context.Abort()
		return
	}

	comment := models.Comment{
		Content:   content,
		CreatedBy: user.Username,
		UserID:    user.ID,
		IDForum:   uint(idForum),
	}

	imageHeader, err := context.FormFile("image")
	if err == nil {
		imagePath := filepath.Join("storage", imageHeader.Filename)
		if err := context.SaveUploadedFile(imageHeader, imagePath); err != nil {
			utils.RespondServerError(context, err)
			return
		}
		comment.Image = imagePath
	}

	if err := database.DB.Db.Create(&comment).Error; err != nil {
		utils.RespondDBError(context, err, "Data tidak ditemukan")
		return
	}

	xp := ForumCommentXP
	capReached := forumCommentsToday(user.ID) > ForumCommentDailyCap
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
		"commentId":    comment.ID,
		"content":      comment.Content,
		"created_by":   comment.CreatedBy,
		"image":        comment.Image,
		"id_forum":     comment.IDForum,
		"gamification": gamify,
	})
}

func GetComments(context *gin.Context) {
	comments := []models.Comment{}
	query := database.DB.Db.Preload("Forum").Preload("User").Preload("User.Tier")

	if idForum := context.Query("idForum"); idForum != "" {
		query = query.Where("id_forum = ?", idForum)
	}

	if err := query.Order("created_at asc").Find(&comments).Error; err != nil {
		utils.RespondServerError(context, err)
		return
	}
	context.JSON(http.StatusOK, gin.H{"data": comments, "count": len(comments)})
}

func GetCommentByID(context *gin.Context) {
	id := context.Param("id")
	var comment models.Comment
	if err := database.DB.Db.Preload("Forum").Preload("User").Preload("User.Tier").First(&comment, id).Error; err != nil {
		context.JSON(http.StatusNotFound, gin.H{"error": "Comment not found"})
		context.Abort()
		return
	}
	context.JSON(http.StatusOK, comment)
}

func UpdateComment(context *gin.Context) {
	id := context.Param("id")
	var comment models.Comment
	if err := database.DB.Db.First(&comment, id).Error; err != nil {
		context.JSON(http.StatusNotFound, gin.H{"error": "Comment not found"})
		context.Abort()
		return
	}

	user, ok := currentUser(context)
	if !ok {
		return
	}
	if !isOwnerOrAdmin(user, comment.UserID) {
		context.JSON(http.StatusForbidden, gin.H{"error": "Anda tidak berhak mengubah komentar ini"})
		context.Abort()
		return
	}

	var input struct {
		Content string `json:"content"`
	}
	if err := context.ShouldBindJSON(&input); err != nil {
		utils.RespondValidationError(context, "Data yang Anda masukkan tidak valid, mohon periksa kembali")
		return
	}

	if input.Content != "" {
		comment.Content = input.Content
	}

	if err := database.DB.Db.Save(&comment).Error; err != nil {
		utils.RespondDBError(context, err, "Comment tidak ditemukan")
		return
	}
	context.JSON(http.StatusOK, comment)
}

func DeleteComment(context *gin.Context) {
	id := context.Param("id")
	var comment models.Comment
	if err := database.DB.Db.First(&comment, id).Error; err != nil {
		context.JSON(http.StatusNotFound, gin.H{"error": "Comment not found"})
		context.Abort()
		return
	}

	user, ok := currentUser(context)
	if !ok {
		return
	}
	if !isOwnerOrAdmin(user, comment.UserID) {
		context.JSON(http.StatusForbidden, gin.H{"error": "Anda tidak berhak menghapus komentar ini"})
		context.Abort()
		return
	}

	if err := database.DB.Db.Delete(&comment).Error; err != nil {
		utils.RespondServerError(context, err)
		return
	}
	context.JSON(http.StatusOK, gin.H{"message": "Comment deleted successfully"})
}

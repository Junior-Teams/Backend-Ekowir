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
	createdBy := context.PostForm("createdBy")
	idForumRaw := context.PostForm("idForum")

	if content == "" || createdBy == "" || idForumRaw == "" {
		context.JSON(http.StatusBadRequest, gin.H{"error": "content, createdBy and idForum are required"})
		context.Abort()
		return
	}

	idForum, err := strconv.ParseUint(idForumRaw, 10, 64)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": "idForum must be a valid number"})
		context.Abort()
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
		CreatedBy: createdBy,
		IDForum:   uint(idForum),
	}

	imageHeader, err := context.FormFile("image")
	if err == nil {
		imagePath := filepath.Join("storage", imageHeader.Filename)
		if err := context.SaveUploadedFile(imageHeader, imagePath); err != nil {
			context.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save image file"})
			context.Abort()
			return
		}
		comment.Image = imagePath
	}

	record := database.DB.Db.Create(&comment)
	if record.Error != nil {
		utils.RespondDBError(context, record.Error, "Data tidak ditemukan")
		return
	}
	context.JSON(http.StatusCreated, gin.H{"commentId": comment.ID, "content": comment.Content, "created_by": comment.CreatedBy, "image": comment.Image, "id_forum": comment.IDForum})
}

func GetComments(context *gin.Context) {
	var comments []models.Comment
	query := database.DB.Db.Preload("Forum")

	if idForum := context.Query("idForum"); idForum != "" {
		query = query.Where("id_forum = ?", idForum)
	}

	if err := query.Find(&comments).Error; err != nil {
		utils.RespondServerError(context, err)
		return
	}
	context.JSON(http.StatusOK, gin.H{"data": comments})
}

func GetCommentByID(context *gin.Context) {
	id := context.Param("id")
	var comment models.Comment
	if err := database.DB.Db.Preload("Forum").First(&comment, id).Error; err != nil {
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

	if err := database.DB.Db.Delete(&comment).Error; err != nil {
		utils.RespondServerError(context, err)
		return
	}
	context.JSON(http.StatusOK, gin.H{"message": "Comment deleted successfully"})
}
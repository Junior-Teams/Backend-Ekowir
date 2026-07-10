package handlers

import (
	"net/http"
	"path/filepath"

	"github.com/ALZEE23/ApiGo/database"
	"github.com/ALZEE23/ApiGo/models"
	"github.com/gin-gonic/gin"
)

func Comment(context *gin.Context) {
	var input struct {
		Content   string `json:"content" binding:"required"`
		CreatedBy string `json:"created_by" binding:"required"`
		Image     string `json:"image"`
		IDForum   uint   `json:"id_forum" binding:"required"`
	}
	if err := context.ShouldBindJSON(&input); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		context.Abort()
		return
	}

	imageHeader, err := context.FormFile("image")
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": "Failed to get image file"})
		return
	}

	imagePath := filepath.Join("storage", imageHeader.Filename)
	if err := context.SaveUploadedFile(imageHeader, imagePath); err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save image file"})
		return
	}

	comment := models.Comment{
		Content:   input.Content,
		CreatedBy: input.CreatedBy,
		Image:     imagePath,
		IDForum:   input.IDForum,
	}

	record := database.DB.Db.Create(&comment)
	if record.Error != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": record.Error.Error()})
		context.Abort()
		return
	}
	context.JSON(http.StatusCreated, gin.H{"commentId": comment.ID, "content": comment.Content, "created_by": comment.CreatedBy, "id_forum": comment.IDForum})
}
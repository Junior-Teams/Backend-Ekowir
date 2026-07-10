package handlers

import (
	"net/http"

	"github.com/ALZEE23/ApiGo/database"
	"github.com/ALZEE23/ApiGo/models"
	"github.com/gin-gonic/gin"
)

func Forum(context *gin.Context) {
	var input struct {
		Title       string `json:"title" binding:"required"`
		Description string `json:"description" binding:"required"`
		CreatedBy   string `json:"created_by" binding:"required"`
		IDModule    uint   `json:"id_module" binding:"required"`
	}
	if err := context.ShouldBindJSON(&input); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		context.Abort()
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
		CreatedBy:   input.CreatedBy,
		IDModule:    input.IDModule,
	}

	record := database.DB.Db.Create(&forum)
	if record.Error != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": record.Error.Error()})
		context.Abort()
		return
	}
	context.JSON(http.StatusCreated, gin.H{"forumId": forum.ID, "title": forum.Title, "description": forum.Description, "created_by": forum.CreatedBy, "id_module": forum.IDModule})
}

func GetForums(context *gin.Context) {
	var forums []models.Forum
	query := database.DB.Db.Preload("Module")

	if idModule := context.Query("idModule"); idModule != "" {
		query = query.Where("id_module = ?", idModule)
	}

	if err := query.Find(&forums).Error; err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		context.Abort()
		return
	}
	context.JSON(http.StatusOK, gin.H{"data": forums})
}

func GetForumByID(context *gin.Context) {
	id := context.Param("id")
	var forum models.Forum
	if err := database.DB.Db.Preload("Module").First(&forum, id).Error; err != nil {
		context.JSON(http.StatusNotFound, gin.H{"error": "Forum not found"})
		context.Abort()
		return
	}
	context.JSON(http.StatusOK, forum)
}

func UpdateForum(context *gin.Context) {
	id := context.Param("id")
	var forum models.Forum
	if err := database.DB.Db.First(&forum, id).Error; err != nil {
		context.JSON(http.StatusNotFound, gin.H{"error": "Forum not found"})
		context.Abort()
		return
	}

	var input struct {
		Title       string `json:"title"`
		Description string `json:"description"`
	}
	if err := context.ShouldBindJSON(&input); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		context.Abort()
		return
	}

	if input.Title != "" {
		forum.Title = input.Title
	}
	if input.Description != "" {
		forum.Description = input.Description
	}

	if err := database.DB.Db.Save(&forum).Error; err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		context.Abort()
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

	if err := database.DB.Db.Delete(&forum).Error; err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		context.Abort()
		return
	}
	context.JSON(http.StatusOK, gin.H{"message": "Forum deleted successfully"})
}

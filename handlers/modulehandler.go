package handlers

import (
	"net/http"
	"path/filepath"

	"github.com/ALZEE23/ApiGo/database"
	"github.com/ALZEE23/ApiGo/models"
	"github.com/ALZEE23/ApiGo/utils"
	"github.com/gin-gonic/gin"
)

func Module(context *gin.Context) {
	codeModule := context.PostForm("codeModule")
	nameModule := context.PostForm("nameModule")
	descriptionModule := context.PostForm("descriptionModule")

	if codeModule == "" || nameModule == "" || descriptionModule == "" {
		context.JSON(http.StatusBadRequest, gin.H{"error": "codeModule, nameModule and descriptionModule are required"})
		context.Abort()
		return
	}

	imageHeader, err := context.FormFile("image")
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": "Failed to get image file"})
		context.Abort()
		return
	}

	imagePath := filepath.Join("storage", imageHeader.Filename)
	if err := context.SaveUploadedFile(imageHeader, imagePath); err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save image file"})
		context.Abort()
		return
	}

	module := models.Module{
		CodeModule:        codeModule,
		NameModule:        nameModule,
		DescriptionModule: descriptionModule,
		Image:             imagePath,
	}

	record := database.DB.Db.Create(&module)
	if record.Error != nil {
		utils.RespondDBError(context, record.Error, "Data tidak ditemukan")
		return
	}

	context.JSON(http.StatusCreated, gin.H{"moduleId": module.ID, "codeModule": module.CodeModule, "nameModule": module.NameModule, "descriptionModule": module.DescriptionModule, "image": module.Image})
}

func GetModules(context *gin.Context) {
	var modules []models.Module
	if err := database.DB.Db.Find(&modules).Error; err != nil {
		utils.RespondServerError(context, err)
		return
	}
	context.JSON(http.StatusOK, gin.H{"data": modules})
}

func GetModuleByID(context *gin.Context) {
	id := context.Param("id")
	var module models.Module
	if err := database.DB.Db.First(&module, id).Error; err != nil {
		context.JSON(http.StatusNotFound, gin.H{"error": "Module not found"})
		context.Abort()
		return
	}
	context.JSON(http.StatusOK, module)
}

func UpdateModule(context *gin.Context) {
	id := context.Param("id")
	var module models.Module
	if err := database.DB.Db.First(&module, id).Error; err != nil {
		context.JSON(http.StatusNotFound, gin.H{"error": "Module not found"})
		context.Abort()
		return
	}

	var input struct {
		CodeModule        string `json:"code_module"`
		NameModule        string `json:"name_module"`
		DescriptionModule string `json:"description_module"`
	}
	if err := context.ShouldBindJSON(&input); err != nil {
		utils.RespondValidationError(context, "Data yang Anda masukkan tidak valid, mohon periksa kembali")
		return
	}

	if input.CodeModule != "" {
		module.CodeModule = input.CodeModule
	}
	if input.NameModule != "" {
		module.NameModule = input.NameModule
	}
	if input.DescriptionModule != "" {
		module.DescriptionModule = input.DescriptionModule
	}

	if err := database.DB.Db.Save(&module).Error; err != nil {
		utils.RespondDBError(context, err, "Module tidak ditemukan")
		return
	}
	context.JSON(http.StatusOK, module)
}

func DeleteModule(context *gin.Context) {
	id := context.Param("id")
	var module models.Module
	if err := database.DB.Db.First(&module, id).Error; err != nil {
		context.JSON(http.StatusNotFound, gin.H{"error": "Module not found"})
		context.Abort()
		return
	}

	if err := database.DB.Db.Delete(&module).Error; err != nil {
		utils.RespondServerError(context, err)
		return
	}
	context.JSON(http.StatusOK, gin.H{"message": "Module deleted successfully"})
}

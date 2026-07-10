package handlers

import (
	"net/http"

	"github.com/ALZEE23/ApiGo/database"
	"github.com/ALZEE23/ApiGo/models"
	"github.com/ALZEE23/ApiGo/utils"
	"github.com/gin-gonic/gin"
)

func Materi(context *gin.Context) {
	var input struct {
		Name         string `json:"name" binding:"required"`
		ArrayElement string `json:"array_element" binding:"required"`
		IDModule     uint   `json:"id_module" binding:"required"`
	}
	if err := context.ShouldBindJSON(&input); err != nil {
		utils.RespondValidationError(context, "Data yang Anda masukkan tidak valid, mohon periksa kembali")
		return
	}

	var module models.Module
	if err := database.DB.Db.First(&module, input.IDModule).Error; err != nil {
		context.JSON(http.StatusNotFound, gin.H{"error": "Module not found"})
		context.Abort()
		return
	}

	materi := models.Materi{
		Name:         input.Name,
		ArrayElement: input.ArrayElement,
		IDModule:     input.IDModule,
	}

	record := database.DB.Db.Create(&materi)
	if record.Error != nil {
		utils.RespondDBError(context, record.Error, "Data tidak ditemukan")
		return
	}
	context.JSON(http.StatusCreated, gin.H{"materiId": materi.ID, "name": materi.Name, "array_element": materi.ArrayElement, "id_module": materi.IDModule})
}

func GetMateris(context *gin.Context) {
	var materis []models.Materi
	query := database.DB.Db.Preload("Module")

	if idModule := context.Query("idModule"); idModule != "" {
		query = query.Where("id_module = ?", idModule)
	}

	if err := query.Find(&materis).Error; err != nil {
		utils.RespondServerError(context, err)
		return
	}
	context.JSON(http.StatusOK, gin.H{"data": materis})
}

func GetMateriByID(context *gin.Context) {
	id := context.Param("id")
	var materi models.Materi
	if err := database.DB.Db.Preload("Module").First(&materi, id).Error; err != nil {
		context.JSON(http.StatusNotFound, gin.H{"error": "Materi not found"})
		context.Abort()
		return
	}
	context.JSON(http.StatusOK, materi)
}

func UpdateMateri(context *gin.Context) {
	id := context.Param("id")
	var materi models.Materi
	if err := database.DB.Db.First(&materi, id).Error; err != nil {
		context.JSON(http.StatusNotFound, gin.H{"error": "Materi not found"})
		context.Abort()
		return
	}

	var input struct {
		Name         string `json:"name"`
		ArrayElement string `json:"array_element"`
		IDModule     uint   `json:"id_module"`
	}
	if err := context.ShouldBindJSON(&input); err != nil {
		utils.RespondValidationError(context, "Data yang Anda masukkan tidak valid, mohon periksa kembali")
		return
	}

	if input.IDModule != 0 {
		var module models.Module
		if err := database.DB.Db.First(&module, input.IDModule).Error; err != nil {
			context.JSON(http.StatusNotFound, gin.H{"error": "Module not found"})
			context.Abort()
			return
		}
		materi.IDModule = input.IDModule
	}
	if input.Name != "" {
		materi.Name = input.Name
	}
	if input.ArrayElement != "" {
		materi.ArrayElement = input.ArrayElement
	}

	if err := database.DB.Db.Save(&materi).Error; err != nil {
		utils.RespondDBError(context, err, "Materi tidak ditemukan")
		return
	}
	context.JSON(http.StatusOK, materi)
}

func DeleteMateri(context *gin.Context) {
	id := context.Param("id")
	var materi models.Materi
	if err := database.DB.Db.First(&materi, id).Error; err != nil {
		context.JSON(http.StatusNotFound, gin.H{"error": "Materi not found"})
		context.Abort()
		return
	}

	if err := database.DB.Db.Delete(&materi).Error; err != nil {
		utils.RespondServerError(context, err)
		return
	}
	context.JSON(http.StatusOK, gin.H{"message": "Materi deleted successfully"})
}

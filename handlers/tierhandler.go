package handlers

import (
	"net/http"
	"strconv"

	"github.com/ALZEE23/ApiGo/database"
	"github.com/ALZEE23/ApiGo/models"
	"github.com/gin-gonic/gin"
)

func Tier(context *gin.Context) {
	var input struct {
		Name  string `json:"name" binding:"required"`
		Point int    `json:"point" binding:"required"`
	}
	if err := context.ShouldBindJSON(&input); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		context.Abort()
		return
	}

	tier := models.Tier{Name: input.Name, Point: input.Point}
	if err := database.DB.Db.Create(&tier).Error; err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		context.Abort()
		return
	}
	context.JSON(http.StatusCreated, tier)
}

func GetTiers(context *gin.Context) {
	tiers := []models.Tier{}
	if err := database.DB.Db.Order("point asc").Find(&tiers).Error; err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		context.Abort()
		return
	}
	context.JSON(http.StatusOK, gin.H{"data": tiers, "count": len(tiers)})
}

func GetTierByID(context *gin.Context) {
	id := context.Param("id")
	var tier models.Tier
	if err := database.DB.Db.First(&tier, id).Error; err != nil {
		context.JSON(http.StatusNotFound, gin.H{"error": "Tier not found"})
		context.Abort()
		return
	}
	context.JSON(http.StatusOK, tier)
}

func UpdateTier(context *gin.Context) {
	id := context.Param("id")
	var tier models.Tier
	if err := database.DB.Db.First(&tier, id).Error; err != nil {
		context.JSON(http.StatusNotFound, gin.H{"error": "Tier not found"})
		context.Abort()
		return
	}

	var input struct {
		Name  string `json:"name"`
		Point int    `json:"point"`
	}
	if err := context.ShouldBindJSON(&input); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		context.Abort()
		return
	}

	if input.Name != "" {
		tier.Name = input.Name
	}
	pointChanged := input.Point != 0 && input.Point != tier.Point
	if input.Point != 0 {
		tier.Point = input.Point
	}

	if err := database.DB.Db.Save(&tier).Error; err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		context.Abort()
		return
	}

	if pointChanged {
		resyncAllUserTiers()
	}

	context.JSON(http.StatusOK, tier)
}

func DeleteTier(context *gin.Context) {
	id := context.Param("id")
	var tier models.Tier
	if err := database.DB.Db.First(&tier, id).Error; err != nil {
		context.JSON(http.StatusNotFound, gin.H{"error": "Tier not found"})
		context.Abort()
		return
	}

	if err := database.DB.Db.Delete(&tier).Error; err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		context.Abort()
		return
	}

	resyncAllUserTiers()

	context.JSON(http.StatusOK, gin.H{"message": "Tier deleted successfully"})
}

// resyncAllUserTiers re-checks every user's tier. Used when a tier's
// threshold changes or a tier is deleted, since that can shift who
// qualifies for what without anyone's xp actually changing.
func resyncAllUserTiers() {
	var ids []uint
	database.DB.Db.Model(&models.User{}).Pluck("id", &ids)
	for _, id := range ids {
		syncUserTier(id)
	}
}

func GetLeaderboard(context *gin.Context) {
	limit := 20
	if l := context.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	var users []models.User
	// Admins aren't participants in the gamification track, so they're
	// excluded rather than just hidden client-side - that keeps ranks
	// contiguous (1, 2, 3, ...) among real competitors.
	if err := database.DB.Db.Preload("Tier").Where("role != ?", "admin").Order("xp desc").Limit(limit).Find(&users).Error; err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		context.Abort()
		return
	}

	type entry struct {
		Rank     int          `json:"rank"`
		ID       uint         `json:"id"`
		Username string       `json:"username"`
		Name     string       `json:"name"`
		Picture  string       `json:"picture"`
		Xp       int          `json:"xp"`
		Tier     *models.Tier `json:"tier,omitempty"`
	}

	result := make([]entry, 0, len(users))
	for i, u := range users {
		result = append(result, entry{
			Rank:     i + 1,
			ID:       u.ID,
			Username: u.Username,
			Name:     u.Name,
			Picture:  u.Picture,
			Xp:       u.Xp,
			Tier:     u.Tier,
		})
	}
	context.JSON(http.StatusOK, gin.H{"data": result, "count": len(result)})
}

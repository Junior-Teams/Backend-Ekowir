package handlers

import (
	"net/http"

	"github.com/ALZEE23/ApiGo/database"
	"github.com/ALZEE23/ApiGo/models"
	"github.com/gin-gonic/gin"
)

func Reward(context *gin.Context) {
	var input struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
		RequiredXp  int    `json:"required_xp" binding:"required"`
	}
	if err := context.ShouldBindJSON(&input); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		context.Abort()
		return
	}

	reward := models.Reward{Name: input.Name, Description: input.Description, RequiredXp: input.RequiredXp}
	if err := database.DB.Db.Create(&reward).Error; err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		context.Abort()
		return
	}
	context.JSON(http.StatusCreated, reward)
}

func GetRewards(context *gin.Context) {
	var rewards []models.Reward
	if err := database.DB.Db.Order("required_xp asc").Find(&rewards).Error; err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		context.Abort()
		return
	}
	context.JSON(http.StatusOK, gin.H{"data": rewards})
}

func GetRewardByID(context *gin.Context) {
	id := context.Param("id")
	var reward models.Reward
	if err := database.DB.Db.First(&reward, id).Error; err != nil {
		context.JSON(http.StatusNotFound, gin.H{"error": "Reward not found"})
		context.Abort()
		return
	}
	context.JSON(http.StatusOK, reward)
}

func UpdateReward(context *gin.Context) {
	id := context.Param("id")
	var reward models.Reward
	if err := database.DB.Db.First(&reward, id).Error; err != nil {
		context.JSON(http.StatusNotFound, gin.H{"error": "Reward not found"})
		context.Abort()
		return
	}

	var input struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		RequiredXp  int    `json:"required_xp"`
	}
	if err := context.ShouldBindJSON(&input); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		context.Abort()
		return
	}

	if input.Name != "" {
		reward.Name = input.Name
	}
	if input.Description != "" {
		reward.Description = input.Description
	}
	if input.RequiredXp != 0 {
		reward.RequiredXp = input.RequiredXp
	}

	if err := database.DB.Db.Save(&reward).Error; err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		context.Abort()
		return
	}
	context.JSON(http.StatusOK, reward)
}

func DeleteReward(context *gin.Context) {
	id := context.Param("id")
	var reward models.Reward
	if err := database.DB.Db.First(&reward, id).Error; err != nil {
		context.JSON(http.StatusNotFound, gin.H{"error": "Reward not found"})
		context.Abort()
		return
	}

	if err := database.DB.Db.Delete(&reward).Error; err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		context.Abort()
		return
	}
	context.JSON(http.StatusOK, gin.H{"message": "Reward deleted successfully"})
}

// GetMyRewards lists every reward alongside whether the current user is
// eligible for it (xp-wise) and whether they've already claimed it.
func GetMyRewards(context *gin.Context) {
	email := context.GetString("email")
	var user models.User
	if err := database.DB.Db.Where("email = ?", email).First(&user).Error; err != nil {
		context.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		context.Abort()
		return
	}

	var rewards []models.Reward
	if err := database.DB.Db.Order("required_xp asc").Find(&rewards).Error; err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		context.Abort()
		return
	}

	var claims []models.RewardClaim
	database.DB.Db.Where("user_id = ?", user.ID).Find(&claims)
	claimed := make(map[uint]bool, len(claims))
	for _, c := range claims {
		claimed[c.RewardID] = true
	}

	type entry struct {
		models.Reward
		Eligible bool `json:"eligible"`
		Claimed  bool `json:"claimed"`
	}

	result := make([]entry, 0, len(rewards))
	for _, r := range rewards {
		result = append(result, entry{
			Reward:   r,
			Eligible: user.Xp >= r.RequiredXp,
			Claimed:  claimed[r.ID],
		})
	}
	context.JSON(http.StatusOK, gin.H{"data": result})
}

func ClaimReward(context *gin.Context) {
	id := context.Param("id")
	var reward models.Reward
	if err := database.DB.Db.First(&reward, id).Error; err != nil {
		context.JSON(http.StatusNotFound, gin.H{"error": "Reward not found"})
		context.Abort()
		return
	}

	email := context.GetString("email")
	var user models.User
	if err := database.DB.Db.Where("email = ?", email).First(&user).Error; err != nil {
		context.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		context.Abort()
		return
	}

	if user.Xp < reward.RequiredXp {
		context.JSON(http.StatusBadRequest, gin.H{"error": "not enough xp to claim this reward"})
		context.Abort()
		return
	}

	var existing models.RewardClaim
	if err := database.DB.Db.Where("user_id = ? AND reward_id = ?", user.ID, reward.ID).First(&existing).Error; err == nil {
		context.JSON(http.StatusConflict, gin.H{"error": "reward already claimed"})
		context.Abort()
		return
	}

	claim := models.RewardClaim{UserID: user.ID, RewardID: reward.ID}
	if err := database.DB.Db.Create(&claim).Error; err != nil {
		// the unique index on (user_id, reward_id) is the real guard against
		// double-claims under a race; this just surfaces it as a normal 409
		context.JSON(http.StatusConflict, gin.H{"error": "reward already claimed"})
		context.Abort()
		return
	}

	context.JSON(http.StatusCreated, gin.H{"message": "Reward claimed successfully", "reward": reward})
}

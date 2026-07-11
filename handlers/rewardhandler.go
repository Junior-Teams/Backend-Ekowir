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

// saveRewardImage stores the optional "image" form file and returns its
// storage path. An empty path with ok=true means no image was uploaded.
func saveRewardImage(context *gin.Context) (path string, ok bool) {
	imageHeader, err := context.FormFile("image")
	if err != nil {
		return "", true
	}
	imagePath := filepath.Join("storage", imageHeader.Filename)
	if err := context.SaveUploadedFile(imageHeader, imagePath); err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save image file"})
		context.Abort()
		return "", false
	}
	return imagePath, true
}

func Reward(context *gin.Context) {
	name := context.PostForm("name")
	description := context.PostForm("description")
	requiredXp, err := strconv.Atoi(context.PostForm("required_xp"))
	if name == "" || err != nil || requiredXp <= 0 {
		utils.RespondValidationError(context, "Data yang Anda masukkan tidak valid, mohon periksa kembali")
		return
	}

	imagePath, ok := saveRewardImage(context)
	if !ok {
		return
	}

	reward := models.Reward{Name: name, Description: description, Image: imagePath, RequiredXp: requiredXp}
	if err := database.DB.Db.Create(&reward).Error; err != nil {
		utils.RespondDBError(context, err, "Data tidak ditemukan")
		return
	}
	context.JSON(http.StatusCreated, reward)
}

func GetRewards(context *gin.Context) {
	rewards := []models.Reward{}
	if err := database.DB.Db.Order("required_xp asc").Find(&rewards).Error; err != nil {
		utils.RespondServerError(context, err)
		return
	}
	context.JSON(http.StatusOK, gin.H{"data": rewards, "count": len(rewards)})
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

	if name := context.PostForm("name"); name != "" {
		reward.Name = name
	}
	if description := context.PostForm("description"); description != "" {
		reward.Description = description
	}
	if rawXp := context.PostForm("required_xp"); rawXp != "" {
		requiredXp, err := strconv.Atoi(rawXp)
		if err != nil || requiredXp <= 0 {
			utils.RespondValidationError(context, "Data yang Anda masukkan tidak valid, mohon periksa kembali")
			return
		}
		reward.RequiredXp = requiredXp
	}
	if imagePath, ok := saveRewardImage(context); !ok {
		return
	} else if imagePath != "" {
		reward.Image = imagePath
	}

	if err := database.DB.Db.Save(&reward).Error; err != nil {
		utils.RespondDBError(context, err, "Hadiah tidak ditemukan")
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
		utils.RespondServerError(context, err)
		return
	}
	context.JSON(http.StatusOK, gin.H{"message": "Reward deleted successfully"})
}

// GetMyRewards lists every reward alongside whether the current user is
// eligible for it (xp-wise) and whether they've already claimed it.
func GetMyRewards(context *gin.Context) {
	user, ok := currentUser(context)
	if !ok {
		return
	}

	rewards := []models.Reward{}
	if err := database.DB.Db.Order("required_xp asc").Find(&rewards).Error; err != nil {
		utils.RespondServerError(context, err)
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
	context.JSON(http.StatusOK, gin.H{"data": result, "count": len(result), "xp": user.Xp, "tier": user.Tier})
}

func ClaimReward(context *gin.Context) {
	id := context.Param("id")
	var reward models.Reward
	if err := database.DB.Db.First(&reward, id).Error; err != nil {
		context.JSON(http.StatusNotFound, gin.H{"error": "Reward not found"})
		context.Abort()
		return
	}

	user, ok := currentUser(context)
	if !ok {
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

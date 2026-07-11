package handlers

import (
	"net/http"
	"time"

	"github.com/ALZEE23/ApiGo/database"
	"github.com/ALZEE23/ApiGo/models"
	"github.com/ALZEE23/ApiGo/utils"
	"github.com/gin-gonic/gin"
)

const dayFormat = "2006-01-02"

// countCreatedPerDay buckets rows of model created since `from` into per-day
// counts (keyed YYYY-MM-DD). Pass userID to scope to one user's rows.
func countCreatedPerDay(model interface{}, from time.Time, userID *uint) (map[string]int, error) {
	query := database.DB.Db.Model(model).Select("created_at").Where("created_at >= ?", from)
	if userID != nil {
		query = query.Where("user_id = ?", *userID)
	}

	var rows []struct{ CreatedAt time.Time }
	if err := query.Find(&rows).Error; err != nil {
		return nil, err
	}

	counts := make(map[string]int)
	for _, row := range rows {
		counts[row.CreatedAt.Format(dayFormat)]++
	}
	return counts, nil
}

// activityPerDay builds the shared 7-day activity series (materi completions,
// quiz passes, comments per day) used by both the admin dashboard and the
// per-user beranda chart.
func activityPerDay(userID *uint) ([]gin.H, error) {
	now := time.Now()
	from := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).AddDate(0, 0, -6)

	materi, err := countCreatedPerDay(&models.MateriCompletion{}, from, userID)
	if err != nil {
		return nil, err
	}
	quiz, err := countCreatedPerDay(&models.QuizCompletion{}, from, userID)
	if err != nil {
		return nil, err
	}
	comments, err := countCreatedPerDay(&models.Comment{}, from, userID)
	if err != nil {
		return nil, err
	}

	activity := make([]gin.H, 0, 7)
	for i := 0; i < 7; i++ {
		day := from.AddDate(0, 0, i).Format(dayFormat)
		activity = append(activity, gin.H{
			"date":     day,
			"materi":   materi[day],
			"quiz":     quiz[day],
			"comments": comments[day],
		})
	}
	return activity, nil
}

// GetAdminDashboard aggregates the numbers behind the admin dashboard: entity
// counts for the overview cards plus the series feeding its charts.
func GetAdminDashboard(context *gin.Context) {
	db := database.DB.Db

	counts := gin.H{}
	for key, model := range map[string]interface{}{
		"users":         &models.User{},
		"modules":       &models.Module{},
		"materis":       &models.Materi{},
		"quizzes":       &models.Quiz{},
		"forums":        &models.Forum{},
		"comments":      &models.Comment{},
		"rewards":       &models.Reward{},
		"reward_claims": &models.RewardClaim{},
	} {
		var count int64
		if err := db.Model(model).Count(&count).Error; err != nil {
			utils.RespondServerError(context, err)
			return
		}
		counts[key] = count
	}

	// User registrations bucketed per month for the last 6 months
	// (including the current one).
	now := time.Now()
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()).AddDate(0, -5, 0)
	var userRows []struct{ CreatedAt time.Time }
	if err := db.Model(&models.User{}).Select("created_at").Where("created_at >= ?", monthStart).Find(&userRows).Error; err != nil {
		utils.RespondServerError(context, err)
		return
	}
	perMonth := make(map[string]int)
	for _, row := range userRows {
		perMonth[row.CreatedAt.Format("2006-01")]++
	}
	registrations := make([]gin.H, 0, 6)
	for i := 0; i < 6; i++ {
		month := monthStart.AddDate(0, i, 0).Format("2006-01")
		registrations = append(registrations, gin.H{"month": month, "count": perMonth[month]})
	}

	activity, err := activityPerDay(nil)
	if err != nil {
		utils.RespondServerError(context, err)
		return
	}

	// Users grouped per tier, ordered by tier points; users without a tier
	// land in a leading "Tanpa Tier" bucket.
	var tiers []models.Tier
	if err := db.Order("point asc").Find(&tiers).Error; err != nil {
		utils.RespondServerError(context, err)
		return
	}
	var tierRows []struct {
		IDTier *uint
		Count  int64
	}
	if err := db.Model(&models.User{}).Select("id_tier, count(*) as count").Group("id_tier").Scan(&tierRows).Error; err != nil {
		utils.RespondServerError(context, err)
		return
	}
	perTier := make(map[uint]int64)
	var withoutTier int64
	for _, row := range tierRows {
		if row.IDTier == nil {
			withoutTier = row.Count
		} else {
			perTier[*row.IDTier] = row.Count
		}
	}
	usersPerTier := []gin.H{{"tier": "Tanpa Tier", "count": withoutTier}}
	for _, tier := range tiers {
		usersPerTier = append(usersPerTier, gin.H{"tier": tier.Name, "count": perTier[tier.ID]})
	}

	context.JSON(http.StatusOK, gin.H{
		"counts":                  counts,
		"registrations_per_month": registrations,
		"activity_per_day":        activity,
		"users_per_tier":          usersPerTier,
	})
}

// GetMyActivity returns the current user's last-7-days activity series for
// the beranda chart.
func GetMyActivity(context *gin.Context) {
	user, ok := currentUser(context)
	if !ok {
		return
	}

	activity, err := activityPerDay(&user.ID)
	if err != nil {
		utils.RespondServerError(context, err)
		return
	}
	context.JSON(http.StatusOK, gin.H{"data": activity})
}

package handlers

import (
	"time"

	"github.com/ALZEE23/ApiGo/database"
	"github.com/ALZEE23/ApiGo/models"
	"gorm.io/gorm"
)

// XP paid out for community participation. Kept small relative to quiz XP
// (the primary progression path) so forum activity nudges engagement
// without letting XP-farming eclipse actually learning the material.
const (
	ForumPostXP    = 15
	ForumCommentXP = 5
)

// Daily caps stop a user from farming XP by spamming low-effort posts or
// comments. Once the cap is hit, the post/comment still goes through - only
// the XP payout stops - so participation itself is never blocked.
const (
	ForumPostDailyCap    = 5
	ForumCommentDailyCap = 20
)

func startOfToday() time.Time {
	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
}

// forumPostsToday/forumCommentsToday count how many XP-eligible actions the
// user has already made today, run *after* the new row is inserted so the
// action just taken is included in the count.
func forumPostsToday(userID uint) int64 {
	var count int64
	database.DB.Db.Model(&models.Forum{}).
		Where("user_id = ? AND created_at >= ?", userID, startOfToday()).
		Count(&count)
	return count
}

func forumCommentsToday(userID uint) int64 {
	var count int64
	database.DB.Db.Model(&models.Comment{}).
		Where("user_id = ? AND created_at >= ?", userID, startOfToday()).
		Count(&count)
	return count
}

// gamificationResult carries everything the frontend needs to show XP/tier/
// reward feedback right after an XP-earning action, mirroring how quiz
// completion reports it so the forum feels like part of the same game.
type gamificationResult struct {
	XpEarned        int             `json:"xp_earned"`
	DailyCapReached bool            `json:"daily_cap_reached"`
	Tier            *models.Tier    `json:"tier,omitempty"`
	TierChanged     bool            `json:"tier_changed,omitempty"`
	NewlyUnlocked   []models.Reward `json:"newly_unlocked_rewards,omitempty"`
}

// awardXp credits xp to the user (xp of 0 is a valid no-op, used once a
// daily cap is hit), re-syncs their tier, and surfaces any reward whose
// threshold they just crossed - so callers can show an "unlocked!" moment
// in the same response instead of making the client poll for it.
func awardXp(userID uint, xp int) (gamificationResult, error) {
	var user models.User
	if err := database.DB.Db.First(&user, userID).Error; err != nil {
		return gamificationResult{}, err
	}
	xpBefore := user.Xp

	if xp > 0 {
		err := database.DB.Db.Transaction(func(tx *gorm.DB) error {
			// COALESCE: legacy rows created before the xp column existed hold
			// NULL, and NULL + n stays NULL - which would silently swallow
			// every payout for those users.
			return tx.Model(&user).UpdateColumn("xp", gorm.Expr("COALESCE(xp, 0) + ?", xp)).Error
		})
		if err != nil {
			return gamificationResult{}, err
		}
	}

	result := gamificationResult{XpEarned: xp}
	if tier, changed, err := syncUserTier(userID); err == nil && tier != nil {
		result.Tier = tier
		result.TierChanged = changed
	}

	if xp > 0 {
		rewards := []models.Reward{}
		database.DB.Db.
			Where("required_xp > ? AND required_xp <= ?", xpBefore, xpBefore+xp).
			Order("required_xp asc").
			Find(&rewards)
		if len(rewards) > 0 {
			result.NewlyUnlocked = rewards
		}
	}

	return result, nil
}

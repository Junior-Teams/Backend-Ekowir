package handlers

import (
	"github.com/ALZEE23/ApiGo/database"
	"github.com/ALZEE23/ApiGo/models"
)

// syncUserTier re-evaluates which tier a user belongs to based on their
// current xp, and persists it if it changed. Call this any time xp is
// modified so the stored tier never goes stale.
func syncUserTier(userID uint) (tier *models.Tier, changed bool, err error) {
	var user models.User
	if err = database.DB.Db.First(&user, userID).Error; err != nil {
		return nil, false, err
	}

	var best models.Tier
	err = database.DB.Db.Where("point <= ?", user.Xp).Order("point desc").First(&best).Error
	if err != nil {
		if user.IDTier != nil {
			user.IDTier = nil
			database.DB.Db.Model(&user).Update("id_tier", nil)
			return nil, true, nil
		}
		return nil, false, nil
	}

	if user.IDTier != nil && *user.IDTier == best.ID {
		return &best, false, nil
	}

	if err = database.DB.Db.Model(&user).Update("id_tier", best.ID).Error; err != nil {
		return nil, false, err
	}
	return &best, true, nil
}

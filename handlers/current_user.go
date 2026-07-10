package handlers

import (
	"net/http"

	"github.com/ALZEE23/ApiGo/database"
	"github.com/ALZEE23/ApiGo/models"
	"github.com/gin-gonic/gin"
)

// currentUser loads the user tied to the session that middlewares.Auth()
// already validated for this request. On failure it writes the response
// itself, so callers can just `return` when ok is false.
func currentUser(context *gin.Context) (user models.User, ok bool) {
	email := context.GetString("email")
	if err := database.DB.Db.Preload("Tier").Where("email = ?", email).First(&user).Error; err != nil {
		context.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		context.Abort()
		return models.User{}, false
	}
	return user, true
}

// isOwnerOrAdmin reports whether user may modify a resource created by
// ownerID: either they created it themselves, or they're an admin.
func isOwnerOrAdmin(user models.User, ownerID uint) bool {
	return user.ID == ownerID || user.Role == "admin"
}

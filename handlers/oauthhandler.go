package handlers

import (
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/ALZEE23/ApiGo/auth"
	"github.com/ALZEE23/ApiGo/database"
	"github.com/ALZEE23/ApiGo/models"
	"github.com/gin-gonic/gin"
)

const oauthStateCookie = "oauthstate"

func GoogleLogin(context *gin.Context) {
	state, err := auth.GenerateOAuthState()
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": "failed to start google login"})
		context.Abort()
		return
	}

	secure := strings.HasPrefix(auth.GoogleOAuthConfig().RedirectURL, "https://")
	context.SetCookie(oauthStateCookie, state, 5*60, "/", "", secure, true)
	context.Redirect(http.StatusTemporaryRedirect, auth.GoogleOAuthConfig().AuthCodeURL(state))
}

func GoogleCallback(context *gin.Context) {
	frontendURL := strings.TrimRight(os.Getenv("FRONTEND_URL"), "/")

	redirectWithError := func(reason string) {
		context.SetCookie(oauthStateCookie, "", -1, "/", "", false, true)
		context.Redirect(http.StatusTemporaryRedirect, frontendURL+"/oauth/callback?error="+url.QueryEscape(reason))
	}

	if googleErr := context.Query("error"); googleErr != "" {
		redirectWithError(googleErr)
		return
	}

	stateCookie, err := context.Cookie(oauthStateCookie)
	if err != nil || stateCookie == "" || stateCookie != context.Query("state") {
		redirectWithError("invalid_state")
		return
	}

	info, err := auth.FetchGoogleUserInfo(context.Request.Context(), context.Query("code"))
	if err != nil {
		redirectWithError("google_auth_failed")
		return
	}

	user, err := findOrCreateGoogleUser(info)
	if err != nil {
		redirectWithError("user_lookup_failed")
		return
	}

	tokenString, err := auth.GenerateJWT(user.Email, user.Username, user.Role)
	if err != nil {
		redirectWithError("token_generation_failed")
		return
	}

	context.SetCookie(oauthStateCookie, "", -1, "/", "", false, true)
	context.Redirect(http.StatusTemporaryRedirect, frontendURL+"/oauth/callback?token="+url.QueryEscape(tokenString))
}

func findOrCreateGoogleUser(info *auth.GoogleUserInfo) (*models.User, error) {
	db := database.DB.Db

	var user models.User
	if err := db.Where("google_id = ?", info.ID).First(&user).Error; err == nil {
		user.Picture = info.Picture
		if err := db.Save(&user).Error; err != nil {
			return nil, err
		}
		return &user, nil
	}

	if err := db.Where("email = ?", info.Email).First(&user).Error; err == nil {
		user.GoogleID = &info.ID
		user.Picture = info.Picture
		if err := db.Save(&user).Error; err != nil {
			return nil, err
		}
		return &user, nil
	}

	username := strings.Split(info.Email, "@")[0]
	user = models.User{
		Name:     info.Name,
		Username: username,
		Email:    info.Email,
		Picture:  info.Picture,
		GoogleID: &info.ID,
	}
	if err := db.Create(&user).Error; err != nil {
		suffix, genErr := auth.GenerateOAuthState()
		if genErr != nil {
			return nil, err
		}
		user.Username = username + "-" + suffix[:6]
		if err := db.Create(&user).Error; err != nil {
			return nil, err
		}
	}
	return &user, nil
}

package middlewares
import (
  "strings"

  "github.com/ALZEE23/ApiGo/auth"
  "github.com/ALZEE23/ApiGo/database"
  "github.com/ALZEE23/ApiGo/models"
  "github.com/gin-gonic/gin"
)
func Auth() gin.HandlerFunc{
  return func(context *gin.Context) {
    tokenString := context.GetHeader("Authorization")
    if tokenString == "" {
      context.JSON(401, gin.H{"error": "request does not contain an access token"})
      context.Abort()
      return
    }
    tokenString = strings.TrimPrefix(tokenString, "Bearer ")
    claims, err := auth.ValidateToken(tokenString)
    if err != nil {
      context.JSON(401, gin.H{"error": "Sesi Anda tidak valid atau telah berakhir, silahkan login kembali"})
      context.Abort()
      return
    }

    var revoked models.RevokedToken
    if err := database.DB.Db.Where("jti = ?", claims.ID).First(&revoked).Error; err == nil {
      context.JSON(401, gin.H{"error": "token has been revoked"})
      context.Abort()
      return
    }

    context.Set("email", claims.Email)
    context.Set("username", claims.Username)
    context.Set("role", claims.Role)
    context.Set("jti", claims.ID)
    context.Set("exp", claims.ExpiresAt.Time)
    context.Next()
  }
}

func RequireRole(roles ...string) gin.HandlerFunc {
  return func(context *gin.Context) {
    role := context.GetString("role")
    for _, allowed := range roles {
      if role == allowed {
        context.Next()
        return
      }
    }
    context.JSON(403, gin.H{"error": "insufficient permissions"})
    context.Abort()
  }
}

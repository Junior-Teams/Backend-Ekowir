package middlewares
import (
  "strings"

  "github.com/ALZEE23/ApiGo/auth"
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
      context.JSON(401, gin.H{"error": err.Error()})
      context.Abort()
      return
    }
    context.Set("email", claims.Email)
    context.Set("username", claims.Username)
    context.Next()
  }
}

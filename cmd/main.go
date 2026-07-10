package main

import (
	"log"
	"time"

	"github.com/ALZEE23/ApiGo/database"
	"github.com/ALZEE23/ApiGo/handlers"
	"github.com/ALZEE23/ApiGo/middlewares"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func setupRoutes() *gin.Engine {
	router := gin.Default()
	router.MaxMultipartMemory = 8 << 20

	config := cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "http://localhost:5173", "https://web.pplg-game.com"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}

	router.Use(cors.New(config))
	router.Static("/storage", "./storage")
	api := router.Group("/api")
	{
		api.GET("/", handlers.Test)
		api.POST("/token", handlers.GenerateToken)
		api.POST("/user/register", handlers.RegisterUser)
		api.GET("/apk", handlers.GetApk)
		api.GET("/auth/google/login", handlers.GoogleLogin)
		api.GET("/auth/google/callback", handlers.GoogleCallback)
		secured := api.Group("/secured").Use(middlewares.Auth())
		{
			secured.GET("/ping", handlers.Ping)
			secured.GET("/me", handlers.GetMe)
			secured.POST("/apk", handlers.Apk)
			secured.GET("/users", handlers.GetUsers)
			secured.GET("/users/:id", handlers.GetUserByID)
			secured.PUT("/users/:id", handlers.UpdateUser)
			secured.DELETE("/users/:id", handlers.DeleteUser)
		}
	}
	return router
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found, relying on process environment")
	}

	database.ConnectDb()
	app := setupRoutes()

	app.Run(":3000")
}

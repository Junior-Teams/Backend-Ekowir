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
		api.GET("/comments", handlers.GetComments)
		api.GET("/comments/:id", handlers.GetCommentByID)
		api.GET("/modules", handlers.GetModules)
		api.GET("/modules/:id", handlers.GetModuleByID)
		api.GET("/materis", handlers.GetMateris)
		api.GET("/materis/:id", handlers.GetMateriByID)
		api.GET("/quizzes", handlers.GetQuizzes)
		api.GET("/quizzes/:id", handlers.GetQuizByID)
		api.GET("/questions", handlers.GetQuestions)
		api.GET("/questions/:id", handlers.GetQuestionByID)
		api.GET("/forums", handlers.GetForums)
		api.GET("/forums/:id", handlers.GetForumByID)
		api.GET("/tiers", handlers.GetTiers)
		api.GET("/tiers/:id", handlers.GetTierByID)
		api.GET("/leaderboard", handlers.GetLeaderboard)
		api.GET("/rewards", handlers.GetRewards)
		api.GET("/rewards/:id", handlers.GetRewardByID)
		secured := api.Group("/secured")
		secured.Use(middlewares.Auth())
		{
			secured.GET("/ping", handlers.Ping)
			secured.GET("/me", handlers.GetMe)
			secured.PUT("/me", handlers.UpdateMe)
			secured.PUT("/me/password", handlers.ChangeMyPassword)
			secured.POST("/logout", handlers.Logout)
			secured.POST("/apk", handlers.Apk)
			secured.POST("/quizzes/:id/submit", handlers.SubmitQuiz)
			secured.POST("/materis/:id/complete", handlers.CompleteMateri)
			secured.GET("/modules/:id/progress", handlers.GetModuleProgress)
			secured.GET("/me/courses", handlers.GetMyCourseHistory)
			secured.GET("/me/activity", handlers.GetMyActivity)
			secured.GET("/rewards", handlers.GetMyRewards)
			secured.POST("/rewards/:id/claim", handlers.ClaimReward)
			secured.POST("/comments", handlers.Comment)
			secured.PUT("/comments/:id", handlers.UpdateComment)
			secured.DELETE("/comments/:id", handlers.DeleteComment)
			secured.POST("/forums", handlers.Forum)
			secured.PUT("/forums/:id", handlers.UpdateForum)
			secured.DELETE("/forums/:id", handlers.DeleteForum)

			admin := secured.Group("/")
			admin.Use(middlewares.RequireRole("admin"))
			{
				admin.GET("/dashboard", handlers.GetAdminDashboard)
				admin.GET("/users", handlers.GetUsers)
				admin.GET("/users/:id", handlers.GetUserByID)
				admin.POST("/users", handlers.CreateUser)
				admin.PUT("/users/:id", handlers.UpdateUser)
				admin.DELETE("/users/:id", handlers.DeleteUser)
				admin.POST("/modules", handlers.Module)
				admin.PUT("/modules/:id", handlers.UpdateModule)
				admin.DELETE("/modules/:id", handlers.DeleteModule)
				admin.POST("/materis", handlers.Materi)
				admin.PUT("/materis/:id", handlers.UpdateMateri)
				admin.DELETE("/materis/:id", handlers.DeleteMateri)
				admin.POST("/quizzes", handlers.Quiz)
				admin.PUT("/quizzes/:id", handlers.UpdateQuiz)
				admin.DELETE("/quizzes/:id", handlers.DeleteQuiz)
				admin.GET("/questions", handlers.GetQuestionsAdmin)
				admin.GET("/questions/:id", handlers.GetQuestionByIDAdmin)
				admin.POST("/questions", handlers.Question)
				admin.PUT("/questions/:id", handlers.UpdateQuestion)
				admin.DELETE("/questions/:id", handlers.DeleteQuestion)
				admin.POST("/tiers", handlers.Tier)
				admin.PUT("/tiers/:id", handlers.UpdateTier)
				admin.DELETE("/tiers/:id", handlers.DeleteTier)
				admin.POST("/rewards", handlers.Reward)
				admin.PUT("/rewards/:id", handlers.UpdateReward)
				admin.DELETE("/rewards/:id", handlers.DeleteReward)
			}
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
package main

import (
	"io"
	"log"
	"os"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/merinovvvv/momentic-backend/controllers"
	"github.com/merinovvvv/momentic-backend/initializers"
	"github.com/merinovvvv/momentic-backend/repository"
	"github.com/merinovvvv/momentic-backend/service"
	"gopkg.in/natefinch/lumberjack.v2"
	"github.com/merinovvvv/momentic-backend/middleware"
)

func init() {
	initializers.LoadEnvVariables()
	initializers.ConnectToDb()
}

func main() {
	logRotator := &lumberjack.Logger{
		Filename:   "log/log.txt",
		MaxSize:    100, // Размер в МБ до ротации
		MaxBackups: 3,
		MaxAge:     28, // количество дней для хранения архивов
		Compress:   true,
	}
	multiWriter := io.MultiWriter(os.Stdout, logRotator)
	log.SetOutput(multiWriter)
	gin.DefaultWriter = multiWriter

	router := gin.Default()
	// router.GET("/ping", func(c *gin.Context) {
	// 	c.JSON(200, gin.H{
	// 		"message": "pong",
	// 	})
	// })

	db := initializers.DB
	videoRepo := repository.NewVideoRepository(db)
	reactionRepo := repository.NewReactionRepository(db)

	videoService := service.NewVideoService(videoRepo)
	reactionService := service.NewReactionService(reactionRepo)

	videoController := controllers.NewVideoController(videoService)
	reactionController := controllers.NewReactionController(reactionService)
	//curl -X POST http://localhost:8080/videos -F "author_id=2" -F "description=Тестовое видео" -F "video_file=@file_path"
	router.POST("/videos", videoController.UploadVideo)

	router.GET("/users/:user_id/friends/videos", videoController.GetTodayFeedByUserID)

	router.PATCH("/videos/:video_id", videoController.UpdateVideoDescription)

	router.DELETE("/videos/:video_id", videoController.DeleteVideo)

	router.POST("/videos/:video_id/reactions", reactionController.HandleReaction)
	router.DELETE("/videos/:video_id/reactions", reactionController.RemoveReaction)
	router.GET("/videos/:video_id/reactions", reactionController.GetVideoReactions)
	log.Println("INFO: Server started.")
	router.POST("/auth/register", controllers.SignUp)
	router.PATCH("/auth/register", middleware.RequireAuth)
	router.POST("/auth/login", controllers.Login)
	router.POST("/auth/verify-code", controllers.VerifyEmail)
	router.POST("/auth/refresh", controllers.Refresh)
	router.POST("/auth/resend-verify-code", controllers.ResendEmailVerification)
	router.GET("/validate", middleware.RequireAuth, controllers.Validate)
	fmt.Println(router.Routes())
	router.Run() // listens on 0.0.0.0:8080 by default
}

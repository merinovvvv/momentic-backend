package main

import (
	"io"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/merinovvvv/momentic-backend/controllers"
	"github.com/merinovvvv/momentic-backend/initializers"
	"github.com/merinovvvv/momentic-backend/repository"
	"github.com/merinovvvv/momentic-backend/service"
	"gopkg.in/natefinch/lumberjack.v2"
)

func init() {
	initializers.LoadEnvVariables()
	initializers.ConnectToDb()
}

func main() {
	logRotator := &lumberjack.Logger{
		Filename:   "log/log.txt", // Путь к файлу
		MaxSize:    100,           // Размер в мегабайтах (МБ) до ротации
		MaxBackups: 3,             // Макс. количество архивных файлов (log.txt.1, log.txt.2 и т.д.)
		MaxAge:     28,            // Макс. количество дней для хранения архивов
		Compress:   true,          // Сжимать старые файлы
	}
	multiWriter := io.MultiWriter(os.Stdout, logRotator)
	log.SetOutput(multiWriter)
	gin.DefaultWriter = multiWriter

	router := gin.Default()
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	db := initializers.DB
	videoRepo := repository.NewVideoRepository(db)
	videoService := service.NewVideoService(videoRepo)
	videoController := controllers.NewVideoController(videoService)
	//curl -X POST http://localhost:8080/videos -F "author_id=2" -F "description=Тестовое видео" -F "video_file=@file_path"
	router.POST("/videos", videoController.UploadVideo)

	router.GET("/users/:user_id/friends/videos", videoController.GetTodayFeedByUserID)

	router.PATCH("/videos/:video_id", videoController.UpdateVideoDescription)

	router.DELETE("/videos/:video_id", videoController.DeleteVideo)
	log.Println("INFO: Server started.")
	router.Run() // listens on 0.0.0.0:8080 by default
}

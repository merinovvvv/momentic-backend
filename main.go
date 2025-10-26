package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/merinovvvv/momentic-backend/controllers"
	"github.com/merinovvvv/momentic-backend/initializers"
	"github.com/merinovvvv/momentic-backend/repository"
	"github.com/merinovvvv/momentic-backend/service"
)

func init() {
	initializers.LoadEnvVariables()
	initializers.ConnectToDb()
}

func main() {
	router := gin.Default()
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	logFile, err := os.OpenFile("log/log.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Ошибка открытия файла логов: %v", err)
	}
	defer logFile.Close()

	log.SetOutput(logFile)
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	db := initializers.DB
	videoRepo := repository.NewVideoRepository(db)
	videoService := service.NewVideoService(videoRepo)
	videoController := controllers.NewVideoController(videoService)
	//curl -X POST http://localhost:8080/videos -F "author_id=2" -F "description=Тестовое видео" -F "video_file=@file_path"
	router.POST("/videos", videoController.UploadVideo)

	router.GET("/users/:user_id/friends/videos", videoController.GetTodayFeedByUserID)

	router.PATCH("/videos/:video_id", videoController.UpdateVideoDescription)

	router.DELETE("/videos/:video_id", videoController.DeleteVideo)
	router.Run() // listens on 0.0.0.0:8080 by default
}

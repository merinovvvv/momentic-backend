package main

import (
	"github.com/gin-gonic/gin"
	"github.com/merinovvvv/momentic-backend/controllers"
	"github.com/merinovvvv/momentic-backend/initializers"
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
	router.POST("/videos", controllers.UploadVideo) //curl -X POST http://localhost:8080/videos -F "author_id=2" -F "description=Тестовое видео" -F "video_file=@file_path"
	router.GET("/users/:user_id/friends/videos", controllers.GetTodayFeedByUserID)
	router.DELETE("/videos/:video_id", controllers.DeleteVideo)
	router.PATCH("/videos/:video_id", controllers.UpdateVideoDescription)
	router.Run() // listens on 0.0.0.0:8080 by default
}

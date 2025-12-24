package main

import (
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/merinovvvv/momentic-backend/controllers"
	"github.com/merinovvvv/momentic-backend/initializers"
	"github.com/merinovvvv/momentic-backend/repository"
	"github.com/merinovvvv/momentic-backend/service"
	"github.com/merinovvvv/momentic-backend/ws"
	"gopkg.in/natefinch/lumberjack.v2"
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
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	// Global chat via WebSocket (for iOS)
	router.GET("/ws/chat", ws.ServeChatWs)
	db := initializers.DB
	videoRepo := repository.NewVideoRepository(db)
	reactionRepo := repository.NewReactionRepository(db)

	videoService := service.NewVideoService(videoRepo)
	reactionService := service.NewReactionService(reactionRepo)

	commentRepo := repository.NewCommentRepository(db)
	commentController := controllers.NewCommentController(commentRepo)

	videoController := controllers.NewVideoController(videoService)
	reactionController := controllers.NewReactionController(reactionService)
	//curl -X POST http://localhost:8080/videos -F "author_id=2" -F "description=Тестовое видео" -F "video_file=@file_path"
	router.POST("/videos", videoController.UploadVideo)

	router.POST("/videos/:video_id/comments", commentController.PostComment)
	router.GET("/videos/:video_id/comments", commentController.GetCommentsByVideoID)

	router.GET("/users/:user_id/friends/videos", videoController.GetTodayFeedByUserID)

	router.PATCH("/videos/:video_id", videoController.UpdateVideoDescription)

	router.DELETE("/videos/:video_id", videoController.DeleteVideo)

	router.POST("/videos/:video_id/reactions", reactionController.HandleReaction)
	router.DELETE("/videos/:video_id/reactions", reactionController.RemoveReaction)
	router.GET("/videos/:video_id/reactions", reactionController.GetVideoReactions)

	hub := ws.NewHub()
	go hub.Run() // запускаем hub в горутине

	websocketController := controllers.NewWebSocketController(hub)
	router.GET("/ws/broadcast", websocketController.ServeWs)
	//curl -X POST http://localhost:8080/admin/broadcast -H "Content-Type: application/json" -d "{\"content\": \"Это сообщение для рассылки!\"}"
	router.POST("/admin/broadcast", func(c *gin.Context) {
		var msg struct {
			Content string `json:"content" binding:"required"`
		}
		if err := c.ShouldBindJSON(&msg); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Content required"})
			return
		}

		hub.Broadcast <- &ws.Message{
			Sender:    "Tech Support",
			Content:   msg.Content,
			Timestamp: time.Now(),
		}
		c.JSON(http.StatusOK, gin.H{"message": "Broadcast sent"})
	})

	videoHub := ws.NewVideoHub()
	go videoHub.Run()

	log.Println("INFO: Server started.")

	router.Run() // listens on 0.0.0.0:8080 by default
}

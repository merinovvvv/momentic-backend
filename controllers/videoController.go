package controllers

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/merinovvvv/momentic-backend/service"
)

const MaxFileSize = 200 << 20

// VideoController содержит зависимость от сервиса
type VideoController struct {
	service service.VideoService
}

func NewVideoController(s service.VideoService) *VideoController {
	return &VideoController{service: s}
}

type UpdateDescriptionRequest struct {
	Description string `json:"description" binding:"required"`
}

// --- UploadVideo (POST) ---
func (vc *VideoController) UploadVideo(c *gin.Context) {
	// 1. HTTP-логика: Извлечение данных и файла
	authorIDStr := c.PostForm("author_id")
	description := c.PostForm("description")

	authorID, err := strconv.ParseInt(authorIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат author_id"})
		return
	}

	file, err := c.FormFile("video_file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Ошибка при получении файла: %s", err.Error())})
		return
	}

	if file.Size > MaxFileSize {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Размер файла превышает лимит"})
		return
	}

	ext := filepath.Ext(file.Filename)
	filename := fmt.Sprintf("%d_%d%s", authorID, time.Now().UnixNano(), ext)
	savePath := filepath.Join("uploads", filename)

	if err := c.SaveUploadedFile(file, savePath); err != nil {
		log.Printf("ERROR: Failed to save file to disk at %s: %v", savePath, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Не удалось сохранить файл на диске: %s", err.Error())})
		return
	}

	video, err := vc.service.UploadVideo(c.Request.Context(), savePath, authorID, description)

	if err != nil {
		os.Remove(savePath)
		if errors.Is(err, service.ErrAuthorIDRequired) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		log.Printf("FATAL: Service failed during UploadVideo: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при сохранении данных о видео в БД"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":  "Видео успешно загружено и опубликовано",
		"video_id": video.VideoID,
		"filepath": video.Filepath,
	})
}

// --- GetTodayFeedByUserID (GET) ---
func (vc *VideoController) GetTodayFeedByUserID(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат ID пользователя"})
		return
	}

	videos, err := vc.service.GetTodayFeed(c.Request.Context(), userID)

	if err != nil {
		if errors.Is(err, service.ErrNoFriends) {
			c.JSON(http.StatusOK, []interface{}{}) // 200 OK с пустым массивом
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при получении ленты видео"})
		return
	}

	c.JSON(http.StatusOK, videos)
}

// --- DeleteVideo (DELETE) ---
func (vc *VideoController) DeleteVideo(c *gin.Context) {
	videoIDStr := c.Param("video_id")
	videoID, err := strconv.ParseInt(videoIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат ID видео"})
		return
	}

	err = vc.service.DeleteVideo(c.Request.Context(), videoID)

	if err != nil {
		if errors.Is(err, service.ErrVideoNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Видео не найдено"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при удалении видео"})
		return
	}

	c.Status(http.StatusNoContent) // 204
}

// --- UpdateVideoDescription (PATCH) ---
func (vc *VideoController) UpdateVideoDescription(c *gin.Context) {
	videoIDStr := c.Param("video_id")
	videoID, err := strconv.ParseInt(videoIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат ID видео"})
		return
	}

	var req UpdateDescriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Требуется поле 'description'"})
		return
	}

	// Вызов сервиса
	err = vc.service.UpdateDescription(c.Request.Context(), videoID, req.Description)

	// Обработка ошибок
	if err != nil {
		if errors.Is(err, service.ErrVideoNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Видео не найдено"})
			return
		}
		if errors.Is(err, service.ErrDescriptionTooLong) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось обновить описание"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Описание успешно обновлено"})
}

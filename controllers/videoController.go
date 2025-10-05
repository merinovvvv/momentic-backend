package controllers

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/merinovvvv/momentic-backend/initializers"
	"github.com/merinovvvv/momentic-backend/models"
)

func UploadVideo(c *gin.Context) {
	authorIDStr := c.PostForm("author_id")
	description := c.PostForm("description")
	if authorIDStr == "" {
		c.JSON(400, gin.H{"error": "Поле author_id обязательно"})
		return
	}
	authorID, err := strconv.ParseInt(authorIDStr, 10, 64)
	if err != nil {
		c.JSON(400, gin.H{"error": "Неверный формат author_id"})
		return
	}
	file, err := c.FormFile("video_file") // "video_file" - это имя поля в форме (multipart/form-data)
	if err != nil {
		c.JSON(400, gin.H{"error": fmt.Sprintf("Ошибка при получении файла: %s", err.Error())})
		return
	}

	ext := filepath.Ext(file.Filename)
	filename := fmt.Sprintf("%d_%d%s", authorID, time.Now().UnixNano(), ext)
	savePath := filepath.Join("uploads", filename) // Путь для сохранения (нужно создать папку "uploads")

	// Сохраняем файл на диске
	if err := c.SaveUploadedFile(file, savePath); err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("Не удалось сохранить файл на диске: %s", err.Error())})
		return
	}

	// 4. Запись метаданных в базу данных (GORM)
	newVideo := models.Video{
		Filepath:    savePath,
		AuthorID:    authorID,
		Description: description,
	}

	result := initializers.DB.Create(&newVideo)
	if result.Error != nil {
		os.Remove(savePath) //удаление файла при ошибке
		c.JSON(500, gin.H{"error": "Ошибка при сохранении данных о видео в БД"})
		return
	}

	c.JSON(201, gin.H{
		"message":  "Видео успешно загружено и опубликовано",
		"video_id": newVideo.VideoID,
		"filepath": newVideo.Filepath,
	})
}

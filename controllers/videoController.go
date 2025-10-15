package controllers

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/merinovvvv/momentic-backend/initializers"
	"github.com/merinovvvv/momentic-backend/models"
	"gorm.io/gorm"
)

// POST запрос на отправку видео
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
	file, err := c.FormFile("video_file") // (multipart/form-data)
	if err != nil {
		c.JSON(400, gin.H{"error": fmt.Sprintf("Ошибка при получении файла: %s", err.Error())})
		return
	}

	ext := filepath.Ext(file.Filename)
	filename := fmt.Sprintf("%d_%d%s", authorID, time.Now().UnixNano(), ext)
	savePath := filepath.Join("uploads", filename) //создать папку "uploads"

	// Сохраняем файл на диске
	if err := c.SaveUploadedFile(file, savePath); err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("Не удалось сохранить файл на диске: %s", err.Error())})
		return
	}

	// 4. Запись  в базу данных
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

// GET запрос на получение видео друзей пользователя
func GetTodayFeedByUserID(c *gin.Context) {
	var videos []models.Video

	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат ID пользователя"})
		return
	}
	friendIDs, err := getFriendsIDs(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при получении списка друзей"})
		return
	}
	if len(friendIDs) == 0 {
		c.JSON(http.StatusOK, []models.Video{})
		return
	}
	startOfDay := time.Now().Truncate(24 * time.Hour) // 00 00
	endOfToday := startOfDay.Add(24 * time.Hour)      // 23 59

	result := initializers.DB.
		Where("author_id IN (?)", friendIDs).
		Where("created_at >= ? AND created_at < ?", startOfDay, endOfToday).
		Order("created_at DESC").
		Find(&videos)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка базы данных при получении видео"})
		return
	}

	c.JSON(http.StatusOK, videos)
}

// Получение списка друзей по userid
func getFriendsIDs(userID int64) ([]int64, error) {
	if userID == 0 {
		return nil, errors.New("invalid user ID")
	}
	var friendships []models.Friendship

	result := initializers.DB.
		Where("status = ?", models.StatusFriends).
		Where("user_id1 = ? OR user_id2 = ?", userID, userID).
		Find(&friendships)

	if result.Error != nil {
		return nil, result.Error
	}

	if len(friendships) == 0 {
		return []int64{}, nil
	}

	friendIDs := make([]int64, 0, len(friendships))

	for _, f := range friendships {
		if f.UserID1 == userID {
			friendIDs = append(friendIDs, f.UserID2)
		} else {
			friendIDs = append(friendIDs, f.UserID1)
		}
	}
	return friendIDs, nil
}

// DELETE запрос на удаления видео по id
func DeleteVideo(c *gin.Context) {
	videoIDStr := c.Param("video_id")

	videoID, err := strconv.ParseInt(videoIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат ID видео"})
		return
	}

	// TODO проверить что пользователь является автором

	var video models.Video

	result := initializers.DB.First(&video, videoID)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Видео не найдено"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка базы данных при поиске"})
		return
	}

	if err := os.Remove(video.Filepath); err != nil {
		c.Error(err)
	}

	deleteResult := initializers.DB.Delete(&video)
	if deleteResult.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка базы данных при удалении записи"})
		return
	}

	c.Status(http.StatusNoContent) //204
}

type UpdateDescriptionRequest struct {
	Description string `json:"description" binding:"required"`
}

// PATCH запрос на изменение описания видео
func UpdateVideoDescription(c *gin.Context) {
	videoIDStr := c.Param("video_id")
	videoID, err := strconv.ParseInt(videoIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат ID видео"})
		return
	}

	var req UpdateDescriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат запроса: требуется поле 'description'"})
		return
	}

	result := initializers.DB.Model(&models.Video{}).
		Where("video_id = ?", videoID).
		Update("description", req.Description)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка базы данных при обновлении"})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Видео не найдено или не было изменено"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Описание успешно обновлено"})
}

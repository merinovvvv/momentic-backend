package controllers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/merinovvvv/momentic-backend/models"
	"github.com/merinovvvv/momentic-backend/repository"
)

type CommentController struct {
	Repo repository.CommentRepository
}

func NewCommentController(repo repository.CommentRepository) *CommentController {
	return &CommentController{Repo: repo}
}

// GET /videos/:video_id/comments
func (cc *CommentController) GetCommentsByVideoID(c *gin.Context) {
	videoIDStr := c.Param("video_id")
	videoID, err := strconv.ParseInt(videoIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid video_id"})
		return
	}
	comments, err := cc.Repo.GetByVideoID(c.Request.Context(), videoID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch comments"})
		return
	}
	c.JSON(http.StatusOK, comments)
}

// POST /videos/:video_id/comments
/*func (cc *CommentController) PostComment(c *gin.Context) {
	videoIDStr := c.Param("video_id")
	var input struct {
		Author    string  `json:"author"`
		Text      string  `json:"text"`
		AvatarURL *string `json:"avatarURL"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	videoID, err := strconv.ParseInt(videoIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid video_id"})
		return
	}

	comment := models.Comment{
		VideoID:   videoID,
		Nickname:  input.Author,
		Content:   input.Text,
		AvatarURL: input.AvatarURL,
		CreatedAt: time.Now(),
	}

	if err := cc.Repo.Create(c.Request.Context(), &comment); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save comment"})
		return
	}

	c.JSON(http.StatusCreated, comment)
}*/
func (cc *CommentController) PostComment(c *gin.Context) {
	videoIDStr := c.Param("video_id")

	// 1. Получаем ID пользователя из JWT (обычно кладется в контекст в AuthMiddleware)
	// Если авторизации пока нет, для теста можно взять константу или передавать в JSON
	// Default test user ID
	var uid uint64 = 1
	if u, exists := c.Get("userID"); exists && u != nil {
		switch v := u.(type) {
		case uint64:
			uid = v
		case int64:
			uid = uint64(v)
		case int:
			uid = uint64(v)
		}
	}

	var input struct {
		Text string `json:"text" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Text is required"})
		return
	}

	videoID, err := strconv.ParseInt(videoIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid video_id"})
		return
	}

	// 2. Создаем объект комментария
	comment := models.Comment{
		VideoID:   videoID,
		UserID:    uid,
		Content:   input.Text,
		CreatedAt: time.Now(),
	}

	// 3. Репозиторий сохранит коммент и через Preload ("User")
	// сам заполнит поле Nickname внутри структуры comment
	if err := cc.Repo.Create(c.Request.Context(), &comment); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save comment"})
		return
	}

	// 4. Теперь comment содержит и ID, и подтянутый Nickname автора
	c.JSON(http.StatusCreated, comment)
}

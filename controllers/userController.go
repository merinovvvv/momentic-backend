package controllers

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/merinovvvv/momentic-backend/repository"
)

type UserController struct {
	Repo repository.UserRepository // Use concrete type from repository package
}

func NewUserController(repo repository.UserRepository) *UserController {
	return &UserController{Repo: repo}
}

// PATCH /user/avatar/
func (uc *UserController) UpdateAvatar(c *gin.Context) {
	var userID uint64 = 1 // fallback for test/dev
	if v, ok := c.Get("userID"); ok && v != nil {
		switch id := v.(type) {
		case uint64:
			userID = id
		case int64:
			userID = uint64(id)
		case int:
			userID = uint64(id)
		}
	}

	imageData, err := c.GetRawData()
	if err != nil || len(imageData) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No image data provided"})
		return
	}

	if len(imageData) < 2 || imageData[0] != 0xFF || imageData[1] != 0xD8 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Data is not JPEG"})
		return
	}

	avatarDir := "uploads/avatars"
	if err := os.MkdirAll(avatarDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create avatar dir"})
		return
	}
	filePath := fmt.Sprintf("%s/%d.jpg", avatarDir, userID)
	if err := os.WriteFile(filePath, imageData, 0644); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save avatar"})
		return
	}

	// save path to DB
	if err := uc.Repo.UpdateAvatarPath(context.Background(), userID, filePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update avatar path in DB"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"avatar_url": "/static/" + filePath})
}

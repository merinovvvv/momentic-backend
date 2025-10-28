package controllers

import (
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/merinovvvv/momentic-backend/models"
	"github.com/merinovvvv/momentic-backend/service"
)

type ReactionController struct {
	service service.ReactionService
}

func NewReactionController(s service.ReactionService) *ReactionController {
	return &ReactionController{service: s}
}

func assumedGetUserID(c *gin.Context) (int64, error) {
	//TODO
	return 1, nil
}

// POST /videos/{video_id}/reaction
func (rc *ReactionController) HandleReaction(c *gin.Context) {
	videoID, err := strconv.ParseInt(c.Param("video_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid video ID format"})
		return
	}

	userID, err := assumedGetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization failed"})
		return
	}

	var req models.ReactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid reaction kind in request body"})
		return
	}

	err = rc.service.HandleReaction(c.Request.Context(), userID, videoID, req.Kind)
	if err != nil {
		if errors.Is(err, service.ErrInvalidReactionKind) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		log.Printf("FATAL: Service error during HandleReaction: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not set reaction"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Reaction set successfully"})
}

// DELETE /videos/{video_id}/reactions
func (rc *ReactionController) RemoveReaction(c *gin.Context) {
	videoID, err := strconv.ParseInt(c.Param("video_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid video ID format"})
		return
	}

	userID, err := assumedGetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization failed"})
		return
	}

	err = rc.service.RemoveReaction(c.Request.Context(), userID, videoID)
	if err != nil {
		if errors.Is(err, service.ErrReactionNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Reaction not found"})
			return
		}
		log.Printf("FATAL: Service error during RemoveReaction: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not remove reaction"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Reaction removed successfully"})
}

// GET /videos/{videoId}/reactions
func (rc *ReactionController) GetVideoReactions(c *gin.Context) {
	videoID, err := strconv.ParseInt(c.Param("video_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid video ID format"})
		return
	}

	users, err := rc.service.GetVideoReactions(c.Request.Context(), videoID)
	if err != nil {
		log.Printf("FATAL: Service error during GetVideoReactions: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not fetch reactions"})
		return
	}

	c.JSON(http.StatusOK, users)
}

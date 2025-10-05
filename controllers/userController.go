package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/merinovvvv/momentic-backend/initializers"
	"github.com/merinovvvv/momentic-backend/models"
	"golang.org/x/crypto/bcrypt"
)

func SignUp(c *gin.Context) {

	var body struct {
		Email string
		Nickname string
		Password string
	}

	if c.Bind(&body) != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to read body",
		})

		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(body.Password), 10)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to hash password", 
		})

		return
	}

	user := models.User{Password: string(hash), Nickname: body.Nickname, Email: body.Email}
	result := initializers.DB.Create(&user)
	if result.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to create user",
		})
	}
	c.JSON(http.StatusOK, gin.H{})
}
//    Bio            string    `gorm:"size:100;not null;default:''"`
//    CreatedAt      time.Time `gorm:"column:created_at;not null;default:now()"`

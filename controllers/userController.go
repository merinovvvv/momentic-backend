package controllers

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/merinovvvv/momentic-backend/initializers"
	"github.com/merinovvvv/momentic-backend/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func SignUp(c *gin.Context) {

	var body struct {
		Email    string `json:"email" binding:"required,email"`
		Nickname string `json:"nickname" binding:"required"`
		Password string `json:"password" binding:"required,min=8"`
	}

	
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request", "details": err.Error()})
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(body.Password), 10)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to hash password", 
		})
		return
	}

	user := models.User{Password: string(hash), Nickname: body.Nickname, Email: body.Email}
	if err := initializers.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to create user",
		})
	}
	c.JSON(http.StatusOK, gin.H{})
}

func Login(c *gin.Context) {
	var body struct {
		Email string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}	

	var user models.User
	err := initializers.DB.First(&user, "email = ?", body.Email).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":"Invalid email or password",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":"Error while oppening db",
			})
		}
		return
	}
	
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(body.Password)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":"Invalid email or password",
		})
		return
	}

	accessClaims := jwt.RegisteredClaims{
		Subject: fmt.Sprint(user.ID),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24)),
		IssuedAt: jwt.NewNumericDate(time.Now()),
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessString, err := accessToken.SignedString([]byte(os.Getenv("SECRET")))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":"Failed to create access token",
		})
		return
	}

	refreshClaims := jwt.RegisteredClaims{
		Subject: fmt.Sprint(user.ID),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24 * 30)),
		IssuedAt: jwt.NewNumericDate(time.Now()),
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshString, err := refreshToken.SignedString([]byte(os.Getenv("REFRESH_SECRET")))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":"Failed to create refresh token",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access": accessString,
		"refresh": refreshString,
	})
}

func Refresh(c *gin.Context) {
	var body struct {
		Refresh  string `json:"refresh" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request", "details": err.Error()})
		return
	}

	token, err := jwt.Parse(body.Refresh, func(token *jwt.Token) (any, error) {
		return []byte(os.Getenv("SECRET")), nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
	if err != nil || !token.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "invalid refresh token",
		})
		return
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || float64(time.Now().Unix()) > claims["exp"].(float64) {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "invalid refresh token",
		})
		return
	}
	
	var user models.User
	if err := initializers.DB.First(&user, claims["sub"]).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	accessClaims := jwt.RegisteredClaims{
		Subject: fmt.Sprint(user.ID),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24)),
		IssuedAt: jwt.NewNumericDate(time.Now()),
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessString, err := accessToken.SignedString([]byte(os.Getenv("SECRET")))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":"Failed to create access token",
		})
		return
	}

	refreshClaims := jwt.RegisteredClaims{
		Subject: fmt.Sprint(user.ID),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24 * 30)),
		IssuedAt: jwt.NewNumericDate(time.Now()),
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshString, err := refreshToken.SignedString([]byte(os.Getenv("REFRESH_SECRET")))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":"Failed to create refresh token",
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"access": accessString,
		"refresh": refreshString,
	})
}

func Validate(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "no user in context",
		})
	}
	c.JSON(http.StatusOK, gin.H{
		"message": user,
	})
}

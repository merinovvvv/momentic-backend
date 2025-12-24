package controllers
import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/merinovvvv/momentic-backend/initializers"
	"github.com/merinovvvv/momentic-backend/models"
	"github.com/merinovvvv/momentic-backend/util"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"github.com/merinovvvv/momentic-backend/repository"
)

func SignUp(c *gin.Context) { // POST auth/register

	var body struct {
		Email    string `json:"email" binding:"required,email"`
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

	user := models.User{Password: string(hash), Email: body.Email}
	if err := initializers.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create user",
		})
	}

	verificationCode := fmt.Sprintf("%05d", rand.Intn(99999))
	if err := util.SendVerificationEmail(user.Email, verificationCode); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to send verification email",
		})
		return
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
	if user.Verified == false {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":"Email not verified",
		})
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
		c.JSON(http.StatusInternalServerError, gin.H{
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
		c.JSON(http.StatusInternalServerError, gin.H{
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
		c.JSON(http.StatusInternalServerError, gin.H{
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

func ChangeUserInfo(c *gin.Context) {
	var body struct {
		Password string `json:"password" binding:"required,min=8"`
		Name string `json:"name"`
		Surname string `json:"surname"`
		
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request", "details": err.Error()})
		return
	}
	
}

func VerifyEmail(c *gin.Context) {
	var body struct {
		 Email    string `json:"email" binding:"required,email"`
		 // Password string `json:"password" binding:"required,min=8"`
		 Code     string `json:"code" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request", "details": err.Error()})
		return
	}
	var emailVerification models.EmailVerification
	err := initializers.DB.First(&emailVerification, "email = ?", body.Email).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":"No such user or email is already verified",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":"Error while oppening db",
			})
		}
		return
	}
	if emailVerification.ExpiresAt.Before(time.Now()) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":"Verification code expired",
		})
		return
	}
	
	//TODO: make it transaction
	if err := initializers.DB.Delete(&emailVerification).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":"Failed to delete verification code",
		})
		return
	}
	if err := initializers.DB.Model(&models.User{}).Where("email = ?", body.Email).Update("verified", true).Error; err != nil { 
		c.JSON(http.StatusBadRequest, gin.H{
			"error":"No such user",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Email verified",
	})
}

func ResendEmailVerification(c *gin.Context) {
	var body struct {
		Email    string `json:"email" binding:"required,email"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request", "details": err.Error()})
		return
	}
	var user models.User
	err := initializers.DB.First(&user, "email = ?", body.Email).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":"No such user",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":"Error while oppening db",
			})
		}
		return
	}
	verificationCode := fmt.Sprintf("%05d", rand.Intn(99999))
	if err := util.SendVerificationEmail(user.Email, verificationCode); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to send verification email",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{})
}

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

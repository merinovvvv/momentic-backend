package main

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/merinovvvv/momentic-backend/controllers"
	"github.com/merinovvvv/momentic-backend/initializers"
	"github.com/merinovvvv/momentic-backend/middleware"
)

func init() {
	initializers.LoadEnvVariables()
	initializers.ConnectToDb()
}

func main() {
	router := gin.Default()
	router.POST("/auth/register", controllers.SignUp)
	router.PATCH("/auth/register", middleware.RequireAuth)
	router.POST("/auth/login", controllers.Login)
	router.POST("/auth/verify-code", controllers.VerifyEmail)
	router.POST("/auth/refresh", controllers.Refresh)
	router.POST("/auth/resend-verify-code", controllers.ResendEmailVerification)
	router.GET("/validate", middleware.RequireAuth, controllers.Validate)
	fmt.Println(router.Routes())
	router.Run() // listens on 0.0.0.0:8080 by default
}

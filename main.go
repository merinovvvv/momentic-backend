package main

import (
	"github.com/gin-gonic/gin"
	"github.com/merinovvvv/momentic-backend/initializers"
	"github.com/merinovvvv/momentic-backend/controllers"
)

func init() {
	initializers.LoadEnvVariables()
	initializers.ConnectToDb()
}

func main() {
	router := gin.Default()
	router.POST("/users", controllers.SignUp)
	router.Run() // listens on 0.0.0.0:8080 by default
}

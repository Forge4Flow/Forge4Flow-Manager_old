package main

import (
	"github.com/forge4flow/forge4flow-manager/internal/config"
	"github.com/forge4flow/forge4flow-manager/internal/database"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	config.LoadAppConfig()

	database.Connect()

	router := initRouter()

	router.Run(":8000")
}

func initRouter() *gin.Engine {
	router := gin.Default()
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	router.Use(cors.New(config))

	return router
}

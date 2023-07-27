package main

import (
	"github.com/forge4flow/forge4flow-manager/internal/config"
	"github.com/forge4flow/forge4flow-manager/internal/controllers"
	"github.com/forge4flow/forge4flow-manager/internal/services"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	config.LoadAppConfig()

	services.Start()

	router := initRouter()

	router.Run(":8000")
}

func initRouter() *gin.Engine {
	router := gin.Default()
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	router.Use(cors.New(config))

	router.NoRoute(controllers.ProxyRouter)

	api := router.Group("/api")
	{
		enviroments := api.Group("/enviroments")
		{
			enviroments.GET("/", controllers.GetEnvironments)
			enviroments.POST("/", controllers.CreateEnvironment)
			enviroments.GET("/:id", controllers.GetEnvironment)
			enviroments.PUT("/:id", controllers.UpdateEnvironment)
			enviroments.DELETE("/:id", controllers.DeleteEnvironment)
		}
	}

	return router
}

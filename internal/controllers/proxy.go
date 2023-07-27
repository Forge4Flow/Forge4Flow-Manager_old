package controllers

import (
	"errors"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/forge4flow/forge4flow-manager/internal/models"
	"github.com/forge4flow/forge4flow-manager/internal/services"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func ProxyRouter(context *gin.Context) {
	// Check if the Proxy should forward requests
	if strings.HasPrefix(context.Request.RequestURI, "/v1") || strings.HasPrefix(context.Request.RequestURI, "/v2") {
		// Get client key.
		clientKey := context.GetHeader("ClientKey")

		// Look up instance by key
		var instance models.CoreInstance
		if result := services.DB.Where(models.CoreInstance{ClientKey: clientKey}).First(&instance); result.Error != nil {
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				context.AbortWithStatusJSON(404, gin.H{
					"error": "Instance with provided key does not exist",
				})
				return
			}

			context.AbortWithStatusJSON(400, gin.H{
				"error": "Internal Proxy Error",
			})
			return
		}

		// if valid key, proxy request to the instance
		targetUrl, err := url.Parse("http://" + instance.Name + ":8000")
		if err != nil {
			context.AbortWithStatusJSON(500, gin.H{
				"error": "Internal Proxy Error",
			})
			return
		}
		proxy := httputil.NewSingleHostReverseProxy(targetUrl)
		proxy.ServeHTTP(context.Writer, context.Request)
		return
	}

	// Not valid proxy request, send 404
	context.AbortWithStatus(404)
}

package config

import (
	"net/http"

	"ADRIFT-backend/internal/middleware"

	"github.com/gin-gonic/gin"
)

func NewRouter(server *gin.Engine) *gin.Engine {
	// NoRoute handler
	server.NoRoute(func(ctx *gin.Context) {
		ctx.JSON(http.StatusNotFound, gin.H{
			"status":  http.StatusNotFound,
			"message": "Route Not Found",
		})
	})

	// Apply middleware
	server.Use(middleware.CORSMiddleware())

	// Health check endpoint
	server.GET("/api/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "aku sehat, kamu kangen?",
		})
	})

	return server
}

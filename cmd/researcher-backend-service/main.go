package main

import (
	"net/http"
	"time"

	"github.com/coneno/logger"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	v1 "github.com/tekenradar/researcher-backend/pkg/http/v1"
)

func healthCheckHandle(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "researcher backend running"})
}

func main() {

	conf := InitConfig()

	logger.SetLevel(conf.LogLevel)

	// Start webserver
	router := gin.Default()
	router.Use(cors.New(cors.Config{
		// AllowAllOrigins: true,
		AllowOrigins:     conf.AllowOrigins,
		AllowMethods:     []string{"POST", "GET", "PUT", "DELETE"},
		AllowHeaders:     []string{"Origin", "Authorization", "Content-Type", "Content-Length"},
		ExposeHeaders:    []string{"Authorization", "Content-Type", "Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
	router.GET("/health", healthCheckHandle)
	v1Root := router.Group("/v1")

	v1APIHandlers := v1.NewHTTPHandler(conf.SAMLConfig)
	v1APIHandlers.AddAuthAPI(v1Root)

	logger.Info.Printf("Tekenradar researcher backend started, listening on port %s", conf.Port)
	logger.Error.Fatal(router.Run(":" + conf.Port))
}

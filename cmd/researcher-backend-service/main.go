package main

import (
	"net/http"
	"time"

	"github.com/coneno/logger"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"github.com/tekenradar/researcher-backend/internal/config"
	"github.com/tekenradar/researcher-backend/pkg/db"
	"github.com/tekenradar/researcher-backend/pkg/grpc/clients"
	v1 "github.com/tekenradar/researcher-backend/pkg/http/v1"
	"github.com/tekenradar/researcher-backend/pkg/runner"
)

const (
	runnerCooldownInSeconds = 3600 * 6
)

func healthCheckHandle(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "researcher backend running"})
}

func main() {
	conf := config.InitConfig()
	researcherDBService := db.NewResearcherDBService(conf.ResearcherDBConfig)

	grpcClients := &clients.APIClients{}
	studyClient, studyServiceClose := clients.ConnectToStudyService(conf.ServiceURLs.StudyService, conf.MaxMsgSize)
	defer studyServiceClose()
	grpcClients.StudyService = studyClient

	logger.SetLevel(conf.LogLevel)

	if !conf.GinDebugMode {
		gin.SetMode(gin.ReleaseMode)
	}

	// Start runner
	backgroundRunner := runner.NewRunner(researcherDBService, runnerCooldownInSeconds)
	backgroundRunner.Run()

	// Start webserver
	router := gin.Default()
	router.Use(cors.New(cors.Config{
		// AllowAllOrigins: true,
		AllowOrigins:     conf.AllowOrigins,
		AllowMethods:     []string{"POST", "GET", "PUT", "DELETE"},
		AllowHeaders:     []string{"Origin", "Authorization", "Content-Type", "Content-Length", "Api-Key"},
		ExposeHeaders:    []string{"Authorization", "Content-Type", "Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
	router.GET("/health", healthCheckHandle)
	// router.Static("/app", "/var/www/html/webapp")
	v1Root := router.Group("/v1")

	v1APIHandlers := v1.NewHTTPHandler(
		grpcClients,
		researcherDBService,
		conf.SAMLConfig,
		conf.UseDummyLogin,
		conf.LoginSuccessRedirectURL,
		conf.APIKeys,
	)
	v1APIHandlers.AddAuthAPI(v1Root)
	v1APIHandlers.AddStudyEventsAPI(v1Root)
	v1APIHandlers.AddStudyAccessAPI(v1Root)
	v1APIHandlers.AddStudyManagementAPI(v1Root)

	logger.Info.Printf("Tekenradar researcher backend started, listening on port %s", conf.Port)
	logger.Error.Fatal(router.Run(":" + conf.Port))
}

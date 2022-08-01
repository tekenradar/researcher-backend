package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"
	mw "github.com/tekenradar/researcher-backend/pkg/http/middlewares"
)

func (h *HttpEndpoints) AddStudyAccessAPI(rg *gin.RouterGroup) {
	studiesGroup := rg.Group("/study")

	studiesGroup.Use(mw.HasValidAPIKey(h.apiKeys))
	studiesGroup.Use(mw.ValidateToken())
	{
		studiesGroup.GET("/", h.getStudyInfos)

		studyGroup := studiesGroup.Group(":studyKey")
		studyGroup.Use(mw.HasAccessToStudy())
		{
			studyGroup.GET("/", h.getStudyInfo)
			// TODO: fetch notification subscriptions
			// TODO: fetch available datasets
			// TODO: fetch data
			// TODO: fetch contact infos
			// TODO: mark participant contact info as permantent (toggle)
			// TODO: save participant contact note
		}
	}
}

func (h *HttpEndpoints) getStudyInfos(c *gin.Context) {
	// TODO: fetch study infos for all studies the user has access to
	c.JSON(http.StatusNotImplemented, gin.H{"message": "not implemented"})
}

func (h *HttpEndpoints) getStudyInfo(c *gin.Context) {
	// TODO: fetch study info for a particular study
	c.JSON(http.StatusNotImplemented, gin.H{"message": "not implemented"})
}

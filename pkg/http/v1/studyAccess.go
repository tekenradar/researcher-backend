package v1

import (
	"net/http"

	"github.com/coneno/logger"
	"github.com/gin-gonic/gin"
	mw "github.com/tekenradar/researcher-backend/pkg/http/middlewares"
	"github.com/tekenradar/researcher-backend/pkg/jwt"
	"github.com/tekenradar/researcher-backend/pkg/types"
)

func (h *HttpEndpoints) AddStudyAccessAPI(rg *gin.RouterGroup) {
	studiesGroup := rg.Group("/study")

	studiesGroup.Use(mw.HasValidAPIKey(h.apiKeys))
	studiesGroup.Use(mw.ValidateToken())
	{
		studiesGroup.GET("/infos", h.getStudyInfos)

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
	token := c.MustGet("validatedToken").(*jwt.UserClaims)

	studyInfos, err := h.researcherDB.FindStudyInfosByKeys(token.Studies)
	if err != nil {
		logger.Error.Printf("%v", err)
		c.JSON(http.StatusOK, gin.H{"studyInfos": []types.StudyInfo{}})
		return
	}
	logger.Info.Printf("study infos fetched by '%s'", token.ID)

	c.JSON(http.StatusOK, gin.H{"studyInfos": studyInfos})
}

func (h *HttpEndpoints) getStudyInfo(c *gin.Context) {
	// TODO: fetch study info for a particular study
	c.JSON(http.StatusNotImplemented, gin.H{"message": "not implemented"})
}

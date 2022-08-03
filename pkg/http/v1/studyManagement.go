package v1

import (
	"net/http"

	"github.com/coneno/logger"
	"github.com/gin-gonic/gin"
	mw "github.com/tekenradar/researcher-backend/pkg/http/middlewares"
	"github.com/tekenradar/researcher-backend/pkg/jwt"
	"github.com/tekenradar/researcher-backend/pkg/types"
)

func (h *HttpEndpoints) AddStudyManagementAPI(rg *gin.RouterGroup) {
	studyManagementGroup := rg.Group("/study-management")

	studyManagementGroup.Use(mw.HasValidAPIKey(h.apiKeys))
	studyManagementGroup.Use(mw.ValidateToken())
	studyManagementGroup.Use(mw.IsAdmin())
	{
		studyManagementGroup.GET("/study-info", h.SM_getAllStudyInfos) // fetch all study infos (even if not explicitly member of it, since admin role)
		studyManagementGroup.POST("/study-info", h.SM_saveStudyInfo)   // save study info (create or overwrite)
		studyManagementGroup.DELETE("/study-info/:studyKey", h.SM_deleteStudyInfo)
	}
}

func (h *HttpEndpoints) SM_getAllStudyInfos(c *gin.Context) {
	token := c.MustGet("validatedToken").(*jwt.UserClaims)

	studyInfos, err := h.researcherDB.FindAllStudyInfos()
	if err != nil {
		logger.Error.Printf("%v", err)
		c.JSON(http.StatusOK, gin.H{"studyInfos": []types.StudyInfo{}})
		return
	}
	logger.Info.Printf("all study infos (ADMIN) fetched by '%s'", token.ID)

	c.JSON(http.StatusOK, gin.H{"studyInfos": studyInfos})
}

func (h *HttpEndpoints) SM_saveStudyInfo(c *gin.Context) {
	token := c.MustGet("validatedToken").(*jwt.UserClaims)

	var req types.StudyInfo
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error.Printf("error: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	si, err := h.researcherDB.SaveStudyInfo(req)
	if err != nil {
		logger.Error.Printf("error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	logger.Info.Printf("study info for '%s' saved by '%s'", req.Key, token.ID)
	c.JSON(http.StatusOK, si)
}

func (h *HttpEndpoints) SM_deleteStudyInfo(c *gin.Context) {
	// TODO: delete study info
	// TODO: delete email notifications for study info
	c.JSON(http.StatusNotImplemented, gin.H{"message": "not implemented"})
}

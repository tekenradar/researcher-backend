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
	studyManagementGroup := rg.Group("/substudy-management")

	studyManagementGroup.Use(mw.HasValidAPIKey(h.apiKeys))
	studyManagementGroup.Use(mw.ValidateToken())
	studyManagementGroup.Use(mw.IsAdmin())
	{
		studyManagementGroup.GET("", h.SM_getAllSubstudyInfos) // fetch all substudy infos (even if not explicitly member of it, since admin role)
		studyManagementGroup.POST("", h.SM_saveSubstudyInfo)   // save study info (create or overwrite)
		studyManagementGroup.DELETE("/:subsubstudyKey", h.SM_deleteSubstudyInfo)
	}
}

func (h *HttpEndpoints) SM_getAllSubstudyInfos(c *gin.Context) {
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

func (h *HttpEndpoints) SM_saveSubstudyInfo(c *gin.Context) {
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

func (h *HttpEndpoints) SM_deleteSubstudyInfo(c *gin.Context) {
	token := c.MustGet("validatedToken").(*jwt.UserClaims)
	substudyKey := c.Param("substudyKey")

	count, err := h.researcherDB.DeleteStudyInfo(substudyKey)
	if err != nil {
		logger.Error.Printf("error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if count < 1 {
		logger.Error.Printf("study not deleted: %s", substudyKey)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "study could not be deleted"})
		return
	}

	err = h.researcherDB.DeleteEmailAllNotificationsForStudy(substudyKey)
	if err != nil {
		logger.Error.Printf("error when removing study's email notifications for study key: %s", substudyKey)
	}

	logger.Info.Printf("study info for '%s' deleted by '%s'", substudyKey, token.ID)
	c.JSON(http.StatusOK, gin.H{"message": "study deleted"})
}

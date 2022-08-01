package v1

import (
	"github.com/gin-gonic/gin"
	mw "github.com/tekenradar/researcher-backend/pkg/http/middlewares"
)

func (h *HttpEndpoints) AddStudyManagementAPI(rg *gin.RouterGroup) {
	studyManagementGroup := rg.Group("/study-management")

	studyManagementGroup.Use(mw.HasValidAPIKey(h.apiKeys))
	studyManagementGroup.Use(mw.ValidateToken())
	studyManagementGroup.Use(mw.IsAdmin())
	{
		// TODO: save study info (create or overwrite)
	}
}

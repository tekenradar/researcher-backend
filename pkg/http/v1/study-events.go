package v1

import (
	"net/http"
	"time"

	"github.com/coneno/logger"
	"github.com/gin-gonic/gin"
	"github.com/influenzanet/study-service/pkg/studyengine"
	"github.com/tekenradar/researcher-backend/pkg/types"
)

func (h *HttpEndpoints) AddStudyEventsAPI(rg *gin.RouterGroup) {

	studyEventsGroup := rg.Group("/study-events")

	studyEventsGroup.POST("/t0-invite", h.t0InviteEventHandl)
}

func (h *HttpEndpoints) t0InviteEventHandl(c *gin.Context) {
	var req studyengine.ExternalEventPayload
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error.Printf("error: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	studyInfos, err := h.researcherDB.FindAllStudyInfos()
	if err != nil {
		logger.Error.Printf("error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	pc, err := extractParticipantContactInfosFromEvent(req)
	if err != nil {
		logger.Error.Printf("error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	for _, studyInfo := range studyInfos {
		if !studyInfo.Features.Contacts {
			continue
		}

		_, err := h.researcherDB.AddParticipantContact(studyInfo.Key, pc)
		if err != nil {
			logger.Error.Printf("failed to create participant contact object with error: %v", err)
			continue
		}

		// TODO: send notifications
	}

	c.JSON(http.StatusOK, gin.H{"message": "event processed"})
}

func extractParticipantContactInfosFromEvent(event studyengine.ExternalEventPayload) (pc types.ParticipantContact, err error) {
	pc = types.ParticipantContact{
		AddedAt:         time.Now().Unix(),
		ParticipantID:   event.ParticipantState.ParticipantID,
		SessionID:       event.ParticipantState.CurrentStudySession,
		KeepContactData: false,
		Notes:           []types.ContactNote{},
	}
	// TODO: object
	// TODO: general
	// TODO: details

	return
}

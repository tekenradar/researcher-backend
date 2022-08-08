package v1

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/coneno/logger"
	"github.com/gin-gonic/gin"
	"github.com/influenzanet/messaging-service/pkg/api/email_client_service"
	"github.com/influenzanet/study-service/pkg/studyengine"
	mw "github.com/tekenradar/researcher-backend/pkg/http/middlewares"
	"github.com/tekenradar/researcher-backend/pkg/http/utils"
	"github.com/tekenradar/researcher-backend/pkg/types"
)

func (h *HttpEndpoints) AddStudyEventsAPI(rg *gin.RouterGroup) {

	studyEventsGroup := rg.Group("/study-events")
	studyEventsGroup.Use(mw.HasValidAPIKey(h.apiKeys))

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

		if !shouldIncludeParticipantContact(studyInfo, req) {
			continue
		}

		_, err := h.researcherDB.AddParticipantContact(studyInfo.Key, pc)
		if err != nil {
			logger.Error.Printf("failed to create participant contact object with error: %v", err)
			continue
		}

		subs, err := h.researcherDB.FindNotificationSubscriptions(studyInfo.Key, "contact")
		if err != nil {
			logger.Debug.Printf("failed to fetch notification subscriptions: %v", err)
			continue
		}
		for _, sub := range subs {

			_, err := h.clients.EmailClientService.SendEmail(context.TODO(), &email_client_service.SendEmailReq{
				To:      []string{sub.Email},
				Subject: fmt.Sprintf("Tekenradar - new contact entry added in study %s", studyInfo.Name),
				Content: fmt.Sprintf(
					"A participant matching the flags for %s (%s) study just entered contact information. \n\nIf no action is taken, the entry will automatically be removed from the system in 12 weeks. \n\n You are receiving this message because your email address is registered in the tekenradar researcher app for this study. Contact: tekenradar@rivm.nl",
					studyInfo.Name, studyInfo.Key,
				),
			})
			if err != nil {
				logger.Debug.Printf("failed to send notification for %s: %v", sub.Email, err)
				continue
			}
		}

	}

	c.JSON(http.StatusOK, gin.H{"message": "event processed"})
}

func shouldIncludeParticipantContact(studyInfo types.StudyInfo, event studyengine.ExternalEventPayload) bool {
	if len(studyInfo.ContactFeatureConfig.IncludeWithParticipantFlags) == 0 {
		return false
	}
	for k, v := range studyInfo.ContactFeatureConfig.IncludeWithParticipantFlags {
		fv, ok := event.ParticipantState.Flags[k]
		if ok && fv == v {
			continue
		}
		return false
	}
	return true
}

func extractParticipantContactInfosFromEvent(event studyengine.ExternalEventPayload) (pc types.ParticipantContact, err error) {
	pc = types.ParticipantContact{
		AddedAt:         time.Now().Unix(),
		ParticipantID:   event.ParticipantState.ParticipantID,
		SessionID:       event.ParticipantState.CurrentStudySession,
		KeepContactData: false,
		Notes:           []types.ContactNote{},
	}

	// -->
	ageFlag, ok := event.ParticipantState.Flags["ageFromPDiff"]
	if ok {
		age, err := strconv.Atoi(strings.Split(ageFlag, ".")[0])
		if err == nil {
			pc.General.Age = age
		}
	}

	// -->
	gender, ok := event.ParticipantState.Flags["gender"]
	if ok {
		pc.General.Gender = gender
	}

	// -->
	pc.General.OtherStudies = false
	otherStudies, ok := event.ParticipantState.Flags["consentAdditionalStudies"]
	if ok {
		pc.General.OtherStudies = otherStudies == "true"
	}

	pc.ContactData.FirstName, err = utils.ExtractResponseValue(event.Response.Responses, "Contactgegevens.Naam", "rg.cloze.vn")
	if err != nil {
		logger.Debug.Println(err)
	}

	pc.ContactData.LastName, err = utils.ExtractResponseValue(event.Response.Responses, "Contactgegevens.Naam", "rg.cloze.an")
	if err != nil {
		logger.Debug.Println(err)
	}

	pc.ContactData.Email, err = utils.ExtractResponseValue(event.Response.Responses, "Contactgegevens.Email", "rg.ic")
	if err != nil {
		logger.Debug.Println(err)
	}

	pc.ContactData.Phone, err = utils.ExtractResponseValue(event.Response.Responses, "Contactgegevens.Tel", "rg.ic")
	if err != nil {
		logger.Debug.Println(err)
	}

	pc.ContactData.Gender, err = utils.MapSingleChoiceResponse(event.Response.Responses, "Contactgegevens.GENDER", map[string]string{
		"a": "male",
		"b": "female",
		"c": "other",
	})
	if err != nil {
		logger.Debug.Println(err)
	}

	pc.ContactData.Birthday, err = utils.ExtractResponseValueAsNum(event.Response.Responses, "Contactgegevens.Birthday", "rg.date")
	if err != nil {
		logger.Debug.Println(err)
	}

	gpInfos, err := utils.FindSurveyItemResponse(event.Response.Responses, "Contactgegevens.GP")
	if err != nil {
		logger.Debug.Println(err)
	} else {
		pc.ContactData.GP = &types.GPInfos{}

		office, err := utils.FindResponseSlot(gpInfos.Response, "rg.cloze.pn")
		if err != nil {
			logger.Debug.Println(err)
		}
		pc.ContactData.GP.Office = office.Value

		name, err := utils.FindResponseSlot(gpInfos.Response, "rg.cloze.nh")
		if err != nil {
			logger.Debug.Println(err)
		}
		pc.ContactData.GP.Name = name.Value

		tel, err := utils.FindResponseSlot(gpInfos.Response, "rg.cloze.tel")
		if err != nil {
			logger.Debug.Println(err)
		}
		pc.ContactData.GP.Phone = tel.Value

		street, err := utils.FindResponseSlot(gpInfos.Response, "rg.cloze.str")
		if err != nil {
			logger.Debug.Println(err)
		}
		pc.ContactData.GP.Address.Street = street.Value

		hnr, err := utils.FindResponseSlot(gpInfos.Response, "rg.cloze.hnr")
		if err != nil {
			logger.Debug.Println(err)
		}
		pc.ContactData.GP.Address.Nr = hnr.Value

		postcode, err := utils.FindResponseSlot(gpInfos.Response, "rg.cloze.pc")
		if err != nil {
			logger.Debug.Println(err)
		}
		pc.ContactData.GP.Address.Postcode = postcode.Value

		city, err := utils.FindResponseSlot(gpInfos.Response, "rg.cloze.plaats")
		if err != nil {
			logger.Debug.Println(err)
		}
		pc.ContactData.GP.Address.City = city.Value
	}

	return pc, nil
}

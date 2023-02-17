package v1

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/coneno/logger"
	"github.com/gin-gonic/gin"
	mw "github.com/tekenradar/researcher-backend/pkg/http/middlewares"
	"github.com/tekenradar/researcher-backend/pkg/jwt"
	"github.com/tekenradar/researcher-backend/pkg/types"
	"google.golang.org/grpc/status"

	"github.com/influenzanet/go-utils/pkg/api_types"
	studyAPI "github.com/influenzanet/study-service/pkg/api"
)

const (
	instanceID = "tekenradar"
)

func (h *HttpEndpoints) AddStudyAccessAPI(rg *gin.RouterGroup) {
	studiesGroup := rg.Group("/substudy")

	studiesGroup.Use(mw.HasValidAPIKey(h.apiKeys))
	studiesGroup.Use(mw.ValidateToken())
	{
		studiesGroup.GET("/infos", h.getStudyInfos)

		studyGroup := studiesGroup.Group(":substudyKey")
		studyGroup.Use(mw.HasAccessToStudy(h.researcherDB))
		{
			studyGroup.GET("/", h.getStudyInfo)
			studyGroup.GET("/data/:datasetKey", h.downloadDataset) // ? from=1213123&until=12313212
			studyGroup.GET("/participant-contacts", h.getParticipantContacts)
			studyGroup.GET("/participant-contacts/:contactID/keep", h.changeParticipantContactKeepStatus) // ?value=true
			studyGroup.POST("/participant-contacts/:contactID/note", h.addNoteToParticipantContact)
			studyGroup.DELETE("/participant-contacts/:contactID", h.deleteParticipantContact)

			// TODO: fetch notification subscriptions
			studyGroup.GET("/notifications", h.fetchNotificationSubscriptions) // ?topic=value
			studyGroup.POST("/notifications", h.addNotificationSubscription)
			studyGroup.DELETE("/notifications/:notificationID", h.deleteNotificationSubscription)
		}
	}
}

func (h *HttpEndpoints) getStudyInfos(c *gin.Context) {
	token := c.MustGet("validatedToken").(*jwt.UserClaims)

	studyInfos, err := h.researcherDB.FindAllStudyInfos()
	if err != nil {
		logger.Error.Printf("%v", err)
		c.JSON(http.StatusOK, gin.H{"studyInfos": []types.StudyInfo{}})
		return
	}
	logger.Info.Printf("study infos fetched by '%s'", token.ID)

	studyInfos = filterStudyInfos(studyInfos, token.ID)
	c.JSON(http.StatusOK, gin.H{"studyInfos": studyInfos})
}

func filterStudyInfos(studyInfos []types.StudyInfo, email string) []types.StudyInfo {
	var filteredInfos []types.StudyInfo
	for _, info := range studyInfos {
		if contains(info.AccessControl.Emails, email) {
			filteredInfos = append(filteredInfos, info)
		}
	}
	return filteredInfos
}

// method to check if value is in array
func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func (h *HttpEndpoints) getStudyInfo(c *gin.Context) {
	token := c.MustGet("validatedToken").(*jwt.UserClaims)
	substudyKey := c.Param("substudyKey")

	studyInfo, err := h.researcherDB.FindStudyInfo(substudyKey)
	if err != nil {
		logger.Error.Printf("%v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}
	logger.Info.Printf("study info for %s fetched by '%s'", substudyKey, token.ID)

	c.JSON(http.StatusOK, studyInfo)
}

func (h *HttpEndpoints) getParticipantContacts(c *gin.Context) {
	token := c.MustGet("validatedToken").(*jwt.UserClaims)
	substudyKey := c.Param("substudyKey")

	pcs, err := h.researcherDB.FindParticipantContacts(substudyKey)
	if err != nil {
		logger.Error.Printf("%v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}
	logger.Info.Printf("partcipant contacts for %s fetched by '%s'", substudyKey, token.ID)

	c.JSON(http.StatusOK, gin.H{"participantContacts": pcs})
}

func (h *HttpEndpoints) changeParticipantContactKeepStatus(c *gin.Context) {
	token := c.MustGet("validatedToken").(*jwt.UserClaims)
	substudyKey := c.Param("substudyKey")
	contactID := c.Param("contactID")

	keep := c.DefaultQuery("value", "") == "true"

	err := h.researcherDB.UpdateKeepParticipantContactStatus(substudyKey, contactID, keep)
	if err != nil {
		logger.Error.Printf("%v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	pcs, err := h.researcherDB.FindParticipantContacts(substudyKey)
	if err != nil {
		logger.Error.Printf("%v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}
	logger.Info.Printf("partcipant contacts for %s fetched by '%s'", substudyKey, token.ID)

	c.JSON(http.StatusOK, gin.H{"participantContacts": pcs})
}

func (h *HttpEndpoints) addNoteToParticipantContact(c *gin.Context) {
	token := c.MustGet("validatedToken").(*jwt.UserClaims)
	substudyKey := c.Param("substudyKey")
	contactID := c.Param("contactID")

	var req types.ContactNote
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error.Printf("error: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.researcherDB.AddNoteToParticipantContact(substudyKey, contactID, req)
	if err != nil {
		logger.Error.Printf("%v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	pcs, err := h.researcherDB.FindParticipantContacts(substudyKey)
	if err != nil {
		logger.Error.Printf("%v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}
	logger.Info.Printf("partcipant contacts note added in study %s fetched by '%s'", substudyKey, token.ID)

	c.JSON(http.StatusOK, gin.H{"participantContacts": pcs})
}

func (h *HttpEndpoints) deleteParticipantContact(c *gin.Context) {
	token := c.MustGet("validatedToken").(*jwt.UserClaims)
	substudyKey := c.Param("substudyKey")
	contactID := c.Param("contactID")

	err := h.researcherDB.DeleteParticipantContact(substudyKey, contactID)
	if err != nil {
		logger.Error.Printf("%v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	pcs, err := h.researcherDB.FindParticipantContacts(substudyKey)
	if err != nil {
		logger.Error.Printf("%v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}
	logger.Info.Printf("partcipant contacts note added in study %s fetched by '%s'", substudyKey, token.ID)

	c.JSON(http.StatusOK, gin.H{"participantContacts": pcs})
}

func (h *HttpEndpoints) downloadDataset(c *gin.Context) {
	token := c.MustGet("validatedToken").(*jwt.UserClaims)
	substudyKey := c.Param("substudyKey")
	datasetKey := c.Param("datasetKey")

	studyInfo, err := h.researcherDB.FindStudyInfo(substudyKey)
	if err != nil {
		logger.Error.Printf("%v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	var dataset *types.DatasetInfo
	for _, datasetInfo := range studyInfo.AvailableDatasets {
		if datasetKey == datasetInfo.ID {
			dataset = &datasetInfo
			break
		}
	}
	if dataset == nil {
		msg := fmt.Sprintf("no dataset info found in study %s for dataset id %s", substudyKey, datasetKey)
		logger.Error.Println(msg)
		c.JSON(http.StatusBadRequest, gin.H{"error": msg})
		return
	}

	var req studyAPI.ResponseExportQuery
	req.StudyKey = "tekenradar"
	req.SurveyKey = dataset.SurveyKey

	from := c.DefaultQuery("from", "")
	if len(from) > 0 {
		n, err := strconv.ParseInt(from, 10, 64)
		if err == nil {
			req.From = n
		}
	}
	until := c.DefaultQuery("until", "")
	if len(until) > 0 {
		n, err := strconv.ParseInt(until, 10, 64)
		if err == nil {
			req.Until = n
		}
	}
	req.IncludeMeta = &studyAPI.ResponseExportQuery_IncludeMeta{
		Position:       c.DefaultQuery("withPositions", "false") == "true",
		InitTimes:      c.DefaultQuery("withInitTimes", "false") == "true",
		DisplayedTimes: c.DefaultQuery("withDisplayTimes", "false") == "true",
		ResponsedTimes: c.DefaultQuery("withResponseTimes", "false") == "true",
	}
	req.Separator = c.DefaultQuery("sep", "-")
	req.ShortQuestionKeys = c.DefaultQuery("shortKeys", "true") == "true"

	req.Token = &api_types.TokenInfos{
		Id:         token.ID,
		InstanceId: instanceID,
		Payload: map[string]string{
			"roles": "SERVICE",
		},
	}

	// Check if query valid:

	if dataset.EndDate > 0 && req.Until > dataset.EndDate {
		logger.Debug.Println("trying to access data that is later than allowed")
		logger.Error.Printf("user %s tried to access dataset %s", token.ID, datasetKey)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "no permission to access this dataset"})
		return
	}
	if dataset.StartDate > 0 && req.From < dataset.StartDate {
		logger.Debug.Println("trying to access data that is earlier than allowed")
		logger.Error.Printf("user %s tried to access dataset %s", token.ID, datasetKey)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "no permission to access this dataset"})
		return
	}

	req.ItemFilter = &studyAPI.ResponseExportQuery_ItemFilter{
		Mode: studyAPI.ResponseExportQuery_ItemFilter_EXCLUDE,
		Keys: dataset.ExcludeColumns,
	}

	stream, err := h.clients.StudyService.GetResponsesWideFormatCSV(context.Background(), &req)
	if err != nil {
		st := status.Convert(err)
		logger.Error.Printf("user %s tried to access dataset %s resulted in error %s", token.ID, datasetKey, st.Message())
		c.JSON(http.StatusInternalServerError, gin.H{"error": st.Message()})
		return
	}

	content := []byte{}
	for {
		chnk, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			st := status.Convert(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": st.Message()})
			return
		}
		content = append(content, chnk.Chunk...)
	}

	reader := bytes.NewReader(content)
	contentLength := int64(len(content))
	contentType := "text/csv"

	extraHeaders := map[string]string{
		"Content-Disposition": `attachment; filename=` + fmt.Sprintf("%s_%s.csv", substudyKey, dataset.SurveyKey),
	}

	c.DataFromReader(http.StatusOK, contentLength, contentType, reader, extraHeaders)
}

func (h *HttpEndpoints) fetchNotificationSubscriptions(c *gin.Context) {
	token := c.MustGet("validatedToken").(*jwt.UserClaims)
	substudyKey := c.Param("substudyKey")
	topic := c.DefaultQuery("topic", "")

	subs, err := h.researcherDB.FindNotificationSubscriptions(substudyKey, topic)
	if err != nil {
		logger.Error.Printf("%v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}
	logger.Info.Printf("email notification subs for %s fetched by '%s'", substudyKey, token.ID)

	c.JSON(http.StatusOK, gin.H{"emailNotifications": subs})
}

func (h *HttpEndpoints) addNotificationSubscription(c *gin.Context) {
	token := c.MustGet("validatedToken").(*jwt.UserClaims)
	substudyKey := c.Param("substudyKey")

	var req types.NotificationSubscription
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error.Printf("error: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, err := h.researcherDB.AddNotificationSubscription(substudyKey, req)
	if err != nil {
		logger.Error.Printf("%v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}
	logger.Info.Printf("email notification added for %s fetched by '%s'", substudyKey, token.ID)

	subs, err := h.researcherDB.FindNotificationSubscriptions(substudyKey, req.Topic)
	if err != nil {
		logger.Error.Printf("%v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	c.JSON(http.StatusOK, gin.H{"emailNotifications": subs})
}

func (h *HttpEndpoints) deleteNotificationSubscription(c *gin.Context) {
	token := c.MustGet("validatedToken").(*jwt.UserClaims)
	substudyKey := c.Param("substudyKey")
	notificationID := c.Param("notificationID")

	_, err := h.researcherDB.DeleteNotificationSubscription(substudyKey, notificationID)
	if err != nil {
		logger.Error.Printf("%v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}
	logger.Info.Printf("email notification deleted for %s fetched by '%s'", substudyKey, token.ID)

	c.JSON(http.StatusOK, gin.H{"message": "successfully deleted"})
}

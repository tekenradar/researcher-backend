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
	studiesGroup := rg.Group("/study")

	studiesGroup.Use(mw.HasValidAPIKey(h.apiKeys))
	studiesGroup.Use(mw.ValidateToken())
	{
		studiesGroup.GET("/infos", h.getStudyInfos)

		studyGroup := studiesGroup.Group(":studyKey")
		studyGroup.Use(mw.HasAccessToStudy())
		{
			studyGroup.GET("/", h.getStudyInfo)
			studyGroup.GET("/data/:datasetKey", h.downloadDataset) // ? from=1213123&until=12313212
			studyGroup.GET("/participant-contacts", h.getParticipantContacts)
			studyGroup.GET("/participant-contacts/:contactID/keep", h.changeParticipantContactKeepStatus) // ?value=true
			studyGroup.POST("/participant-contacts/:contactID/note", h.addNoteToParticipantContact)

			// TODO: fetch notification subscriptions
			studyGroup.GET("/notifications", h.fetchNotificationSubscriptions) // ?topic=value
			studyGroup.POST("/notifications", h.addNotificationSubscription)
			studyGroup.DELETE("/notifications/:notificationID", h.deleteNotificationSubscription)
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
	token := c.MustGet("validatedToken").(*jwt.UserClaims)
	studyKey := c.Param("studyKey")

	studyInfo, err := h.researcherDB.FindStudyInfo(studyKey)
	if err != nil {
		logger.Error.Printf("%v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}
	logger.Info.Printf("study info for %s fetched by '%s'", studyKey, token.ID)

	c.JSON(http.StatusOK, studyInfo)
}

func (h *HttpEndpoints) getParticipantContacts(c *gin.Context) {
	token := c.MustGet("validatedToken").(*jwt.UserClaims)
	studyKey := c.Param("studyKey")

	pcs, err := h.researcherDB.FindParticipantContacts(studyKey)
	if err != nil {
		logger.Error.Printf("%v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}
	logger.Info.Printf("partcipant contacts for %s fetched by '%s'", studyKey, token.ID)

	c.JSON(http.StatusOK, gin.H{"participantContacts": pcs})
}

func (h *HttpEndpoints) changeParticipantContactKeepStatus(c *gin.Context) {
	token := c.MustGet("validatedToken").(*jwt.UserClaims)
	studyKey := c.Param("studyKey")
	contactID := c.Param("contactID")

	keep := c.DefaultQuery("value", "") == "true"

	err := h.researcherDB.UpdateKeepParticipantContactStatus(studyKey, contactID, keep)
	if err != nil {
		logger.Error.Printf("%v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	pcs, err := h.researcherDB.FindParticipantContacts(studyKey)
	if err != nil {
		logger.Error.Printf("%v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}
	logger.Info.Printf("partcipant contacts for %s fetched by '%s'", studyKey, token.ID)

	c.JSON(http.StatusOK, gin.H{"participantContacts": pcs})
}

func (h *HttpEndpoints) addNoteToParticipantContact(c *gin.Context) {
	token := c.MustGet("validatedToken").(*jwt.UserClaims)
	studyKey := c.Param("studyKey")
	contactID := c.Param("contactID")

	var req types.ContactNote
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error.Printf("error: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.researcherDB.AddNoteToParticipantContact(studyKey, contactID, req)
	if err != nil {
		logger.Error.Printf("%v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	pcs, err := h.researcherDB.FindParticipantContacts(studyKey)
	if err != nil {
		logger.Error.Printf("%v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}
	logger.Info.Printf("partcipant contacts note added in study %s fetched by '%s'", studyKey, token.ID)

	c.JSON(http.StatusOK, gin.H{"participantContacts": pcs})
}

func (h *HttpEndpoints) downloadDataset(c *gin.Context) {
	token := c.MustGet("validatedToken").(*jwt.UserClaims)
	studyKey := c.Param("studyKey")
	datasetKey := c.Param("datasetKey")

	studyInfo, err := h.researcherDB.FindStudyInfo(studyKey)
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
		msg := fmt.Sprintf("no dataset info found in study %s for dataset id %s", studyKey, datasetKey)
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
		ItemVersion:    c.DefaultQuery("withItemVersions", "false") == "true",
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
		"Content-Disposition": `attachment; filename=` + fmt.Sprintf("%s_%s.csv", studyKey, dataset.SurveyKey),
	}

	c.DataFromReader(http.StatusOK, contentLength, contentType, reader, extraHeaders)
}

func (h *HttpEndpoints) fetchNotificationSubscriptions(c *gin.Context) {
	token := c.MustGet("validatedToken").(*jwt.UserClaims)
	studyKey := c.Param("studyKey")
	topic := c.DefaultQuery("topic", "")

	subs, err := h.researcherDB.FindNotificationSubscriptions(studyKey, topic)
	if err != nil {
		logger.Error.Printf("%v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}
	logger.Info.Printf("email notification subs for %s fetched by '%s'", studyKey, token.ID)

	c.JSON(http.StatusOK, gin.H{"emailNotifications": subs})
}

func (h *HttpEndpoints) addNotificationSubscription(c *gin.Context) {
	token := c.MustGet("validatedToken").(*jwt.UserClaims)
	studyKey := c.Param("studyKey")

	var req types.NotificationSubscription
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error.Printf("error: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, err := h.researcherDB.AddNotificationSubscription(studyKey, req)
	if err != nil {
		logger.Error.Printf("%v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}
	logger.Info.Printf("email notification added for %s fetched by '%s'", studyKey, token.ID)

	subs, err := h.researcherDB.FindNotificationSubscriptions(studyKey, req.Topic)
	if err != nil {
		logger.Error.Printf("%v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	c.JSON(http.StatusOK, gin.H{"emailNotifications": subs})
}

func (h *HttpEndpoints) deleteNotificationSubscription(c *gin.Context) {
	token := c.MustGet("validatedToken").(*jwt.UserClaims)
	studyKey := c.Param("studyKey")
	notificationID := c.Param("notificationID")

	_, err := h.researcherDB.DeleteNotificationSubscription(studyKey, notificationID)
	if err != nil {
		logger.Error.Printf("%v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}
	logger.Info.Printf("email notification deleted for %s fetched by '%s'", studyKey, token.ID)

	c.JSON(http.StatusOK, gin.H{"message": "successfully deleted"})
}

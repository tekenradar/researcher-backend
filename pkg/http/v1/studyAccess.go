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
			// TODO: fetch contact infos
			// TODO: mark participant contact info as permantent (toggle)
			// TODO: save participant contact note
			// TODO: fetch notification subscriptions
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

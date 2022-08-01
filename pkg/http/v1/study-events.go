package v1

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *HttpEndpoints) AddStudyEventsAPI(rg *gin.RouterGroup) {

	studyEventsGroup := rg.Group("/study-events")

	studyEventsGroup.POST("/t0-invite", h.t0InviteEventHandl)
}

func (h *HttpEndpoints) t0InviteEventHandl(c *gin.Context) {
	// TODO: implement correct logic to handle T0 invites, by sorting them to correct categories

	// TODO: for debugging, save POST body as a JSON
	resp, err := ioutil.ReadAll(c.Request.Body)

	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "Unable to read request body",
		})
		return
	}

	err1 := ioutil.WriteFile("T0_invite_event.json", resp, 0644)

	if err1 != nil {
		fmt.Println("error:", err1)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "Unable to save the file",
		})
		return
	}

	// File saved successfully. Return proper result
	c.JSON(http.StatusOK, gin.H{
		"message": "Your file has been successfully saved."})
}

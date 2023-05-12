package v1

import (
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/coneno/logger"
	"github.com/tekenradar/researcher-backend/internal/config"
	mw "github.com/tekenradar/researcher-backend/pkg/http/middlewares"
	"github.com/tekenradar/researcher-backend/pkg/http/utils"
	"github.com/tekenradar/researcher-backend/pkg/jwt"

	"github.com/gin-gonic/gin"
)

func (h *HttpEndpoints) AddAuthAPI(rg *gin.RouterGroup) {
	auth := rg.Group("/auth")

	auth.POST("/init-token", mw.HasValidAPIKey(h.apiKeys), h.initToken)
	auth.POST("/renew-token", mw.HasValidAPIKey(h.apiKeys), h.renewToken)
	auth.POST("/logout", h.logout)
}

type InitTokenRequest struct {
	Email string `json:"email" binding:"required"`
}

func (h *HttpEndpoints) initToken(c *gin.Context) {
	var req InitTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error.Printf("error: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	roles := []string{}
	// check if user is a research admin
	researchAdmins := os.Getenv(config.ENV_RESEARCHADMIN_EMAILS)
	if strings.Contains(researchAdmins, req.Email) {
		roles = []string{
			jwt.ROLE_ADMIN,
		}
	}

	// prepare token
	token, err := jwt.GenerateNewToken(
		req.Email,
		utils.TokenMaxAge*time.Second,
		roles,
	)

	if err != nil {
		logger.Error.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	// prepare renew token
	renewToken := "todo"

	c.JSON(http.StatusOK, gin.H{"accessToken": token, "renewToken": renewToken, "expiresIn": utils.TokenMaxAge})
}

func (h *HttpEndpoints) renewToken(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
}

func (h *HttpEndpoints) logout(c *gin.Context) {
	// TODO: invalidate renew token
	c.JSON(http.StatusOK, gin.H{"msg": "logout successful"})
}

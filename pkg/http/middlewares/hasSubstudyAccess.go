package middlewares

import (
	"net/http"

	"github.com/coneno/logger"
	"github.com/gin-gonic/gin"
	"github.com/tekenradar/researcher-backend/pkg/db"
	"github.com/tekenradar/researcher-backend/pkg/jwt"
)

// ValidateToken reads the token from the request and validates it
func HasAccessToStudy(dbRef *db.ResearcherDBService) gin.HandlerFunc {
	return func(c *gin.Context) {
		substudyKey := c.Param("substudyKey")
		token := c.MustGet("validatedToken").(*jwt.UserClaims)

		// check if user has access to substudy
		substudyInfo, err := dbRef.FindStudyInfo(substudyKey)
		if err != nil {
			logger.Error.Printf("error: %v", err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		email := token.ID

		if contains(substudyInfo.AccessControl.Emails, email) {
			c.Next()
			return
		}

		logger.Error.Printf("user %s tried unauthorized access %s study", token.ID, substudyKey)
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
	}
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

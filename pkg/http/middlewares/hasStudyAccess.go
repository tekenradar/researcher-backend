package middlewares

import (
	"net/http"

	"github.com/coneno/logger"
	"github.com/gin-gonic/gin"
	"github.com/tekenradar/researcher-backend/pkg/jwt"
)

// ValidateToken reads the token from the request and validates it
func HasAccessToStudy() gin.HandlerFunc {
	return func(c *gin.Context) {
		studyKey := c.Param("studyKey")
		token := c.MustGet("validatedToken").(*jwt.UserClaims)

		for _, r := range token.Studies {
			if r == studyKey {
				c.Next()
				return
			}
		}
		logger.Error.Printf("user %s tried unauthorized access %s study", token.ID, studyKey)
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
	}
}

package middlewares

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tekenradar/researcher-backend/pkg/jwt"
)

// ValidateToken reads the token from the request and validates it
func IsAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.MustGet("validatedToken").(*jwt.UserClaims)

		for _, r := range token.Roles {
			if r == jwt.ROLE_ADMIN {
				c.Next()
				return
			}
		}
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "admin account required for this feature"})
	}
}

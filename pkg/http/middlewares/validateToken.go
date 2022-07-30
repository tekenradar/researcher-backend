package middlewares

import (
	"net/http"

	"github.com/coneno/logger"
	"github.com/gin-gonic/gin"
	"github.com/tekenradar/researcher-backend/pkg/http/utils"
	"github.com/tekenradar/researcher-backend/pkg/jwt"
)

// ValidateToken reads the token from the request and validates it
func ValidateToken() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := c.Cookie(utils.AuthCookieName)
		if err != nil {
			logger.Error.Println(err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "no Authorization token found"})
			c.Abort()
			return
		}

		parsedToken, valid, err := jwt.ValidateToken(token)
		if err != nil || !valid {
			logger.Error.Printf("invalid token with err: %v", err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			c.Abort()
			return
		}

		c.Set("validatedToken", parsedToken)
		c.Next()
	}
}

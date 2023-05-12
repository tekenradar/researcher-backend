package middlewares

import (
	"net/http"
	"strings"

	"github.com/coneno/logger"
	"github.com/gin-gonic/gin"
	"github.com/tekenradar/researcher-backend/pkg/jwt"
)

// ValidateToken reads the token from the request and validates it
func ValidateToken() gin.HandlerFunc {
	return func(c *gin.Context) {
		var token string
		req := c.Request
		tokens, ok := req.Header["Authorization"]
		if ok && len(tokens) > 0 {
			token = tokens[0]
			token = strings.TrimPrefix(token, "Bearer ")
			if len(token) == 0 {
				c.JSON(http.StatusBadRequest, gin.H{"error": "no Authorization token found"})
				c.Abort()
				return
			}
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "no Authorization token found"})
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

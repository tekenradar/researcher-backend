package middlewares

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func RequireQueryParams(params []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		for _, key := range params {
			value, ok := c.GetQuery(key)
			if !ok || len(value) < 1 {
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": key + " parameter missing"})
				return
			}
		}
		c.Next()
	}
}

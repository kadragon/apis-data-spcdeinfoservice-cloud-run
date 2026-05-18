package proxy

import (
	"crypto/subtle"
	"net/http"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware(expectedKey string) gin.HandlerFunc {
	expected := []byte(expectedKey)
	return func(c *gin.Context) {
		if c.Request.Method == http.MethodOptions {
			writeCORS(c)
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		got := []byte(c.GetHeader("x-api-key"))
		if subtle.ConstantTimeCompare(got, expected) != 1 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error":   "Unauthorized",
				"message": "Invalid or missing x-api-key header",
			})
			return
		}

		c.Next()
	}
}

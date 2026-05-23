package proxy

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func CORSMiddleware() gin.HandlerFunc {
	headers := map[string]string{
		"Access-Control-Allow-Origin":  "*",
		"Access-Control-Allow-Methods": "GET, OPTIONS",
		"Access-Control-Allow-Headers": "x-api-key, Content-Type",
	}
	return func(c *gin.Context) {
		for k, v := range headers {
			c.Header(k, v)
		}
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

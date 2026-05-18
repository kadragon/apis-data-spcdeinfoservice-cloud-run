package proxy

import "github.com/gin-gonic/gin"

var corsHeaders = map[string]string{
	"Access-Control-Allow-Origin":  "*",
	"Access-Control-Allow-Methods": "GET, OPTIONS",
	"Access-Control-Allow-Headers": "x-api-key, Content-Type",
}

func writeCORS(c *gin.Context) {
	for k, v := range corsHeaders {
		c.Header(k, v)
	}
}

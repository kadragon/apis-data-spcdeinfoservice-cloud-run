package proxy

import (
	"errors"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

var sharedClient = &http.Client{
	Timeout: 0,
	Transport: &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 20,
		IdleConnTimeout:     90 * time.Second,
	},
}

func NewHandler(baseURL, upstreamPath, serviceKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		q := c.Request.URL.Query()
		q.Set("serviceKey", serviceKey)
		target := baseURL + upstreamPath + "?" + q.Encode()

		req, err := http.NewRequestWithContext(c.Request.Context(), http.MethodGet, target, nil)
		if err != nil {
			writeCORS(c)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error":   "Internal Server Error",
				"message": "Failed to build upstream request",
			})
			return
		}
		req.Header.Set("User-Agent", RandomUA())

		resp, cancel, err := fetchWithRetry(c.Request.Context(), sharedClient, req)
		if err != nil {
			writeCORS(c)
			var ue *UpstreamError
			if errors.As(err, &ue) {
				errTitle := "Bad Gateway"
				msg := "Unable to process request"
				if ue.IsTimeout {
					errTitle = "Gateway Timeout"
					msg = "Request timed out"
				}
				c.AbortWithStatusJSON(ue.StatusCode, gin.H{"error": errTitle, "message": msg})
				return
			}
			c.AbortWithStatusJSON(http.StatusBadGateway, gin.H{
				"error":   "Bad Gateway",
				"message": "Unable to process request",
			})
			return
		}
		defer cancel()
		defer func() {
			if err := resp.Body.Close(); err != nil {
				log.Printf("close upstream body: %v", err)
			}
		}()

		ct := resp.Header.Get("Content-Type")
		if ct == "" {
			ct = "application/xml"
		}
		writeCORS(c)
		c.Header("Content-Type", ct)
		c.Status(resp.StatusCode)

		if _, err := io.Copy(c.Writer, resp.Body); err != nil {
			log.Printf("pipe error: %v", err)
		}
	}
}

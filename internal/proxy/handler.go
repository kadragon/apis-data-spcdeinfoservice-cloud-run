package proxy

import (
	"context"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kadragon/apis-data-spcdeinfoservice-cloud-run/internal/cache"
)

// HeaderTimeout bounds time to receive response headers per attempt.
// Body streaming is not subject to this deadline. Var so tests can override.
var HeaderTimeout = 10 * time.Second

func NewClient() *http.Client {
	transport := &http.Transport{
		MaxIdleConns:          200,
		MaxIdleConnsPerHost:   200,
		IdleConnTimeout:       90 * time.Second,
		ResponseHeaderTimeout: HeaderTimeout,
	}

	var rt http.RoundTripper = transport

	if os.Getenv("CACHE_ENABLED") == "true" {
		ttlMin := 10
		if ttlStr := os.Getenv("CACHE_TTL_MINUTES"); ttlStr != "" {
			if val, err := strconv.Atoi(ttlStr); err == nil && val > 0 {
				ttlMin = val
			}
		}
		ttl := time.Duration(ttlMin) * time.Minute
		memCache := cache.NewInMemoryCache(1 * time.Minute)
		rt = NewCachingRoundTripper(transport, memCache, ttl)
	}

	return &http.Client{
		Timeout:   0,
		Transport: rt,
	}
}

func NewHandler(baseURL, upstreamPath, serviceKey string, client *http.Client) gin.HandlerFunc {
	if client == nil {
		panic("NewHandler: client must not be nil")
	}
	return func(c *gin.Context) {
		q := c.Request.URL.Query()
		q.Set("serviceKey", serviceKey)
		target := baseURL + upstreamPath + "?" + q.Encode()

		req, err := http.NewRequestWithContext(c.Request.Context(), http.MethodGet, target, nil)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error":   "Internal Server Error",
				"message": "Failed to build upstream request",
			})
			return
		}
		req.Header.Set("User-Agent", RandomUA())

		resp, err := fetchWithRetry(c.Request.Context(), client, req)
		if err != nil {
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
		defer func() {
			if err := resp.Body.Close(); err != nil {
				log.Printf("close upstream body: %v", err)
			}
		}()

		ct := resp.Header.Get("Content-Type")
		if ct == "" {
			ct = "application/xml"
		}
		c.Header("Content-Type", ct)
		c.Status(resp.StatusCode)

		if _, err := io.Copy(c.Writer, resp.Body); err != nil && !errors.Is(err, context.Canceled) {
			log.Printf("pipe error: %v", err)
		}
	}
}

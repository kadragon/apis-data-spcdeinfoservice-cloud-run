package proxy

import (
	"context"
	"errors"
	"io"
	"log/slog"
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

	cacheEnabled := os.Getenv("CACHE_ENABLED")
	if cacheEnabled != "" && cacheEnabled != "true" {
		slog.Warn("cache: unrecognized CACHE_ENABLED", "value", cacheEnabled)
	}
	if cacheEnabled == "true" {
		ttlMin := 10
		if ttlStr := os.Getenv("CACHE_TTL_MINUTES"); ttlStr != "" {
			if val, err := strconv.Atoi(ttlStr); err != nil {
				slog.Warn("cache: invalid CACHE_TTL_MINUTES, using default", "value", ttlStr, "default_min", ttlMin)
			} else if val <= 0 {
				slog.Warn("cache: non-positive CACHE_TTL_MINUTES, using default", "value", val, "default_min", ttlMin)
			} else {
				ttlMin = val
			}
		}
		ttl := time.Duration(ttlMin) * time.Minute
		memCache := cache.NewInMemoryCache(1 * time.Minute)
		rt = NewCachingRoundTripper(transport, memCache, ttl)
		slog.Info("cache enabled", "ttl", ttl.String())
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

		proxyTarget(c, client, target)
	}
}

func proxyTarget(c *gin.Context, client *http.Client, target string) {
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
			slog.Error("close upstream body", "err", err)
		}
	}()

	ct := resp.Header.Get("Content-Type")
	if ct == "" {
		ct = "application/xml"
	}
	c.Header("Content-Type", ct)
	c.Status(resp.StatusCode)

	if _, err := io.Copy(c.Writer, resp.Body); err != nil && !errors.Is(err, context.Canceled) {
		slog.Warn("pipe error", "err", err)
	}
}

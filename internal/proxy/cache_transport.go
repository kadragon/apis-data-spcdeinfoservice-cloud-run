package proxy

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"github.com/kadragon/apis-data-spcdeinfoservice-cloud-run/internal/cache"
)

var _ http.RoundTripper = (*CachingRoundTripper)(nil)

type cachedResponse struct {
	StatusCode int
	Header     http.Header
	Body       []byte
}

// CachingRoundTripper intercepts HTTP requests and caches HTTP 200 OK responses.
type CachingRoundTripper struct {
	underlying http.RoundTripper
	cache      cache.Cache
	ttl        time.Duration
}

// NewCachingRoundTripper wraps an existing RoundTripper with cache support.
func NewCachingRoundTripper(underlying http.RoundTripper, c cache.Cache, ttl time.Duration) *CachingRoundTripper {
	return &CachingRoundTripper{
		underlying: underlying,
		cache:      c,
		ttl:        ttl,
	}
}

// RoundTrip executes the HTTP request, returning cached response if present, otherwise fetching and caching.
func (c *CachingRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	// Only cache GET requests
	if req.Method != http.MethodGet {
		return c.underlying.RoundTrip(req)
	}

	key := generateCacheKey(req.Method, req.URL)
	if val, found := c.cache.Get(key); found {
		if cached, ok := val.(*cachedResponse); ok {
			resp := &http.Response{
				StatusCode:    cached.StatusCode,
				Proto:         "HTTP/1.1",
				ProtoMajor:    1,
				ProtoMinor:    1,
				Header:        cloneHeader(cached.Header),
				Body:          io.NopCloser(bytes.NewReader(cached.Body)),
				ContentLength: int64(len(cached.Body)),
				Request:       req,
			}
			return resp, nil
		}
	}

	resp, err := c.underlying.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	// Only cache HTTP 200 OK responses
	if resp.StatusCode == http.StatusOK {
		bodyBytes, err := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if err != nil {
			return nil, err
		}

		cached := &cachedResponse{
			StatusCode: resp.StatusCode,
			Header:     cloneHeader(resp.Header),
			Body:       bodyBytes,
		}
		c.cache.Set(key, cached, c.ttl)
		// Note: data.go.kr may return HTTP 200 with an error body (e.g. resultCode != "00").
		// This proxy does not inspect response bodies, so such errors are cached for the full TTL.
		slog.Debug("cache stored", "key", key, "ttl", c.ttl.String())

		resp.Body = io.NopCloser(bytes.NewReader(bodyBytes))
	}

	return resp, nil
}

func generateCacheKey(method string, u *url.URL) string {
	q := u.Query()
	q.Del("serviceKey")        // Strip sensitive service API key
	encodedQuery := q.Encode() // Note: Encode() natively sorts parameters by key

	base := u.Scheme + "://" + u.Host + u.Path
	if encodedQuery != "" {
		base += "?" + encodedQuery
	}
	return method + ":" + base
}

func cloneHeader(h http.Header) http.Header {
	h2 := make(http.Header, len(h))
	for k, vv := range h {
		vv2 := make([]string, len(vv))
		copy(vv2, vv)
		h2[k] = vv2
	}
	return h2
}

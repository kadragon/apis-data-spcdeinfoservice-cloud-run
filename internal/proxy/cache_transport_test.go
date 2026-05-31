package proxy

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/url"
	"sync/atomic"
	"testing"
	"time"

	"github.com/kadragon/apis-data-spcdeinfoservice-cloud-run/internal/cache"
)

type mockRoundTripper struct {
	calls atomic.Int32
	fn    func(req *http.Request) (*http.Response, error)
}

func (m *mockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	m.calls.Add(1)
	return m.fn(req)
}

func TestCachingRoundTripper_HitAndMiss(t *testing.T) {
	memCache := cache.NewInMemoryCache(10 * time.Minute)
	defer memCache.Close()

	// Mock upstream that returns a counter
	var upstreamCalls int32
	mockUpstream := func(req *http.Request) (*http.Response, error) {
		atomic.AddInt32(&upstreamCalls, 1)
		resp := &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body:       io.NopCloser(bytes.NewReader([]byte("response-body"))),
			Request:    req,
		}
		resp.Header.Set("Content-Type", "application/xml")
		return resp, nil
	}

	transport := NewCachingRoundTripper(
		&mockRoundTripper{fn: mockUpstream},
		memCache,
		50*time.Millisecond,
	)

	// 1. First Request (Miss)
	req1, _ := http.NewRequestWithContext(context.Background(), "GET", "http://example.com/test?serviceKey=secret123&a=1&b=2", nil)
	resp1, err := transport.RoundTrip(req1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	body1, _ := io.ReadAll(resp1.Body)
	resp1.Body.Close()

	if string(body1) != "response-body" {
		t.Fatalf("want response-body, got %s", body1)
	}
	if resp1.Header.Get("Content-Type") != "application/xml" {
		t.Fatalf("want application/xml, got %s", resp1.Header.Get("Content-Type"))
	}
	if atomic.LoadInt32(&upstreamCalls) != 1 {
		t.Fatalf("expected 1 upstream call, got %d", atomic.LoadInt32(&upstreamCalls))
	}

	// 2. Second Request (Hit) - serviceKey is different, param order is different, but should hit
	req2, _ := http.NewRequestWithContext(context.Background(), "GET", "http://example.com/test?b=2&serviceKey=diffkey&a=1", nil)
	resp2, err := transport.RoundTrip(req2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	body2, _ := io.ReadAll(resp2.Body)
	resp2.Body.Close()

	if string(body2) != "response-body" {
		t.Fatalf("want cached response-body, got %s", body2)
	}
	if resp2.Header.Get("Content-Type") != "application/xml" {
		t.Fatalf("want application/xml, got %s", resp2.Header.Get("Content-Type"))
	}
	// Upstream calls should still be 1 (Cache Hit!)
	if atomic.LoadInt32(&upstreamCalls) != 1 {
		t.Fatalf("expected cached hit to skip upstream call, but count is %d", atomic.LoadInt32(&upstreamCalls))
	}

	// 3. Expire cache
	time.Sleep(60 * time.Millisecond)

	req3, _ := http.NewRequestWithContext(context.Background(), "GET", "http://example.com/test?a=1&b=2", nil)
	resp3, err := transport.RoundTrip(req3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	resp3.Body.Close()

	// Upstream call should be 2 (Cache Miss after expiration)
	if atomic.LoadInt32(&upstreamCalls) != 2 {
		t.Fatalf("expected upstream call after expiration, but count is %d", atomic.LoadInt32(&upstreamCalls))
	}
}

func TestCachingRoundTripper_Non200NoCache(t *testing.T) {
	memCache := cache.NewInMemoryCache(10 * time.Minute)
	defer memCache.Close()

	var upstreamCalls int32
	mockUpstream := func(req *http.Request) (*http.Response, error) {
		atomic.AddInt32(&upstreamCalls, 1)
		return &http.Response{
			StatusCode: http.StatusInternalServerError,
			Header:     make(http.Header),
			Body:       io.NopCloser(bytes.NewReader([]byte("error"))),
			Request:    req,
		}, nil
	}

	transport := NewCachingRoundTripper(
		&mockRoundTripper{fn: mockUpstream},
		memCache,
		10*time.Minute,
	)

	req1, _ := http.NewRequestWithContext(context.Background(), "GET", "http://example.com/test", nil)
	resp1, _ := transport.RoundTrip(req1)
	resp1.Body.Close()

	req2, _ := http.NewRequestWithContext(context.Background(), "GET", "http://example.com/test", nil)
	resp2, _ := transport.RoundTrip(req2)
	resp2.Body.Close()

	if atomic.LoadInt32(&upstreamCalls) != 2 {
		t.Fatalf("5xx response should not be cached, got %d calls", atomic.LoadInt32(&upstreamCalls))
	}
}

func TestGenerateCacheKey_Sanitization(t *testing.T) {
	u, _ := url.Parse("http://apis.data.go.kr/test?serviceKey=abc%3D%3D&a=2&b=1")
	key := generateCacheKey("GET", u)
	expected := "GET:http://apis.data.go.kr/test?a=2&b=1"
	if key != expected {
		t.Fatalf("expected cache key %q, got %q", expected, key)
	}

	u2, _ := url.Parse("http://apis.data.go.kr/test?b=1&a=2&serviceKey=XYZ")
	key2 := generateCacheKey("GET", u2)
	if key2 != expected {
		t.Fatalf("expected sorted cache key to match: got %q", key2)
	}
}

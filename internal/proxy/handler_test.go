package proxy

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/gin-gonic/gin"
)

var handlerTestClient = NewClient()

func newHandlerEngine(upstream *httptest.Server) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(CORSMiddleware())
	r.GET("/svc/getThing", NewHandler(upstream.URL, "/getThing", "test-svc-key", handlerTestClient))
	return r
}

func TestHandler_ProxiesAndInjectsServiceKey(t *testing.T) {
	var gotURL string
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotURL = r.URL.String()
		w.Header().Set("Content-Type", "application/xml")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("<ok/>"))
	}))
	defer upstream.Close()

	r := newHandlerEngine(upstream)
	req := httptest.NewRequestWithContext(context.Background(), "GET", "/svc/getThing?numOfRows=10", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", rec.Code)
	}
	if !strings.Contains(gotURL, "numOfRows=10") {
		t.Fatalf("upstream URL missing numOfRows: %s", gotURL)
	}
	if !strings.Contains(gotURL, "serviceKey=test-svc-key") {
		t.Fatalf("upstream URL missing serviceKey: %s", gotURL)
	}
	if rec.Header().Get("Content-Type") != "application/xml" {
		t.Fatalf("wrong Content-Type: %s", rec.Header().Get("Content-Type"))
	}
	if rec.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Fatal("missing CORS header")
	}
	if rec.Body.String() != "<ok/>" {
		t.Fatalf("wrong body: %s", rec.Body.String())
	}
}

func TestHandler_ContentTypeFallback(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		// No body write: Go's http package only sniffs Content-Type on Write(), so
		// omitting Write() ensures the upstream response has no Content-Type header,
		// letting our handler apply the "application/xml" fallback.
		w.WriteHeader(http.StatusOK)
	}))
	defer upstream.Close()

	r := newHandlerEngine(upstream)
	req := httptest.NewRequestWithContext(context.Background(), "GET", "/svc/getThing", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Header().Get("Content-Type") != "application/xml" {
		t.Fatalf("expected application/xml fallback, got: %s", rec.Header().Get("Content-Type"))
	}
}

func TestHandler_4xxNoRetry(t *testing.T) {
	var calls atomic.Int32
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusNotFound)
	}))
	defer upstream.Close()

	r := newHandlerEngine(upstream)
	req := httptest.NewRequestWithContext(context.Background(), "GET", "/svc/getThing", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("want 404, got %d", rec.Code)
	}
	if calls.Load() != 1 {
		t.Fatalf("4xx must not retry: got %d calls", calls.Load())
	}
}

func TestHandler_5xxRetryExhaustion(t *testing.T) {
	var calls atomic.Int32
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer upstream.Close()

	r := newHandlerEngine(upstream)
	req := httptest.NewRequestWithContext(context.Background(), "GET", "/svc/getThing", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadGateway {
		t.Fatalf("want 502, got %d", rec.Code)
	}
	if calls.Load() != 2 {
		t.Fatalf("want 2 upstream calls, got %d", calls.Load())
	}
	if rec.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Fatal("missing CORS on error response")
	}
}

func TestHandler_5xxThenSuccess(t *testing.T) {
	var calls atomic.Int32
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		n := calls.Add(1)
		if n < 2 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.Header().Set("Content-Type", "application/xml")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("<ok/>"))
	}))
	defer upstream.Close()

	r := newHandlerEngine(upstream)
	req := httptest.NewRequestWithContext(context.Background(), "GET", "/svc/getThing", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("want 200 after retry, got %d", rec.Code)
	}
	if calls.Load() != 2 {
		t.Fatalf("want 2 upstream calls, got %d", calls.Load())
	}
}

func TestHandler_UserAgentSet(t *testing.T) {
	var gotUA string
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUA = r.Header.Get("User-Agent")
		w.WriteHeader(http.StatusOK)
	}))
	defer upstream.Close()

	r := newHandlerEngine(upstream)
	req := httptest.NewRequestWithContext(context.Background(), "GET", "/svc/getThing", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if gotUA == "" {
		t.Fatal("User-Agent not set on upstream request")
	}
}

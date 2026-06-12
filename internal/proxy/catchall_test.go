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

const catchAllAuthKey = "test-auth-key"

// newCatchAllEngine mirrors main.go wiring: CORS → /health → auth → NoRoute
// catch-all proxying to the given upstream.
func newCatchAllEngine(upstream *httptest.Server) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(CORSMiddleware())
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	r.Use(AuthMiddleware(catchAllAuthKey))
	r.NoRoute(NewCatchAllHandler(upstream.URL, "test-svc-key", handlerTestClient))
	return r
}

func catchAllRequest(method, target string) *http.Request {
	req := httptest.NewRequestWithContext(context.Background(), method, target, nil)
	req.Header.Set("x-api-key", catchAllAuthKey)
	return req
}

func TestCatchAll_ProxiesArbitraryPathAndInjectsServiceKey(t *testing.T) {
	var gotURL string
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotURL = r.URL.String()
		w.Header().Set("Content-Type", "application/xml")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("<ok/>"))
	}))
	defer upstream.Close()

	r := newCatchAllEngine(upstream)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, catchAllRequest("GET", "/1360000/VilageFcstInfoService_2.0/getUltraSrtNcst?pageNo=1"))

	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", rec.Code)
	}
	if !strings.HasPrefix(gotURL, "/1360000/VilageFcstInfoService_2.0/getUltraSrtNcst?") {
		t.Fatalf("upstream URL missing original path: %s", gotURL)
	}
	if !strings.Contains(gotURL, "serviceKey=test-svc-key") {
		t.Fatalf("upstream URL missing serviceKey: %s", gotURL)
	}
	if !strings.Contains(gotURL, "pageNo=1") {
		t.Fatalf("upstream URL missing original query: %s", gotURL)
	}
	if rec.Body.String() != "<ok/>" {
		t.Fatalf("wrong body: %s", rec.Body.String())
	}
}

func TestCatchAll_RequiresAuth(t *testing.T) {
	var calls atomic.Int32
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer upstream.Close()

	r := newCatchAllEngine(upstream)
	rec := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(context.Background(), "GET", "/1360000/SomeService/getThing", nil)
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d", rec.Code)
	}
	if calls.Load() != 0 {
		t.Fatalf("upstream must not be called without auth: got %d calls", calls.Load())
	}
}

func TestCatchAll_Upstream4xxPassesThrough(t *testing.T) {
	var calls atomic.Int32
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls.Add(1)
		w.Header().Set("Content-Type", "application/xml")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte("<error>SERVICE ERROR</error>"))
	}))
	defer upstream.Close()

	r := newCatchAllEngine(upstream)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, catchAllRequest("GET", "/9999999/UnknownService/getThing"))

	if rec.Code != http.StatusNotFound {
		t.Fatalf("want upstream 404 passed through, got %d", rec.Code)
	}
	if rec.Body.String() != "<error>SERVICE ERROR</error>" {
		t.Fatalf("upstream error body must pass through, got: %s", rec.Body.String())
	}
	if calls.Load() != 1 {
		t.Fatalf("4xx must not retry: got %d calls", calls.Load())
	}
}

func TestCatchAll_RejectsPathTraversal(t *testing.T) {
	var calls atomic.Int32
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer upstream.Close()

	r := newCatchAllEngine(upstream)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, catchAllRequest("GET", "/1360000/../secret"))

	if rec.Code != http.StatusNotFound {
		t.Fatalf("want 404 for traversal path, got %d", rec.Code)
	}
	if calls.Load() != 0 {
		t.Fatalf("upstream must not be called for traversal path: got %d calls", calls.Load())
	}
}

func TestCatchAll_RejectsSingleSegmentPath(t *testing.T) {
	var calls atomic.Int32
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer upstream.Close()

	r := newCatchAllEngine(upstream)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, catchAllRequest("GET", "/favicon.ico"))

	if rec.Code != http.StatusNotFound {
		t.Fatalf("want 404 for single-segment path, got %d", rec.Code)
	}
	if calls.Load() != 0 {
		t.Fatalf("upstream must not be called for single-segment path: got %d calls", calls.Load())
	}
}

// data.go.kr paths are plain ASCII; percent-encoded segments decode to
// characters outside the whitelist and are rejected rather than forwarded.
func TestCatchAll_RejectsEncodedSpecialChars(t *testing.T) {
	var calls atomic.Int32
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer upstream.Close()

	r := newCatchAllEngine(upstream)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, catchAllRequest("GET", "/1360000/Some%20Service/op"))

	if rec.Code != http.StatusNotFound {
		t.Fatalf("want 404 for encoded special char path, got %d", rec.Code)
	}
	if calls.Load() != 0 {
		t.Fatalf("upstream must not be called: got %d calls", calls.Load())
	}
}

func TestCatchAll_NonGETMethodNotAllowed(t *testing.T) {
	var calls atomic.Int32
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer upstream.Close()

	r := newCatchAllEngine(upstream)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, catchAllRequest("POST", "/1360000/SomeService/getThing"))

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("want 405 for POST, got %d", rec.Code)
	}
	if calls.Load() != 0 {
		t.Fatalf("upstream must not be called for POST: got %d calls", calls.Load())
	}
}

func TestCatchAll_HealthRemainsOpen(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer upstream.Close()

	r := newCatchAllEngine(upstream)
	rec := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(context.Background(), "GET", "/health", nil)
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("/health must not require auth: got %d", rec.Code)
	}
}

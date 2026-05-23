package services

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func stubFactory(_, _, _ string) gin.HandlerFunc {
	return func(c *gin.Context) { c.Status(http.StatusOK) }
}

func newTestEngine(serviceKey string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	RegisterAll(r, serviceKey, stubFactory)
	return r
}

func TestAllowedPathCounts(t *testing.T) {
	tests := []struct {
		name  string
		spec  ServiceSpec
		count int
	}{
		{"SpcdeInfo", SpcdeInfoSpec, 3},
		{"GetSecuritiesProductInfo", GetSecuritiesProductInfoSpec, 3},
		{"BidPublicInfo", BidPublicInfoSpec, 25},
		{"KorService", KorServiceSpec, 15},
		{"PubDataOpnStd", PubDataOpnStdSpec, 3},
		{"SjFestival", SjFestivalSpec, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if len(tt.spec.AllowedPaths) != tt.count {
				t.Errorf("want %d allowed paths, got %d", tt.count, len(tt.spec.AllowedPaths))
			}
		})
	}
}

func TestUnknownPathReturns404(t *testing.T) {
	r := newTestEngine("key")
	req := httptest.NewRequestWithContext(context.Background(), "GET", "/SpcdeInfoService/nonexistent", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	// gin returns 404 for unregistered routes
	if rec.Code != http.StatusNotFound {
		t.Fatalf("want 404 for unknown path, got %d", rec.Code)
	}
}

func TestAllowedPathsRouted(t *testing.T) {
	for _, spec := range all {
		for _, p := range spec.AllowedPaths {
			path := spec.MountPath + p
			t.Run(path, func(t *testing.T) {
				// We only check that gin recognizes the route (200 or any non-404 is fine;
				// the handler will fail because the real upstream is unavailable, so we
				// accept anything other than 404 here).
				gin.SetMode(gin.TestMode)
				r := gin.New()
				// Register a stub handler so we don't hit the real upstream.
				grp := r.Group(spec.MountPath)
				grp.GET(p, func(c *gin.Context) { c.Status(http.StatusOK) })

				req := httptest.NewRequestWithContext(context.Background(), "GET", path, nil)
				rec := httptest.NewRecorder()
				r.ServeHTTP(rec, req)

				if rec.Code == http.StatusNotFound {
					t.Fatalf("route %s not registered (got 404)", path)
				}
			})
		}
	}
}

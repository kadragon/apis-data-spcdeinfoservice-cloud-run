package proxy

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func newAuthEngine(key string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(AuthMiddleware(key))
	r.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})
	return r
}

func TestAuthMiddleware_MissingKey(t *testing.T) {
	r := newAuthEngine("secret")
	req := httptest.NewRequestWithContext(context.Background(), "GET", "/test", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d", rec.Code)
	}
}

func TestAuthMiddleware_WrongKey(t *testing.T) {
	r := newAuthEngine("secret")
	req := httptest.NewRequestWithContext(context.Background(), "GET", "/test", nil)
	req.Header.Set("x-api-key", "wrong")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d", rec.Code)
	}
}

func TestAuthMiddleware_CorrectKey(t *testing.T) {
	r := newAuthEngine("secret")
	req := httptest.NewRequestWithContext(context.Background(), "GET", "/test", nil)
	req.Header.Set("x-api-key", "secret")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", rec.Code)
	}
}

func TestAuthMiddleware_OptionsPreflight(t *testing.T) {
	r := newAuthEngine("secret")
	req := httptest.NewRequestWithContext(context.Background(), "OPTIONS", "/test", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("want 204, got %d", rec.Code)
	}
	if rec.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Fatal("missing CORS header")
	}
}

func TestAuthMiddleware_LengthMismatch(t *testing.T) {
	r := newAuthEngine("secret")
	req := httptest.NewRequestWithContext(context.Background(), "GET", "/test", nil)
	req.Header.Set("x-api-key", "secret-extra-chars")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d", rec.Code)
	}
}

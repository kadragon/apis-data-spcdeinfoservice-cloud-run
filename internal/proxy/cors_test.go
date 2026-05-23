package proxy

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func newCORSEngine() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(CORSMiddleware())
	r.GET("/test", func(c *gin.Context) { c.Status(http.StatusOK) })
	return r
}

func TestCORSMiddleware_SetsHeadersOnGet(t *testing.T) {
	r := newCORSEngine()
	req := httptest.NewRequestWithContext(context.Background(), "GET", "/test", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Fatal("missing CORS header on GET")
	}
}

func TestCORSMiddleware_OptionsPreflight(t *testing.T) {
	r := newCORSEngine()
	req := httptest.NewRequestWithContext(context.Background(), "OPTIONS", "/test", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("want 204, got %d", rec.Code)
	}
	if rec.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Fatal("missing CORS header on OPTIONS")
	}
}

func TestCORSAndAuth_OptionsPreflightSkipsAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(CORSMiddleware())
	r.Use(AuthMiddleware("secret"))
	r.GET("/test", func(c *gin.Context) { c.Status(http.StatusOK) })

	req := httptest.NewRequestWithContext(context.Background(), "OPTIONS", "/test", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("OPTIONS must return 204 without x-api-key, got %d", rec.Code)
	}
}

func TestCORSMiddleware_HeadersPresentOnUnauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(CORSMiddleware())
	r.Use(AuthMiddleware("secret"))
	r.GET("/test", func(c *gin.Context) { c.Status(http.StatusOK) })

	req := httptest.NewRequestWithContext(context.Background(), "GET", "/test", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d", rec.Code)
	}
	if rec.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Fatal("CORS header missing on 401 — browser will see CORS error instead of auth error")
	}
}

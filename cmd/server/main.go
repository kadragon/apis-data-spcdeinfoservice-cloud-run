package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/kadragon/apis-data-spcdeinfoservice-cloud-run/internal/proxy"
	"github.com/kadragon/apis-data-spcdeinfoservice-cloud-run/internal/services"
)

func main() {
	authKey := mustEnv("AUTH_API_KEY")
	serviceKey := mustEnv("DATAGOKR_SERVICEKEY")

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	r.Use(proxy.CORSMiddleware())
	r.Use(proxy.AuthMiddleware(authKey))

	client := proxy.NewClient()
	services.RegisterAll(r, serviceKey, func(baseURL, path, svcKey string) gin.HandlerFunc {
		return proxy.NewHandler(baseURL, path, svcKey, client)
	})

	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Not Found"})
	})

	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           r,
		ReadHeaderTimeout: 10 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("Proxy server running on port %s", port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("listen: %v", err)
		}
	}()

	<-ctx.Done()
	stop()
	log.Println("Shutting down gracefully...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	if err := srv.Shutdown(shutdownCtx); err != nil {
		cancel()
		log.Fatalf("forced shutdown: %v", err)
	}
	cancel()
}

func mustEnv(k string) string {
	v := os.Getenv(k)
	if v == "" {
		log.Fatalf("FATAL: missing required env var %s", k)
	}
	return v
}

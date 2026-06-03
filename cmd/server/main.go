package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/kadragon/apis-data-spcdeinfoservice-cloud-run/internal/proxy"
	"github.com/kadragon/apis-data-spcdeinfoservice-cloud-run/internal/services"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		ReplaceAttr: func(_ []string, a slog.Attr) slog.Attr {
			if a.Key == slog.LevelKey {
				a.Key = "severity"
				// GCP Cloud Logging expects "WARNING"; slog emits "WARN".
				sev := strings.ToUpper(a.Value.String())
				if sev == "WARN" {
					sev = "WARNING"
				}
				a.Value = slog.StringValue(sev)
			}
			return a
		},
	}))
	slog.SetDefault(logger)

	authKey := mustEnv("AUTH_API_KEY")
	serviceKey := mustEnv("DATAGOKR_SERVICEKEY")

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(proxy.CORSMiddleware())

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	r.Use(proxy.AuthMiddleware(authKey))

	client := proxy.NewClient()
	services.RegisterAll(r, serviceKey, func(p services.HandlerParams) gin.HandlerFunc {
		return proxy.NewHandler(p.BaseURL, p.Path, p.ServiceKey, client)
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
		slog.Info("server starting", "port", port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("listen failed", "err", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	stop()
	slog.Info("shutting down gracefully")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	if err := srv.Shutdown(shutdownCtx); err != nil {
		cancel()
		slog.Error("forced shutdown", "err", err)
		os.Exit(1)
	}
	cancel()
}

func mustEnv(k string) string {
	v := os.Getenv(k)
	if v == "" {
		slog.Error("missing required env var", "name", k)
		os.Exit(1)
	}
	return v
}

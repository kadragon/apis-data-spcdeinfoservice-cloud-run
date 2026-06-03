package proxy

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func testClient() *http.Client {
	return &http.Client{Timeout: 0}
}

func TestFetchWithRetry_Success(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer upstream.Close()

	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, upstream.URL, nil)
	resp, err := fetchWithRetry(context.Background(), testClient(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("want 200, got %d", resp.StatusCode)
	}
}

func TestFetchWithRetry_5xxExhausted(t *testing.T) {
	var calls atomic.Int32
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer upstream.Close()

	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, upstream.URL, nil)
	_, err := fetchWithRetry(context.Background(), testClient(), req)
	if err == nil {
		t.Fatal("expected error")
	}
	if calls.Load() != 2 {
		t.Fatalf("want 2 attempts, got %d", calls.Load())
	}
	var ue *UpstreamError
	if !errors.As(err, &ue) {
		t.Fatalf("want UpstreamError, got %T", err)
	}
	if ue.StatusCode != http.StatusBadGateway {
		t.Fatalf("want 502, got %d", ue.StatusCode)
	}
}

func TestFetchWithRetry_4xxNoRetry(t *testing.T) {
	var calls atomic.Int32
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer upstream.Close()

	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, upstream.URL, nil)
	resp, err := fetchWithRetry(context.Background(), testClient(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if calls.Load() != 1 {
		t.Fatalf("4xx must not retry: got %d calls", calls.Load())
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", resp.StatusCode)
	}
}

func TestFetchWithRetry_NetworkErrorScrubsServiceKey(t *testing.T) {
	// Unroutable address forces a connection error whose *url.Error embeds
	// the full request URL, including the secret serviceKey query param.
	const secret = "TOPSECRETKEY1234"
	target := "http://127.0.0.1:1/path?serviceKey=" + secret + "&a=1"

	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, target, nil)
	client := &http.Client{Timeout: 100 * time.Millisecond}
	_, err := fetchWithRetry(context.Background(), client, req)
	if err == nil {
		t.Fatal("expected network error")
	}
	var ue *UpstreamError
	if !errors.As(err, &ue) {
		t.Fatalf("want UpstreamError, got %T", err)
	}
	if strings.Contains(ue.Message, secret) {
		t.Fatalf("serviceKey leaked in error message: %s", ue.Message)
	}
}

func TestFetchWithRetry_Timeout(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer upstream.Close()

	client := &http.Client{
		Transport: &http.Transport{
			ResponseHeaderTimeout: 50 * time.Millisecond,
		},
	}

	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, upstream.URL, nil)
	_, err := fetchWithRetry(context.Background(), client, req)
	if err == nil {
		t.Fatal("expected timeout error")
	}
	var ue *UpstreamError
	if !errors.As(err, &ue) {
		t.Fatalf("want UpstreamError, got %T", err)
	}
	if ue.StatusCode != http.StatusGatewayTimeout {
		t.Fatalf("want 504, got %d", ue.StatusCode)
	}
	if !ue.IsTimeout {
		t.Fatal("IsTimeout should be true")
	}
}

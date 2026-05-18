package proxy

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	MaxRetries    = 2
	BackoffBaseMs = 1000
)

// AttemptTimeout is a var (not const) so tests can override it.
var AttemptTimeout = 10 * time.Second

type UpstreamError struct {
	Message    string
	StatusCode int
	IsTimeout  bool
}

func (e *UpstreamError) Error() string { return e.Message }

// fetchWithRetry retries on network errors and 5xx responses.
// On success, returns the cancel func for the per-attempt ctx — caller must defer it
// alongside resp.Body.Close() to keep the ctx alive during streaming.
func fetchWithRetry(parent context.Context, client *http.Client, baseReq *http.Request) (*http.Response, context.CancelFunc, error) {
	var lastErr error

	for attempt := 0; attempt <= MaxRetries; attempt++ {
		ctx, cancel := context.WithTimeout(parent, AttemptTimeout)
		req := baseReq.Clone(ctx)

		resp, err := client.Do(req)
		switch {
		case err != nil:
			cancel()
			isTimeout := errors.Is(err, context.DeadlineExceeded)
			msg := err.Error()
			if isTimeout {
				msg = "Request timed out"
			}
			lastErr = &UpstreamError{
				Message:    msg,
				StatusCode: timeoutStatus(isTimeout),
				IsTimeout:  isTimeout,
			}
		case resp.StatusCode < 500:
			// Success — caller owns cancel and body.
			return resp, cancel, nil
		default:
			// 5xx — drain body so the connection can be reused, then retry.
			_, _ = io.Copy(io.Discard, resp.Body)
			_ = resp.Body.Close()
			cancel()
			lastErr = &UpstreamError{
				Message:    fmt.Sprintf("upstream returned %d", resp.StatusCode),
				StatusCode: http.StatusBadGateway,
			}
		}

		if attempt < MaxRetries {
			backoff := time.Duration(BackoffBaseMs*(1<<attempt)) * time.Millisecond
			select {
			case <-time.After(backoff):
			case <-parent.Done():
				return nil, nil, parent.Err()
			}
		}
	}

	return nil, nil, lastErr
}

func timeoutStatus(isTimeout bool) int {
	if isTimeout {
		return http.StatusGatewayTimeout
	}
	return http.StatusBadGateway
}

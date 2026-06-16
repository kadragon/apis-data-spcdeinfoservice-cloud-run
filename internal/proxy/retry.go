package proxy

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"time"
)

const MaxRetries = 1

// BackoffBaseMs is a var so tests can override it without timing flakiness.
var BackoffBaseMs = 1000

type UpstreamError struct {
	Message    string
	StatusCode int
	IsTimeout  bool
}

func (e *UpstreamError) Error() string { return e.Message }

// fetchWithRetry retries on network errors and 5xx responses.
// The per-attempt header deadline is governed by the client transport's
// ResponseHeaderTimeout; once headers arrive the body streams without
// an additional timeout so large responses are not truncated.
func fetchWithRetry(parent context.Context, client *http.Client, baseReq *http.Request) (*http.Response, error) {
	var lastErr error

	for attempt := 0; attempt <= MaxRetries; attempt++ {
		req := baseReq.Clone(parent)

		resp, err := client.Do(req)
		switch {
		case err != nil:
			if parent.Err() != nil {
				return nil, parent.Err()
			}
			isTimeout := isTimeoutErr(err)
			msg := scrubServiceKey(err)
			if isTimeout {
				msg = "Request timed out"
			}
			lastErr = &UpstreamError{
				Message:    msg,
				StatusCode: timeoutStatus(isTimeout),
				IsTimeout:  isTimeout,
			}
		case resp.StatusCode < 500:
			return resp, nil
		default:
			// 5xx — drain body so the connection can be reused, then retry.
			_, _ = io.Copy(io.Discard, resp.Body)
			_ = resp.Body.Close()
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
				return nil, parent.Err()
			}
		}
	}

	return nil, lastErr
}

// scrubServiceKey returns err's message with any embedded upstream URL
// redacted so the secret serviceKey query param never reaches logs or
// responses. *url.Error embeds the full request URL; we rewrite its URL
// field before stringifying. Falls back to the raw message for other errors.
func scrubServiceKey(err error) string {
	var ue *url.Error
	if errors.As(err, &ue) {
		redacted := *ue
		redacted.URL = redactURL(ue.URL)
		return redacted.Error()
	}
	return err.Error()
}

func redactURL(raw string) string {
	u, parseErr := url.Parse(raw)
	if parseErr != nil {
		return "[redacted url]"
	}
	q := u.Query()
	if q.Has("serviceKey") {
		q.Set("serviceKey", "REDACTED")
		u.RawQuery = q.Encode()
	}
	return u.String()
}

func isTimeoutErr(err error) bool {
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}
	var ne net.Error
	return errors.As(err, &ne) && ne.Timeout()
}

func timeoutStatus(isTimeout bool) int {
	if isTimeout {
		return http.StatusGatewayTimeout
	}
	return http.StatusBadGateway
}

# Tasks

## Deferred from PR #45 review

- [x] **[P3] `HandlerFactory` positional string args** (`internal/services/registry.go`) — Replaced three positional `string` params with `HandlerParams` struct (`BaseURL, Path, ServiceKey`) to prevent silent swaps.

## Deferred from PR #47 review

- [x] **[P3] `retry.go` latent serviceKey leak** (`internal/proxy/retry.go`) — `scrubServiceKey` unwraps `*url.Error` and redacts the `serviceKey` query param before building `UpstreamError`. Covered by `TestFetchWithRetry_NetworkErrorScrubsServiceKey`.
- [x] **[P2] `os.Exit` in goroutine bypasses graceful shutdown** (`cmd/server/main.go`) — `ListenAndServe` failure now sends on a `serveErr` channel; main goroutine `select`s on it and exits cleanly instead of `os.Exit`-ing mid-stack inside the goroutine.
- [x] **[P3] slog default handler in tests** — `TestMain` in `internal/proxy/setup_test.go` sets a discard handler, silencing package-level `slog.*` output under test.

## Deferred from PR #49 review

- [ ] **[P3] `TestMain` slog discard only covers `internal/proxy`** — `internal/cache` and `internal/services` still emit default-handler `slog.*` output under test. Add the same `TestMain` discard pattern to those packages (separate PR).

## Review Backlog

### PR #51 — Catch-all proxy for any data.go.kr API + Cloud Run cost trim (2026-06-12)

- [ ] [debt] Add `message` field to catch-all path-validation 404 for response-shape consistency with other error JSON (source: pr-review-toolkit:review-pr) — internal/proxy/catchall.go:32
- [x] [debt] Exempt `context.DeadlineExceeded` from pipe-error warn in `proxyTarget` io.Copy guard (pre-existing, relocated by refactor) (source: pr-review-toolkit:review-pr) — internal/proxy/handler.go:114 *(resolved in PR #52)*
- [x] [debt] Log (with `redactURL`) when `http.NewRequestWithContext` fails before returning 500 (pre-existing) (source: pr-review-toolkit:review-pr) — internal/proxy/handler.go:73 *(resolved in PR #52)*
- [x] [debt] Early-return on raw `context.Canceled`/`DeadlineExceeded` from `fetchWithRetry` backoff instead of generic silent 502 (pre-existing) (source: pr-review-toolkit:review-pr) — internal/proxy/handler.go:82 *(resolved in PR #52)*
- [ ] [constraint] Test that caller-supplied `?serviceKey=` is overwritten (guards Golden Principle 1 against `q.Set`→`q.Add` regression) (source: pr-review-toolkit:review-pr) — internal/proxy/catchall_test.go
- [ ] [constraint] Test upstream 5xx→502 with retry through catch-all wiring (source: pr-review-toolkit:review-pr) — internal/proxy/catchall_test.go
- [ ] [constraint] Assert CORS header in catch-all success test for parity with handler_test.go (source: pr-review-toolkit:review-pr) — internal/proxy/catchall_test.go
- [ ] [debt] `strings.TrimSuffix(baseURL, "/")` guard in `NewCatchAllHandler` against double-slash target URLs (source: agy) — internal/proxy/catchall.go:38

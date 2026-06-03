# Tasks

## Deferred from PR #45 review

- [x] **[P3] `HandlerFactory` positional string args** (`internal/services/registry.go`) — Replaced three positional `string` params with `HandlerParams` struct (`BaseURL, Path, ServiceKey`) to prevent silent swaps.

## Deferred from PR #47 review

- [x] **[P3] `retry.go` latent serviceKey leak** (`internal/proxy/retry.go`) — `scrubServiceKey` unwraps `*url.Error` and redacts the `serviceKey` query param before building `UpstreamError`. Covered by `TestFetchWithRetry_NetworkErrorScrubsServiceKey`.
- [x] **[P2] `os.Exit` in goroutine bypasses graceful shutdown** (`cmd/server/main.go`) — `ListenAndServe` failure now sends on a `serveErr` channel; main goroutine `select`s on it and exits cleanly instead of `os.Exit`-ing mid-stack inside the goroutine.
- [x] **[P3] slog default handler in tests** — `TestMain` in `internal/proxy/setup_test.go` sets a discard handler, silencing package-level `slog.*` output under test.

## Deferred from PR #49 review

- [ ] **[P3] `TestMain` slog discard only covers `internal/proxy`** — `internal/cache` and `internal/services` still emit default-handler `slog.*` output under test. Add the same `TestMain` discard pattern to those packages (separate PR).

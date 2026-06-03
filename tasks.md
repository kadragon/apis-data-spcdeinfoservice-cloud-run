# Tasks

## Deferred from PR #45 review

- [x] **[P3] `HandlerFactory` positional string args** (`internal/services/registry.go`) — Replaced three positional `string` params with `HandlerParams` struct (`BaseURL, Path, ServiceKey`) to prevent silent swaps.

## Deferred from PR #47 review

- [ ] **[P3] `retry.go` latent serviceKey leak** (`internal/proxy/retry.go:40`) — `msg := err.Error()` captures `*url.Error` which embeds the full upstream URL incl. `serviceKey`. Not logged today (returned only as generic client message), but a landmine if `fetchWithRetry` error is ever slog'd. Scrub URL when building `UpstreamError` from network errors.
- [ ] **[P2] `os.Exit` in goroutine bypasses graceful shutdown** (`cmd/server/main.go`) — `ListenAndServe` failure calls `os.Exit(1)` inside the goroutine, skipping signal handler / defers. Pre-existing behavior (was `log.Fatalf`); propagate error via channel to main goroutine if a clean shutdown path is wanted.
- [ ] **[P3] slog default handler in tests** — package-level `slog.*` calls in `internal/proxy` use Go default text handler under test (JSON only set in `main()`). Cosmetic test noise; set a discard handler in `TestMain` if it bothers.

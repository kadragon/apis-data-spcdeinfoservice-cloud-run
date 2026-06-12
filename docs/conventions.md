# Conventions

## Adding a New Proxied API

No code change needed by default — the catch-all `NoRoute` handler proxies any
`apis.data.go.kr` path verbatim. Create a `ServiceSpec` only when a short alias
route is wanted:

1. Create `internal/services/<name>.go` — one file, one var:
   ```go
   var MyServiceSpec = services.ServiceSpec{
       MountPath:    "/my-service",
       BaseURL:      "https://api.data.go.kr/openapi/MyService",
       AllowedPaths: []string{"/getMyData"},
   }
   ```
2. Add `MyServiceSpec` to the `all` slice in `internal/services/registry.go`.
3. Done. No changes to `cmd/server/main.go` needed.

**Never** put two `ServiceSpec` vars in the same file.

## Error Handling

- Use `fmt.Errorf("context: %w", err)` for wrapping — never `errors.New` when wrapping.
- `golangci-lint` enforces `errorlint` (unwrapped errors in switch/if) and `nilerr` (returning nil error when err != nil).
- HTTP handler errors: use `c.JSON(status, gin.H{"error": msg})` and `return` immediately.

## Context Usage

- All outbound HTTP calls must use a context. `golangci-lint noctx` enforces this.
- `fetchWithRetry` manages its own per-attempt timeout context (10s). Caller passes parent ctx from gin.

## Body Closing

- `golangci-lint bodyclose` enforces that every `http.Response.Body` is closed.
- Pattern: `defer resp.Body.Close()` immediately after error check.

## Logging

- Use gin's default logger for request/response logging.
- `log.Printf` for startup/shutdown messages only.
- **Never log secret values.** `DATAGOKR_SERVICEKEY` and `AUTH_API_KEY` must not appear in any log statement.

## Test Files

- Live alongside source: `internal/proxy/handler_test.go`, `internal/services/registry_test.go`
- Run with `-race` flag: `go test ./... -race`
- Test exclusions in `.golangci.yml`: `_test.go` files skip `errcheck`, `bodyclose`, `errchkjson`, `forcetypeassert`

## Naming

- Service spec vars: `<PascalCaseName>Spec` (e.g., `BidPublicInfoSpec`)
- Handler constructors: `NewHandler` (single factory, no per-service variants)
- Test helpers: unexported, scoped to `_test.go`

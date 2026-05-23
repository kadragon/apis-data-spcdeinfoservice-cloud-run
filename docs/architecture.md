# Architecture

## Stack

| Layer | Technology |
|-------|-----------|
| Language | Go 1.25 |
| Framework | gin v1.12 |
| Build | `go build ./...` |
| CI/CD | Cloud Build → Artifact Registry → Cloud Run |
| Runtime | distroless/static-debian12:nonroot |

## Source Layout

```
cmd/server/main.go          gin engine, env validation, route registration, graceful shutdown
internal/
  proxy/
    handler.go              NewHandler — builds gin.HandlerFunc for a proxied route
    retry.go                fetchWithRetry — HTTP client with 3-attempt retry + backoff
    auth.go                 AuthMiddleware — x-api-key constant-time compare
    cors.go                 writeCORS — sets CORS response headers
    useragent.go            RandomUA() — random User-Agent from curated pool
  services/
    registry.go             ServiceSpec struct, all slice, RegisterAll()
    *.go                    one ServiceSpec per file (one var per file)
```

## Layer Rules

### Dependency Direction

```
cmd/server → internal/services → internal/proxy
```

`internal/proxy` knows nothing about `internal/services`. `cmd/server` wires them.

### Boundaries

- `internal/proxy/` is reusable — no service-specific logic
- `internal/services/` is config-only — no HTTP client logic, only `ServiceSpec` declarations
- Auth, CORS, and retry logic live exclusively in `internal/proxy/`; no duplicates elsewhere
- Secret values (`serviceKey`, `AUTH_API_KEY`) flow only through function parameters, never globals

## Proxy Flow

```
Request → gin router → AuthMiddleware → gin.HandlerFunc (NewHandler)
  → fetchWithRetry (3 attempts, ResponseHeaderTimeout=10s per attempt, no body timeout, 1s/2s backoff on 5xx/network)
  → inject serviceKey query param + random User-Agent
  → stream upstream body via io.Copy
  → writeCORS headers
```

## Key Abstractions

1. **`ServiceSpec`** — declarative config for one upstream API: `MountPath`, `BaseURL`, `AllowedPaths`. Adding a new API = one new file with one `ServiceSpec`.
2. **`NewHandler`** — factory returning `gin.HandlerFunc`; takes `(baseURL, upstreamPath, serviceKey)`. Stateless — no shared mutable state.
3. **`fetchWithRetry`** — signature `(ctx, client, req) → (*http.Response, error)`. No `CancelFunc` returned. Timeout is `ResponseHeaderTimeout=10s` on the HTTP transport; body streaming has no deadline.
4. **`AuthMiddleware`** — wraps the entire router group (excluding `/health`). `OPTIONS` → 204 pass-through for preflight.

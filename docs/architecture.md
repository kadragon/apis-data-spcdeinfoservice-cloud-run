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
    handler.go              NewHandler — builds gin.HandlerFunc for a proxied route; proxyTarget — shared forwarding core; NewClient() — default HTTP client constructor
    catchall.go             NewCatchAllHandler — NoRoute fallback proxying any valid path to apis.data.go.kr
    retry.go                fetchWithRetry — HTTP client with 3-attempt retry + backoff
    auth.go                 AuthMiddleware — x-api-key constant-time compare
    cors.go                 CORSMiddleware — sets CORS headers on all responses, handles OPTIONS preflight
    useragent.go            RandomUA() — random User-Agent from curated pool
  services/
    registry.go             ServiceSpec struct, HandlerFactory type, all slice, RegisterAll()
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
Request → gin router → CORSMiddleware (sets CORS headers; OPTIONS → 204 abort)
  → AuthMiddleware (x-api-key check; 401 already has CORS from above)
  → gin.HandlerFunc (NewHandler for alias routes; NewCatchAllHandler via NoRoute for everything else)
    → fetchWithRetry (3 attempts, ResponseHeaderTimeout=10s per attempt, no body timeout, 1s/2s backoff on 5xx/network)
    → inject serviceKey query param + random User-Agent
    → stream upstream body via io.Copy
```

Unmatched paths hit `NewCatchAllHandler` (registered as gin `NoRoute`; global CORS/Auth
middleware still apply). It forwards `GET /<≥2 URL-safe segments>` verbatim to
`https://apis.data.go.kr`, rejecting other methods (405) and dot-segment/odd paths (404).
Upstream responses — including 4xx and data.go.kr error payloads — pass through unchanged.

## Key Abstractions

1. **`ServiceSpec`** — declarative short-alias config for one upstream API: `MountPath`, `BaseURL`, `AllowedPaths`. Optional since the catch-all route covers any `apis.data.go.kr` path; add one only to expose a short alias. One new file with one `ServiceSpec`.
2. **`HandlerFactory`** — `func(HandlerParams) gin.HandlerFunc` where `HandlerParams{BaseURL, Path, ServiceKey}` (named fields prevent silent swaps of same-typed args). Passed to `RegisterAll`; `main.go` closes over the HTTP client. Swap the factory to change handler strategy (caching, rate-limiting) without touching the registry.
3. **`NewHandler`** — factory returning `gin.HandlerFunc`; takes `(baseURL, upstreamPath, serviceKey string, client *http.Client)`. Stateless — no shared mutable state.
4. **`NewClient`** — constructs the default `*http.Client` with tuned transport. `main.go` calls this once and closes over it in the factory.
5. **`fetchWithRetry`** — signature `(ctx, client, req) → (*http.Response, error)`. No `CancelFunc` returned. Timeout is `ResponseHeaderTimeout=10s` on the HTTP transport; body streaming has no deadline.
6. **`CORSMiddleware`** — registered before `AuthMiddleware`; sets CORS headers on every response including 401 and handles `OPTIONS` preflight (204). Do not add CORS headers in handlers.
7. **`AuthMiddleware`** — wraps the entire router group (excluding `/health` and OPTIONS preflights handled by `CORSMiddleware`).

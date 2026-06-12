# apis-data-spcdeinfoservice

API proxy server that forwards requests to Korean public data APIs ([data.go.kr](https://data.go.kr)), injecting service keys and handling CORS. Deployed to Google Cloud Run via Cloud Build.

## Proxy Routes

**Catch-all:** any unmatched `GET /<agency>/<service...>/<operation>` path is forwarded
verbatim to `https://apis.data.go.kr/<same path>` with the service key injected. APIs not
registered to the key return data.go.kr's own error payload unchanged. Example:

```bash
curl -H "x-api-key: YOUR_API_KEY" \
  "http://localhost:3000/1360000/VilageFcstInfoService_2.0/getUltraSrtNcst?pageNo=1"
```

**Short aliases** (kept for backward compatibility):

| Route | Description | Upstream API |
|---|---|---|
| `/SpcdeInfoService` | Korean special day info (holidays, anniversaries, 24 divisions) | `B090041/openapi/service/SpcdeInfoService` |
| `/GetSecuritiesProductInfoService` | Securities price info (ETF, ETN, ELW) | `1160100/service/GetSecuritiesProductInfoService` |
| `/BidPublicInfoService` | Public bid/procurement info (construction, service, goods, etc.) | `1230000/ad/BidPublicInfoService` |
| `/KorService2` | Tourism API | `B551011/KorService2` |
| `/PubDataOpnStdService` | Nara-jangteo open standard | `1230000/ao/PubDataOpnStdService` |
| `/sjFestival` | Sejong festival info | `5690000/sjFestival` |

## Environment Variables

| Variable | Required | Description |
|---|---|---|
| `AUTH_API_KEY` | Yes | Client authentication key, checked via `x-api-key` header |
| `DATAGOKR_SERVICEKEY` | Yes | Service key injected into proxied requests to data.go.kr |
| `PORT` | No | Server port (default: 3000, Dockerfile sets 8080) |

## Commands

```bash
go build ./...        # Build
go run ./cmd/server   # Run locally
go test ./...         # Test
golangci-lint run     # Lint
```

## Usage

All requests require an `x-api-key` header matching `AUTH_API_KEY`. `/health` is public.

```bash
curl -H "x-api-key: YOUR_API_KEY" \
  "http://localhost:3000/SpcdeInfoService/getRestDeInfo?solYear=2026&solMonth=01"

curl -H "x-api-key: YOUR_API_KEY" \
  "http://localhost:3000/BidPublicInfoService/getBidPblancListInfoCnstwk?numOfRows=10&pageNo=1"
```

## Adding a New Proxied API

No code change needed — the catch-all route proxies any `apis.data.go.kr` path.
To add a short alias instead:

1. Create `internal/services/<name>.go` with a `var <Name>Spec = ServiceSpec{...}`.
2. Add the spec to the `all` slice in `internal/services/registry.go`.

## Deployment

Cloud Build (`cloudbuild.yaml`) → Artifact Registry → Cloud Run `asia-northeast3` (project `workflow-knue`).
Secrets managed via Google Secret Manager.

> **Debug image:** distroless has no shell. Swap `gcr.io/distroless/static-debian12:nonroot`
> to `:debug-nonroot` for BusyBox-based debugging.

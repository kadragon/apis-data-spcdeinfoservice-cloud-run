# apis-data-spcdeinfoservice

API proxy server that forwards requests to Korean public data APIs ([data.go.kr](https://data.go.kr)), injecting service keys and handling CORS. Deployed to Google Cloud Run via Cloud Build.

## Proxy Routes

| Route | Description | Upstream API |
|---|---|---|
| `/SpcdeInfoService` | Korean special day info (holidays, anniversaries, 24 divisions) | `B090041/openapi/service/SpcdeInfoService` |
| `/GetSecuritiesProductInfoService` | Securities price info (ETF, ETN, ELW) | `1160100/service/GetSecuritiesProductInfoService` |
| `/BidPublicInfoService` | Public bid/procurement info (construction, service, goods, etc.) | `1230000/ad/BidPublicInfoService` |

## Getting Started

### Prerequisites

- Node.js (see `Dockerfile` — uses `node:slim`)
- npm

### Installation

```bash
npm install
```

### Environment Variables

| Variable | Required | Description |
|---|---|---|
| `AUTH_API_KEY` | Yes | Client authentication key, checked via `x-api-key` header |
| `DATAGOKR_SERVICEKEY` | Yes | Service key injected into proxied requests to data.go.kr |
| `PORT` | No | Server port (default: 3000, Dockerfile sets 8080) |

### Running

```bash
npm start
```

## Usage

All requests require an `x-api-key` header matching the configured `AUTH_API_KEY`.

```bash
# Local (default port 3000)
curl -H "x-api-key: YOUR_API_KEY" \
  "http://localhost:3000/SpcdeInfoService/getRestDeInfo?solYear=2026&solMonth=01"

# Docker (port 8080)
curl -H "x-api-key: YOUR_API_KEY" \
  "http://localhost:8080/SpcdeInfoService/getRestDeInfo?solYear=2026&solMonth=01"
```

Preflight `OPTIONS` requests are handled automatically (CORS).

## Development

```bash
npm test           # Run tests (Vitest)
npm run test:watch # Watch mode
npm run lint       # Lint & format check (Biome)
npm run lint:fix   # Auto-fix lint issues
```

## Adding a New Proxied API

1. Create a new file in `src/services/` following the existing pattern:
   - Define `baseUrl` and `allowedPaths`
   - Export a function that calls `createService(baseUrl, allowedPaths)`
2. Mount the new service in `src/index.js`

## Deployment

Deployed via Cloud Build (`cloudbuild.yaml`) to Google Cloud Run in `asia-northeast3`.

Secrets (`AUTH_API_KEY`, `DATAGOKR_SERVICEKEY`) are managed through Google Secret Manager.

## License

ISC

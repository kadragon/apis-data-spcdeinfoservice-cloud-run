# Runbook

## Quick Start

### Prerequisites

- Go 1.25 (`go version`)
- golangci-lint v2 (`golangci-lint --version`)
- lefthook (`lefthook --version`) — pre-commit hooks
- gcloud CLI (for deploy only)

### Setup

```bash
git clone https://github.com/kadragon/apis-data-spcdeinfoservice-cloud-run
cd apis-data-spcdeinfoservice-cloud-run
lefthook install                # install pre-commit hooks

# set required env vars
export AUTH_API_KEY="your-key"
export DATAGOKR_SERVICEKEY="your-servicekey"

go run ./cmd/server             # starts on :3000
```

### Verify

```bash
curl http://localhost:3000/health          # expect: 200 OK
curl -H "x-api-key: $AUTH_API_KEY" http://localhost:3000/SpcdeInfoService/getAnniversaryInfo
```

## Build

| Command | Purpose |
|---------|---------|
| `go build ./...` | Compile check |
| `go run ./cmd/server` | Run locally (PORT=3000) |
| `PORT=8080 go run ./cmd/server` | Run on specific port |

## Test

| Command | Purpose |
|---------|---------|
| `go test ./... -race` | Full test suite with race detector |
| `go test ./internal/proxy/... -race -v` | Proxy tests verbose |
| `go test ./internal/services/... -race -v` | Service registry tests verbose |

Test files: `*_test.go` alongside source. No external test DB needed — all tests are unit tests with httptest.

## Lint & Format

| Command | Purpose |
|---------|---------|
| `golangci-lint run ./...` | Full lint check |
| `goimports -w .` | Format + organize imports |
| `gofumpt -w .` | Strict Go formatting |

Pre-commit hook (lefthook) runs lint + test + build automatically on `git commit`.

## Deploy

### Environments

| Environment | URL | Branch | Deploy |
|-------------|-----|--------|--------|
| Production | Cloud Run `apis-data-spcdeinfoservice` (`asia-northeast3`) | `main` | Cloud Build trigger on push |

### Deploy Steps

Cloud Build triggers automatically on push to `main`. Manual trigger:

```bash
gcloud builds submit --config cloudbuild.yaml --project workflow-knue
```

Secrets injected by Cloud Build from Secret Manager: `AUTH_API_KEY`, `DATAGOKR_SERVICEKEY`.

### Debug Container

Swap base image in Dockerfile: `:nonroot` → `:debug-nonroot` for BusyBox shell access.

## Environment Variables

| Variable | Required | Description |
|----------|----------|-------------|
| `AUTH_API_KEY` | Yes | Client auth key, checked via `x-api-key` header |
| `DATAGOKR_SERVICEKEY` | Yes | Injected as query param into data.go.kr requests |
| `PORT` | No | HTTP listen port (default: 3000; Dockerfile sets 8080) |

Both secrets managed via Google Secret Manager (`workflow-knue` project).

## Common Failures

### `mustEnv: missing AUTH_API_KEY`

**Symptom:** Server exits immediately at startup with `missing required env var: AUTH_API_KEY`
**Cause:** Required env var not set
**Fix:** `export AUTH_API_KEY="<value>"` before running

### `golangci-lint: forcetypeassert`

**Symptom:** `forcetypeassert: type assertion must be checked`
**Fix:** Replace `x := val.(Type)` with `x, ok := val.(Type); if !ok { ... }`

### lefthook pre-commit fails

**Symptom:** `git commit` blocked by lint/test/build failure
**Fix:** Fix the reported issue. Do NOT use `--no-verify` to bypass. Investigate root cause.

## Sweep Trigger Policy

**Manual.** Run between features:

```bash
# No tools/sweep.sh yet — manual checks:
golangci-lint run ./...
go test ./... -race
grep -r "serviceKey\|AUTH_API_KEY" internal/ --include="*.go"  # verify no secrets in logs
```

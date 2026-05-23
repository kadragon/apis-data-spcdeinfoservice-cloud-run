# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

API proxy server (Go + gin) that forwards requests to Korean public data APIs (data.go.kr), injecting service keys and handling CORS. Deployed to Google Cloud Run via Cloud Build.

## Commands

- **Build:** `go build ./...`
- **Run:** `go run ./cmd/server`
- **Test:** `go test ./... -race`
- **Lint:** `golangci-lint run ./...`

## Architecture

Go 1.24 + gin. cmd/internal layout.

**Proxy flow:** gin route → `proxy.NewHandler` → `fetchWithRetry` (3 attempts, 10s timeout each, 1s/2s backoff on 5xx/network error) → stream upstream body via `io.Copy`.

**Key files:**

- `cmd/server/main.go` — gin engine, env validation (`mustEnv`), `/health` (public), `AuthMiddleware`, `RegisterAll`, graceful shutdown (SIGTERM/SIGINT, 5s timeout)
- `internal/proxy/handler.go` — `NewHandler(baseURL, upstreamPath, serviceKey)` gin.HandlerFunc; injects `serviceKey` query param, randomizes User-Agent, streams response
- `internal/proxy/retry.go` — `fetchWithRetry`: 3 attempts, 10s per-attempt ctx timeout, exponential backoff, retries on network error and 5xx, not on 4xx; returns `cancel` func for caller to defer
- `internal/proxy/auth.go` — `AuthMiddleware`: `OPTIONS` → 204, `x-api-key` constant-time compare (`crypto/subtle`), 401 on mismatch
- `internal/proxy/cors.go` — `writeCORS` helper, CORS headers constant
- `internal/proxy/useragent.go` — curated UA pool (20 strings) + `RandomUA()`
- `internal/services/registry.go` — `ServiceSpec` struct, `all` slice, `RegisterAll(r, serviceKey)` mounts each allowed path as a gin GET route
- `internal/services/*.go` — one `var <Name>Spec = ServiceSpec{...}` per upstream API

**Adding a new proxied API:**
1. Create `internal/services/<name>.go` with a `var <Name>Spec = ServiceSpec{MountPath, BaseURL, AllowedPaths}`.
2. Add the spec to the `all` slice in `internal/services/registry.go`.

## Environment Variables

- `AUTH_API_KEY` — Required. Client authentication key checked via `x-api-key` header.
- `DATAGOKR_SERVICEKEY` — Required. Injected into proxied requests to data.go.kr.
- `PORT` — Optional (default: 3000, Dockerfile sets 8080).

Both secrets managed via Google Secret Manager (see `cloudbuild.yaml`).

## Deployment

Cloud Build (`cloudbuild.yaml`): multi-stage Docker build (golang:1.24-alpine → distroless/static-debian12:nonroot), push to Artifact Registry (`asia-northeast3-docker.pkg.dev/workflow-knue/cloud-run-images`), deploy to Cloud Run in `asia-northeast3`. Project: `workflow-knue`.

Debug image: swap `:nonroot` to `:debug-nonroot` for BusyBox shell.

## Lint

`.golangci.yml` v2 config. Key linters: `bodyclose`, `noctx`, `errorlint`, `nilerr`, `forcetypeassert`. Formatters: `goimports`, `gofumpt`.

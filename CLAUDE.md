# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

API proxy server that forwards requests to Korean public data APIs (data.go.kr), injecting service keys and handling CORS. Deployed to Google Cloud Run via Cloud Build.

## Commands

- **Start server:** `npm start`
- **Lint:** `npm run lint` (Biome — linting + formatting)
- **Lint fix:** `npm run lint:fix`
- **Test:** `npm test` (Vitest)
- **Test watch:** `npm run test:watch`

## Architecture

ES modules (`"type": "module"` in package.json). Express 5 server with two proxy routes:

- `/SpcdeInfoService` — Korean special day info (holidays, anniversaries, 24 divisions)
- `/GetSecuritiesProductInfoService` — Securities price info (ETF, ETN, ELW)

**Key files:**

- `src/index.js` — Express app setup, API key auth middleware (`x-api-key` header, constant-time comparison), route mounting
- `src/common.js` — `createService(baseUrl, allowedPaths)` factory that creates proxy middleware: validates path against allowlist, appends `DATAGOKR_SERVICEKEY`, randomizes User-Agent, streams response back
- `src/services/*.js` — Each service defines its base URL and allowed endpoint paths, then calls `createService`

**Adding a new proxied API:** Create a new file in `src/services/` following the existing pattern (define baseUrl + allowedPaths, export a function calling `createService`), then mount it in `index.js`.

## Environment Variables

- `AUTH_API_KEY` — Required. Client authentication key checked via `x-api-key` header.
- `DATAGOKR_SERVICEKEY` — Required. Injected into proxied requests to data.go.kr.
- `PORT` — Optional (default: 3000, Dockerfile sets 8080).

Both secrets are managed via Google Secret Manager (see `cloudbuild.yaml`).

## Deployment

Cloud Build (`cloudbuild.yaml`): builds Docker image, pushes to GCR, deploys to Cloud Run in `asia-northeast3`. Project: `workflow-knue`.

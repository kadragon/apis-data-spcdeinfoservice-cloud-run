# Backlog

## Now

<!-- promote to [>] when sprint starts -->

## Next

- [ ] Add `tools/sweep.sh` for periodic harness checks
- [ ] Align retry budget with Cloud Run request timeout: `MaxRetries=2` + `ResponseHeaderTimeout=10s` + backoffs = ~33s exceeds `--timeout 30s`. Either lower Cloud Run timeout to ≥90s or cap retries to 1. See `internal/proxy/retry.go` + `cloudbuild.yaml`.

## Someday

- [ ] Add structured logging (slog) to replace `log.Printf`
- [ ] Add response caching layer for upstream API calls

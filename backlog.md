# Backlog

## Now

<!-- promote to [>] when sprint starts -->

## Next

- [x] Add `tools/sweep.sh` for periodic harness checks and integrated with lefthook pre-commit
- [x] Align retry budget with Cloud Run request timeout: capped `MaxRetries=1`, giving a worst-case header phase of ~21s (2×10s `ResponseHeaderTimeout` + 1s backoff), well below the `--timeout 30s` limit.

## Someday

- [x] Add structured logging (slog) to replace `log.Printf`
- [x] Add response caching layer for upstream API calls (done in ec5f2cc)

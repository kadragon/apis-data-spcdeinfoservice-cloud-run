# Backlog

## Now

<!-- promote to [>] when sprint starts -->

## Next

- [ ] Add `tools/sweep.sh` for periodic harness checks
- [x] Align retry budget with Cloud Run request timeout: capped `MaxRetries=1`, reducing maximum duration budget to ~11s, which is well below the `--timeout 30s` limit.

## Someday

- [ ] Add structured logging (slog) to replace `log.Printf`
- [ ] Add response caching layer for upstream API calls

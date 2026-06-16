# Tasks

## Deferred from PR #49 review

- [ ] **[P3] `TestMain` slog discard only covers `internal/proxy`** — `internal/cache` and `internal/services` still emit default-handler `slog.*` output under test. Add the same `TestMain` discard pattern to those packages (separate PR).

## Review Backlog

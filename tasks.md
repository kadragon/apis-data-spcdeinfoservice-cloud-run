# Tasks

## Deferred from PR #45 review

- [x] **[P3] `HandlerFactory` positional string args** (`internal/services/registry.go`) — Replaced three positional `string` params with `HandlerParams` struct (`BaseURL, Path, ServiceKey`) to prevent silent swaps.

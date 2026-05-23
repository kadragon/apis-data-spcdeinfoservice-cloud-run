# apis-data-spcdeinfoservice Agent Rules

Go 1.25 + gin API proxy forwarding requests to Korean public data APIs (data.go.kr), deployed to Cloud Run.

## Docs Index (read on demand)

| File | When to read |
|------|--------------|
| `docs/architecture.md` | Before modifying source structure or proxy flow |
| `docs/conventions.md` | Before writing new service specs or proxy logic |
| `docs/workflows.md` | When starting any implementation cycle |
| `docs/delegation.md` | Before delegating to sub-agents |
| `docs/eval-criteria.md` | When evaluating completed features |
| `docs/runbook.md` | For build, test, deploy commands and troubleshooting |

## Golden Principles

Mechanical enforcement noted per-rule. Violations should block commits (lefthook).

1. **Service keys never in logs or responses** — `DATAGOKR_SERVICEKEY` and `AUTH_API_KEY` must not appear in gin log output, error responses, or debug output. Convention only — no automated check yet.
2. **Auth middleware on all non-health routes** — Every route except `/health` must pass through `AuthMiddleware`. Verified by `auth_test.go` (unit); no integration-level registry enforcement yet.
3. **New service API = new file** — One `ServiceSpec` per file in `internal/services/`. Never add a second spec to an existing service file. Convention only — no automated check yet.
4. **4xx responses from upstream are never retried** — `fetchWithRetry` in `internal/proxy/retry.go` must not retry 4xx status. Enforced by `TestFetchWithRetry_4xxNoRetry` in `retry_test.go`.

## Delegation

Read `docs/delegation.md` for full routing table. Objective triggers only.

| Trigger (objective) | Delegate | Gate |
|---------------------|----------|------|
| Target module has >5 files or >300 LOC | Explore agent (sonnet) | Mandatory |
| Change touches ≥3 directories | Architecture analysis (opus) | Mandatory |
| Same failure x2 | `codex:rescue` | Escalation |

## Token Economy

1. Do not re-read a file already read this session. Check diff/region only.
2. Do not call tools to confirm information already in context.
3. Run independent tool calls in parallel.
4. Delegate analysis producing >20 lines of output to a sub-agent; return only conclusion.
5. Do not restate what the user just said.

## Working with Existing Code

- Adding a new proxied API: create `internal/services/<name>.go` with one `ServiceSpec`, then add to `all` slice in `registry.go`. See `docs/conventions.md`.
- `fetchWithRetry` signature: `(ctx, client, req) → (*http.Response, error)`. No cancel func returned; timeout is `ResponseHeaderTimeout=10s` on the transport.
- CORS headers are set in `writeCORS` — do not add headers in handlers.
- `RandomUA()` is called inside `NewHandler` — no need to pass UA explicitly.

## Language Policy

- Code, commits, docs: English
- User communication: Korean

## Maintenance

Update this file **only** when ALL of the following are true:

1. Information is not directly discoverable from code / config / manifests / docs
2. It is operationally significant — affects build, test, deploy, or runtime safety
3. It would likely cause mistakes if left undocumented
4. It is stable and not task-specific

**Never add:** architecture summaries, directory overviews, style conventions
already enforced by tooling, anything already visible in the repo, or
temporary / task-specific instructions.

Prefer modifying or removing outdated entries over appending. When unsure, add
a short inline `TODO:` comment rather than inventing guidance.

Size budget: target ≤100 lines, hard warn >200. Move long content to
`docs/*.md` and leave a pointer line here.

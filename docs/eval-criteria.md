# Evaluation Criteria

For this proxy service, evaluation is lightweight. No separate evaluator agent needed for routine `ServiceSpec` additions — the exit criterion is green tests + lint. Use full evaluation for proxy logic changes or auth changes.

## Criteria

### 1. Correctness (50%)

Proxied routes return upstream data correctly; auth blocks unauthenticated requests; `/health` passes unauthenticated.

| Score | Description |
|-------|-------------|
| 5 | All tests pass with `-race`; manual curl confirms proxy round-trip; auth rejects bad key |
| 4 | Tests pass; minor edge case not covered by test |
| 3 | Tests pass; manual verification pending |
| 2 | Tests fail or auth bypass possible |
| 1 | Panic, build failure, or data.go.kr service key exposed |

**How to test:** `go test ./... -race`. Then: `curl -H "x-api-key: $KEY" localhost:3000/<route>`.

### 2. Security Compliance (30%)

No secrets in logs, responses, or source. Auth enforced on all non-health routes.

| Score | Description |
|-------|-------------|
| 5 | `grep -r "DATAGOKR_SERVICEKEY\|AUTH_API_KEY" --include="*.go"` returns no hardcoded values; all routes except `/health` require auth |
| 3 | Secrets not hardcoded but present in a log statement |
| 1 | Secret value visible in any response body or log output |

**How to test:** `grep -r 'serviceKey\|AUTH_API_KEY' internal/ --include="*.go"` should show only parameter passing, never string literals with real values.

### 3. Convention Compliance (20%)

One `ServiceSpec` per file, `cancel` func deferred, `bodyclose` satisfied, no raw error strings.

| Score | Description |
|-------|-------------|
| 5 | `golangci-lint run ./...` exits 0; one spec per file; cancel deferred at call site |
| 3 | Lint passes but convention violations caught in review |
| 1 | Lint fails or two specs in same file |

**How to test:** `golangci-lint run ./...`.

## Sprint Contract Template

```markdown
### Sprint Contract: {Feature Name}

Generator proposes:
- I will build: {specific scope}
- Success looks like: `go test ./... -race` passes + {concrete check}
- Out of scope: {explicit exclusions}

Agreed criteria:
- [ ] `go test ./... -race` exits 0
- [ ] `golangci-lint run ./...` exits 0
- [ ] {feature-specific check}
```

## Pass Threshold

- All criteria ≥ 3
- Security criterion must be ≥ 5 (no exceptions for secret exposure)

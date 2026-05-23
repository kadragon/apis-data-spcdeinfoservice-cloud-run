# Workflows

Six workflows. Pick one per cycle. This is a small solo repo — most work uses `code` only.

## `plan` — Spec Generation

For non-trivial features only (new service categories, retry logic changes, auth changes).

1. Write `docs/design/{feature}.md`: what changes, why, acceptance criteria.
2. Review with user before writing code.
3. Add backlog items from approved spec.

Skip for: adding a new `ServiceSpec` (mechanical, no planning needed).

## `code` — Implementation

**Step 0: Branch** — never commit to `main`. Branch first: `git checkout -b <type>/<slug>`.

**Step 1: Scope check**
Check objective delegation triggers from `docs/delegation.md`:
- Target module >5 files or >300 LOC? → Explore agent (sonnet) before proceeding.
- Change touches ≥3 directories? → Architecture analysis (opus).
If none match, proceed directly.

**Step 2: Exit criterion** — state in one sentence what "done" looks like before writing code. Example: "`go test ./... -race` passes, new service appears in `/health` or routes."

**Step 3: Implement** — ≤2 files: implement directly. >2 files: delegate to subagent.

**Step 4: Verify** — run `go test ./... -race` + `golangci-lint run ./...`. Both must pass. The implementer runs this check; do NOT self-declare done without it.

**Step 5: PR** — lefthook pre-commit runs lint + test + build automatically on commit.

## `draft` — Documentation

Write or update `docs/`. Ground every claim in current code. Never modify production code during this workflow.

## `constrain` — Architectural Enforcement

1. Write structural test or lint rule first.
2. Run it.
3. If current code violates → add to `backlog.md`, don't fix here.
4. Update `docs/architecture.md`.

## `sweep` — Garbage Collection

Run between features:

```bash
bash tools/sweep.sh  # if created; otherwise manual checks below
```

Manual checks:
- `golangci-lint run ./...` — address any new findings
- Check `docs/` for stale content (proxy flow changed? service added?)
- Verify all `ServiceSpec` vars are in `all` slice in `registry.go`
- Harness freshness: is any constraint now enforced by the compiler that was previously manual?

## `explore` — Research

State question → research/prototype → report options and tradeoffs → **do not commit**. If approved, flows into `plan` or `code`.

---

## Context Anxiety

For multi-session work, write `handoff-{feature}.md` at the **start** of the session (when context is fresh), not when context is degraded. Delete on feature completion.

## Sweep Trigger Policy

**Manual** — run `bash tools/sweep.sh` between features. No automated hook (repo is small enough that manual cadence is sufficient). Record any sweep findings in `tasks.md`.

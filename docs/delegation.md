# Delegation

This is a small solo repo (~800 LOC). Most work is done in a single session without delegation. Delegate only when objective triggers fire.

## Pattern Selection

```
Q1. Does the task decompose into >1 genuinely parallel subtask?
    No  → single session. No delegation.
    Yes → Q2.
Q2. Is there a clear, verifiable exit criterion?
    Yes → spawn subagent with 4-field Spawn Prompt Contract.
    No  → clarify with user first.
```

## Spawn Prompt Contract (mandatory for all spawns)

Every subagent spawn must include all four fields:

```
- Objective: {what specifically should the subagent accomplish?}
- Output format: {diff / report / table / verdict — be concrete}
- Tools to use: {subset of tools to prioritize}
- Boundaries: {files/modules/workflows this spawn MUST NOT touch}
```

## Routing Table

### Mandatory Gates (blocking)

| Trigger (objective) | Delegate to | Model | Context to pass |
|---------------------|-------------|-------|-----------------|
| Target module has >5 files or >300 LOC | Explore agent | sonnet | Module path, `docs/architecture.md` |
| Change touches ≥3 directories | Architecture analysis agent | opus | Changed paths, `docs/architecture.md` |
| First edit in `internal/proxy/` this session | Explore agent | sonnet | Directory path, `docs/architecture.md` |

### Escalation

| Trigger | Delegate to | Notes |
|---------|-------------|-------|
| Same failure x2 | `codex:rescue` | Include: what was tried, exact error output, relevant files |
| Design decision needed | Design exploration agent (opus) | Blocking — don't proceed without decision |

### Background (non-blocking)

| Trigger | Action |
|---------|--------|
| Every commit | `/code-review` skill (background, optional) |

## Context Manifest

### Explore Agent

**Purpose:** Map a module before modifying it.

**Required context:**
- Module path (pass as prompt)
- `docs/architecture.md` — layer rules, dependency direction
- `docs/conventions.md` — naming and patterns

**Expected output:** File list with purpose of each file, entry points, dependencies.

### Architecture Analysis Agent

**Purpose:** Assess impact of cross-directory changes.

**Required context:**
- List of changed paths
- `docs/architecture.md`
- `docs/conventions.md`

**Expected output:** Impact report — which invariants are affected, what must stay consistent.

## Applying Sub-Agent Output

- **Structural fix** (typo, missing import) → apply in current cycle.
- **Behavioral change** (new feature, changed logic) → add to `backlog.md`. Never apply directly.
- **Contradicts design doc** → report both options to user. Do not choose.

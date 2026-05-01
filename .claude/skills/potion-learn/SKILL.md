---
name: potion-learn
description: "Learns from getprobo/probo PR reviews, free-form feedback, and codebase drift to evolve the Probo project guidelines under .claude/guidelines/. Mines human review comments from a specific PR (preferring the 3 senior reviewers identified in shared.md § 13), parses free-form text input (CodeRabbit exports, meeting notes), detects when the codebase has drifted from documented contrib/claude/ rules, challenges every finding with devil's-advocate reasoning, and stages learnings for reviewed merge into the multi-file guidelines tree (shared.md, go-backend/, typescript-frontend/). Use when the user asks to learn from a PR, absorb feedback, update guidelines from reviews, check for guideline drift, or evolve project conventions."
effort: high
argument-hint: "[--pr NUMBER] [--text \"...\"] [--file path] [--drift-only] [--merge]"
allowed-tools: Read, Write, Edit, Glob, Grep, Bash, Agent, AskUserQuestion
---

# Potion Learn — Probo Guidelines Evolution

Continuously evolve `.claude/guidelines/` (Probo's multi-file guidelines
tree) from getprobo/probo PR reviews, free-form feedback, and detected
drift between `contrib/claude/` docs and code reality.

## When to use this skill

- Learn conventions from a Probo PR's review comments (e.g.
  `--pr 1132`)
- Absorb team decisions, meeting notes, or CodeRabbit exports
- Check whether `contrib/claude/*.md` documented rules still match the
  code (drift detection)
- Evolve `.claude/guidelines/shared.md` and the per-stack guidelines
  (`go-backend/`, `typescript-frontend/`) without losing the
  `<!-- user-edited -->` Team Notes / Open Questions sections

## Prerequisites check

Before anything else, verify the multi-file guidelines exist:

```
Glob: .claude/guidelines/shared.md
Glob: .claude/guidelines/go-backend/index.md
Glob: .claude/guidelines/typescript-frontend/index.md
```

If any are missing, stop and tell the user:

> No multi-file guidelines found at `.claude/guidelines/`. Run the
> `potion-skill-generator` skill first to generate the initial Probo
> guidelines tree, then come back to evolve them with `potion-learn`.

## Parse input mode

Parse `$ARGUMENTS` to determine the mode:

| Flag | Mode | Example |
|------|------|---------|
| `--pr NUMBER` | PR mode | `--pr 1132` |
| `--text "..."` | Text mode (inline) | `--text "Always use baseurl"` |
| `--file PATH` | Text mode (from file) | `--file feedback.md` |
| `--drift-only` | Drift-only mode | `--drift-only` |
| `--merge` | Merge pending learnings into guidelines | `--merge` |
| (none) | Auto-detect PR from current branch (`gh pr view`) | — |

`--pr` and `--text`/`--file` can be combined (mixed mode). `--drift-only`
and `--merge` are exclusive.

**Merge mode** — skip Phase 1 + 2; jump to reading existing
`.claude/learnings.md` and presenting User Gate 2 (merge decision).

**Auto-detect** — if no flag and not `--drift-only`:
```bash
gh pr view --json number -q '.number' 2>/dev/null
```
If a PR is found, use it. Otherwise ask the user.

## Setup workspace

```bash
mkdir -p .skill-gen-workspace/learn
```

## Phase 1 — GATHER

Launch gather agents in parallel based on input mode. Drift detection
always runs. In drift-only mode, only the Drift Detector launches.

**Guidelines path resolution:** Probo uses the multi-file form. Pass
`.claude/guidelines/` (the directory) plus the three top-level entry
points (`shared.md`, `go-backend/index.md`, `typescript-frontend/index.md`)
to every agent so it knows the structure.

### PR Miner (if PR mode)

Launch the `pr-miner` agent:

```
Agent: pr-miner
Prompt: |
  Mine review comments from getprobo/probo PR #{pr_number}.

  Repo: getprobo/probo
  Senior reviewers (weighted higher confidence) — see shared.md § 13:
  the 3 senior reviewers identified in the PR-mining sample.
  Bots to skip: cubic-dev-ai, github-advanced-security.

  Domain hints — categorize each finding by the most likely target file:
  - SQL / coredata rules → go-backend/patterns.md or pitfalls.md
  - Service / Request+Validate / IAM → go-backend/patterns.md
  - Logging PII rules → shared.md § 8
  - Four-surface API rule → shared.md § 3
  - GraphQL nullability, Relay store, local TS types → typescript-frontend/patterns.md or shared.md § 13
  - Naming, license headers, struct tags → respective conventions.md
  - Worker patterns, FOR UPDATE SKIP LOCKED → go-backend/patterns.md
  - n8n action shape → typescript-frontend/patterns.md or module-notes/packages-n8n-node.md

  output_path: .skill-gen-workspace/learn/pr-findings.json
```

### Text Parser (if text mode)

```
Agent: text-parser
Prompt: |
  Parse the following text into convention findings for the Probo project.
  Use the same domain-hint mapping as the PR Miner.
  {text content or file_path}
  output_path: .skill-gen-workspace/learn/text-findings.json
```

### Drift Detector (always)

```
Agent: drift-detector
Prompt: |
  Check for drift between the Probo guidelines tree and the codebase.

  Guidelines tree:
  - .claude/guidelines/shared.md
  - .claude/guidelines/go-backend/{index,patterns,conventions,testing,pitfalls}.md
  - .claude/guidelines/go-backend/module-notes/*.md
  - .claude/guidelines/typescript-frontend/{index,patterns,conventions,testing,pitfalls}.md
  - .claude/guidelines/typescript-frontend/module-notes/*.md

  Authoritative source-of-truth: contrib/claude/*.md (28 docs).
  When the guidelines disagree with contrib/claude/, the doc wins —
  flag those as drift items targeting the guideline file.

  Known active drift to recheck (already documented in shared.md § 14):
  - pkg/probo/agent_run.go:472 hardcoded SQL `'PENDING'`
  - pkg/server/api/csp.go missing WithSSRFProtection()
  - OIDC error_description PII leak
  - apps/console/src/routes/ deprecated loaderFromQueryLoader / withQueryRef
  - contrib/claude/react-components.md "props for configuration, data from hooks"
  - No root .golangci.yml / eslint.config.* / tsconfig.json — confirm still true

  Also recheck by sampling: do recent files in pkg/coredata, pkg/probo,
  pkg/server/api/{console,trust,connect,mcp}/v1, apps/console/src/pages,
  packages/ui still match the documented patterns?

  output_path: .skill-gen-workspace/learn/drift-report.json
```

**Wait for all gather agents to complete.**

### Merge findings

If both PR and text findings exist, merge:
1. Read both JSON files
2. Re-number IDs sequentially (F-001, F-002, …) across both sources
3. Write merged result to `.skill-gen-workspace/learn/merged-findings.json`

If only one source, copy it as `merged-findings.json`.

## Phase 2 — CHALLENGE

```
Agent: challenger
Prompt: |
  Challenge each finding and drift item for the Probo project. Be rigorous.

  Devil's-advocate questions to ask of every finding:
  - Is this an enduring rule or a one-off PR comment? (frequency ≥ 2 across
    the PR-mining history elevates confidence — see shared.md § 13.)
  - Does it conflict with an existing contrib/claude/ doc?
  - Is it a stack-specific rule mistakenly proposed for shared.md?
  - Is it codifying drift instead of fixing the code? (e.g. proposing
    "inline SQL is fine in pkg/probo" would conflict with shared.md § 13 #1
    — reject.)
  - Does it touch a reviewer hot zone (pkg/iam/oauth2server, GraphQL
    nullability, agent/checkpoint, webhook payloads)? Hot-zone rules
    should not weaken existing constraints.

  findings_path: .skill-gen-workspace/learn/merged-findings.json
  drift_path: .skill-gen-workspace/learn/drift-report.json
  guidelines_path: .claude/guidelines/
  output_path: .skill-gen-workspace/learn/challenged-findings.json
```

**Wait for the challenger to complete.**

## User Gate 1 — Review challenged findings

Read challenged findings and present a summary table:

```
## Learning Summary

Source: {sources used}

| ID | Convention | Source | Verdict | Confidence | Target |
|----|-----------|--------|---------|------------|--------|
| F-001 | {convention} | PR #1132 | accept | high | shared.md § 13 |
| ... | ... | ... | ... | ... | ... |
| D-001 | {claim} drift | drift scan | accept | high | go-backend/pitfalls.md |

Accepted: N | Modified: N | Rejected: N

Review each finding? [y/n/abort]
```

Use `AskUserQuestion`:
- **y** — show each finding one by one, allow override
- **n** — accept all verdicts as-is
- **abort** — stop, no changes

If overrides happen, update the challenged findings JSON before
proceeding.

## Phase 3 — WRITE

Write learnings into `.claude/learnings.md` for staging. The
`learnings-writer` agent must respect Probo's multi-file structure:

```
Agent: learnings-writer
Prompt: |
  Stage challenged findings into .claude/learnings.md for the Probo
  project. The guidelines are multi-file:

  - .claude/guidelines/shared.md (cross-cutting)
  - .claude/guidelines/go-backend/{index,patterns,conventions,testing,pitfalls}.md
  - .claude/guidelines/go-backend/module-notes/*.md
  - .claude/guidelines/typescript-frontend/{index,patterns,conventions,testing,pitfalls}.md
  - .claude/guidelines/typescript-frontend/module-notes/*.md

  Each pending learning must specify a target_section that names the
  exact file + section within the tree. Examples:
  - "shared.md § 13 — Code-Review-Enforced Standards"
  - "go-backend/patterns.md § 1 — Service / TenantService"
  - "typescript-frontend/pitfalls.md — new entry"
  - "go-backend/module-notes/coredata.md — top pitfalls"

  Preserve all <!-- user-edited --> blocks (Team Notes, Open Questions).

  challenged_path: .skill-gen-workspace/learn/challenged-findings.json
  guidelines_path: .claude/guidelines/
  learnings_path: .claude/learnings.md
  archive_dir: .claude/learnings-archive
  mode: write
```

**Wait for the writer to complete.**

## User Gate 2 — Merge decision

```
## Merge into Guidelines?

{N} learnings staged in .claude/learnings.md
{N} drift alerts staged

Merge candidates:
- [L-001] → shared.md § 13 ({summary})
- [L-002] → go-backend/patterns.md § 4 ({summary})
- [D-001] → go-backend/pitfalls.md ({summary})

Options:
1. Merge all into guidelines now
2. Select which to merge
3. Keep in learnings only (merge later with --merge)
```

Use `AskUserQuestion` for the user's choice.

### If merge approved (option 1 or 2)

Determine which IDs to merge.

```
Agent: learnings-writer
Prompt: |
  Merge approved learnings into the Probo multi-file guidelines tree.

  Each learning's target_section names the exact file + section. Insert
  into that file at the right place. For new pitfalls, append to the
  numbered list in pitfalls.md and renumber. For new module-notes
  entries, edit the module-note file directly. Never edit content
  inside <!-- user-edited --> blocks.

  challenged_path: .skill-gen-workspace/learn/challenged-findings.json
  guidelines_path: .claude/guidelines/
  learnings_path: .claude/learnings.md
  archive_dir: .claude/learnings-archive
  mode: merge
  merge_ids: [{selected IDs}]
```

After merge, archive the consumed learnings with a timestamp under
`.claude/learnings-archive/{YYYY-MM-DD}/`.

### If kept for later (option 3)

> Learnings staged in `.claude/learnings.md`. Run `potion-learn --merge`
> later to merge selected items into `.claude/guidelines/`.

## Done

Summarize:
- How many findings gathered (by source: PR, text, drift)
- How many survived challenge
- How many were staged / merged / rejected
- Any drift alerts found
- Per-stack breakdown (Go backend vs TypeScript frontend vs shared)

## Probo-specific notes for the agents

- **`contrib/claude/` is authoritative.** Whenever a guideline file
  disagrees with a `contrib/claude/*.md` doc, the doc wins. Generate a
  drift item targeting the guideline file, not the doc.
- **PII rule is non-negotiable.** Findings that propose logging emails,
  names, IPs, or `error_description` verbatim should be rejected by the
  challenger.
- **Four-surface rule.** Findings that propose adding a backend operation
  without all four surfaces (GraphQL ↔ MCP ↔ CLI ↔ n8n) should be
  flagged as a drift item, not codified.
- **Free-form commits.** Findings that propose Conventional Commits
  prefixes (`feat:`, `fix:`) should be rejected — Probo explicitly does
  not use them (`shared.md` § 5).
- **No `Co-Authored-By` for AI.** Findings that suggest adding such
  trailers to commits should be rejected.
- **Two-environment Relay split.** Findings about the frontend that
  ignore the core/iam Vite/Babel boundary are likely incorrect.

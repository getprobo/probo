---
name: access-review
description: Run a semi-automated Probo access review campaign. Use when the user wants to review access entries, decide approve/revoke/escalate, or resume an in-progress campaign with MCP and .probo/access-reviews/ notes.
compatibility: Requires Probo MCP (OAuth 2.0) and file write access for .probo/access-reviews/
---

# Access review

Run a **semi-automated** access review for campaign `$ARGUMENTS` (or ask the
user for the campaign name). Review entries only — do not create, start,
cancel, or close campaigns.

Before executing, read these files **relative to this skill directory**:

- `references/mcp-tools.md` — MCP tool names, inputs, pagination
- `references/decision-rubric.md` — semi-auto decision rules
- `references/notes-format.md` — working memory file schema

## Preconditions

1. Probo MCP must be connected. If tools fail with auth errors, stop and tell
   the user to complete OAuth sign-in for the Probo MCP server in their agent
   (Claude Code: `/mcp` or `claude mcp login probo`; Codex: `codex mcp login
   probo`; OpenCode/Cursor: configure MCP in settings then authenticate).
2. Resolve the campaign from `$ARGUMENTS` (name match or GID). If ambiguous,
   list `listAccessReviewCampaigns` results and ask the user to pick one.
3. Campaign `status` must be `IN_PROGRESS` or `PENDING_ACTIONS`. Stop with a
   clear message for `DRAFT`, `COMPLETED`, or `CANCELLED`.

## Working notes file

Create or resume `.probo/access-reviews/<campaign-slug>.md` per
`references/notes-format.md`. Create `.probo/access-reviews/` if missing.

## Workflow

### 1. Orient

- Call `getAccessReviewStatistics` for the campaign.
- Summarize totals and pending count for the user.
- If no pending entries, report completion and stop.

### 2. Fetch batch

- Call `listAccessEntries` with `campaign_id`, `filter.decision: PENDING`,
  `size: 50`.
- Use `last_cursor` from the notes file when resuming.

### 3. Classify each entry

Apply `references/decision-rubric.md`:

| Class | Action |
| --- | --- |
| **Auto** | Queue for `recordAccessReviewEntryDecisions` |
| **Ambiguous** | Present to user; do not write yet |
| **Skip** | Log in notes only |

Append each auto decision to the notes file before writing.

### 4. Write auto decisions

- Batch via `recordAccessReviewEntryDecisions` when possible.
- Non-`APPROVED` decisions **must** include `decision_note`.
- On MCP error, stop and do not advance `last_cursor`.

### 5. Present ambiguous entries

Show email, roles, flags, proposed decision, rationale. Record only after
explicit user confirmation.

### 6. Checkpoint

Update notes: `last_cursor`, session log, `updated_at`. Ask to continue if
`next_cursor` is set.

## Hard rules

- Never call `closeAccessReviewCampaign` or campaign setup mutations unless
  the user explicitly requests setup work outside this skill.
- Never invent entry IDs or decisions — use MCP responses only.
- Never record non-`APPROVED` without `decision_note`.

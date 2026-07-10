---
name: missing-signatures
description: Report who is missing document signatures or quorum approvals in Probo. Use when the user wants a per-person recap of unsigned signature requests, pending approval decisions, or document signing compliance status.
compatibility: Requires Probo MCP (OAuth 2.0) and file write access for .probo/missing-signatures/
---

# Missing signatures report

Build a **read-only, per-person recap** of outstanding document obligations for
organization `$ARGUMENTS` (or ask the user for the organization name). Cover:

1. **Signature requests** — `REQUESTED` but not `SIGNED` on a published version
2. **Quorum approvals** — `PENDING` decisions on a `PENDING` approval quorum
   while the version is `PENDING_APPROVAL`

Before executing, read these files **relative to this skill directory**:

- `references/mcp-tools.md` — MCP tool names, inputs, pagination
- `references/report-format.md` — per-person recap and summary layout
- `references/notes-format.md` — working memory for resumable scans

## Preconditions

1. Probo MCP must be connected. If tools fail with auth errors, stop and tell
   the user to complete OAuth sign-in for the Probo MCP server in their agent
   (Claude Code: `/mcp` or `claude mcp login probo`; Codex: `codex mcp login
   probo`; OpenCode/Cursor: configure MCP in settings then authenticate).
2. Resolve the organization from `$ARGUMENTS` (name match or GID). If ambiguous,
   call `listOrganizations` and ask the user to pick one.
3. This skill is **reporting only**. Do not request signatures, cancel requests,
   publish documents, or submit approval decisions unless the user explicitly
   asks for follow-up actions outside this report.

## Working notes file

Create or resume `.probo/missing-signatures/<org-slug>.md` per
`references/notes-format.md`. Create `.probo/missing-signatures/` if missing.

## Workflow

### 1. Orient

- Record `organization_id` in the notes file.
- If resuming, read `last_document_cursor` and cached `profile_cache` from
  notes.
- Tell the user whether this is a fresh scan or a resume.

### 2. Scan documents (paginated)

For each page from `listDocuments`:

- Skip `ARCHIVED` documents.
- Call `listDocumentVersions` for the document.
- For each version, branch on `status`:

| Version status | What to check |
| --- | --- |
| `PUBLISHED` | `listDocumentVersionSignatures` with `filter.states: ["REQUESTED"]` |
| `PENDING_APPROVAL` | `listDocumentVersionApprovalQuorums`; for each quorum with `status: PENDING`, `listDocumentVersionApprovalDecisions` with `filter.states: ["PENDING"]` |

Prefer the **current published** version (`current_published_major` /
`current_published_minor` on the document) when multiple published minors exist.
Still scan all `PUBLISHED` versions if the user asked for exhaustive coverage.

Paginate every list call. Store `last_document_cursor` after each document page
so a large org can be resumed.

### 3. Resolve people

Collect unique `signed_by` and `approver_id` profile GIDs. Resolve each once via
`getUser` and cache `full_name` and `email_address` in the notes file. Never
invent names.

### 4. Build per-person recap

Aggregate findings by profile. Each item is either:

| Type | Meaning |
| --- | --- |
| `signature` | Signature requested, not yet signed |
| `approval` | Quorum approval decision still pending |

Use `references/report-format.md` for the output layout. Sort people by
`full_name`, then email.

### 5. Present summary

Show:

- Total people with at least one outstanding item
- Count by type (signatures vs approvals)
- Count by document
- The per-person recap tables

If the scan is incomplete (more document pages), say so and offer to continue.

### 6. Checkpoint

Update notes: `last_document_cursor`, `profile_cache`, session log, `updated_at`.
Clear `last_document_cursor` when the document pagination is exhausted.

## Hard rules

- Never call write mutations (`requestDocumentVersionSignature`,
  `cancelSignatureRequest`, `publishDocument`, `voidDocumentVersionApproval`)
  unless the user explicitly requests action after reviewing the report.
- Never invent profile IDs, document titles, or states — use MCP responses only.
- Include document title, version (`major.minor`), and `requested_at` or quorum
  `created_at` on every line item so the report is audit-ready.

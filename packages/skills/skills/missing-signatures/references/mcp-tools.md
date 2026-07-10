# Missing signatures MCP tools

All tools are on the Probo MCP server (`probo`). Read each tool schema before
calling.

## Organization scope

### `listOrganizations`

List organizations the caller can access. Use to resolve `$ARGUMENTS` when the
user provides a name instead of a GID.

### `listUsers`

Resolve profile display fields. `users[]` entries are `Profile` objects.

| Field | Usage |
| --- | --- |
| `organization_id` | Organization GID |
| `size` | Page size; use `100` when prefetching is helpful |
| `cursor` | Pagination |

Prefer `getUser` for individual lookups when building the recap; use `listUsers`
only when bulk prefetch is faster.

### `getUser`

Required: `id` (profile GID)

Returns `user` with `full_name`, `email_address`, `state`. Cache results in the
notes file to avoid repeat calls.

## Documents and versions

### `listDocuments`

Primary iterator for the scan.

| Field | Usage |
| --- | --- |
| `organization_id` | Required |
| `size` | Use `25` per page |
| `cursor` | Store in notes as `last_document_cursor` |
| `filter.status` | Optional — omit archived docs in post-processing or pass active-only if supported |

Returns `documents[]` and `next_cursor`. Each document includes
`current_published_major`, `current_published_minor`, `status`.

### `getDocument`

Use when you need fresh `current_published_*` fields for one document.

### `listDocumentVersions`

Required: `document_id`

| Field | Usage |
| --- | --- |
| `filter.statuses` | `["PUBLISHED"]` or `["PENDING_APPROVAL"]` to narrow |
| `size` / `cursor` | Paginate when a document has many versions |

Returns `document_versions[]` with `title`, `major`, `minor`, `status`.

### `getDocumentVersion`

Use when you need a single version's metadata without listing all versions.

## Signature gaps (published versions)

### `listDocumentVersionSignatures`

Required: `document_version_id`

| Field | Usage |
| --- | --- |
| `filter.states` | `["REQUESTED"]` for outstanding signature requests |
| `size` / `cursor` | Paginate |

Returns `document_version_signatures[]`. Key fields:

| Field | Report use |
| --- | --- |
| `signed_by` | Profile GID — group by person |
| `state` | `REQUESTED` = missing signature |
| `requested_at` | When the request was sent |
| `signed_at` | `null` while outstanding |

### `getDocumentVersionSignature`

Use for detail on a single signature row if needed.

## Quorum approval gaps (pending approval versions)

### `listDocumentVersionApprovalQuorums`

Required: `document_version_id`

Returns `approval_quorums[]`. Only process quorums with `status: PENDING`.

### `getDocumentVersionApprovalQuorum`

Use when you need quorum metadata for one version.

### `listDocumentVersionApprovalDecisions`

Required: `quorum_id`

| Field | Usage |
| --- | --- |
| `filter.states` | `["PENDING"]` for outstanding approvers |
| `size` / `cursor` | Paginate |

Returns `approval_decisions[]`. Key fields:

| Field | Report use |
| --- | --- |
| `approver_id` | Profile GID — group by person |
| `state` | `PENDING` = missing approval |
| `decided_at` | `null` while outstanding |
| `comment` | Include if present |

### `getDocumentVersionApprovalDecision`

Use for detail on a single decision row if needed.

## Out of scope for this skill

Do not call unless the user explicitly asks after reviewing the report:

- `requestDocumentVersionSignature`
- `cancelSignatureRequest`
- `publishDocument`
- `voidDocumentVersionApproval`

Signing and approving are GraphQL-only today (`signDocument`,
`approveDocumentVersion`, `rejectDocumentVersion`) — not available via MCP.

## Pagination and resume

1. Outer loop: `listDocuments` — persist `last_document_cursor` in notes.
2. Inner loops: paginate versions, signatures, quorums, and decisions.
3. When `listDocuments` returns no `next_cursor`, clear `last_document_cursor`
   and mark the scan complete in the session log.

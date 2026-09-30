# Document MCP tools

Read each tool schema before calling it. All tools are on the connected Probo
MCP server (`probo-us`, `probo-eu`, or a self-hosted server).

## Find

### `listDocuments`

| Field | Usage |
| --- | --- |
| `organization_id` | Required |
| `filter.query` | Search the latest version title and body |
| `filter.document_types` | For example `["POLICY"]` |
| `filter.published` | `true` when a published version must exist |
| `filter.status` | Omit to get `ACTIVE` only. `["ARCHIVED"]` returns archived documents only; `["ACTIVE", "ARCHIVED"]` returns both |
| `size` / `cursor` | Paginate |

Each `documents[]` entry has `id`, `title`, `document_type`, `status`,
`current_published_major`, and `current_published_minor`. No markdown.

### `getDocument`

Required: `id` (document id). Same metadata as a list entry, for one document.
Call `readDocument` for the text.

## Read and edit

### `readDocument`

Required: `document_id` (document id, never a version id) and `version`.

| `version` | Returns |
| --- | --- |
| `LATEST` | Tip revision, the one `updateDocument` edits |
| `PUBLISHED` | The revision at `current_published_major` and `current_published_minor`. Fails when nothing is published |
| `DRAFT` | Tip revision only when its status is `DRAFT` |

`document_version.content` is markdown. Keep `document_version.id` when a later
call needs a version id.

### `updateDocument`

Required: `id` (document id). Optional `title`, `content` (markdown),
`classification`, `document_type`, `default_approver_ids`. Writes the draft.

### `addDocument`

Creates a document and its first draft. Required: `organization_id`, `title`,
`content` (markdown), `classification`, `document_type`.

### `publishDocument`

Required: `document_id`, `minor` (boolean), `changelog`. No `approver_ids`
field. Major publishes (`minor: false`) email `default_approver_ids` from
`updateDocument` when that list is set.

### `deleteDocumentDraft`

Required: `id` (document id). Discards the latest draft. The document stays.
Cannot delete the initial v0.1 draft.

### `deleteDocument`

Required: `document_id`. Deletes the document and every version.

### `archiveDocument` / `unarchiveDocument`

Required: `id` (document id).

## Version id only

Use these only when you already have `document_version.id`.

| Tool | Id field |
| --- | --- |
| `getDocumentVersion` | `id` |
| `listDocumentVersions` | `document_id` is still the document id; the results are versions |
| `listDocumentVersionSignatures` | `document_version_id` |
| `requestDocumentVersionSignature` | `document_version_id` |
| `signDocument` | `document_version_id` |
| `approveDocumentVersion` | `document_version_id` |
| `rejectDocumentVersion` | `document_version_id` |
| `voidDocumentVersionApproval` | `document_version_id` |
| `cancelSignatureRequest` | `document_version_signature_id` |

`listDocumentVersions` is for scanning every revision. To read one known
revision, call `readDocument` instead.

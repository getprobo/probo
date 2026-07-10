# Missing signatures notes file

Path: `.probo/missing-signatures/<org-slug>.md`

`<org-slug>` — lowercase organization name with non-alphanumerics replaced by
hyphens (e.g. `Acme Corp` → `acme-corp`).

## Template

```markdown
# Missing signatures: <organization name>

organization_id: <gid>
last_document_cursor:
scan_status: <in_progress|complete>
updated_at: <ISO-8601 UTC>

## Profile cache

| profile_id | full_name | email_address |
| --- | --- | --- |
| gid://… | Jane Doe | jane@example.com |

## Session log

- <ISO-8601> — Started scan.
- <ISO-8601> — Processed document page. Cursor: <cursor or done>.
- <ISO-8601> — Scan complete. <n> people, <s> signatures, <a> approvals.

## Findings (raw)

| profile_id | type | document_id | version_id | title | version | since | record_id |
| --- | --- | --- | --- | --- | --- | --- | --- |
```

## Field rules

| Field | Rule |
| --- | --- |
| `last_document_cursor` | Empty on fresh run. Set to `listDocuments` `next_cursor` after each document page. Clear when pagination is done. |
| `scan_status` | `in_progress` until all document pages are processed |
| `updated_at` | Update on every file write |
| Profile cache | Append rows as `getUser` resolves profiles; reuse on resume |
| Findings table | Append rows as gaps are discovered; dedupe by `record_id` on resume |

## Resume behavior

1. If the file exists, read `organization_id`, `last_document_cursor`, profile
   cache, and findings.
2. Confirm with the user that resuming the same organization is intended.
3. Continue `listDocuments` from `last_document_cursor` if set.
4. Do not duplicate findings rows with the same `record_id`.

## Git

Do not commit or push this file unless the user asks. It is working memory for
the reporting session.

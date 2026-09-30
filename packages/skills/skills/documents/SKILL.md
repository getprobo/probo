---
name: documents
description: >-
  Read, edit, and publish Probo documents without mixing a document id up with a
  document version id. Use when the user wants to find a policy or other document,
  read its markdown, change a draft, publish a major or minor version, archive a
  document, discard a draft, or see who must sign or approve a revision.
compatibility: Requires Probo MCP (OAuth 2.0).
---

# Probo documents

A **document id** and a **document version id** are different. Passing one where
the other belongs is the usual failure.

Before calling tools, read `references/mcp-tools.md` in this skill directory.

## Which id

| You have | Use it as |
| --- | --- |
| Document id from `listDocuments` or `getDocument` | `id` on `getDocument`, `updateDocument`, `deleteDocumentDraft`, `archiveDocument`, `unarchiveDocument`. `document_id` on `readDocument`, `listDocumentVersions`, `publishDocument`, `deleteDocument`. |
| Document version id from `readDocument` (`document_version.id`) or `listDocumentVersions` | `id` on `getDocumentVersion`. `document_version_id` on signature, approval, and `signDocument` tools. |

`listDocuments` and `getDocument` return metadata, including `title` and
`document_type` from the latest version. They do not return markdown.

## Workflow

1. **Find the server and organization.** The plugin ships `probo-us` and
   `probo-eu`. Call `listOrganizations` on each connected server until the named
   company appears, then stay on that server.
2. **Find the document.** `listDocuments` with `organization_id`. Search with
   `filter.query` (latest title and body) or `filter.document_types` (for example
   `["POLICY"]`). Use `filter.published: true` when only published documents
   matter. Read `title` on each result. Do not invent a title.
3. **Read the text.** `readDocument` with that document id and `version`:
   - `PUBLISHED` — the current published revision
   - `LATEST` — the tip, which is what `updateDocument` edits
   - `DRAFT` — the tip only when it is still a draft
4. **Edit.** `updateDocument` with the document id in `id` and markdown in
   `content`. This writes the draft. It does not publish. Set
   `default_approver_ids` here when a later major publish should email approvers.
5. **Publish.** `publishDocument` with `document_id`, `minor`, and `changelog`.
   `minor: true` publishes a minor version and needs an existing published major.
   `minor: false` publishes a new major, or emails the default approvers when
   those are set. There is no `approver_ids` argument.
6. **Sign or approve.** Use the version id from `document_version.id`.
   `approveDocumentVersion` is an approver's vote on a revision already pending
   approval. It is not how you request that approval.

`deleteDocumentDraft` drops the unpublished draft and keeps the document.
`deleteDocument` deletes the document and every version.
`deleteCompliancePortalDocument` only removes a portal link.

For a session that only works on documents, connect to the same MCP URL with
`?toolset=documents`. The OAuth resource stays `/api/mcp/v1`.

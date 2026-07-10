# Access review notes file

Path: `.probo/access-reviews/<campaign-slug>.md`

`<campaign-slug>` — lowercase campaign name with non-alphanumerics replaced by
hyphens (e.g. `Q3 GitHub Review` → `q3-github-review`).

## Template

```markdown
# Access review: <campaign name>

campaign_id: <gid>
organization_id: <gid>
campaign_status: <IN_PROGRESS|PENDING_ACTIONS>
last_cursor:
updated_at: <ISO-8601 UTC>

## Session log

- <ISO-8601> — Started review. Pending: <n>.
- <ISO-8601> — Batch complete. Auto: <a> approved, <r> revoked, <e> escalated, <d> deferred. Ambiguous: <u>. Cursor: <cursor or done>.

## Entry notes

| entry_id | email | decision | auto | rationale |
|----------|-------|----------|------|-----------|
| gid://… | user@example.com | REVOKE | yes | Terminated user flag |

## Ambiguous (awaiting user)

| entry_id | email | flags | suggested | question |
|----------|-------|-------|-----------|----------|
| gid://… | admin@example.com | NEW, PRIVILEGED_ACCESS | ESCALATE | New admin — approve or revoke? |
```

## Field rules

| Field | Rule |
| --- | --- |
| `last_cursor` | Empty on fresh run. Set to `listAccessEntries` `next_cursor` after each successful batch. Clear when null (pagination done). |
| `updated_at` | Update on every file write |
| `auto` column | `yes` if written via semi-auto rubric; `no` if user confirmed |
| Session log | Append-only; one line per batch or major event |
| Ambiguous table | Remove rows after user confirms and decision is recorded |

## Resume behavior

1. If the file exists, read `campaign_id`, `last_cursor`, and ambiguous rows.
2. Confirm with the user that resuming the same campaign is intended.
3. Continue `listAccessEntries` from `last_cursor` if set; otherwise start from
   the first pending page.
4. Do not duplicate entry notes for IDs already in the table with a final
   decision.

## Git

Do not commit or push this file unless the user asks. It is working memory for
the review session.

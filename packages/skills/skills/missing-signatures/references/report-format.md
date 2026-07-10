# Missing signatures report format

Present the report in this order: executive summary, per-person recap, optional
document index.

## Executive summary

```markdown
# Missing signatures: <organization name>

organization_id: <gid>
generated_at: <ISO-8601 UTC>
scan_status: <complete|in_progress>

## Summary

| Metric | Count |
| --- | --- |
| People with outstanding items | <n> |
| Outstanding signature requests | <n> |
| Pending quorum approvals | <n> |
| Documents affected | <n> |

<One paragraph in plain language for auditors or managers.>
```

## Per-person recap (primary view)

Sort alphabetically by `full_name`. One section per person with outstanding
items. Omit people with zero items.

```markdown
## <full_name> (<email_address>)

profile_id: <gid>
profile_state: <ACTIVE|INACTIVE>

| Type | Document | Version | Since | ID |
| --- | --- | --- | --- | --- |
| signature | Information Security Policy | 2.1 | 2026-06-01 | <signature gid> |
| approval | Data Retention Policy | 3.0 | 2026-06-10 | <decision gid> |
```

### Column rules

| Column | Rule |
| --- | --- |
| Type | `signature` or `approval` |
| Document | `document_versions[].title` from MCP |
| Version | `major.minor` (e.g. `2.1`) |
| Since | `requested_at` for signatures; quorum `created_at` or decision `created_at` for approvals (ISO date, UTC) |
| ID | `document_version_signature.id` or `approval_decision.id` for traceability |

### Person with multiple items

Keep one table per person. Do not split signatures and approvals into separate
top-level sections — the per-person view is the reporting unit.

## Document index (optional appendix)

When the user wants a document-centric view, add after the per-person recap:

```markdown
## By document

### <document title> (<document_id>)

| Person | Type | Version | Since |
| --- | --- | --- | --- |
| Jane Doe | signature | 2.1 | 2026-06-01 |
```

## Empty result

If no outstanding signatures or approvals exist:

```markdown
No outstanding signature requests or pending quorum approvals were found for
<organization name> as of <generated_at>.
```

Still write the notes file with `scan_status: complete`.

## Presentation tips

- Lead with the summary table, then the per-person recap.
- Flag `INACTIVE` profiles with outstanding items — they may need escalation.
- For `PENDING_APPROVAL` versions, note that publication is blocked until the
  quorum resolves (unanimous approval required).

# Access review MCP tools

All tools are on a Probo MCP server (`probo-us`, `probo-eu`, or a self-hosted
server configured in the agent). Read each tool schema before calling.

## Organization scope

### `listOrganizations`

List organizations the caller can access on the server it is called on. No
required input, so it serves two distinct purposes:

1. **Choosing a server** when the region is unknown — call it on each connected
   server. A non-empty result does not identify the region, since the caller can
   have organizations on both. Only a unique match for the organization the user
   named settles it; otherwise show each server with its organizations and ask.
2. **Choosing an organization** on a server that is already settled — including
   a single self-hosted instance serving several organizations. Take the unique
   name match, or the only organization when the server returns one, and
   otherwise list them and ask.

Either way the organization must be resolved before `listAccessReviewCampaigns`,
which requires `organization_id`.

## Read

### `listAccessReviewCampaigns`

List campaigns for an organization. Use to resolve `$ARGUMENTS` to a campaign
when the user provides a name instead of a GID.

Required: `organization_id` — resolve the organization before calling this.

### `listAccessEntries`

List entries for a campaign. Primary data source for this command.

| Field | Usage |
| --- | --- |
| `campaign_id` | Campaign GID |
| `filter.decision` | Use `PENDING` for review batches |
| `filter.flag` | Optional — focus on a flag (e.g. `TERMINATED_USER`) |
| `filter.incremental_tag` | Optional — `NEW`, `REMOVED`, `UNCHANGED` |
| `filter.is_admin` | Optional boolean |
| `filter.active` | Optional boolean |
| `size` | Page size; use `50` per batch |
| `cursor` | Resume pagination; store in notes file as `last_cursor` |

Returns `entries[]` and `next_cursor`. Each entry includes `source_name` and
nullable `connector_id`; use these fields to identify the source tool. Never
infer the tool from account IDs, roles, email patterns, or other entry data.

### `getAccessReviewStatistics`

Required: `campaign_id`

Returns `statistics` with `total_count`, `decision_counts`, `flag_counts`,
`incremental_tag_counts`. Call at the start of each run and after large batches.

## Write (semi-auto command)

### `recordAccessReviewEntryDecisions`

Preferred for auto batch. Input `decisions[]` with:

- `access_review_entry_id` (required)
- `decision` — `APPROVED`, `REVOKE`, `DEFER`, `ESCALATE` (not `PENDING`)
- `decision_note` — required for non-`APPROVED`

### `recordAccessReviewEntryDecision`

Use for single entries after user confirms an ambiguous case.

### `flagAccessReviewEntry`

Optional when the user agrees a flag is missing. Input:

- `access_review_entry_id`
- `flags[]` — see rubric for valid values
- `flag_reasons[]` — optional strings

## Out of scope for this command

Do not call unless the user explicitly asks for campaign setup:

- `closeAccessReviewCampaign`
- `cancelAccessReviewCampaign`
- `startAccessReviewCampaign`
- `createAccessReviewCampaign`
- `createAccessReviewSource` / source mutations

## Entry fields (review signals)

| Field | Review use |
| --- | --- |
| `source_name`, `connector_id` | Authoritative source tool; do not infer it from account data |
| `email`, `full_name` | Identity |
| `roles`, `job_title` | Access level |
| `is_admin` | Heightened scrutiny |
| `active` | `false` often supports revoke |
| `mfa_status` | `DISABLED` on privileged access → escalate |
| `auth_method` | `SSO`, `PASSWORD`, `API_KEY`, `OAUTH2`, `SSH`, `SERVICE_ACCOUNT` context |
| `account_type` | `SERVICE_ACCOUNT` vs `USER` |
| `last_login` | Dormancy signal |
| `incremental_tag` | `NEW` needs extra scrutiny |
| `flags`, `flag_reasons` | Primary risk signals |
| `decision` | Target `PENDING` entries only |

## Pagination and resume

1. Read `last_cursor` from the notes file.
2. Pass it to `listAccessEntries` to continue where the last batch stopped.
3. Write the new `next_cursor` back after each successful batch.
4. When `next_cursor` is null/empty, the pending page is exhausted — refresh
   statistics to confirm remaining `PENDING` count.

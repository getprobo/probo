# Portal mechanics: creating and updating commitments via the Probo MCP

Read this before making any write call. All tool names below are Probo MCP tools; the server prefix
varies by environment (there may be more than one Probo server, e.g. US and EU). Use the server whose
`listOrganizations` returns the target organization.

## Tool sequence

1. **Resolve the organization** — `listOrganizations`, match by name, keep the `id`.
2. **Resolve the compliance portal** — `getCompliancePortal` with `organization_id`. Keep
   `compliance_portal.id`; every group call still takes it as the `trust_center_id` argument (the resource
   was renamed from "trust center" to "compliance portal" but the ID parameter kept its old name).
3. **Read what already exists** — `listCommitmentGroups` with the `trust_center_id`, and
   `listCommitments` with a `group_id`. Do this so you reuse the existing group instead of duplicating it.
   Order by `{field: "RANK", direction: "ASC"}`.
4. **Create the single group** — all commitments live under one group titled "Security at <company name>".
   If it does not exist yet, `addCommitmentGroup` with `trust_center_id`, `title`, `description`; if it (or
   any other group) already exists, reuse it. Capture the returned `commitment_group.id`. You need it to
   attach commitments.
5. **Add commitments** — `addCommitment` with `group_id`, `icon`, `eyebrow`, `title`, `description`. These
   can be sent in parallel once you have the group id.

## Updating and deleting

- `updateCommitmentGroup` — takes `id`; any of `title`, `description`, `rank` are optional. Null/omitted
  fields are left unchanged.
- `updateCommitment` — takes `id`; any of `icon`, `eyebrow`, `title`, `description`, `rank` are optional.
- `deleteCommitment` — takes the commitment `id`.
- `deleteCommitmentGroup` — takes the group `id`. Deleting a group removes the commitments inside it, so
  to remove a whole theme (and its commitments) you can delete the group directly rather than each
  commitment first.

## Ordering

Both groups and commitments have a `rank` (1-based) that sets display order. New items get the next rank
in creation order. To reorder, pass `rank` to the update call.

## Icon enum

`icon` must be exactly one of these values:

```
LOCK_KEY, EYE_SLASH, FINGERPRINT, SHIELD_WARNING, SHIELD_CHECK, SIREN, KEY, LOCK,
CLOUD, DATABASE, GLOBE, EYE, USERS, CERTIFICATE, GAVEL, HEARTBEAT, BELL, BUG, CODE, SERVER
```

Sensible mappings:

- Encryption / data at rest: `LOCK`
- Data deletion / privacy: `EYE_SLASH`
- Data inventory / storage: `DATABASE`
- Authentication (SSO/MFA): `KEY`
- Identity / least privilege: `FINGERPRINT`
- Production / infrastructure access: `SERVER`
- Source control / code: `CODE`
- Vulnerability scanning / testing: `BUG`
- Monitoring / logging: `EYE`
- Alerting: `BELL` or `SIREN`
- Threat / risk: `SHIELD_WARNING`
- Controls in place / audited: `SHIELD_CHECK` or `CERTIFICATE`
- Governance / legal / compliance: `GAVEL`
- Vendors / people: `USERS`
- Availability / uptime: `HEARTBEAT`
- Cloud / hosting: `CLOUD`
- Networking / public surface: `GLOBE`

## Gotchas seen in practice

- **`insufficient scope`.** The connected MCP token may be read-only for the compliance portal, or may not yet
  have commitment write scope at all. If creates fail with `insufficient scope`, nothing was written.
  Tell the user to re-authorize / refresh the Probo MCP connection with compliance-portal write scope, then
  retry the same batch. Do not keep retrying the identical call; the scope has to change first.
- **One group only.** All commitments belong under a single group titled "Security at <company name>". If
  the portal already has a group (whatever its title), reuse it rather than creating a second one; rename it
  with `updateCommitmentGroup` if its title is not "Security at <company name>". If earlier runs left several
  groups, consolidate: move commitments into the one group and delete the extras (confirm before deleting).
- **Publishing is public.** Creates, updates, and deletes change what visitors to the compliance portal see.
  Confirm the copy with the user before writing.

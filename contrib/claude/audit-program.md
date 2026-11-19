# Audit programs

An audit program is the certificate cycle for one organization and one framework. It groups the audits that happen during that cycle, and it can open the next audit when the current one is finished and due.

A program is not an audit. It has no report, no findings, and no audit state. The certificate window (`valid_from`, `valid_until`) is a fact about the cycle. Each linked audit keeps its own firm, report, findings, controls, state, validity, and fieldwork dates.

Linking is optional. An audit with no program is a complete record. A one-off engagement stays that way.

## Decisions

Three shapes were considered. This spec is the one to build.

- A parent audit (pull request 1563) made the cycle itself an audit. The cycle is the certificate, so it is its own record. That pull request stays closed.
- A program with no scheduling (pull request 1564) is the right container and the wrong lifecycle. It also assumed one `framework_id` on `audits` and entity type 114. Audits now link frameworks through `audits_frameworks` and carry a `firm`. Entity type 114 is `BusinessFunctionEntityType`. Do not revive that branch. Implement this spec from current `main`.
- Task recurrence clones the current task, moves the interval onto the clone, clears it on the old row, ignores state, and collapses missed cycles into one future occurrence. There is no series to list. A certificate cycle has to keep every engagement, so the interval stays on the program and each spawn is a new empty audit. Earlier audits stay linked.

## Model

New table `audit_programs`:

| Column | Type | Notes |
| --- | --- | --- |
| `id` | `TEXT` primary key | GID, entity type below |
| `tenant_id` | `TEXT NOT NULL` | |
| `organization_id` | `TEXT NOT NULL` | FK to `organizations`, `ON DELETE CASCADE` |
| `framework_id` | `TEXT NOT NULL` | FK to `frameworks`, `ON DELETE CASCADE` |
| `name` | `TEXT NOT NULL` | |
| `valid_from` | `DATE` | Certificate window. Nullable |
| `valid_until` | `DATE` | Nullable. When both bounds are set, `valid_until` is strictly after `valid_from` |
| `recurrence` | `INTERVAL` | Nullable. Same `TimeSpan` storage as `tasks.recurrence` |
| `created_at` | `TIMESTAMPTZ NOT NULL` | |
| `updated_at` | `TIMESTAMPTZ NOT NULL` | |

`audits` gains `audit_program_id TEXT` referencing `audit_programs(id)` `ON DELETE SET NULL`.

Index `audits(audit_program_id)`. The recurrence worker and the program's audit connection filter on it on every poll. That is the justification for the index. No other new indexes.

Entity type: the next free value in `pkg/coredata/entity_type_reg.go`. When this spec was written, `CMPDomainSurfaceEntityType` was 147, so the program is 148. Confirm the registry at implementation time. Do not reuse 114.

`AuthorizationAttributes` returns `organization_id`, same as `Audit`.

Order fields: `CREATED_AT`, `NAME`, `VALID_FROM`, `VALID_UNTIL`.

The framework is immutable after create. A certificate is for one framework. Changing it would invalidate every link.

Deleting a framework deletes its programs. The audit foreign key then sets `audit_program_id` to null, so the audits remain.

## Linking

An audit has at most one program. A program has many audits.

Linking is allowed when all of these hold:

- The program loads in the caller's tenant scope.
- `program.organization_id` equals `audit.organization_id`. Any other organization, including a program that exists in another tenant, is `NotFound`. The error does not reveal that the other row exists.
- The program's framework is one of the audit's rows in `audits_frameworks`. A same-organization mismatch is `Invalid` on `audit_program_id`. The caller can already see both records.
- The audit has at least one framework. An audit with an empty framework set cannot be linked.

An audit that covers several frameworks can still belong to only one program: the cycle whose framework is in that set. A combined engagement that must sit in two cycles is a later change (the link becomes a junction). This version does not do that.

Updating an audit's frameworks so the set no longer contains the linked program's framework is `Invalid`. The caller unlinks first. The console clears the program control in the form when the selected frameworks no longer qualify, and the server still enforces the rule.

Clearing `audit_program_id` unlinks. The audit stays.

Deleting a program sets `audit_program_id` to null on its audits. It does not delete audits, reports, findings, or controls.

## Certificate window and recurrence

`valid_from` / `valid_until` are the certificate window, usually several years. `recurrence` is how often an audit inside that window is opened, usually `P1Y`. They are independent.

The worker does not extend the certificate window. It may open an audit after `valid_until`. The program then shows as past its certificate end. Renewing the certificate is a person's update of those two dates.

A partial update of one bound validates the merged pair. `valid_until` must be strictly after `valid_from` when both are set. Clearing a bound is allowed. The same merge rule applies to an audit's own validity and fieldwork dates; this spec does not change that.

`name` is required, trimmed, and checked with `SafeTextNoNewLine(TitleMaxLength)`.

`recurrence` uses the GraphQL `TimeSpan` scalar, ISO-8601, the same type as `Task.recurrenceInterval`. Null means the program does not open further audits.

A set interval must satisfy all of these:

- It is strictly greater than zero.
- It is at least one day. Hourly and shorter intervals are `Invalid`.
- Adding it to 31 January 2000 produces a later date. This rejects spans that look positive and clamp backwards, the same class `pkg/task/recurrence.go` refuses.

Clearing the interval stops future spawns. Audits already linked stay linked. Changing the interval does not rewrite existing audits. The next due date is recomputed from the latest audit.

## When the next audit opens

A worker `audit-program-recurrence` follows `contrib/claude/go-worker.md`. Poll interval is a seconds field on the probod config, next to `task-recurrence`. Claim with `FOR UPDATE SKIP LOCKED` inside a transaction. One program per claim. The worker is a system action: it does not run as a user and does not check an IAM permission. It writes only the tenant on the locked program row.

The anchor of a linked audit is `audit_end_date`, then `valid_until`, then `created_at`.

The latest audit is the linked audit with the greatest anchor. Tie-break on `id` descending.

An open audit has state `TO_BOOK`, `AUDIT_BOOKED`, `NOT_STARTED`, or `IN_PROGRESS`. A terminal audit has state `COMPLETED`, `REJECTED`, or `OUTDATED`.

The program is due when all of these hold:

- `recurrence` is not null.
- At least one audit is linked. The worker never creates the first audit. The person creates or links that one, which avoids spawning a duplicate while they are still attaching history.
- No linked audit is open. An engagement still in progress, including one past its planned end, blocks the next spawn. `REJECTED` is terminal, so it does not by itself block, and it does not by itself spawn a replacement. A re-audit before the interval is a manual create.
- `latest.anchor + recurrence <= now`.

The claim query expresses that predicate, so a program that is not due is not locked. After a successful spawn the new audit is `TO_BOOK`, so the program drops out of the due set.

One spawn per claim. A program that is several intervals behind gets one new audit, not one per missed year. That audit stays open until a person finishes it, so the following ticks do not fill the gap. The missed years stay visible because the new dates are one interval after the previous audit, even when that period is already over.

If the interval cannot produce a later date from the latest anchor, set `recurrence` to null and log the program id. That is the only automatic clear. It exists so a row that can never advance cannot sit at the head of the queue. A normal `P1Y` does not hit this path. Any other insert failure rolls the transaction back and retries.

Log program id and audit id only.

## The audit that gets created

The spawned audit is empty engagement, not a clone of the previous one.

| Field | Value |
| --- | --- |
| Organization | The program's organization |
| Program | This program |
| Frameworks | Exactly the program's framework. Extra frameworks on the previous audit stay on that audit |
| Name | `{program name} {year}`, where the year is the calendar year of the new fieldwork start. When that start is empty, use the calendar year of `anchor + recurrence` |
| Firm | Copied from the latest audit, including null |
| State | `TO_BOOK` |
| Report | None |
| Findings and controls | None |
| Validity | Both bounds of the latest audit shifted by the interval, when both are set and the shifted end is still after the shifted start. Otherwise both null |
| Fieldwork dates | Same shift rule as validity, using `audit_start_date` and `audit_end_date` |

Shift with `TimeSpan.AddTo`, the same addition task recurrence uses. A name collision is allowed. Audits have no unique name.

Unlinking or deleting the spawned audit makes the program due again on a later tick, when the predicate still holds. That is intentional: discarding the placeholder does not stop the cycle.

## Authorization

Actions:

- `core:audit-program:get`
- `core:audit-program:list`
- `core:audit-program:create`
- `core:audit-program:update`
- `core:audit-program:delete`

`ViewerPolicy` and `AuditorPolicy` get `get` and `list` only.

Fold the actions into the existing OAuth scopes, rather than adding a scope:

- `v1:audit:read` gets `get` and `list`.
- `v1:audit` gets all five.

Creating, updating, and deleting a program require the matching program action. Linking or unlinking an audit requires `core:audit:update` plus a visible program (`get`). The link checks in [Linking](#linking) still apply.

## API surface

Expose the resource on all four surfaces in `contrib/claude/api-surface.md`.

GraphQL, in `pkg/server/api/console/v1/graphql/`:

- `AuditProgram implements Node`, with `organization`, `framework`, `validFrom`, `validUntil`, `recurrenceInterval`, a cursor connection of `audits`, `createdAt`, `updatedAt`, and `permission(action:)`.
- `Organization.auditPrograms` connection.
- Mutations `createAuditProgram`, `updateAuditProgram`, `deleteAuditProgram`.
- Create input: `organizationId`, `frameworkId`, `name`, optional `validFrom`, `validUntil`, `recurrenceInterval`.
- Update input: `id`, optional `name`, and omittable nullable `validFrom`, `validUntil`, `recurrenceInterval`. No `frameworkId`.
- `Audit.auditProgram` and omittable nullable `auditProgramId` on audit create and update. Null clears the link. Omitted leaves it.

Map errors with the helpers in `contrib/claude/api-surface.md`. Cross-organization and invisible ids are `NotFound`. Validation is `Invalid`. Anything else is the generic internal error. Batch `auditProgram` loads with a dataloader. Paginate the program's audits and the program picker.

MCP tools: `listAuditPrograms`, `getAuditProgram`, `addAuditProgram`, `updateAuditProgram`, `deleteAuditProgram`. Audit add and update accept omittable nullable `audit_program_id`. Program update accepts omittable nullable `name`, `valid_from`, `valid_until`, and `recurrence_interval`, so a caller who sets only the interval does not wipe the name.

CLI: `prb audit-program` with `list`, `view`, `create`, `update`, `delete`. `prb audit create` and `prb audit update` gain `--program` and `--clear-program`. Program create and update gain `--recurrence` and `--clear-recurrence`. Reject a non-positive `--limit` before pagination. `--clear-program` and `--program` together are `Invalid`, and the same for the recurrence pair.

n8n: an Audit Program resource with create, get, get-all, update, and delete. Audit create and update accept `auditProgramId`. Reads include the linked program.

## Console

Programs live under Audits. The governance nav keeps a single Audits entry (`navDestinations`, id `audits`). There is no Audit Programs item.

Routes, under the existing audits route:

- `/organizations/:organizationId/audits` — audit list.
- `/organizations/:organizationId/audits/programs` — program list.
- `/organizations/:organizationId/audits/programs/:auditProgramId` — program details.

Tabs on the audits layout: Audits, then Programs. Each tab is shown only when the organization permission for that list action is true. There is no redirect from `/audit-programs`. That path never shipped.

The list shows name, framework, certificate window, recurrence, and audit count. Create and delete are gated on `canCreate` / `canDelete`.

Details edit name, certificate window, and recurrence. Disable those controls when `permission(action: "core:audit-program:update")` is false. Show linked audits with their own fragment, paginated, linking to the audit.

Audit create and audit details get a program select. Options are programs of this organization whose framework is among the audit's selected frameworks. Paginate the options. When the audit is already linked to a program that is not on the loaded page, resolve that program by id so the control shows it. Changing the framework set to one that excludes the selected program clears the control before submit.

Strings for the program pages live in a colocated `_locales` namespace for `en-US`, `fr-FR`, and `nl-NL`. Tab labels live with the audits area.

## Tests

End-to-end coverage in `e2e/`, for the console GraphQL API and MCP:

- Create a program, link an audit, and read the link from both sides.
- A second program link replaces the first. Clearing the link leaves the audit in place.
- Delete the program and the audit remains with `auditProgram` null.
- A cross-organization program id is `NotFound`.
- A same-organization program whose framework is not on the audit is `Invalid`.
- Dropping that framework while linked is `Invalid`.
- A viewer can list and get, and cannot create, update, or delete.
- Update with only `recurrenceInterval` leaves `name` and the certificate dates unchanged.
- A non-positive CLI `--limit` is an error.
- Worker: a program with no audits does not spawn.
- Worker: an open latest audit does not spawn, even when the anchor plus the interval is in the past.
- Worker: a terminal latest audit spawns one `TO_BOOK` audit when due. The new audit is linked, has the program's framework only, copies `firm`, has no report, and has dates shifted by the interval.
- Worker: the next tick does not spawn another audit while that one is open.
- Worker: a program several intervals behind still spawns one audit, dated one interval after the previous one.
- Worker: an interval that cannot advance the anchor is cleared, and the program is not retried forever.

## Out of scope

- Generating the first audit.
- Backfilling one audit per missed interval.
- Copying the previous audit's report, findings, controls, extra frameworks, or name.
- Putting a recurrence interval on `audits`.
- Letting one audit belong to two programs.
- Auto-extending `valid_until` when a spawn falls past it.
- Editing the program's framework.

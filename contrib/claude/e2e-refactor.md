# E2E Suite Refactoring Plan

The console E2E suite is the primary product harness. Refactoring must preserve
its black-box boundary: tests continue to use real HTTP endpoints,
authentication, email delivery, workers, object storage, and PostgreSQL.

## Goals

- Make tests read as product behavior.
- Keep failures actionable for humans and agents.
- Increase coverage without making runtime grow linearly.
- Reduce repeated signup, invitation, and organization setup.
- Preserve tenant isolation and independent mutable resources.

The baseline CI run on 2026-08-04 executed the console tests in 5 minutes and
37 seconds, excluding stack and binary setup.

## Test shapes

Choose one shape for each behavior:

1. **Contract matrix** — required fields, validation, enum values, RBAC, and
   pagination variants. Keep these table-driven and parallel.
2. **Isolated scenario** — CRUD, filtering, ordering, and resolver behavior
   that can share one organization while each case owns its mutable resources.
3. **Journey** — ordered, stateful, or multi-actor product workflows. Use
   `e2e/internal/journey` and describe user-visible actions.
4. **Protocol flow** — OAuth, SCIM, file upload, and trust-center protocols.
   Keep protocol-specific drivers and expose meaningful stages in diagnostics.

Do not wrap every assertion in a journey step. Journey steps are for causal
workflow boundaries, not a replacement for table-driven tests.

## Fixture rules

- One RBAC suite uses one `testutil.OrganizationRoles` fixture unless the
  behavior explicitly tests organization boundaries.
- Each parallel mutation case creates its own uniquely named resource.
- Tenant-isolation suites use two independent organizations.
- Unfiltered list, ordering, and pagination assertions use dedicated
  organizations so unrelated parallel data cannot affect their results.
- Stateful workflows reuse actors and resources but execute ordered actions,
  not parallel subtests.
- Never share an authenticated client across top-level tests through package
  globals.
- Never replace product setup with direct SQL merely to improve runtime.

## Source structure

- GraphQL documents are declared once per operation.
- Role expectations and validation variants are tables.
- Domain helpers describe actions and return typed results.
- Helpers fail with the operation, actor, resource identifier, GraphQL code,
  and response path when available.
- Calls follow the repository multiline-argument rule even when `gofmt` or a
  linter would accept a more compact form.
- Test names read from broad behavior to specific expectation, for example:
  `TestFramework_RBAC/update/viewer_cannot_update`.

## Migration sequence

Each batch must pass formatting, the pinned Go linter, targeted tests, console
test compilation, and live CI E2E before the next batch starts.

1. **Harness and pilot**
   - Correct journey harness style and diagnostics.
   - Migrate framework RBAC to the reusable role fixture.
   - Compare bootstrap count and CI runtime against the baseline.
2. **High-cost RBAC suites**
   - Measure, document, datum, audit, and processing activity.
   - Use one role fixture per entity and declarative operation matrices.
3. **Core governance domains**
   - Framework, control, task, risk, obligation, finding, and asset.
   - Consolidate compatible isolated scenarios without polluting list tests.
4. **Third-party and privacy domains**
   - Third parties, contacts, services, compliance reports, processing
     activities, DPIA/TIA, and rights requests.
5. **Workflow-heavy domains**
   - Document lifecycle, publication, access reviews, cookie banners, and risk
     assessments.
   - Introduce domain drivers where ordered journeys otherwise repeat GraphQL.
6. **Protocol suites**
   - OAuth, SCIM, connectors, trust center, and uploads.
   - Preserve protocol-level assertions while adding stage diagnostics.
7. **Final consolidation**
   - Remove superseded local fixture helpers.
   - Audit duplicate RBAC coverage without deleting distinct wiring checks.
   - Tune `-parallel` and database pool size from measured contention.

## Batch acceptance

For every batch, record:

- `NewClient` and `NewClientInOrg` counts before and after.
- Test-process wall time from GitHub Actions.
- Slowest top-level tests from JUnit.
- Any case retaining a dedicated organization and why.
- Formatting, lint, compile, and live E2E outcomes.

Do not continue a mechanical migration when runtime regresses, isolation
becomes ambiguous, or the resulting Go is harder to read than the original.

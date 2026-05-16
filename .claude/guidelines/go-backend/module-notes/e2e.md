# Probo — Go Backend — e2e/

**Purpose.** End-to-end test suite (~65 files) running against a live
`probod` binary backed by Docker Compose (Postgres, Mailpit, SeaweedFS,
optionally Pebble ACME, Keycloak). Exercises the full GraphQL surface
(console + connect + mcp), authentication flows, RBAC matrix, and
tenant isolation.

> See [testing.md § 3 E2E mechanics](../testing.md#3-e2e-mechanics).

**Key files.**

- `e2e/internal/testutil/client.go` — `Client` (cookie-jar HTTP
  transport, `Execute`, `Do`, `ExecuteConnect`, `ExecuteWithFile`,
  `ExecuteShouldFail`).
- `e2e/internal/testutil/graphql.go` — `GraphQLErrors`, transport.
- `e2e/internal/testutil/assert.go` — assertion helpers
  (`RequireForbiddenError`, `RequireErrorCode`, `AssertNodeNotAccessible`,
  `AssertFirstPage`/`AssertMiddlePage`/`AssertLastPage`,
  `AssertOrderedAscending`/`AssertTimesOrderedDescending`,
  `AssertTimestampsOnCreate`/`AssertTimestampsOnUpdate`).
- `e2e/internal/testutil/mailpit.go` — `SearchMails`,
  `CheckMessageLinks`.
- `e2e/internal/testutil/env.go` — `GetBaseURL`, `GetMailpitBaseURL`.
- `e2e/internal/factory/<entity>.go` — flat `CreateXxx(c, factory.Attrs{...})`
  + builder `NewXxx(c).WithField(...).Create()`.
- `e2e/internal/factory/factory.go` — `Attrs`, `SafeName`, `SafeEmail`,
  `SafeOrigin` (gofakeit-backed).
- `e2e/console/<entity>_test.go` — one file per entity. `vendor_test.go`
  is the canonical reference.
- `e2e/mcp/<tool>_test.go` — MCP tool tests.

**How to extend (a new e2e test).**

1. Add factory entries under `e2e/internal/factory/<entity>.go` (both
   the flat and builder forms).
2. Create `e2e/console/<entity>_test.go` (or `e2e/mcp/`) with
   `package console_test`.
3. Top-level test functions and **every** `t.Run(...)` subtest call
   `t.Parallel()`.
4. Cover at minimum: happy path, RBAC matrix (all 5 roles × all
   verbs), tenant isolation (two `NewClient` calls + `AssertNodeNotAccessible`),
   pagination (use `AssertFirstPage` etc.), and timestamps on
   create/update.
5. For email side effects, poll Mailpit via
   `c.SearchMails(t, ctx, "to:"+email)`.

**Top pitfalls.**

- Missing `t.Parallel()` at any subtest level — serialises the suite,
  called out in review.
- Sharing factory output across tests by package-level vars — every
  test creates its own org via `NewClient`; cross-test sharing breaks
  isolation.
- Hard-coded names/emails — collisions with the parallel suite. Use
  `factory.SafeName` / `factory.SafeEmail`.
- Asserting on Mailpit immediately without polling — mail delivery is
  async; allow a few retries.
- Forgetting RBAC + tenant-isolation tests on a new mutation — these
  are the project's safety net for the `Scoper` invariant.

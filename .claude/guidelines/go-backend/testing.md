# Probo — Go Backend — Testing

> Cross-stack rules (PII in test data, license headers, CI gates) live in
> [shared.md](../shared.md). This file documents Go test mechanics: unit
> tests, e2e tests, factory builders, RBAC matrix tests, tenant isolation,
> pagination assertions, and coverage expectations.
>
> Authoritative references:
> [`contrib/claude/go-testing.md`](../../../contrib/claude/go-testing.md),
> [`contrib/claude/e2e.md`](../../../contrib/claude/e2e.md).

---

## 1. Frameworks

| Layer | Framework | Used by |
| --- | --- | --- |
| Unit & e2e assertions | `github.com/stretchr/testify` (`require` + `assert`) | every package with tests |
| Integration target | `make test-e2e` against live `probod` + Docker Compose stack | `e2e/console`, `e2e/mcp` |
| Mail assertions | Mailpit (HTTP API at `GetMailpitBaseURL()`) | `e2e/internal/testutil/mailpit.go` |
| Random data | `github.com/brianvoe/gofakeit/v7` via `factory.SafeName`/`SafeEmail` | factories |
| OAuth2 / HTTP fakes | `httptest.NewServer` + `httpclient.WithSSRFAllowLoopback()` | `pkg/connector` |

> **No Ginkgo / GoConvey / mock libraries.** Plain `*testing.T`,
> testify, and table-driven `t.Run`.

---

## 2. Unit-test mechanics

### Mandatory `t.Parallel()` at every level

```go
func TestVendor_Validate(t *testing.T) {
    t.Parallel()

    cases := []struct {
        name    string
        request CreateVendorRequest
        wantErr bool
    }{
        {"required name missing", CreateVendorRequest{}, true},
        // ...
    }

    for _, c := range cases {
        c := c
        t.Run(c.name, func(t *testing.T) {
            t.Parallel()
            err := c.request.Validate()
            if c.wantErr {
                require.Error(t, err)
            } else {
                require.NoError(t, err)
            }
        })
    }
}
```

Every test function and **every** subtest calls `t.Parallel()`. The CI
runs with `-race`; missing `t.Parallel()` is called out in review.

### `require` vs `assert`

- **`require.*`** — fail-fast preconditions. The first failure aborts the
  test. Use for: `require.NoError(t, err)`, factory creation, "the row
  must exist" guards.
- **`assert.*`** — collecting assertions that should keep running. Use
  for value comparisons after the precondition has been met:
  `assert.Equal(t, want, got)`, `assert.Len(t, list, 3)`,
  `assert.True(t, found)`.

The mix is intentional: `require` proves you have a valid object to
inspect; `assert` then inspects all of its fields without short-circuiting
on the first mismatch.

### Black-box test packages

Tests live in **`package <name>_test`** by default — frequency-2
reviewer rule (PR #1023). This forces tests to use the public API and
prevents creeping reliance on internals. See
[conventions.md § 11](./conventions.md#11-test-packages).

### File naming

- `<source>_test.go` next to `<source>.go`.
- Test function name: `Test<Type>_<Scenario>` (e.g. `TestVendor_Validate`,
  `TestCookieBanner_LoadByOrganizationID_WithFilter`).
- Subtest names: lowercase descriptive strings —
  `"with full details"`, `"viewer cannot create"`,
  `"unknown model returns false"`.

### Table-driven shape (canonical)

`pkg/llm/registry_test.go` and `pkg/connector/oauth2_test.go` are good
references for table-driven testify tests with `t.Parallel()` at both
levels.

---

## 3. E2E mechanics

E2E tests live in `e2e/console/` (~43 files), `e2e/mcp/` (~22 files),
plus the shared `e2e/internal/testutil` and `e2e/internal/factory`.
They run against a real `probod` binary backed by the Docker Compose
stack (`make stack-up && make test-e2e`).

### Client setup

```go
func TestVendor_Create(t *testing.T) {
    t.Parallel()

    owner := testutil.NewClient(t, testutil.RoleOwner)         // signs up a user, creates an org
    viewer := testutil.NewClientInOrg(t, testutil.RoleViewer, owner) // joins same org via Mailpit invite flow
    // ...
}
```

- `testutil.NewClient(t, role)` provisions a fresh user **and** a fresh
  organization. Returns a `*testutil.Client` carrying cookie-jar HTTP
  transport, `OrganizationID`, `UserID`, `ProfileID`, `Role`.
- `testutil.NewClientInOrg(t, role, owner)` provisions a new user in an
  existing org. Drives the full **invite → Mailpit lookup → activate →
  set password** sequence. Tests that need multiple roles in the same
  org use this.
- Each test gets its own org — there is no shared fixture data.
  Parallel tests are isolated by tenant scoping in the daemon.

### Factory builders (two-tier)

```go
// Builder pattern — preferred when the test reads the entity multiple times
v := factory.NewVendor(owner).
    WithName(factory.SafeName("vendor")).
    WithDescription("imported from CSV").
    Create()

// Flat form — preferred for one-liners
vid := factory.CreateVendor(owner, factory.Attrs{"name": factory.SafeName("vendor")})
```

- Both forms live under `e2e/internal/factory/`.
- **`factory.SafeName(prefix)`** and **`factory.SafeEmail()`** generate
  unique values per call (gofakeit-backed) — required because tests run
  in parallel against a shared database.
- `factory.Attrs` is `map[string]any` with typed accessor methods
  (`getString`, `getInt`, `getBool`, `getStringPtr`).
- Adding a new entity to the factory: implement both forms in the same
  file (`e2e/internal/factory/<entity>.go`).

### GraphQL transport

```go
type vendorResponse struct {
    Node struct {
        ID   string
        Name string
    }
}

var resp vendorResponse
err := owner.Execute(t, ctx, vendorQuery, map[string]any{"id": vid}, &resp)
require.NoError(t, err)
assert.Equal(t, vid, resp.Node.ID)
```

- `Client.Execute` → unmarshals into the response struct, fails fast on
  GraphQL errors.
- `Client.Do` → returns the raw `GraphQLResponse` for tests that need
  to inspect partial results.
- `Client.ExecuteShouldFail` → expects errors, returns the typed
  `GraphQLErrors` slice (each carries `extensions.code`).
- `Client.ExecuteConnect` → hits the `connect` GraphQL endpoint
  (`/api/connect/v1/graphql`).
- `Client.ExecuteWithFile` → multipart upload following the
  graphql-multipart-request-spec.

### Inline response structs

E2E tests use **inline anonymous structs** for GraphQL response shapes
inside each test function. The `console` and `mcp` test packages do
not import generated TypeScript-style types. Keeping the struct local
makes the test self-documenting and avoids cross-test coupling.

---

## 4. Test patterns

### RBAC matrix tests

For every mutating endpoint, parametrise over roles to assert that
authorization is enforced:

```go
func TestVendor_Create_RBAC(t *testing.T) {
    t.Parallel()

    cases := []struct {
        role    testutil.TestRole
        canCreate bool
    }{
        {testutil.RoleOwner, true},
        {testutil.RoleAdmin, true},
        {testutil.RoleEmployee, false},
        {testutil.RoleViewer, false},
        {testutil.RoleAuditor, false},
    }

    for _, c := range cases {
        c := c
        t.Run(string(c.role), func(t *testing.T) {
            t.Parallel()
            owner := testutil.NewClient(t, testutil.RoleOwner)
            actor := testutil.NewClientInOrg(t, c.role, owner)

            err := actor.Execute(t, ctx, createVendorMutation, vars, nil)
            if c.canCreate {
                require.NoError(t, err)
            } else {
                testutil.RequireForbiddenError(t, err)
            }
        })
    }
}
```

- **`testutil.RequireForbiddenError(t, err)`** asserts the error has
  a `FORBIDDEN` extensions.code.
- **`testutil.RequireErrorCode(t, err, "INVALID")`** for any other
  category (`NOT_FOUND`, `INVALID`, `CONFLICT`, `UNAUTHENTICATED`).

### Tenant isolation

```go
func TestVendor_TenantIsolation(t *testing.T) {
    t.Parallel()

    owner1 := testutil.NewClient(t, testutil.RoleOwner)
    owner2 := testutil.NewClient(t, testutil.RoleOwner) // different org

    vid := factory.CreateVendor(owner1, factory.Attrs{"name": factory.SafeName("v")})

    // Owner of org 2 must not see org 1's vendor.
    testutil.AssertNodeNotAccessible(t, owner2, vid)
}
```

- **Two `testutil.NewClient` calls** create two independent organizations.
- **`testutil.AssertNodeNotAccessible(t, client, gid)`** accepts both a
  nil node response and a non-nil error — both are valid access-denied
  signals from the GraphQL `node(id: ID!)` field.
- Apply this to every entity that has a public ID. It is the single
  most important regression test for the `Scoper` invariant.

### Pagination assertions

```go
testutil.AssertFirstPage(t, pageInfo)   // hasPrev=false, hasNext=true
testutil.AssertMiddlePage(t, pageInfo)  // hasPrev=true, hasNext=true
testutil.AssertLastPage(t, pageInfo)    // hasPrev=true, hasNext=false
```

For ordering:

```go
testutil.AssertOrderedAscending(t, ids)
testutil.AssertTimesOrderedDescending(t, timestamps)
```

For timestamps on create/update flows:

```go
testutil.AssertTimestampsOnCreate(t, createdAt, updatedAt)
testutil.AssertTimestampsOnUpdate(t, before, after)
```

### Mailpit-backed assertions

```go
mails, err := owner.SearchMails(t, ctx, "to:"+inviteeEmail)
require.NoError(t, err)
require.Len(t, mails, 1)

links := owner.CheckMessageLinks(t, ctx, mails[0].ID)
activationURL := links[0]
// ...drive activation flow...
```

Used in invite tests, password-reset tests, document-published
notifications, vendor-assessment dispatch, etc.

---

## 5. Coverage expectations

- **Default**: there is no project-wide coverage threshold. CI emits
  `coverage.out` but does not fail on a percentage.
- **Auth-sensitive code requires 100% unit-test coverage.**
  Frequency-3 reviewer rule (PR #957): *"this file must have unit test
  100%"*. This applies to:
  - `pkg/iam/oauth2server/`
  - `pkg/iam/oidc/`
  - `pkg/iam/saml/`
  - PKCE / ID-token verification helpers
  - any new file that issues, validates, or rotates tokens
- **New features need at least one e2e test in `e2e/<surface>/`.**
  Frequency-2 reviewer rule (PR #1102: *"Maybe add some e2e tests?"*).
  The minimum bar is one happy-path test per new mutation; richer
  coverage (RBAC matrix, tenant isolation, pagination) is encouraged.

---

## 6. Running tests

```bash
# All Go unit tests with race detector and coverage
make test
# One package
make test MODULE=./pkg/llm

# E2E suite (requires Docker Compose stack up + Lima sandbox if
# you do not have a Linux host)
make stack-up
make test-e2e
```

E2E target reads `GetBaseURL()` and `GetMailpitBaseURL()` from env vars
set by the Make target — see `e2e/internal/testutil/env.go`.

---

## 7. A complete example

`e2e/console/vendor_test.go` is the canonical reference: it covers
create / update / delete / list / view, RBAC matrix, tenant isolation,
pagination, timestamp assertions, and webhook side-effects, all using
the conventions above. When in doubt about a new e2e test, copy its
shape.

For unit testing, `pkg/llm/registry_test.go` shows the table-driven
form, `pkg/connector/oauth2_test.go` shows the `httptest` + state-token
fake-server form, and `pkg/page/cursor_test.go` shows the
generics-with-test-stub form (`testOrderField` satisfies the
`OrderField` constraint for isolation).

# e2e

End-to-end tests against a running `bin/probod` instance.

## Prerequisites

Build the binary first: `make build` (or `SKIP_APPS=1 make build` for backend-only).

## Running

```
make test-e2e
```

## Test setup

`testutil.Setup()` starts `bin/probod` as a subprocess (once per test run) and waits for the GraphQL endpoint to be healthy.

## Client

```go
c := testutil.NewClient(t, testutil.RoleOwner)
```

The client carries organization/tenant context and provides:
- `c.Execute(query, variables, &result)` — GraphQL queries
- `c.ExecuteConnect(query, variables, &result)` — Connect API queries
- `c.GetOrganizationID()` — current org

## Factory pattern

Test data created via `factory.Create*` functions:

```go
vendorID := factory.CreateVendor(c, factory.Attrs{"name": "Acme"})
userID := factory.CreateUser(c)
```

- `factory.SafeName(prefix)` — random unique names
- `factory.SafeEmail()` — random unique emails
- `factory.Attrs` map for overriding defaults

## Test structure

- Always `t.Parallel()` at the test function level
- Inline GraphQL queries as string constants
- Typed result structs per query
- `require.NoError` for mutation/query errors, `assert.Equal` for value checks

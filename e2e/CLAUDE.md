# e2e

End-to-end tests against a running `bin/probod` instance.

## Prerequisites

Build the binary first: `make build` (or `SKIP_APPS=1 make build` for backend-only).

## Running

```
make test-e2e
```

## Test setup

`testutil.Setup()` starts `bin/probod` as a subprocess (once per test run via `sync.Once`) and waits for the GraphQL endpoint to be healthy. No explicit teardown is needed — each test gets its own organization/user, so tests never interfere with each other.

## Client

```go
owner := testutil.NewClient(t, testutil.RoleOwner)
admin := testutil.NewClientInOrg(t, testutil.RoleAdmin, owner)
viewer := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)
```

`NewClient` creates a standalone user with their own organization. `NewClientInOrg` adds a user to an existing owner's organization with a downgraded role.

The client provides:
- `c.Execute(query, variables, &result)` — Console API (authenticated)
- `c.ExecuteConnect(query, variables, &result)` — Connect API (sign-up, sign-in)
- `c.ExecuteShouldFail(query, variables, &result)` — expects an error
- `c.GetOrganizationID()` — current org

## Factory pattern

Test data created via `factory.Create*` or the builder pattern:

```go
// Simple — returns ID string
vendorID := factory.CreateVendor(c, factory.Attrs{"name": "Acme"})

// Builder — chainable for optional fields
vendorID := factory.NewVendor(owner).
    WithName("Test").
    WithDescription("Desc").
    Create()
```

- `factory.SafeName(prefix)` — random unique names
- `factory.SafeEmail()` — random unique emails
- `factory.Attrs` map for overriding defaults

## Writing a test

Every test follows this structure:

```go
func TestVendor_Create(t *testing.T) {
    t.Parallel()

    owner := testutil.NewClient(t, testutil.RoleOwner)

    t.Run(
        "create a vendor",
        func(t *testing.T) {
            t.Parallel()

            const query = `
                mutation CreateVendor($input: CreateVendorInput!) {
                    createVendor(input: $input) {
                        vendorEdge {
                            node {
                                id
                                name
                                description
                            }
                        }
                    }
                }
            `

            var result struct {
                CreateVendor struct {
                    VendorEdge struct {
                        Node struct {
                            ID          string `json:"id"`
                            Name        string `json:"name"`
                            Description string `json:"description"`
                        } `json:"node"`
                    } `json:"vendorEdge"`
                } `json:"createVendor"`
            }

            name := factory.SafeName("vendor")

            err := owner.Execute(
                query,
                map[string]any{
                    "input": map[string]any{
                        "organizationId": owner.GetOrganizationID(),
                        "name":           name,
                        "description":    "A test vendor",
                    },
                },
                &result,
            )

            require.NoError(t, err)
            assert.NotEmpty(t, result.CreateVendor.VendorEdge.Node.ID)
            assert.Equal(t, name, result.CreateVendor.VendorEdge.Node.Name)
        },
    )
}
```

Key rules:
- Always `t.Parallel()` at both test function and subtest level
- Inline GraphQL queries as string constants
- Typed result structs with `json` tags per query
- Variables as `map[string]any`
- `require.NoError` for GraphQL call errors, `assert.Equal` for value checks

## Authorization testing

Test that roles are properly enforced and tenants are isolated:

```go
t.Run(
    "viewer cannot create vendor",
    func(t *testing.T) {
        t.Parallel()

        viewer := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)
        err := viewer.Execute(query, variables, &result)
        testutil.RequireForbiddenError(t, err)
    },
)

t.Run(
    "other org cannot access vendor",
    func(t *testing.T) {
        t.Parallel()

        otherOwner := testutil.NewClient(t, testutil.RoleOwner)
        err := otherOwner.Execute(query, variables, &result)
        require.Error(t, err)
    },
)
```

## Assertion helpers

| Helper | Purpose |
|--------|---------|
| `RequireForbiddenError(t, err)` | Verifies FORBIDDEN error code |
| `RequireErrorCode(t, err, code)` | Checks specific GraphQL error code |
| `AssertTimestampsOnCreate(t, created, updated)` | `createdAt == updatedAt` |
| `AssertTimestampsOnUpdate(t, created, updated)` | `createdAt` unchanged, `updatedAt` advances |
| `AssertFirstPage(t, pageInfo)` | First page of a paginated result |
| `AssertLastPage(t, pageInfo)` | Last page of a paginated result |
| `AssertOrderedAscending(t, items)` | Items in ascending order |
| `AssertOrderedDescending(t, items)` | Items in descending order |
| `AssertNodeNotAccessible(t, client, id)` | Tenant isolation check |

## File organization

```
e2e/
├── console/                    # Test files (package console_test)
│   ├── vendor_test.go
│   ├── framework_test.go
│   ├── audit_test.go
│   └── ...
└── internal/
    ├── factory/
    │   └── factory.go          # Test data builders
    └── testutil/
        ├── testutil.go         # Server setup/teardown
        ├── client.go           # Client and auth
        ├── graphql.go          # GraphQL request/response
        ├── assert.go           # Assertion helpers
        └── mailpit.go          # Email service integration
```

One test file per entity (e.g. `vendor_test.go`). Test function names follow `TestEntity_Operation` (e.g. `TestVendor_Create`, `TestVendor_Update`).

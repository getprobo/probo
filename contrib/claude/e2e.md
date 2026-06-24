# End-to-End Testing

E2E tests live in `e2e/console/` (package `console_test`) and run against a live `bin/probod` instance. The test infrastructure handles server lifecycle, authentication, and test data creation.

## Prerequisites

E2e uses the local [Pebble](https://github.com/letsencrypt/pebble) ACME server over HTTPS. Pebble’s TLS certificate is minted with [mkcert](https://github.com/FiloSottile/mkcert); register mkcert’s root CA in your system trust store once per machine:

```bash
mkcert -install
```

Without this step, probod cannot verify Pebble’s HTTPS endpoint when it registers an ACME account at startup.

You also need the Docker stack running and `bin/probod` built. `make stack-up` generates Pebble TLS material under `compose/pebble/certs/` (via mkcert):

```bash
make stack-up
make build
```

E2e config is built at test startup in `e2e/internal/testutil/testutil.go` (`generateConfig` → `probod-bootstrap`). It points ACME at Pebble but does not set `PROBOD_ACME_ROOT_CA`; local runs rely on the system trust store populated by `mkcert -install`. CI passes `PROBOD_ACME_ROOT_CA` in the workflow instead.

## Running tests

```bash
make test-e2e # Run all e2e tests
```

## Client setup

**Standalone user (new organization):**
```go
owner := testutil.NewClient(t, testutil.RoleOwner)
```

**User in existing organization:**
```go
admin := testutil.NewClientInOrg(t, testutil.RoleAdmin, owner)
viewer := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)
```

Each call creates a unique identity with a fresh email. Available roles: `RoleOwner`, `RoleAdmin`, `RoleViewer`.

## Client methods

| Method                                                | API     | Purpose                           |
| ----------------------------------------------------- | ------- | --------------------------------- |
| `c.Execute(query, vars, &result)`                     | Console | Execute and unmarshal into result |
| `c.MustExecute(query, vars, &result)`                 | Console | Execute, fail test on error       |
| `c.ExecuteShouldFail(query, vars)`                    | Console | Expect error, fail if succeeds    |
| `c.Do(query, vars)`                                   | Console | Low-level, returns raw response   |
| `c.ExecuteConnect(query, vars, &result)`              | Connect | For auth operations               |
| `c.ExecuteWithFile(query, vars, path, file, &result)` | Console | Single file upload                |
| `c.GetOrganizationID()`                               | —       | Current org GID                   |
| `c.GetUserID()`                                       | —       | Current user GID                  |

## Test data factories

Two patterns in `e2e/internal/factory/`:

**Builder pattern (preferred):**
```go
thirdPartyID := factory.NewThirdParty(owner).
	WithName("Stripe").
	WithCategory("CLOUD_PROVIDER").
	Create()

frameworkID := factory.NewFramework(owner).
	WithName("SOC 2").
	Create()

controlID := factory.NewControl(owner, frameworkID).
	WithName("Access Control").
	Create()
```

**Simple factory functions:**
```go
thirdPartyID := factory.CreateThirdParty(c, factory.Attrs{"name": "Acme"})
taskID := factory.CreateTask(c, &measureID, factory.Attrs{"name": "Task 1"})
```

Use `factory.SafeName("prefix")` for unique names and `factory.SafeEmail()` for unique emails.

## Test structure

Every test and subtest **must** call `t.Parallel()`. One test file per entity in `e2e/console/`. Function naming: `TestEntity_Operation`.

```go
func TestThirdParty_Create(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	t.Run("with required fields", func(t *testing.T) {
		t.Parallel()

		const query = `
			mutation CreateThirdParty($input: CreateThirdPartyInput!) {
				createThirdParty(input: $input) {
					thirdPartyEdge {
						node { id name }
					}
				}
			}
		`

		var result struct {
			CreateThirdParty struct {
				ThirdPartyEdge struct {
					Node struct {
						ID   string `json:"id"`
						Name string `json:"name"`
					} `json:"node"`
				} `json:"thirdPartyEdge"`
			} `json:"createThirdParty"`
		}

		err := owner.Execute(query, map[string]any{
			"input": map[string]any{
				"organizationId": owner.GetOrganizationID().String(),
				"name":           factory.SafeName("ThirdParty"),
			},
		}, &result)

		require.NoError(t, err)
		assert.NotEmpty(t, result.CreateThirdParty.ThirdPartyEdge.Node.ID)
	})
}
```

## Authorization (RBAC) testing

Test each role's access to each operation:

```go
t.Run("viewer cannot create", func(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	viewer := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)

	_, err := viewer.Do(createQuery, map[string]any{
		"input": map[string]any{
			"organizationId": viewer.GetOrganizationID().String(),
			"name":           "Test",
		},
	})
	testutil.RequireForbiddenError(t, err, "viewer cannot create")
})
```

## Tenant isolation testing

```go
t.Run("other org cannot access", func(t *testing.T) {
	t.Parallel()
	owner1 := testutil.NewClient(t, testutil.RoleOwner)
	owner2 := testutil.NewClient(t, testutil.RoleOwner)

	thirdPartyID := factory.NewThirdParty(owner1).WithName("ThirdParty").Create()

	var result struct {
		Node *struct{ ID string } `json:"node"`
	}
	err := owner2.Execute(nodeQuery, map[string]any{"id": thirdPartyID}, &result)
	testutil.AssertNodeNotAccessible(t, err, result.Node == nil, "ThirdParty")
})
```

## Assertion helpers

**Pagination:**
```go
testutil.AssertFirstPage(t, edgeCount, pageInfo, expectedCount, expectMore)
testutil.AssertMiddlePage(t, edgeCount, pageInfo, expectedCount)
testutil.AssertLastPage(t, edgeCount, pageInfo, expectedCount, expectPrevious)
```

**Timestamps:**
```go
testutil.AssertTimestampsOnCreate(t, createdAt, updatedAt, beforeCreate)
testutil.AssertTimestampsOnUpdate(t, createdAt, updatedAt, origCreatedAt, origUpdatedAt)
```

**Ordering:**
```go
testutil.AssertOrderedAscending[T](t, values, "fieldName")
testutil.AssertOrderedDescending[T](t, values, "fieldName")
testutil.AssertTimesOrderedDescending(t, times, "createdAt")
```

**Authorization:**
```go
testutil.RequireForbiddenError(t, err, "message")
testutil.RequireErrorCode(t, err, "CODE_NAME", "message")
```

**Optional fields:**
```go
testutil.AssertOptionalStringEqual(t, expected, actual, "fieldName")
```

## Validation testing

Use table-driven tests for validation scenarios:

```go
tests := []struct {
	name              string
	input             map[string]any
	wantErrorContains string
}{
	{name: "missing name", input: map[string]any{}, wantErrorContains: "name"},
	{name: "HTML injection", input: map[string]any{"name": "<script>xss</script>"}, wantErrorContains: "HTML"},
	{name: "control char", input: map[string]any{"name": "Test\x00"}, wantErrorContains: "control"},
}

for _, tt := range tests {
	t.Run(tt.name, func(t *testing.T) {
		input := map[string]any{"organizationId": owner.GetOrganizationID().String()}
		maps.Copy(input, tt.input)

		_, err := owner.Do(query, map[string]any{"input": input})
		require.Error(t, err)
		assert.Contains(t, err.Error(), tt.wantErrorContains)
	})
}
```

## File uploads

```go
err := owner.ExecuteWithFile(
	uploadQuery,
	map[string]any{"input": map[string]any{"thirdPartyId": thirdPartyID, "file": nil}},
	"input.file",
	testutil.UploadFile{
		Filename:    "report.pdf",
		ContentType: "application/pdf",
		Content:     pdfBytes,
	},
	&result,
)
```

## New entity e2e test checklist

1. **File** — `e2e/console/<entity>_test.go`, package `console_test`
2. **CRUD** — create (required fields, all fields), update, delete, get by ID, list
3. **Validation** — required fields, empty strings, HTML injection, control chars, max length, invalid enums
4. **RBAC** — owner/admin/viewer access for create, update, delete, read
5. **Tenant isolation** — cross-org user cannot access resource
6. **Timestamps** — `createdAt == updatedAt` on create, `updatedAt` advances on update
7. **Sub-resolvers** — parent references, child collections
8. **Parallel** — `t.Parallel()` on every test and subtest

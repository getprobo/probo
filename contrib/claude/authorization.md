# Authorization — IAM & Policy

Policy-based authorization in `pkg/iam/` using an evaluation model similar to AWS IAM. Explicit deny > explicit allow > implicit deny.

**Policies are Go code, not database rows.** All policy logic is assembled from Go structs at startup (`pkg/probo/policies.go`, `pkg/iam/iam_policies.go`). The database only stores the `authz_role` enum and membership rows — there is no `policies` or `permissions` table. Never create migrations for policy storage.

## Core concepts

**Policy** — a named collection of statements:
```go
policy.NewPolicy("vendor-crud", "Vendor CRUD",
	policy.Allow(ActionVendorGet, ActionVendorList).WithSID("read-vendors"),
	policy.Deny(ActionVendorDelete).WithSID("deny-vendor-delete"),
).WithDescription("Standard vendor access")
```

**Statement** — a single permission rule with effect (allow/deny), actions, optional resources, and optional conditions.

**Action format** — `SERVICE:RESOURCE:OPERATION` with wildcard support:
```
core:vendor:create      # specific action
core:vendor:*           # all vendor actions
core:*                  # all core actions
*                       # everything
```

## Policy evaluation

The evaluator processes all statements against a request:

1. If any statement explicitly denies → `DecisionDeny`
2. If any statement explicitly allows → `DecisionAllow`
3. No match → `DecisionNoMatch` (implicit deny)

## Authorizer flow

`Authorizer` is the main orchestrator in `pkg/iam/authorizer.go`:

```go
err := iamService.Authorizer.Authorize(ctx, iam.AuthorizeParams{
	Principal:          identityID,    // who
	Resource:           vendorID,      // what
	Action:             probo.ActionVendorGet,  // which action
	ResourceAttributes: map[string]string{},    // optional extra attributes
})
```

The flow:
1. Load organization membership for the resource's organization
2. Load principal attributes (identity + membership role)
3. Load resource attributes via `AuthorizationAttributes()` on the entity
4. Build policies: identity-scoped + role-specific
5. Evaluate all policies
6. Return `ErrInsufficientPermissions` if no allow match

## PolicySet

Policies are organized into identity-scoped (applied to all authenticated users) and role-based:

```go
ps := iam.NewPolicySet().
	AddRolePolicy("OWNER", OwnerPolicy).
	AddRolePolicy("ADMIN", AdminPolicy).
	AddRolePolicy("VIEWER", ViewerPolicy).
	AddIdentityScopedPolicy(SelfManagePolicy)
```

Register during service initialization:
```go
iamService.Authorizer.RegisterPolicySet(ProboPolicySet())
```

## Conditions (attribute-based access control)

Conditions constrain when a statement applies. All conditions must be satisfied.

```go
// Users can only access resources in their organization
organizationCondition := policy.Equals("principal.organization_id", "resource.organization_id")

policy.Allow(ActionVendorGet).
	WithSID("view-vendor").
	When(organizationCondition)
```

| Operator | Purpose |
|----------|---------|
| `policy.Equals(key, value)` | Key equals value |
| `policy.NotEquals(key, value)` | Key does not equal value |
| `policy.In(key, value)` | Key in list (supports comma-separated DB fields) |
| `policy.NotIn(key, value)` | Key not in list |

Key paths use `principal.ATTR` or `resource.ATTR` (e.g., `principal.organization_id`, `resource.source`).

## AuthorizationAttributer interface

Resources that support authorization must implement this interface in `pkg/coredata/`:

```go
func (v *Vendor) AuthorizationAttributes(ctx context.Context, conn pg.Conn) (map[string]string, error) {
	q := `SELECT organization_id FROM vendors WHERE id = $1 LIMIT 1;`
	var organizationID gid.GID
	if err := conn.QueryRow(ctx, q, v.ID).Scan(&organizationID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrResourceNotFound
		}
		return nil, fmt.Errorf("cannot query vendor authorization attributes: %w", err)
	}
	return map[string]string{"organization_id": organizationID.String()}, nil
}
```

The returned map provides attributes for condition evaluation (e.g., `resource.organization_id`).

## Error types

```go
var (
	ErrInsufficientPermissions // access denied
	ErrAssumptionRequired      // session assumption needed
	ErrUnsupportedPrincipalType // principal is not an Identity
)
```

## Integration in resolvers

**GraphQL resolvers** use `AuthorizeFunc` from `pkg/server/api/authz/`:
```go
if err := authorize(ctx, vendorID, probo.ActionVendorGet); err != nil {
	return nil, err
}
```

**MCP resolvers** use `MustAuthorize` which panics (caught by middleware):
```go
r.MustAuthorize(ctx, input.ID, probo.ActionVendorGet)
```

## File locations

| What | File |
|------|------|
| Product action constants (`core:*`) | `pkg/probo/actions.go` |
| IAM action constants (`iam:*`) | `pkg/iam/iam_actions.go` |
| Product role policies (`ProboPolicySet`) | `pkg/probo/policies.go` |
| IAM role policies (`IAMPolicySet`) | `pkg/iam/iam_policies.go` |
| Authorizer + `AuthorizationAttributer` | `pkg/iam/authorizer.go` |
| PolicySet registration | `pkg/iam/policy_set.go` |
| GraphQL authz helper | `pkg/server/api/authz/authorization.go` |
| MCP authz + recovery | `pkg/server/api/mcp/v1/resolver.go`, `mcputils/recovery.go` |

## Action constants

IAM actions live in `pkg/iam/iam_actions.go`, probo actions in `pkg/probo/actions.go`. Follow the naming pattern:

```go
const (
	ActionVendorGet    = "core:vendor:get"
	ActionVendorList   = "core:vendor:list"
	ActionVendorCreate = "core:vendor:create"
	ActionVendorUpdate = "core:vendor:update"
	ActionVendorDelete = "core:vendor:delete"
)
```

## Built-in role policies

| Role | Access level |
|------|-------------|
| `OWNER` | Full access to all features including org management |
| `ADMIN` | Full access to core features, restricted org management |
| `VIEWER` | Read-only access to most entities |
| `AUDITOR` | Read-only, excludes internal/employee content |
| `EMPLOYEE` | Can sign documents and view internal content |

## New entity IAM wiring

When adding a new entity that needs authorization:

1. **Action constants** — add `core:<entity>:<verb>` constants in `pkg/probo/actions.go` (get, list, create, update, delete)
2. **Role policies** — wire actions into the appropriate role policies in `pkg/probo/policies.go` (`OwnerPolicy`, `AdminPolicy`, `ViewerPolicy`, etc.) with `organization_id` condition
3. **`AuthorizationAttributes`** — implement on the `coredata` entity struct, returning at minimum `{"organization_id": ...}` (use the denormalized `OrganizationID` field — see coredata doc)
4. **Entity type registry** — register in `pkg/coredata/entity_type_reg.go` and `NewEntityFromID` so the authorizer can construct the entity from its GID
5. **Resolver calls** — add `r.authorize(ctx, id, probo.ActionEntityGet)` in GraphQL resolvers and `r.MustAuthorize(ctx, id, probo.ActionEntityGet)` in MCP resolvers

## Key patterns

- **Always use `organization_id` condition** — most policies scope access to the principal's organization
- **SID every statement** — `.WithSID("description")` for debugging
- **Explicit denies for restrictions** — even if allow would match, deny takes precedence
- **Identity-scoped for self-management** — cross-org permissions like managing own profile
- **Role-based for org features** — CRUD operations on domain entities

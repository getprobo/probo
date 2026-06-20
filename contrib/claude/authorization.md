# Authorization — IAM & Policy

Policy-based authorization in `pkg/iam/` using an evaluation model similar to AWS IAM. Explicit deny > explicit allow > implicit deny.

**Policies are Go code, not database rows.** All policy logic is assembled from Go structs at startup (`pkg/probo/policies.go`, `pkg/iam/iam_policies.go`). The database only stores the `authz_role` enum and membership rows — there is no `policies` or `permissions` table. Never create migrations for policy storage.

## Core concepts

**Policy** — a named collection of statements:
```go
policy.NewPolicy("thirdParty-crud", "ThirdParty CRUD",
	policy.Allow(ActionThirdPartyGet, ActionThirdPartyList).WithSID("read-thirdParties"),
	policy.Deny(ActionThirdPartyDelete).WithSID("deny-thirdParty-delete"),
).WithDescription("Standard third party access")
```

**Statement** — a single permission rule with effect (allow/deny), actions, optional resources, and optional conditions.

**Action format** — `SERVICE:RESOURCE:OPERATION` with wildcard support:
```
core:thirdParty:create      # specific action
core:thirdParty:*           # all third party actions
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
scope, err := iamService.Authorizer.Authorize(ctx, iam.AuthorizeParams{
	Principal:          identityID,    // who
	Resource:           thirdPartyID,      // what
	Action:             probo.ActionThirdPartyGet,  // which action
	ResourceAttributes: map[string]string{},    // optional extra attributes
})
```

The flow:
1. Load organization membership for the resource's organization
2. Load principal attributes (identity + membership role)
3. Load resource attributes via `AuthorizationAttributes()` on the entity
4. Build policies: identity-scoped + role-specific
5. Evaluate all policies
6. Return an authorization scope (`*coredata.Scope`) for downstream data access
7. Return `ErrInsufficientPermissions` if no allow match

## Batch authorization

Use batch authorization when a caller needs all-or-nothing authorization across
multiple resources for the same action:

```go
scope, err := iamService.Authorizer.AuthorizeBatch(ctx, iam.AuthorizeBatchParams{
	Principal: identityID,
	Action:    probo.ActionTaskDelete,
	Resources: taskIDs, // all resources must have same entity type + organization
})
```

Batch semantics:
- **All-or-nothing** — the first denied resource returns `ErrInsufficientPermissions`
- **Single-entity-type batch** — mixed entity types return `ErrMixedEntityTypeBatch`
- **Single-organization batch** — mixed or missing `organization_id` attributes return `ErrMixedOrganizationBatch`
- **Empty resource list** returns `ErrEmptyResourceBatch`
- **Batch attributes are required** — each resource type in `AuthorizeBatch` must implement batch attributes loading or it returns `ErrBatchAuthorizationUnsupportedResourceType`
- **Shared `ResourceAttributes` map** is applied to every resource in the batch
- **Audit logs** are written per resource only when all resources are authorized (`DryRun` skips logs)

GraphQL wrappers can use `authz.NewBatchAuthorizeFunc(...)` with
`authz.WithBatchAttr`, `authz.WithBatchDryRun`, and
`authz.WithBatchSkipAssumptionCheck`.

MCP resolvers can use `Resolver.AuthorizeBatch(ctx, resourceIDs, action)` for
the same behavior and error mapping as single-resource authorization.

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

policy.Allow(ActionThirdPartyGet).
	WithSID("view-thirdParty").
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
func (v *ThirdParty) AuthorizationAttributes(
	ctx context.Context,
	conn pg.Querier,
	resourceIDs []gid.GID,
) (map[gid.GID]map[string]string, error) {
	q := `SELECT id, organization_id FROM third_parties WHERE id = ANY(@resource_ids::text[])`

	rows, err := conn.Query(ctx, q, pgx.StrictNamedArgs{"resource_ids": resourceIDs})
	if err != nil {
		return nil, fmt.Errorf("cannot query third party authorization attributes: %w", err)
	}
	defer rows.Close()

	attrsByID := make(map[gid.GID]map[string]string)
	for rows.Next() {
		var id, organizationID gid.GID
		if err := rows.Scan(&id, &organizationID); err != nil {
			return nil, fmt.Errorf("cannot scan third party authorization attributes: %w", err)
		}
		attrsByID[id] = map[string]string{"organization_id": organizationID.String()}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("cannot iterate third party authorization attributes: %w", err)
	}

	return attrsByID, nil
}
```

The returned map provides attributes for condition evaluation (e.g., `resource.organization_id`).
`AuthorizationAttributes` implementers should assume caller-side preconditions:
- `resourceIDs` is non-empty
- `resourceIDs` is deduplicated
- in `AuthorizeBatch`, all resources are the same entity type

Implementers should return only found rows keyed by id. Missing resources are handled by caller-side per-resource existence checks.

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
scope, err := r.authorize(ctx, thirdPartyID, probo.ActionThirdPartyGet)
if err != nil {
	return nil, err
}
```

**MCP resolvers** use `Authorize` and return early on error:
```go
scope, err := r.Authorize(ctx, input.ID, probo.ActionThirdPartyGet)
if err != nil {
	return nil, types.GetThirdPartyOutput{}, err
}
```

### Always take `scope` from `authorize` — never reconstruct it

`authorize` (and `Authorize` in MCP) returns a `*coredata.Scope` that has been
resolved from the resource's `organization_id` attribute. Pass that scope
straight to the service/coredata layer instead of building a new one with
`coredata.NewScopeFromObjectID(...)` after the authorize call.

The two are not strictly identical: `NewScopeFromObjectID(id)` only reads the
tenant component of the GID, while the authorizer derives the scope from the
loaded resource attributes (and may be extended to compute it differently in
the future). Reconstructing the scope from the GID bypasses that and silently
drifts when the resource lookup changes.

```go
// GOOD — scope comes from authorize, fed straight to the service
scope, err := r.authorize(ctx, obj.ID, probo.ActionThirdPartyList)
if err != nil {
	return nil, err
}

thirdPartyIDs, err := r.cookieBanner.LoadDistinctThirdPartyIDsByCookieBannerID(ctx, scope, obj.ID)

// BAD — authorize discards scope, then we rebuild it from the same GID
if _, err := r.authorize(ctx, obj.ID, probo.ActionThirdPartyList); err != nil {
	return nil, err
}

scope := coredata.NewScopeFromObjectID(obj.ID)
thirdPartyIDs, err := r.cookieBanner.LoadDistinctThirdPartyIDsByCookieBannerID(ctx, scope, obj.ID)
```

The only time it is acceptable to write `if _, err := r.authorize(...)` is when
**no downstream call needs a scope** — typically authorize calls against the
caller's `identity.ID` for global / cross-tenant catalogs (e.g.
`ActionCommonThirdPartyList`, `ActionCommonThirdPartyGet`) whose service
methods are unscoped. In that case the returned scope would be derived from
the identity (a nil-tenant principal) and is useless to the caller, so
discarding it with `_` is correct:

```go
// GOOD — global catalog, downstream is unscoped
identity := authn.IdentityFromContext(ctx)
if _, err := r.authorize(ctx, identity.ID, probo.ActionCommonThirdPartyList); err != nil {
	return nil, err
}

parties, err := r.thirdParty.Search(ctx, name) // no scope argument
```

For batch authorization, the same rule applies to `r.batchAuthorize` (GraphQL)
and `r.AuthorizeBatch` (MCP) — keep the returned scope and pass it down.

## File locations

| What | File |
|------|------|
| Product action constants (`core:*`) | `pkg/probo/actions.go` |
| IAM action constants (`iam:*`) | `pkg/iam/iam_actions.go` |
| Product role policies (`ProboPolicySet`) | `pkg/probo/policies.go` |
| Per-service policy sets (e.g. `accessreview.PolicySet`, `agentrun.PolicySet`) | `pkg/<service>/actions.go`, `pkg/<service>/policies.go` |
| IAM role policies (`IAMPolicySet`) | `pkg/iam/iam_policies.go` |
| Authorizer + `AuthorizationAttributer` | `pkg/iam/authorizer.go` |
| PolicySet registration | `pkg/iam/policy_set.go` |
| OAuth2 scope registry (`oauth2scope.Registry`) | `pkg/iam/oauth2scope/registry.go` |
| OAuth2 scope constants (per domain) | `pkg/<service>/oauth2_scopes.go` |
| OAuth2 discovery + request context | `pkg/iam/oauth2/` |
| GraphQL authz helper | `pkg/server/api/authz/authorization.go` |
| MCP authz + recovery | `pkg/server/api/mcp/v1/resolver.go`, `mcputils/recovery.go` |

## Action constants

IAM actions live in `pkg/iam/iam_actions.go`, probo actions in `pkg/probo/actions.go`. Follow the naming pattern:

```go
const (
	ActionThirdPartyGet    = "core:thirdParty:get"
	ActionThirdPartyList   = "core:thirdParty:list"
	ActionThirdPartyCreate = "core:thirdParty:create"
	ActionThirdPartyUpdate = "core:thirdParty:update"
	ActionThirdPartyDelete = "core:thirdParty:delete"
)
```

## OAuth2 API scopes

OAuth2 scopes for API access are defined as `coredata.OAuth2Scope` constants in each owning package (for example [`pkg/probo/oauth2_scopes.go`](../../pkg/probo/oauth2_scopes.go), [`pkg/iam/oauth2_scopes.go`](../../pkg/iam/oauth2_scopes.go)). [`pkg/coredata/oauth2_scope.go`](../../pkg/coredata/oauth2_scope.go) defines the persistence type. Standard OIDC scopes live in [`pkg/iam/oauth2/scope.go`](../../pkg/iam/oauth2/scope.go). Register scope sets with `Authorizer.RegisterScopes`.

**Format:**

- Read: `v1:<namespace>:read` (e.g. `v1:privacy:read`, `v1:document:read`, `v1:org:read`)
- Write / full: `v1:<namespace>` without the `:read` suffix (e.g. `v1:org`, `v1:connector`, `v1:agent`)

Scopes are namespace- or product-level only — no resource segments (e.g. `v1:privacy:dpia` is not supported).

**Discovery:**

- Authorization server (RFC 8414): `scopes_supported` on `/.well-known/oauth-authorization-server` lists OIDC + all API scopes; `protected_resources` links to the resource metadata document
- Protected resource (RFC 9728): `scopes_supported` on `/.well-known/oauth-protected-resource` lists `openid` plus API scopes

**Enforcement:** OAuth2 bearer-token requests carry the validated access token on the request context (`pkg/iam/oauth2/request_context.go`). Before IAM policy evaluation, `iam.Authorizer` checks registered `oauth2scope.Registry` mappings via `Registry.Allows`. Each domain package exports `OAuth2ScopeMappings` in its `oauth2_scopes.go`; `probod` registers all domain mappings on the shared registry before `iam.NewService`. The check uses explicit scope→action lists — no `:read` / `:get` heuristics at enforcement time. Session, personal API key, and SCIM auth skip the check (no access token on context). Unmapped IAM actions **deny** OAuth requests (fail closed). Enforcement reads scopes from the access token directly.

Add new namespace-level scope constants in the owning package's `oauth2_scopes.go`, map their IAM actions in that package's `OAuth2ScopeMappings`, and add the mapping to `probod` wiring alongside the other domain registrations. Write scopes are registered only when their mutating IAM actions are mapped.

**Well-known Probo CLI client:** `iam_oauth2_clients` scopes for `AAAAAAAAAAAASwAAAAAAAAAAcHJiY2xp` must match `CLIClientScopes` in `pkg/cli/config/config.go` (requested by `prb auth login`). When adding API scopes, update the client migration, `CLIClientScopes`, and scope registration together.

### Personal OAuth2 access tokens

Manual bearer tokens created from the console are stored in `iam_oauth2_access_tokens` with a `NULL` `client_id` and are scoped to the creating identity. They are managed via Connect GraphQL on the signed-in user's `Identity`, similar to personal API keys. IAM actions:

| Action | Purpose |
|--------|---------|
| `iam:oauth2-access-token:create` | Create a manual token |
| `iam:oauth2-access-token:list` | List your tokens |
| `iam:oauth2-access-token:get` | Read token metadata |
| `iam:oauth2-access-token:delete` | Revoke (delete) a token |

**Policies:** `IAMSelfManageIdentityPolicy` allows listing on your identity; `IAMSelfManageOAuth2AccessTokenPolicy` allows create/get/delete when `principal.id == resource.identity_id`. **OAuth2 scope gate:** create/list/get/delete map to `v1:iam:read` / `v1:iam` in `pkg/iam/oauth2_scopes.go`.

## Built-in role policies

| Role | Access level |
|------|-------------|
| `OWNER` | Full access to all features including org management |
| `ADMIN` | Full access to core features, restricted org management |
| `VIEWER` | Read-only access to most entities |
| `AUDITOR` | Read-only, excludes internal/employee content |
| `EMPLOYEE` | Can sign documents and view internal content |

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
5. **Resolver calls** — add `scope, err := r.authorize(ctx, id, probo.ActionEntityGet)` in GraphQL resolvers and `scope, err := r.Authorize(ctx, id, probo.ActionEntityGet)` in MCP resolvers, then pass `scope` to services

## Decision logging

Every authorization evaluation (allow and deny) emits a structured `authz decision`
log line through the authorizer logger with opaque IDs only:

- `effect` — `allow`, `deny`, `no_match`, or `error`
- `action`, `principal_id`, `resource_id`
- `policy_id` — statement SID when available
- `reason` — human-readable explanation for operators (never returned to clients)
- `latency` — PDP evaluation duration

Audit log entries remain **allow-only**. Denials are visible in application logs,
not the product audit trail.

## Key patterns

- **Always use `organization_id` condition** — most policies scope access to the principal's organization
- **SID every statement** — `.WithSID("description")` for debugging
- **Explicit denies for restrictions** — even if allow would match, deny takes precedence
- **Identity-scoped for self-management** — cross-org permissions like managing own profile
- **Role-based for org features** — CRUD operations on domain entities

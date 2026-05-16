# Probo — Go Backend — pkg/server/api/* (console, trust, connect, mcp, cookiebanner)

**Purpose.** All HTTP and GraphQL/MCP API surfaces. Five sibling
packages, each with its own router, schema, and resolver layer:

| Surface | Package | Generator | Mounted at |
| --- | --- | --- | --- |
| Console (authenticated) | `pkg/server/api/console/v1` | gqlgen | `/api/console/v1/graphql` |
| Trust Center (public) | `pkg/server/api/trust/v1` | gqlgen | served by separate Trust HTTP/HTTPS server |
| Connect (sign-up, invite) | `pkg/server/api/connect/v1` | gqlgen | `/api/connect/v1/graphql` |
| MCP (model context protocol) | `pkg/server/api/mcp/v1` | mcpgen | `/api/mcp/v1` |
| Cookie banner (public) | `pkg/server/api/cookiebanner` | none | `/api/cookiebanner/...` |

Plus the shared error helpers in `pkg/server/gqlutils` (typed error
constructors with extensions.code) and the cross-resolver
`pkg/server/api/authz` (`AuthorizeFunc`).

> See [patterns.md § 7 GraphQL resolver shape](../patterns.md#7-graphql-resolver-shape)
> and [§ 8 MCP resolver shape](../patterns.md#8-mcp-resolver-shape).

**Key files.**

- `console/v1/resolver.go` — `Resolver` struct, `r.authorize`,
  `r.ProboService`, chi router (authn middleware, dataloader middleware,
  /graphql, connector OAuth2 routes).
- `console/v1/graphql/*.graphql` — schema (`base.graphql` defines
  directives + `Query` + empty `Mutation`; one file per entity
  uses `extend type Mutation`).
- `console/v1/types/*.go` — hand-written `types.NewVendor(coredata)` etc.
  + `OrderBy[T]` generic, Connection/Edge structs with
  `Resolver`+`ParentID` fields for totalCount.
- `console/v1/dataloader/dataloader.go` — per-request batch loaders.
- `console/v1/vendor_resolvers.go` — canonical resolver file (one per
  schema file by gqlgen `follow-schema` layout).
- `mcp/v1/specification.yaml` — MCP tool catalog (regenerate with
  `go generate ./pkg/server/api/mcp/v1`).
- `mcp/v1/types/*.go` — one file per entity, builders + Omittable
  helpers.
- `gqlutils/errors.go` — `NotFound`, `Forbidden`, `Conflict`, `Invalid`,
  `Unauthenticated`, `Internal`, `Unavailable`, `AlreadyAuthenticated`.

**How to extend (a new GraphQL field/mutation).**

1. Edit the `.graphql` schema file (use `extend type Mutation` only;
   never `extend type` anything else).
2. Run `go generate ./pkg/server/api/console/v1` (or the relevant
   surface).
3. Implement the new resolver method in `<entity>_resolvers.go`:
   `r.authorize` first line, error switch with mandatory `default:`.
4. Add the matching MCP tool in `specification.yaml`, regenerate, and
   implement the body next to the generated stub. Use `MustAuthorize`.
5. Add the matching `prb` CLI verb file under
   `pkg/cmd/<resource>/<verb>/<verb>.go`.
6. Add the n8n action — see [shared.md § 3](../../shared.md#3-the-four-surface-api-rule).
7. Add an e2e test (`e2e/console/<entity>_test.go` or `e2e/mcp/...`).

**Top pitfalls.**

- Missing `default:` in resolver error switches — leaks internal errors.
- Forgetting `@goModel` on a Connection type — `totalCount` returns 0
  or panics.
- Adding the GraphQL field but skipping MCP/CLI/n8n — frequency-2
  reviewer rule (PR #1132).
- Direct service calls in field resolvers instead of DataLoader — N+1.
- `extend type X` for `X` other than `Mutation` — gqlgen mis-routes the
  resolver.
- Surfacing GraphQL fields as non-null (`!`) when the resolver can fail —
  use Relay `@required` instead (frequency-4 reviewer rule).

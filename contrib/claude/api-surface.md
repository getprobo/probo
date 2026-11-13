# API Surface Rules

Every feature must be exposed through **all four interfaces**: GraphQL, MCP, CLI, and n8n. When adding a new endpoint or editing an existing type, keep all four in sync:

- **GraphQL** — `pkg/server/api/console/v1/graphql/*.graphql` (+ codegen) — see [`contrib/claude/graphql.md`](graphql.md)
- **MCP** — `pkg/server/api/mcp/v1/` (+ codegen) — see [`contrib/claude/mcp.md`](mcp.md)
- **CLI** — `pkg/cmd/` — see [`contrib/claude/cli.md`](cli.md)
- **n8n** — `packages/n8n-node/` — see [`contrib/claude/n8n.md`](n8n.md)

If you add a mutation in GraphQL, add the corresponding MCP tool, CLI command, and n8n node. If you rename or change a type, update it everywhere.

Every new Go API endpoint must have end-to-end tests in `e2e/`.

## Exception — no credentials on MCP

The MCP API does not accept credentials. Do not add an MCP tool whose input is a customer secret or a connector binding (API key, client secret, workload-identity role, or the same material on an organization connector). Connector create stays on GraphQL, the CLI, and n8n. An agent connects the provider in the console, then uses the connector id with the remaining MCP tools.

## Error handling — never leak internal details

By default every error returned to the end user **must be an opaque internal error**. Only errors that are explicitly matched and mapped to a known category may surface a meaningful message. Unrecognized or unexpected errors are always replaced with a generic "internal server error" response — never expose stack traces, SQL errors, file paths, or any implementation detail.

### Allowed user-facing error categories

| Category | GraphQL helper | HTTP helper | When to use |
|---|---|---|---|
| Not found | `gqlx.NotFound` / `NotFoundf` | `jsonx.RenderNotFound` | Resource does not exist or is not visible to the caller |
| Forbidden | `gqlx.Forbidden` / `Forbiddenf` | `jsonx.RenderForbidden` | Caller lacks permission (after authentication) |
| Invalid | `gqlx.Invalid` / `Invalidf` / `InvalidValidationErrors` | `jsonx.RenderBadRequest` | Validation failure on user-supplied input |
| Conflict | `gqlx.Conflict` / `Conflictf` | — | Unique constraint or state conflict |
| Unauthenticated | `gqlx.Unauthenticated` / `Unauthenticatedf` | `jsonx.RenderUnauthorized` | Missing or expired credentials |

### Catch-all is always internal

Any error that does **not** match one of the categories above must be returned as:

- **GraphQL** — `gqlx.Internal(ctx)` (fixed generic message, no error details)
- **HTTP** — `jsonx.RenderInternalServerError(w)` (fixed 500 body, no error details)
- **MCP** — return a generic "internal error" string; never forward `err.Error()`

Log the original error server-side (with request/trace IDs) so it can be investigated, but **never include it in the response**.

### Pattern in resolvers

A GraphQL resolver must not return `fmt.Errorf`, `errors.New`, or `graphql.ErrorOnPath`. Those have no `extensions.code`, and the tracing logger records them as `UNKNOWN` at error level. Do not build `&gqlerror.Error` by hand either: it can set a code, but it skips the helpers. Return a `gqlx` helper. `authorize` already returns one; pass that error through.

```go
result, err := s.doSomething(ctx, req)
if err != nil {
    switch {
    case errors.Is(err, probo.ErrNotFound):
        return nil, gqlx.NotFoundf(ctx, "thing %q not found", id)
    case errors.Is(err, probo.ErrConflict):
        return nil, gqlx.Conflictf(ctx, "thing already exists")
    default:
        logger.ErrorCtx(ctx, "cannot do something", log.Error(err))
        return nil, gqlx.Internal(ctx)
    }
}
```

The `default` branch must **always** be present and must **always** return the generic internal error.

`gqlx.MapError` is that switch for the common case: `coredata.ErrResourceNotFound` becomes `NOT_FOUND`, `validator.ValidationErrors` become `INVALID`, and every other error is logged and returned as `INTERNAL`. `NOT_FOUND` shows `err.Error()`, so wrap the sentinel only with client-safe text. Return the sentinel unwrapped, or `gqlx.NotFoundf(ctx, "resource not found")`, when the prefix could reveal internals. When not-found is already handled, log and return `gqlx.Internal(ctx)` for what remains. A field the schema models as optional (no alias, no generated document) returns `nil, nil` instead of an error.

A new resolver-owned code needs a helper in `pkg/server/gqlx/errors.go` and an entry in `safeGraphQLErrorCodes` (`pkg/server/gqlx/tracing.go`). A gqlgen protocol code such as `GRAPHQL_PARSE_FAILED` or `COMPLEXITY_LIMIT_EXCEEDED` only needs the allowlist entry. A code missing from that list is logged as `UNKNOWN` at error level. `INTERNAL` stays at error. Every other allowlisted code logs at info. See [`graphql.md`](graphql.md).

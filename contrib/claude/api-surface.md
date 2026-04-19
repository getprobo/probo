# API Surface Rules

Every feature must be exposed through **all four interfaces**: GraphQL, MCP, CLI, and n8n. When adding a new endpoint or editing an existing type, keep all four in sync:

- **GraphQL** — `pkg/server/api/console/v1/graphql/*.graphql` (+ codegen) — see [`contrib/claude/graphql.md`](graphql.md)
- **MCP** — `pkg/server/api/mcp/v1/` (+ codegen) — see [`contrib/claude/mcp.md`](mcp.md)
- **CLI** — `pkg/cmd/` — see [`contrib/claude/cli.md`](cli.md)
- **n8n** — `packages/n8n-node/` — see [`contrib/claude/n8n.md`](n8n.md)

If you add a mutation in GraphQL, add the corresponding MCP tool, CLI command, and n8n node. If you rename or change a type, update it everywhere.

Every new Go API endpoint must have end-to-end tests in `e2e/`.

## Error handling — never leak internal details

By default every error returned to the end user **must be an opaque internal error**. Only errors that are explicitly matched and mapped to a known category may surface a meaningful message. Unrecognized or unexpected errors are always replaced with a generic "internal server error" response — never expose stack traces, SQL errors, file paths, or any implementation detail.

### Allowed user-facing error categories

| Category | GraphQL helper | HTTP helper | When to use |
|---|---|---|---|
| Not found | `gqlutils.NotFound` / `NotFoundf` | `jsonutil.RenderNotFound` | Resource does not exist or is not visible to the caller |
| Forbidden | `gqlutils.Forbidden` / `Forbiddenf` | `jsonutil.RenderForbidden` | Caller lacks permission (after authentication) |
| Invalid | `gqlutils.Invalid` / `Invalidf` / `InvalidValidationErrors` | `jsonutil.RenderBadRequest` | Validation failure on user-supplied input |
| Conflict | `gqlutils.Conflict` / `Conflictf` | — | Unique constraint or state conflict |
| Unauthenticated | `gqlutils.Unauthenticated` / `Unauthenticatedf` | — | Missing or expired credentials |

### Catch-all is always internal

Any error that does **not** match one of the categories above must be returned as:

- **GraphQL** — `gqlutils.Internal(ctx)` (fixed generic message, no error details)
- **HTTP** — `jsonutil.RenderInternalServerError(w)` (fixed 500 body, no error details)
- **MCP** — return a generic "internal error" string; never forward `err.Error()`

Log the original error server-side (with request/trace IDs) so it can be investigated, but **never include it in the response**.

### Pattern in resolvers

```go
result, err := s.doSomething(ctx, req)
if err != nil {
    switch {
    case errors.Is(err, probo.ErrNotFound):
        return nil, gqlutils.NotFoundf(ctx, "thing %q not found", id)
    case errors.Is(err, probo.ErrConflict):
        return nil, gqlutils.Conflictf(ctx, "thing already exists")
    default:
        logger.ErrorCtx(ctx, "cannot do something", log.Error(err))
        return nil, gqlutils.Internal(ctx)
    }
}
```

The `default` branch must **always** be present and must **always** return the generic internal error.

---
name: potion-implementer
description: >
  General implementation agent for Probo. Creates new code following project
  patterns across both Go backend and TypeScript frontend. This agent
  delegates from the implement skill for tasks that benefit from a fresh
  context window.
tools: Read, Write, Edit, Glob, Grep, Bash
model: inherit
color: green
effort: high
maxTurns: 25
---

# Probo Implementer

You implement features in Probo following its established patterns.

## Before writing code

1. Read `.claude/guidelines/shared.md` for cross-stack conventions
2. Determine which stack you are working in using the module maps below
3. Read the relevant stack guidelines:
   - Go Backend: `.claude/guidelines/go-backend/patterns.md`, `.claude/guidelines/go-backend/conventions.md`
   - TypeScript Frontend: `.claude/guidelines/typescript-frontend/patterns.md`, `.claude/guidelines/typescript-frontend/conventions.md`
4. Read the canonical example for that module
5. Check for existing similar code (Grep) -- avoid reinventing

## Module map -- Go Backend

| Module | Path | Purpose |
|--------|------|---------|
| pkg/coredata | `pkg/coredata/` | All raw SQL, entity types, filters, migrations |
| pkg/probo | `pkg/probo/` | Core business logic (40+ sub-services) |
| pkg/iam | `pkg/iam/` | IAM, auth, policy evaluation |
| pkg/server/api/console/v1 | `pkg/server/api/console/v1/` | Console GraphQL API |
| pkg/server/api/mcp/v1 | `pkg/server/api/mcp/v1/` | MCP API |
| pkg/cmd | `pkg/cmd/` | CLI commands |
| e2e | `e2e/` | End-to-end tests |

## Module map -- TypeScript Frontend

| Module | Path | Purpose |
|--------|------|---------|
| apps/console | `apps/console/` | Admin dashboard SPA |
| apps/trust | `apps/trust/` | Public trust center SPA |
| packages/ui | `packages/ui/` | Shared design system |
| packages/helpers | `packages/helpers/` | Domain formatters |

## Key patterns -- Go Backend

- **Two-level service tree:** `Service` (global) -> `WithTenant(tenantID)` -> `TenantService`
- **Request struct + Validate():** every mutating method takes `*Request` with fluent validation
- **All SQL in pkg/coredata only:** no other package may contain SQL
- **pgx.StrictNamedArgs:** never NamedArgs
- **Error wrapping:** `fmt.Errorf("cannot <action>: %w", err)` -- never "failed to"
- **Scoper:** entity structs have no TenantID field; tenant isolation via Scoper
- **Functional options:** `With*` functions for optional config
- **Grouped declarations:** `type ()`, `const ()`, `var ()` blocks
- **One arg per line:** fully inline or fully expanded, never mixed

## Key patterns -- TypeScript Frontend

- **Relay colocated operations:** all GraphQL in component files
- **Loader component:** `useQueryLoader` + `useEffect` (not deprecated `withQueryRef`)
- **tv() variants:** tailwind-variants for component styling
- **useMutation + useToast:** for mutations with user feedback
- **Permission fragments:** `canX: permission(action: "core:entity:verb")`
- **Named exports:** everywhere except lazy-loaded pages (default export)
- **Import ordering:** external, aliased (#/), relative

## Error handling

### Go Backend
```go
// Wrap errors with "cannot" prefix
return nil, fmt.Errorf("cannot load widget: %w", err)

// Map coredata sentinels to domain errors
if errors.Is(err, coredata.ErrResourceNotFound) {
    return nil, NewErrWidgetNotFound(id)
}

// GraphQL resolvers: log then return gqlutils error
r.logger.ErrorCtx(ctx, "cannot load widget", log.Error(err))
return nil, gqlutils.Internal(ctx)
```

### TypeScript Frontend
```tsx
// Mutations with onCompleted/onError callbacks
const [doAction, isLoading] = useMutation<ActionMutation>(mutation);
doAction({
  variables: { input },
  onCompleted() {
    toast({ title: __("Success"), variant: "success" });
  },
  onError(error) {
    toast({ title: __("Error"), description: formatError(__("Failed"), error as GraphQLError), variant: "error" });
  },
});
```

## File placement

### Go Backend
- Entity data access: `pkg/coredata/<entity>.go` + `_filter.go` + `_order_field.go`
- Business logic: `pkg/probo/<entity>_service.go`
- GraphQL schema: `pkg/server/api/console/v1/schema.graphql`
- GraphQL resolver: `pkg/server/api/console/v1/v1_resolver.go` (or per-type resolver files)
- MCP spec: `pkg/server/api/mcp/v1/specification.yaml`
- CLI: `pkg/cmd/<resource>/<verb>/<verb>.go`
- E2E tests: `e2e/console/<entity>_test.go`
- Migrations: `pkg/coredata/migrations/<YYYYMMDDTHHMMSSZ>.sql`

### TypeScript Frontend
- Pages: `apps/console/src/pages/organizations/<domain>/<Page>.tsx`
- Loaders: `apps/console/src/pages/organizations/<domain>/<Page>Loader.tsx`
- Routes: `apps/console/src/routes/<domain>Routes.ts`
- Dialogs: `apps/console/src/pages/organizations/<domain>/dialogs/<Dialog>.tsx`
- UI atoms: `packages/ui/src/Atoms/<Name>/<Name>.tsx`
- UI molecules: `packages/ui/src/Molecules/<Name>/<Name>.tsx`
- Helpers: `packages/helpers/src/<domain>.ts`

## Testing

### Go Backend
- Framework: testify (`require` for fatal, `assert` for non-fatal)
- Naming: `TestEntity_Operation`, subtests with lowercase descriptions
- Run: `make test MODULE=./pkg/foo` or `make test-e2e`
- Always: `t.Parallel()` at top-level AND every subtest
- E2E: factory builders, RBAC testing, tenant isolation

### TypeScript Frontend
- Storybook: stories for UI atoms/molecules in `packages/ui`
- Vitest: unit tests for helpers in `packages/helpers`
- Run: `cd packages/ui && npm run storybook` or `cd packages/helpers && npx vitest run`

## After writing code

- [ ] Tests written and passing
- [ ] Error handling follows the project pattern
- [ ] File naming matches conventions
- [ ] No debug prints or temporary code left behind
- [ ] Types properly defined (no untyped escape hatches)
- [ ] ISC license header on all new files with current year
- [ ] Three-interface sync: if new feature, GraphQL + MCP + CLI all present

## Common mistakes

- **pgx.NamedArgs** -- always use `pgx.StrictNamedArgs` (approval blocker)
- **"failed to" errors** -- always use "cannot" prefix (approval blocker)
- **Missing t.Parallel()** -- required at all test levels (approval blocker)
- **withQueryRef** -- use Loader component pattern instead (approval blocker)
- **Mixed multiline style** -- one arg per line or all inline, never mixed (approval blocker)
- **Wrong Relay environment** -- IAM pages use `iamEnvironment`, everything else uses `coreEnvironment`
- **Editing generated files** -- never edit `schema/schema.go`, `types/types.go`, or `server/server.go`
- **Missing entity type registration** -- always add `NewEntityFromID` switch case for new entities

# Probo -- TypeScript Frontend -- apps/console

> Module-specific notes for `apps/console` (`@probo/console`)
> For stack-wide patterns, see [patterns.md](../patterns.md) and [conventions.md](../conventions.md)

## Purpose

The primary admin dashboard for compliance managers. A React 19 + Vite SPA (port 5173) that communicates with two GraphQL endpoints over Relay. It covers all compliance domains: risks, measures, documents, vendors, frameworks, audits, assets, data, tasks, processing activities, rights requests, snapshots, obligations, findings, and a public compliance/trust-center page manager.

## Dual Relay Environments

This is the only app with two Relay environments. See [patterns.md -- Relay Environments](../patterns.md#relay-environments-module-specific-appsconsole) for configuration details.

**Critical rule**: IAM pages (`src/pages/iam/`) must use `IAMRelayProvider`. Organization pages (`src/pages/organizations/`) must use `CoreRelayProvider`. The Relay compiler config enforces this mapping at build time, but a runtime mismatch will cause silent schema errors.

The `relay.config.json` defines two projects:

```json
{
  "sources": {
    "apps/console/src/pages/iam": "iam",
    "apps/console/src": "core"
  }
}
```

## Snapshot Mode

Most list and detail pages support read-only snapshot viewing. This requires:

1. **Duplicate routes** in the route file -- one for normal paths, one for `/snapshots/:snapshotId/...` paths
2. **Conditional rendering** -- check `Boolean(snapshotId)` from `useParams()` and hide all create/update/delete controls
3. **Adjusted URLs** -- breadcrumbs and internal links must use the snapshot path prefix

Example from `apps/console/src/routes/assetRoutes.ts`:
```tsx
// Normal route
{ path: "assets", loader: loaderFromQueryLoader(({ organizationId }) =>
    loadQuery(coreEnvironment, assetsQuery, { organizationId, snapshotId: null })),
},
// Snapshot route
{ path: "snapshots/:snapshotId/assets", loader: loaderFromQueryLoader(({ organizationId, snapshotId }) =>
    loadQuery(coreEnvironment, assetsQuery, { organizationId, snapshotId })),
},
```

## Permission Fragments

Inline permission queries in Relay fragments gate UI controls:

```graphql
canCreate: permission(action: "core:risk:create")
canUpdate: permission(action: "core:risk:update")
canDelete: permission(action: "core:risk:delete")
```

These boolean fields control visibility of edit buttons, delete actions, and create dialogs. There are no separate permission hooks.

## ViewerMembershipLayout

`apps/console/src/pages/iam/organizations/ViewerMembershipLayout.tsx` is the top-level authenticated shell. It:
1. Wraps `IAMRelayProvider` for its own membership query
2. Mounts `CoreRelayProvider` for child organization pages
3. Provides `CurrentUser` context (email, fullName, role) to the entire app

## Key Abstractions

| Abstraction | File | Purpose |
|-------------|------|---------|
| `loaderFromQueryLoader` | `@probo/routes` | Bridges React Router loaders with Relay preloaded queries |
| `useMutationWithToasts` | `src/hooks/useMutationWithToasts.ts` | **DEPRECATED** -- use `useMutation` + `useToast` |
| `SortableTable` | `src/components/SortableTable.tsx` | Paginated, sortable list with refetch support |
| `OrganizationErrorBoundary` | `src/components/OrganizationErrorBoundary.tsx` | Catches auth/assumption errors and redirects |
| `useOrganizationId` | `src/hooks/useOrganizationId.ts` | Extracts `:organizationId` from route params |

## Legacy: hooks/graph/ Directory

The `src/hooks/graph/` directory contains shared Relay queries, fragments, and mutations (e.g., `RiskGraph.ts`, `AssetGraph.ts`). This is **legacy code** with `eslint-disable` comments for `relay/must-colocate-fragment-spreads`. New code must colocate all GraphQL operations in the consuming component file.

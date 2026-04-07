# App arborescence (folder and file layout)

Conventions for organising pages, routes, and supporting files in Probo frontend apps (`apps/console`). The guiding principle is **one arborescence**: the route hierarchy is expressed once, through the `pages/` folder tree, and everything related to a route lives next to it.

**The codebase does not fully match these rules yet.** Some route definitions still live in a separate `src/routes/` folder. Treat this guide as the target for new work and refactors.

## Related guides

| Topic | Guide |
|-------|-------|
| `@probo/ui`, Tailwind, `tailwind-variants`, folders, skeletons, compound modules | [`contrib/claude/ui.md`](ui.md) |
| React component shape, props, file/export conventions | [`contrib/claude/react-components.md`](react-components.md) |
| Relay queries, fragments, loaders, `queryRef` | [`contrib/claude/relay.md`](relay.md) |

## Single arborescence principle

The `pages/` folder **is** the route tree. Every route segment maps to a folder under `pages/`, and route definitions live inside that folder as `routes.ts`. No other root-level folder should replicate the same hierarchy.

### Do / don't: route file placement

```text
// Bad — separate routes/ folder duplicates pages/ structure
src/
  routes/
    vendorRoutes.ts          # route definitions for vendors
    assetRoutes.ts           # route definitions for assets
  pages/
    organizations/
      vendors/
        VendorsPage.tsx
      assets/
        AssetsPage.tsx
```

```text
// Good — routes.ts colocated with the pages it references
src/
  pages/
    organizations/
      vendors/
        routes.ts            # route definitions for vendors
        VendorsPage.tsx
      assets/
        routes.ts            # route definitions for assets
        AssetsPage.tsx
```

Existing examples that already follow this pattern: `pages/organizations/compliance-page/routes.ts` and `pages/iam/organizations/people/routes.ts`. The parent route file (`routes.tsx` at the app root) imports and spreads them:

```tsx
import { compliancePageRoutes } from "./pages/organizations/compliance-page/routes";

// inside the route tree array
...compliancePageRoutes,
```

## Special files

Each page folder may contain a subset of these files. Names use PascalCase matching the feature.

| File | Role |
|------|------|
| `routes.ts` | Route definitions for this folder. Exports an array spread into the parent route tree. Uses `lazy()` from `@probo/react-lazy` to point at loaders / pages. |
| `MyPageLoader.tsx` | Bundle entry point imported by `lazy()` in the route. **Default export.** loads data via Relay, renders a skeleton while loading, then mounts the page with `queryRef`. |
| `MyPage.tsx` | The actual page component. Receives `queryRef` from the loader, calls `usePreloadedQuery`. |
| `MyPageSkeleton.tsx` | `Suspense` fallback rendered while the page is still receiving data. Also used as the route-level `Fallback`. |
| `MyPageError.tsx` | Error boundary rendering component for this page's error state. |
| `_components/` | Sub-components scoped to this page (see [below](#_components-folder)). |

### `routes.ts`

Contains route objects for the current folder's feature, exported as a named array and spread into the parent. Keep imports minimal — only `lazy`, skeleton components, and typing.

```ts
// pages/organizations/vendors/routes.ts
import { lazy } from "@probo/react-lazy";
import type { AppRoute } from "@probo/routes";

import { VendorsPageSkeleton } from "./VendorsPageSkeleton";

export const vendorRoutes = [
  {
    path: "vendors",
    Fallback: VendorsPageSkeleton,
    Component: lazy(() => import("./VendorsPageLoader")),
  },
  {
    path: "vendors/:vendorId",
    Fallback: VendorsPageSkeleton,
    Component: lazy(() => import("./VendorDetailPageLoader")),
    children: [
      {
        path: "overview",
        Component: lazy(() => import("./tabs/VendorOverviewTab")),
      },
    ],
  },
] satisfies AppRoute[];
```

### `MyPageLoader.tsx`

The loader is the **lazy bundle entry point**. It sets up providers, triggers the Relay query, shows a skeleton until the query resolves, then renders the page.

```tsx
// pages/organizations/vendors/VendorsPageLoader.tsx
import { Suspense, useEffect } from "react";
import { useQueryLoader } from "react-relay";

import type { VendorsPageQuery } from "#/__generated__/core/VendorsPageQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import VendorsPage, { vendorsPageQuery } from "./VendorsPage";
import { VendorsPageSkeleton } from "./VendorsPageSkeleton";

function VendorsPageQueryLoader() {
  const organizationId = useOrganizationId();
  const [queryRef, loadQuery] = useQueryLoader<VendorsPageQuery>(vendorsPageQuery);

  useEffect(() => {
    loadQuery({ organizationId });
  }, [loadQuery, organizationId]);

  if (!queryRef) {
    return <VendorsPageSkeleton />;
  }

  return <VendorsPage queryRef={queryRef} />
}

export default function VendorsPageLoader() {
  return (
    <CoreRelayProvider>
      <VendorsPageQueryLoader queryRef={queryRef} />
    </CoreRelayProvider>
  );
}
```

### `MyPage.tsx`

Receives the `queryRef` from the loader and renders the UI. Default export so `lazy()` can import it.

```tsx
// pages/organizations/vendors/VendorsPage.tsx
export default function VendorsPage({ queryRef }: VendorsPageProps) {
  const data = usePreloadedQuery(vendorsPageQuery, queryRef);
  return (/* … */);
}
```

### `MyPageSkeleton.tsx`

A lightweight loading placeholder. Keep it free of data-fetching logic so it loads instantly.

```tsx
// pages/organizations/vendors/VendorsPageSkeleton.tsx
export function VendorsPageSkeleton() {
  return (/* pulse / skeleton UI */);
}
```

### `MyPageError.tsx`

Rendered by the route error boundary when the page throws.

```tsx
// pages/organizations/vendors/VendorsPageError.tsx
export function VendorsPageError() {
  const error = useRouteError();
  return (/* error UI */);
}
```

## File naming

Component files (`.tsx` that export a React component) use **PascalCase**: `VendorsPage.tsx`, `VendorContactRow.tsx`, `VendorsPageSkeleton.tsx`.

All other helper files (utilities, hooks, constants, configuration) use **camelCase**: `routes.ts`, `useVendorFilters.ts`, `formatCurrency.ts`, `constants.ts`.

### Do / don't: file naming

```text
// Bad — helper file in PascalCase
pages/organizations/vendors/FormatVendorStatus.ts
pages/organizations/vendors/UseVendorFilters.ts
pages/organizations/vendors/Routes.ts

// Good — helpers are camelCase, components are PascalCase
pages/organizations/vendors/formatVendorStatus.ts
pages/organizations/vendors/useVendorFilters.ts
pages/organizations/vendors/routes.ts
pages/organizations/vendors/VendorsPage.tsx
pages/organizations/vendors/VendorsPageSkeleton.tsx
```

## `_components` folder

Sub-components that are used **only** by a single page live in a `_components/` folder next to that page. The underscore prefix visually distinguishes them from route-segment folders.

| Situation | Where the component lives |
|-----------|--------------------------|
| Used by one page only | `pages/organizations/vendors/_components/` |
| Used by multiple pages in the same feature | Nearest common ancestor's `_components/` (e.g. `pages/organizations/_components/`) |
| Reusable UI primitive | `@probo/ui` package |

### Do / don't: component placement

```text
// Bad — shared component buried in a single page's _components
pages/organizations/vendors/_components/StatusBadge.tsx    # also used by risks page
pages/organizations/risks/SomeRiskPage.tsx                 # imports ../../vendors/_components/StatusBadge

// Good — shared component hoisted to common ancestor
pages/organizations/_components/StatusBadge.tsx
```

```text
// Bad — page-specific helper placed in a global folder
src/components/VendorContactRow.tsx     # only used by VendorContactsTab

// Good — scoped to the page that uses it
pages/organizations/vendors/_components/VendorContactRow.tsx
```

## Full example tree

Target layout for a `vendors` feature under `pages/organizations/`:

```text
pages/organizations/vendors/
  routes.ts                        # route definitions for vendors
  VendorsPageLoader.tsx            # lazy entry — providers + Suspense + query loader
  VendorsPage.tsx                  # page component (usePreloadedQuery)
  VendorsPageSkeleton.tsx          # loading fallback
  VendorDetailPageLoader.tsx       # lazy entry for detail view
  VendorDetailPage.tsx             # detail page component
  VendorDetailPageSkeleton.tsx     # detail loading fallback
  _components/                     # sub-components used only by vendor pages
    VendorContactRow.tsx
    VendorRiskSummary.tsx
  tabs/                            # tab content for the detail page
    VendorOverviewTab.tsx
    VendorComplianceTab.tsx
    VendorContactsTab.tsx
```

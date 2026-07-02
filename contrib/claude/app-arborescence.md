# App arborescence (folder and file layout)

Conventions for organising pages, routes, and supporting files in Probo frontend apps (`apps/compliance-portal`, `apps/console`). The guiding principle is **one arborescence**: the route hierarchy is expressed once, through the `pages/` folder tree, and everything related to a route lives next to it.

These rules are the **source of truth**. Where existing code disagrees (e.g. `apps/console` still keeps some route definitions in a separate `src/routes/` folder), the code is non-compliant and should be migrated — it is not precedent.

## Related guides

| Topic                                                                            | Guide                                                       |
| -------------------------------------------------------------------------------- | ----------------------------------------------------------- |
| `@probo/ui`, Tailwind, `tailwind-variants`, folders, skeletons, compound modules | [`contrib/claude/ui.md`](ui.md)                             |
| React component shape, props, file/export conventions, naming/suffixes           | [`contrib/claude/react-components.md`](react-components.md) |
| Relay queries, fragments, loaders, `queryRef`                                    | [`contrib/claude/relay.md`](relay.md)                       |
| Error boundaries at any level, error/fallback props                              | [`contrib/claude/error-handling.md`](error-handling.md)     |
| i18next translations and `_locales` folders                                      | [`contrib/claude/i18n.md`](i18n.md)                         |

## Single arborescence principle

The `pages/` folder **is** the route tree. Every route segment maps to a folder under `pages/`, and route definitions live inside that folder as `routes.ts`. No other root-level folder should replicate the same hierarchy.

### Do / don't: route file placement

```text
// Bad — separate routes/ folder duplicates pages/ structure
src/
  routes/
    thirdPartyRoutes.ts          # route definitions for third parties
    assetRoutes.ts           # route definitions for assets
  pages/
    organizations/
      third-parties/
        ThirdPartiesPage.tsx
      assets/
        AssetsPage.tsx
```

```text
// Good — routes.ts colocated with the pages it references
src/
  pages/
    organizations/
      third-parties/
        routes.ts            # route definitions for third parties
        ThirdPartiesPage.tsx
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

| File                 | Role                                                                                                                                                                                                           |
| -------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `routes.ts`          | Route definitions for this folder. Exports an array spread into the parent route tree. Uses `lazy()` from `@probo/react-lazy` to point at loaders / pages.                                                     |
| `MyLayout.tsx`       | A **layout route** component that renders shared chrome (`Breadcrumb`, `PageHeader`, `Tabs`, …) and an `<Outlet />`. Named with the **`Layout` suffix** — never `Page` — to make its role obvious at a glance. |
| `MyLayoutLoader.tsx` | Loader for a layout that needs data (same pattern as `MyPageLoader`).                                                                                                                                          |
| `MyPageLoader.tsx`   | Bundle entry point imported by `lazy()` in the route. **Default export.** loads data via Relay, renders a skeleton while loading, then mounts the page with `queryRef`. Only needed when the page reads data.  |
| `MyPage.tsx`         | The actual page component. Receives `queryRef` from the loader (when data is loaded), or is the **default export** directly imported by `lazy()` when no data is needed.                                       |
| `MyPageSkeleton.tsx` | `Suspense` fallback rendered while the page is still receiving data. Also used as the route-level `Fallback`. Only needed when the page reads data.                                                            |
| `MyPageError.tsx`    | Error UI rendered by a boundary for this page (see [`error-handling.md`](error-handling.md)). An `*Error` file may exist at **any** level, not just the route root.                                              |
| `_components/`       | Sub-components scoped to this page (see [below](#_components-folder)).                                                                                                                                         |
| `_lib/`              | Non-component helpers scoped to this subtree: hooks, utilities, constants, types (see [below](#_lib-folder)).                                                                                                  |
| `_locales/`          | i18next translation catalogs for this route segment, one file per locale (`en-US.json`). Colocated with `routes.ts` (see [below](#_locales-folder)).                                                          |

### Layout vs Page naming

A component is a **layout** when it renders `<Outlet />` and exists to provide shared UI (breadcrumbs, tabs, page header) around child routes. It is a **page** when it renders final content with no `<Outlet />`.

Use the correct suffix so the role is clear from the file name alone:

```text
// Bad — a layout route named as a "Page"
ThirdPartyDetailPage.tsx          # renders <Outlet />, wraps child routes
CookieBannerConfigPage.tsx    # renders tabs + <Outlet />

// Good — layout routes use the "Layout" suffix
ThirdPartyDetailLayout.tsx
CookieBannerConfigLayout.tsx
```

## When do you need a loader, query, or Relay provider?

The loader / query / provider scaffolding exists to support Relay — it isn't boilerplate to copy into every page. Match the files to what the page actually does:

| Page has…                            | Files needed                                                                                                                             |
| ------------------------------------ | ---------------------------------------------------------------------------------------------------------------------------------------- |
| No Relay data and no mutation        | `MyPage.tsx` only. `lazy()` imports it directly. No loader, no skeleton, no provider, no query.                                          |
| A mutation only (no query)           | `MyPage.tsx` wrapped in the appropriate Relay provider (e.g. `IAMRelayProvider`, `CoreRelayProvider`). No loader, no skeleton, no query. |
| A query (with or without a mutation) | Full pattern: `MyPageLoader.tsx` (provider + `useQueryLoader` + skeleton) → `MyPage.tsx` (`usePreloadedQuery`) + `MyPageSkeleton.tsx`.   |

Layouts follow the same rule. A layout that only renders a `PageHeader` and an `<Outlet />` does **not** need a query, loader, skeleton, or provider — just a plain component imported by `lazy()` in `routes.ts`.

### Example: mutation only, no query

See [`pages/iam/organizations/NewOrganizationPage.tsx`](../../apps/console/src/pages/iam/organizations/NewOrganizationPage.tsx) for a reference. The page is the default export, wraps its inner component in `IAMRelayProvider` so the mutation has a Relay environment, and has no loader / skeleton / query.

```tsx
// pages/iam/organizations/NewOrganizationPage.tsx
function NewOrganizationPageInner() {
  const [createOrganization, isCreating]
    = useMutation<NewOrganizationPageMutation>(createOrganizationMutation);
  // …
}

export default function NewOrganizationPage() {
  return (
    <IAMRelayProvider>
      <NewOrganizationPageInner />
    </IAMRelayProvider>
  );
}
```

### Example: no data at all

```tsx
// pages/organizations/cookie-banners/CookieBannerLayout.tsx
export default function CookieBannerLayout() {
  const { __ } = useTranslate();
  usePageTitle(__("Cookie Banners"));

  return (
    <div className="space-y-6">
      <PageHeader title={__("Cookie Banners")} />
      <Outlet />
    </div>
  );
}
```

`routes.ts` points `lazy()` at `CookieBannerLayout` directly — no `CookieBannerLayoutLoader`, no `cookieBannerLayoutQuery`, no skeleton.

### `routes.ts`

Contains route objects for the current folder's feature, exported as a named array and spread into the parent. Keep imports minimal — only `lazy`, skeleton components, and typing.

```ts
// pages/organizations/third-parties/routes.ts
import { lazy } from "@probo/react-lazy";
import type { AppRoute } from "@probo/routes";

import { ThirdPartiesPageSkeleton } from "./ThirdPartiesPageSkeleton";

export const thirdPartyRoutes = [
  {
    path: "third-parties",
    Fallback: ThirdPartiesPageSkeleton,
    Component: lazy(() => import("./ThirdPartiesPageLoader")),
  },
  {
    path: "third-parties/:thirdPartyId",
    Fallback: ThirdPartiesPageSkeleton,
    Component: lazy(() => import("./ThirdPartyDetailLayoutLoader")),
    children: [
      {
        path: "overview",
        Component: lazy(() => import("./overview/ThirdPartyOverviewPage")),
      },
    ],
  },
] satisfies AppRoute[];
```

### `MyPageLoader.tsx`

The loader is the **lazy bundle entry point**. It sets up providers, triggers the Relay query, shows a skeleton until the query resolves, then renders the page.

```tsx
// pages/organizations/third-parties/ThirdPartiesPageLoader.tsx
import { Suspense, useEffect } from "react";
import { useQueryLoader } from "react-relay";

import type { ThirdPartiesPageQuery } from "#/__generated__/core/ThirdPartiesPageQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import ThirdPartiesPage, { thirdPartiesPageQuery } from "./ThirdPartiesPage";
import { ThirdPartiesPageSkeleton } from "./ThirdPartiesPageSkeleton";

function ThirdPartiesPageQueryLoader() {
  const organizationId = useOrganizationId();
  const [queryRef, loadQuery] = useQueryLoader<ThirdPartiesPageQuery>(thirdPartiesPageQuery);

  useEffect(() => {
    loadQuery({ organizationId });
  }, [loadQuery, organizationId]);

  if (!queryRef) {
    return <ThirdPartiesPageSkeleton />;
  }

  return <ThirdPartiesPage queryRef={queryRef} />
}

export default function ThirdPartiesPageLoader() {
  return (
    <CoreRelayProvider>
      <ThirdPartiesPageQueryLoader />
    </CoreRelayProvider>
  );
}
```

### `MyPage.tsx`

Receives the `queryRef` from the loader and renders the UI. Default export so `lazy()` can import it.

```tsx
// pages/organizations/third-parties/ThirdPartiesPage.tsx
export default function ThirdPartiesPage({ queryRef }: ThirdPartiesPageProps) {
  const data = usePreloadedQuery(thirdPartiesPageQuery, queryRef);
  return (/* … */);
}
```

### `MyPageSkeleton.tsx`

A lightweight loading placeholder. Keep it free of data-fetching logic so it loads instantly.

```tsx
// pages/organizations/third-parties/ThirdPartiesPageSkeleton.tsx
export function ThirdPartiesPageSkeleton() {
  return (/* pulse / skeleton UI */);
}
```

### `MyPageError.tsx`

Rendered by the route error boundary when the page throws.

```tsx
// pages/organizations/third-parties/ThirdPartiesPageError.tsx
export function ThirdPartiesPageError() {
  const error = useRouteError();
  return (/* error UI */);
}
```

## File naming

Component files (`.tsx` that export a React component) use **PascalCase**: `ThirdPartiesPage.tsx`, `ThirdPartyContactListItem.tsx`, `ThirdPartiesPageSkeleton.tsx`.

All other helper files (utilities, hooks, constants, configuration) use **camelCase**: `routes.ts`, `useThirdPartyFilters.ts`, `formatCurrency.ts`, `constants.ts`.

### Do / don't: file naming

```text
// Bad — helper file in PascalCase
pages/organizations/third-parties/FormatThirdPartyStatus.ts
pages/organizations/third-parties/UseThirdPartyFilters.ts
pages/organizations/third-parties/Routes.ts

// Good — helpers are camelCase, components are PascalCase
pages/organizations/third-parties/formatThirdPartyStatus.ts
pages/organizations/third-parties/useThirdPartyFilters.ts
pages/organizations/third-parties/routes.ts
pages/organizations/third-parties/ThirdPartiesPage.tsx
pages/organizations/third-parties/ThirdPartiesPageSkeleton.tsx
```

## Page sections

When a detail page has visually distinct sections (e.g. a properties card and a paginated list), extract each section into its own component in `_components/`. Each section owns a colocated Relay fragment (or pagination fragment) so that field additions never modify the parent page's query.

The page query spreads the section fragments and passes the fragment key to each section component:

```tsx
// TrackerPatternDetailPage.tsx (page — spreads section fragments)
export const trackerPatternDetailPageQuery = graphql`
  query TrackerPatternDetailPageQuery($trackerPatternId: ID!) {
    node(id: $trackerPatternId) {
      ... on TrackerPattern {
        id
        displayName
        ...TrackerPatternPropertiesSection_trackerPattern
        ...TrackerPatternDetectedTrackersSection_trackerPattern
      }
    }
  }
`;

// In JSX:
<TrackerPatternPropertiesSection trackerPatternKey={pattern} />
<TrackerPatternDetectedTrackersSection trackerPatternKey={pattern} />
```

```tsx
// _components/TrackerPatternPropertiesSection.tsx — owns its fragment
const fragment = graphql`
  fragment TrackerPatternPropertiesSection_trackerPattern on TrackerPattern {
    pattern
    matchType
    trackerType
    // ...
  }
`;

export function TrackerPatternPropertiesSection({ trackerPatternKey }: Props) {
  const pattern = useFragment(fragment, trackerPatternKey);
  return <Card padded>{/* PropertyRows */}</Card>;
}
```

Section components follow the same naming and fragment conventions as connection item components (see below), but represent a **logical section** of a page rather than a single list item.

## `_components` folder

Sub-components that are used **only** by a single page live in a `_components/` folder next to that page. The underscore prefix visually distinguishes them from route-segment folders.

| Situation                                  | Where the component lives                                                          |
| ------------------------------------------ | ---------------------------------------------------------------------------------- |
| Used by one page only                      | `pages/organizations/third-parties/_components/`                                         |
| Used by multiple pages in the same feature | Nearest common ancestor's `_components/` (e.g. `pages/organizations/_components/`) |
| Reusable UI primitive                      | `@probo/ui` package                                                                |

### Do / don't: component placement

```text
// Bad — shared component buried in a single page's _components
pages/organizations/third-parties/_components/StatusBadge.tsx    # also used by risks page
pages/organizations/risks/SomeRiskPage.tsx                 # imports ../../third-parties/_components/StatusBadge

// Good — shared component hoisted to common ancestor
pages/organizations/_components/StatusBadge.tsx
```

```text
// Bad — page-specific helper placed in a global folder
src/components/ThirdPartyContactListItem.tsx     # only used by the third-parties feature

// Good — scoped to the page that uses it
pages/organizations/third-parties/_components/ThirdPartyContactListItem.tsx
```

## `_lib` folder

Non-component code scoped to a subtree — hooks, utilities, constants, types — lives in a `_lib/` folder next to the pages that use it. The same hoisting rule as `_components/` applies: shared helpers move to the nearest common ancestor's `_lib/`; truly global helpers live in `src/lib/`. Files in `_lib/` use camelCase (`useThirdPartyFilters.ts`, `formatCurrency.ts`, `constants.ts`). See [`contrib/claude/hooks.md`](hooks.md) for custom-hook conventions.

```text
// Good — feature-scoped helpers under _lib
pages/organizations/third-parties/
  _lib/
    useThirdPartyFilters.ts
    formatThirdPartyStatus.ts
  _components/
    ThirdPartyListItem.tsx
```

## No barrel files

App code imports every module by its **explicit path** (`#/lib/relay/useMutation`, `#/lib/http/endpoint`). Do **not** add `index.ts` re-export barrels inside an app's `src/` to shorten import paths.

`index.ts` barrels are reserved for **package public entrypoints** (`packages/*`), where they define a real published API surface (`@probo/ui`, `@probo/relay`, …). An app has no published surface, so a barrel hides nothing — it only adds an indirection hop and invites Vite/tree-shaking and circular-import problems.

```text
// Bad — barrel inside the app just to shorten a path
src/lib/relay/
  index.ts            # export { useMutation } from "./useMutation"
  useMutation.ts
// import { useMutation } from "#/lib/relay"

// Good — import the module directly; no index.ts
src/lib/relay/
  useMutation.ts
// import { useMutation } from "#/lib/relay/useMutation"
```

## `_locales` folder

Translations are i18next catalogs in a `_locales/` folder, **one file per locale**, named by locale tag: `en-US.json`, `fr-FR.json`. See [`contrib/claude/i18n.md`](i18n.md) for key conventions and setup.

`_locales/` is colocated with a `routes.ts`. The rule of thumb:

- A `_locales/` folder belongs at a folder that **names a resource and owns a `routes.ts`** (e.g. `organizations/routes.ts` + `organizations/_locales/`, `organizations/measures/routes.ts` + `organizations/measures/_locales/`).
- There must be **no more `_locales/` folders than there are `routes.ts` files.** If a folder has no `routes.ts`, it does not get its own `_locales/` — its strings live in the nearest ancestor that does.

```text
// Good — _locales sits beside routes.ts at each resource boundary
pages/organizations/
  routes.ts
  _locales/
    en-US.json
    fr-FR.json
  measures/
    routes.ts
    _locales/
      en-US.json
      fr-FR.json
    MeasuresPage.tsx

// Bad — a _locales folder with no sibling routes.ts (strings belong to the parent resource)
pages/organizations/measures/_components/
  _locales/
    en-US.json
```

## Child-route folder naming

Folders that contain child-route pages are named after the **resource or concept** the page represents, not after the UX component that currently renders them. UX patterns change (tabs become pages, drawers become routes, etc.); the resource name stays stable.

### Do / don't: child-route folders

```text
// Bad — folder named after a UI element
configuration/
  tabs/                            # "tabs" is a UI component, not a resource
    ThirdPartyOverviewTab.tsx
    ThirdPartyComplianceTab.tsx

// Good — folders named after the resource each child route represents
configuration/
  overview/
    ThirdPartyOverviewPage.tsx
  compliance/
    ThirdPartyCompliancePage.tsx
```

This also means child-route components use the `*Page` suffix (not `*Tab`), because they are pages in their own right — the fact that a tab bar navigates between them is an implementation detail of the parent layout.

## Full example tree

Target layout for a `third-parties` feature under `pages/organizations/`:

```text
pages/organizations/third-parties/
  routes.ts                        # route definitions for third parties
  _locales/                        # i18next catalogs (one _locales per routes.ts)
    en-US.json
    fr-FR.json
  ThirdPartiesPageLoader.tsx            # lazy entry — providers + Suspense + query loader
  ThirdPartiesPage.tsx                  # page component (usePreloadedQuery)
  ThirdPartiesPageSkeleton.tsx          # loading fallback
  ThirdPartiesPageError.tsx             # error UI for this page's boundary
  ThirdPartyDetailLayoutLoader.tsx     # lazy entry for detail layout
  ThirdPartyDetailLayout.tsx           # layout — breadcrumbs, tabs, <Outlet />
  ThirdPartyDetailLayoutSkeleton.tsx   # detail loading fallback
  NewThirdPartyPage.tsx                # mutation-only page — default export, wraps itself in the Relay provider
  _lib/                            # hooks / utils / constants scoped to third party pages
    useThirdPartyFilters.ts
  _components/                     # sub-components used only by third party pages
    ThirdPartyContactListItem.tsx
    ThirdPartyRiskSummarySection.tsx
  overview/                        # child route: /third-parties/:thirdPartyId/overview
    ThirdPartyOverviewPage.tsx
  compliance/                      # child route: /third-parties/:thirdPartyId/compliance
    ThirdPartyCompliancePage.tsx
  contacts/                        # child route: /third-parties/:thirdPartyId/contacts
    ThirdPartyContactsPage.tsx
```

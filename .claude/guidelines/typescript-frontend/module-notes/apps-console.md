# Probo — TypeScript Frontend — apps/console

The main compliance SPA. **437 TS/TSX files**, React 19 + Relay 19 + Vite + react-router v7. Two
Relay environments coexist (`coreEnvironment` for `/api/console/v1/graphql`, `iamEnvironment` for
`/api/connect/v1/graphql`); they are split at the **Vite/Babel level** so fragments under
`src/pages/iam/**` compile against `src/__generated__/iam/` and everything else against
`src/__generated__/core/`. The route tree is built in `src/routes.tsx` and assembled from a mix of
**legacy** `src/routes/*Routes.ts` files (deprecated `loaderFromQueryLoader` / `withQueryRef`
pattern) and the **target** colocated `routes.ts` files inside each page folder.

## Key files

- `apps/console/src/main.tsx` — entry, mounts `QueryClientProvider`, `TranslatorProvider`,
  `RouterProvider`.
- `apps/console/src/environments.ts` — both Relay environments + `Network.create` wrapping
  `makeFetchQuery` from `@probo/relay`.
- `apps/console/src/providers/CoreRelayProvider.tsx`,
  `apps/console/src/providers/IAMRelayProvider.tsx` — the only correct ways to mount Relay.
- `apps/console/src/routes.tsx` — top-level route tree, spreads colocated `routes.ts` arrays.
- `apps/console/src/pages/organizations/findings/` — **canonical page folder** to copy from.
- `apps/console/src/components/RootErrorBoundary.tsx` — typed-error redirect chain.
- `apps/console/src/hooks/useOrganizationId.ts` — read `:organizationId` from router params.
- `apps/console/src/types.ts` — `NodeOf<T>`, `ItemOf<T>`.

## How to add a new page

1. Create folder `src/pages/<area>/<feature>/`.
2. Add `routes.ts` exporting an `AppRoute[]` (lazy import the loader).
3. Spread the array into the parent route in `src/routes.tsx`.
4. Implement `<Feature>PageLoader.tsx` (Relay provider + `useQueryLoader` + Suspense wrap).
5. Implement `<Feature>Page.tsx` (`usePreloadedQuery`, fragments).
6. Implement `<Feature>PageSkeleton.tsx` (synchronous, no Relay calls — see
   [pitfalls.md § 10](../pitfalls.md#10-skeleton-importing-root)).
7. Sub-components in `_components/`.

## Top pitfalls

1. Wrong Relay provider for the path (Core vs IAM) — see [pitfalls.md § 1](../pitfalls.md#1-missing-relay-provider-in-a-pageloader).
2. Omitting `filters: []` on `@connection` — see [pitfalls.md § 3](../pitfalls.md#3-omitting-filters--on-connection).

# Probo — TypeScript Frontend — @probo/relay & @probo/routes

These two packages together encapsulate Relay client wiring and route-loader plumbing.

## @probo/relay

Owns the GraphQL fetch path and typed error classification.

- `makeFetchQuery(endpoint)` — returns a Relay `Network.create`-compatible fetch function. Used by
  both `apps/console` (twice — Core and IAM environments) and `apps/trust`.
- **Six typed error classes** thrown after inspecting `errors[].extensions.code`:
  `UnAuthenticatedError`, `ForbiddenError`, `FullNameRequiredError`, `AssumptionRequiredError`,
  `NDASignatureRequiredError`, plus a generic `GraphQLError`.
- Every subclass calls `Object.setPrototypeOf(this, NewClass.prototype)` after `super()`.
  Skipping this breaks `instanceof` through TypeScript's class transpilation — see
  [pitfalls.md § 4](../pitfalls.md#4-proborelay-error-subclass-missing-objectsetprototypeof).

### How to extend

To add a new error code:

1. Add the GraphQL extension code to the server side (Go).
2. Add a new subclass in `packages/relay/src/errors.ts` with `setPrototypeOf`.
3. Map the code in the fetch function's response classifier.
4. Insert the `instanceof` check into `RootErrorBoundary` at the **right priority**
   (see [patterns.md § Error Boundary Chain](../patterns.md#8-error-boundary-chain)).

## @probo/routes

Holds route-related types and **deprecated** loader helpers retained only for legacy
`apps/console/src/routes/*Routes.ts` callers.

- `AppRoute` — the `satisfies AppRoute[]` type for every route array.
- `routeFromAppRoute` — converts the type-checked array into the React Router input.
- `loaderFromQueryLoader` (**DEPRECATED**) — returns a loader that calls Relay
  `loadQuery` and returns `{ queryRef, dispose }`.
- `withQueryRef` (**DEPRECATED**) — HOC that reads `{ queryRef, dispose }` from `useLoaderData`
  and disposes after a **deliberate 1000 ms delay** on unmount.

### How to extend

**Don't.** New routes use the colocated `*PageLoader` pattern — see
[patterns.md § Route Definitions](../patterns.md#2-route-definitions). When migrating a legacy
route, replace `loaderFromQueryLoader` + `withQueryRef` together; never half-migrate.

## Top pitfalls

1. **Forgetting `Object.setPrototypeOf`** in a new error subclass —
   [pitfalls.md § 4](../pitfalls.md#4-proborelay-error-subclass-missing-objectsetprototypeof).
2. **Removing the 1000 ms dispose delay** in `withQueryRef` — it is intentional —
   [pitfalls.md § 12](../pitfalls.md#12-removing-the-1000-ms-dispose-delay-in-withqueryref).

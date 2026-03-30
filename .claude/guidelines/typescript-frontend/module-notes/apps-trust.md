# Probo -- TypeScript Frontend -- apps/trust

> Module-specific notes for `apps/trust` (`@probo/trust`)
> For stack-wide patterns, see [patterns.md](../patterns.md) and [conventions.md](../conventions.md)

## Purpose

A public-facing trust center React SPA (port 5174) that exposes an organization's compliance posture to external visitors. It is read-only except for the auth and NDA signing flows. Pages include Overview, Documents, Subprocessors, Updates, and NDA signing.

## Single Relay Environment

Unlike the console app, trust uses a single Relay environment (`consoleEnvironment`) pointing at `/api/trust/v1/graphql`. Defined in `apps/trust/src/providers/RelayProviders.tsx`.

## Path Prefix Routing

The trust app supports two base URL modes:

| Mode | Base Path | How detected |
|------|-----------|-------------|
| Probo-hosted | `/trust/{slug}` | URL matches `/trust/[^/]+` |
| Custom domain | `/` | Default fallback |

Detection happens in `apps/trust/src/routes.tsx` via `getBasename()`. The `createBrowserRouter` receives the detected basename.

**Critical rule**: Every manually constructed URL must use `getPathPrefix()` from `apps/trust/src/utils/pathPrefix.ts`. Never hardcode `/trust/` or `/`. This applies to redirect URLs, continue URLs, and all navigation targets. See [pitfalls.md -- Trust Center Path Prefix](../pitfalls.md#13-hardcoding-paths-without-getpathprefix-appstrust).

## Auth Flow

External visitors who need to access gated documents go through a multi-step auth flow:

1. `RootErrorBoundary` catches `UnAuthenticatedError` and redirects to `/connect` with a `?continue=` URL
2. `/connect` (ConnectPage) -- magic link or OIDC login
3. `/verify-magic-link` -- validates the magic link token
4. `/full-name` -- captured if `FullNameRequiredError` is thrown
5. `/nda` -- NDA signing if `NDASignatureRequiredError` is thrown

After successful auth, `useRequestAccessCallback` in MainLayout reads `request-document-id` / `request-report-id` / `request-file-id` search params and fires the corresponding access-request mutation automatically (post-login replay).

## NDA Signing

`apps/trust/src/pages/NDAPage.tsx` renders the NDA PDF, records signing events, and polls Relay every 1500ms until the signature status reaches `COMPLETED` before redirecting. There is currently no timeout/max-retry strategy for this polling.

## TrustCenterProvider

`apps/trust/src/providers/TrustCenterProvider.tsx` stores the `currentTrustCenter` Relay data from `MainLayout` so child components can access it without prop-drilling. Sub-pages like `OverviewPage` read trust center data from `useOutletContext` (inherited from MainLayout), not from their own query.

## Route Organization

All routes are defined in a single `apps/trust/src/routes.tsx` file (no separate route slice files like console). Routes are grouped:

- **Auth routes**: Wrapped in `AuthLayout` (connect, verify-magic-link, full-name)
- **Content routes**: Wrapped in `MainLayout` with `RootErrorBoundary` (overview, documents, subprocessors, updates)
- **Standalone routes**: NDA page, document detail page

## Key Abstractions

| Abstraction | File | Purpose |
|-------------|------|---------|
| `getPathPrefix` | `src/utils/pathPrefix.ts` | Computes URL prefix for Probo-hosted vs custom domain mode |
| `RootErrorBoundary` | `src/components/RootErrorBoundary.tsx` | Typed error catch + auth redirect with continue URL |
| `useRequestAccessCallback` | `src/hooks/useRequestAccessCallback.ts` | Post-login access request replay from URL search params |
| `TrustCenterProvider` | `src/providers/TrustCenterProvider.tsx` | React context for trust center data from layout |

## Differences from Console

| Aspect | Console | Trust |
|--------|---------|-------|
| Relay environments | 2 (core + IAM) | 1 (trust) |
| Mutations | Frequent (CRUD for all entities) | Rare (access requests, NDA signing) |
| Route files | Separate per domain | Single `routes.tsx` |
| Snapshot mode | Yes (read-only views) | No (always read-only) |
| Permission fragments | Yes (`canCreate`, `canUpdate`, `canDelete`) | No (public read-only) |
| Auth errors | `UnAuthenticatedError`, `AssumptionRequiredError` | `UnAuthenticatedError`, `FullNameRequiredError`, `NDASignatureRequiredError` |

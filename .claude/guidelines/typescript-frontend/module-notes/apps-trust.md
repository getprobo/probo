# Probo — TypeScript Frontend — apps/trust

Public **unauthenticated** trust-center SPA (~50 TS/TSX files). Visitors view a tenant's published
trust report, request access to gated documents, sign NDAs, and authenticate via magic-link or
OIDC. Backed by `/api/trust/v1/graphql`. Most queries take a public token / slug as a variable
rather than relying on a session cookie.

The auth lazy-pages (magic-link consume, OIDC callback, NDA signature) each mount **their own
Relay environment**. This isolation is intentional (different auth states, different cache
lifetimes) but means a mutation in an auth lazy-page does **not** update the main layout's Relay
store.

## Key files

- `apps/trust/src/main.tsx` — entry.
- `apps/trust/src/environments.ts` — primary trust environment.
- `apps/trust/src/pages/auth/*` — magic-link / OIDC / NDA isolated bundles.
- `apps/trust/src/pages/<slug>/` — trust report pages, public-token-keyed.

## How to extend

- Adding a new public-facing page: same `*PageLoader` + `*Page` + `*PageSkeleton` pattern as
  `apps/console`. Use the trust environment provider.
- Adding an auth flow: branch a new lazy-page under `pages/auth/` with its own environment if
  state isolation is required. Otherwise share the trust environment.

## Top pitfalls

1. **Open-redirect on `continue` URL** — must validate same-origin + path-prefix before
   `window.location.href = continue`. See [pitfalls.md § 6](../pitfalls.md#6-appstrust-open-redirect-via-unvalidated-continue-url).
2. **Isolated Relay stores in auth lazy-pages** — mutations don't update the main layout. See
   [pitfalls.md § 8](../pitfalls.md#8-appstrust-isolated-relay-stores-in-auth-lazy-pages).

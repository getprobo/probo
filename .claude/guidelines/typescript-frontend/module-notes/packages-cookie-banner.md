# Probo — TypeScript Frontend — @probo/cookie-banner

Vanilla **web component** (no React) that shows a cookie-consent banner on the marketing site and
inside the app shell. Implemented as a custom element with **shadow DOM** for style isolation.
Ships **two bundles**: ESM (for module-aware consumers) and IIFE (for `<script>` drop-in).

The banner queries the cookie-pattern catalog via the public Trust API and persists user
preferences in `localStorage`. A dismiss-bit lives in `sessionStorage` so the banner doesn't
re-prompt within a tab.

## Key files

- `packages/cookie-banner/src/index.ts` — custom element registration.
- `packages/cookie-banner/src/element.ts` — `CookieBannerElement` class extending `HTMLElement`.
- `packages/cookie-banner/<build config>` — dual ESM + IIFE outputs.

## How to extend

- New cookie category / pattern: update the GraphQL query against the public Trust API; render in
  the shadow DOM.
- New visual variant: shadow DOM means CSS in `<style>` blocks inside the element — no Tailwind
  here.

## Top pitfalls

1. **`importFunction.toString()` minification collision** — the sessionStorage key is derived
   from a function's stringified body. Production minification can collapse different functions
   to the same string, causing two banner instances to share state. See
   [pitfalls.md § 17](../pitfalls.md#17-packagescookie-banner-sessionstorage-key-minification-collision).
2. **No tests** — visual regressions are easy to miss.

The four-surface API rule applies (cookie-pattern operations were explicitly called out in PR
review #1132). See [shared.md § 3](../shared.md#3-the-four-surface-api-rule).

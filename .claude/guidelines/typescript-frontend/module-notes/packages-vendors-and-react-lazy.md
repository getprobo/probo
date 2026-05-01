# Probo — TypeScript Frontend — @probo/vendors & @probo/react-lazy

Two unrelated small packages grouped here for brevity.

## @probo/vendors

Static catalog of ~100+ third-party vendors (name, country, category, certifications). Shipped as
a **`data.json`** file plus a `data.d.ts` typing. Consumed by `apps/console` via
**MiniSearch** for the vendor picker UI.

### Key files

- `packages/vendors/src/data.json` — catalog data.
- `packages/vendors/src/data.d.ts` — TypeScript shape.
- `packages/vendors/src/index.ts` — re-exports the parsed JSON + `Vendor` type.

### How to extend

- Adding a vendor: append an entry to `data.json` with the required fields. Run the consumer's
  build to validate the shape.

### Top pitfalls

1. **`data.d.ts` references undefined `CountryCode` type** — see
   [pitfalls.md § 18](../pitfalls.md#18-packagesvendors-datadts-references-undefined-countrycode-type).
   Define `CountryCode` (ISO-3166-1 alpha-2 union or `string` brand) before consumers can rely on
   strict typing of the country field.

---

## @probo/react-lazy

Thin wrapper around `React.lazy()` with **two** added behaviours:

1. **Retry on chunk-load failure** (typical after a deploy that invalidates old chunks).
2. **One-shot full page reload** as a fallback, gated by a sessionStorage counter to prevent
   reload loops.

### Key files

- `packages/react-lazy/src/lazy.ts` — the wrapper.

### How to use

```ts
import { lazy } from "@probo/react-lazy";
const FindingsPageLoader = lazy(() => import("./FindingsPageLoader"));
```

### Top pitfalls

1. **Counter key derived from `importFunction.toString()`** — minification can collapse two
   different `import(...)` calls to the same key, sharing the reload counter and either
   suppressing legitimate retries or causing a false reload-loop guard.
   Same hazard as the cookie banner's sessionStorage key — see
   [pitfalls.md § 17](../pitfalls.md#17-packagescookie-banner-sessionstorage-key-minification-collision).
   Mitigation: pass an explicit string key, or hash the import path at build time.

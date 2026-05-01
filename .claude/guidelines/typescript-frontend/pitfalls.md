# Probo — TypeScript Frontend — Pitfalls

> Concrete pitfalls extracted from module profiles and PR reviews. Each item gives **what goes
> wrong**, **why**, **how to avoid**, and a source pointer.

---

## High Severity

### 1. Missing Relay provider in a `*PageLoader`

**What goes wrong.** `useQueryLoader`, `useMutation`, `usePreloadedQuery` throw at runtime, or
silently target the wrong GraphQL endpoint, when the subtree is not wrapped in a
`RelayEnvironmentProvider`.

**Why.** `apps/console` has **two** Relay environments (`coreEnvironment`, `iamEnvironment`) with
different stores. The Vite/Babel plugin invocation that compiles your fragment is selected by
**path** (`src/pages/iam/**` vs everything else), not by which provider you mount.

**How to avoid.** In every `*PageLoader.tsx`:
- Pages outside `src/pages/iam/` → wrap in `<CoreRelayProvider>` (from
  `apps/console/src/providers/CoreRelayProvider.tsx`).
- Pages under `src/pages/iam/` → wrap in `<IAMRelayProvider>`.
- Mutation-only pages still need a provider. See
  `apps/console/src/pages/iam/organizations/NewOrganizationPage.tsx`.

**Source.** `apps-console.json` profile, `apps/console/vite.config.ts`,
[`contrib/claude/relay.md`](../../../contrib/claude/relay.md).

---

### 2. Using deprecated `loaderFromQueryLoader` / `withQueryRef` for new routes

**What goes wrong.** Bundle still works, but you have entrenched a deprecated API that the team is
actively migrating away from. Future refactors will rip it out.

**Why.** The route system is mid-migration. The colocated `*PageLoader` pattern replaces
loader-based queryRef ferrying. The old helpers are kept only to support legacy
`apps/console/src/routes/*Routes.ts`.

**How to avoid.** New routes must follow `apps/console/src/pages/organizations/findings/`:
`routes.ts` + `*PageLoader.tsx` + `*Page.tsx` + `*PageSkeleton.tsx`. Never call
`loaderFromQueryLoader` or `withQueryRef` in new code.

**Source.** `packages-routes.json` profile (helpers explicitly marked deprecated),
[`contrib/claude/app-arborescence.md`](../../../contrib/claude/app-arborescence.md).

---

### 3. Omitting `filters: []` on `@connection`

**What goes wrong.** `ConnectionHandler.getConnection` lookups, `@appendEdge`, `@deleteEdge`, and
`@prependEdge` directives silently miss the connection. The mutation completes successfully but the
list does not update; users see stale data and refreshing fixes it.

**Why.** When `filters` is omitted, Relay includes **all** non-pagination arguments in the
connection store key. The next render passes the same arguments and reads from the same key — but
your mutation directives reference a connection identified by a different (filter-aware) key.

**How to avoid.** Always specify `filters` explicitly:

```graphql
@connection(key: "FindingsPage_findings", filters: [])
@connection(key: "FindingsPage_findings", filters: ["status"])  # status resets the connection
```

**Source.** `apps-console.json` profile,
[`contrib/claude/relay.md`](../../../contrib/claude/relay.md).

---

### 4. `@probo/relay` error subclass missing `Object.setPrototypeOf`

**What goes wrong.** `instanceof UnAuthenticatedError` returns `false`. The error reaches
`RootErrorBoundary` but no branch matches; users see the generic `PageError` instead of being
redirected to login.

**Why.** TypeScript's class transpilation to ES5 (and even some ES2015 paths) loses the prototype
chain on `Error` subclasses. The fix is mandatory in every subclass:

```ts
export class UnAuthenticatedError extends Error {
  constructor(message?: string) {
    super(message);
    Object.setPrototypeOf(this, UnAuthenticatedError.prototype);
    this.name = "UnAuthenticatedError";
  }
}
```

**How to avoid.** When you add a new typed error in `packages/relay/src/`, copy the
`setPrototypeOf` line from an existing class. Do not skip it.

**Source.** `packages-relay.json` profile.

---

### 5. Reordering the `RootErrorBoundary` `instanceof` chain

**What goes wrong.** A more-specific error (e.g. `FullNameRequiredError`, which extends Error
directly but is semantically downstream of `UnAuthenticatedError`) gets caught by the wrong
branch — the user lands on the wrong recovery page (e.g. NDA flow when they should be sent to
login).

**Why.** The chain order is: **UNAUTHENTICATED → FULL_NAME_REQUIRED → ASSUMPTION_REQUIRED →
NDA_SIGNATURE_REQUIRED → FORBIDDEN → default**. The semantic priority encodes which "fix me first"
the user must complete: you can't onboard a name if you're not authenticated; you can't sign an
NDA if you haven't selected an org.

**How to avoid.** Read `apps/console/src/components/RootErrorBoundary.tsx` carefully before
editing. If you add a new typed error, decide its priority and slot it in deliberately — never
append to the bottom unless it's strictly less-specific than `FORBIDDEN`.

**Source.** `packages-relay.json` profile, `apps-console.json` profile.

---

### 6. `apps/trust` open-redirect via unvalidated `continue` URL

**What goes wrong.** Magic-link / OIDC / NDA flows accept a `continue` query parameter to redirect
the user back into the app after auth. If the `continue` URL is not validated, an attacker can craft
a magic-link email pointing to a malicious site.

**How to avoid.** Validate `continue` before redirect:

1. Same-origin only (`new URL(continue, window.location.origin).origin === window.location.origin`).
2. Path-prefix allow-list (e.g. starts with `/trust/`).
3. Reject if the URL has a different scheme/host or contains `@` (URL parser quirks).

**Source.** `apps-trust.json` profile.

---

## Medium Severity

### 7. `useLazyLoadQuery` instead of `useQueryLoader` + `usePreloadedQuery`

**What goes wrong.** Network request is deferred until the component renders, producing
fetch-then-render and a flash of skeleton state on every navigation. Compare against the loader
pattern that starts the request *before* the component is rendered.

**How to avoid.** Use `useQueryLoader` + `useEffect` + `usePreloadedQuery` (the `*PageLoader`
pattern). Existing `apps/console/src/hooks/graph/VendorGraph.ts` still uses `useLazyLoadQuery` —
do not extend it; migrate when you touch it.

**Source.** `apps-console.json` profile.

---

### 8. `apps/trust` isolated Relay stores in auth lazy-pages

**What goes wrong.** A mutation submitted from the magic-link / OIDC / NDA lazy-page does not
update the main layout's Relay store, because each auth lazy-page mounts its **own** Relay
environment. After auth completes, the layout still shows the pre-auth state until a full reload.

**How to avoid.** Either (a) hoist the mutation to the main environment — preferred, by sharing the
trust environment across the auth bundles, or (b) explicitly trigger a reload / re-init of the
layout's queries on auth success.

**Source.** `apps-trust.json` profile.

---

### 9. `packages/ui` legacy `Button.tsx` uses `export const` + arrow

**What goes wrong.** `packages/ui/src/atoms/Button/Button.tsx` is the canonical "we will fix it
soon" example. Copying its shape spreads the anti-pattern.

**How to avoid.** Use `export function Button(...)`. See modern atoms (any added after Button) for
the right shape.

**Source.** `packages-ui.json` profile.

---

### 10. Skeleton importing `*Root`

**What goes wrong.** The skeleton accidentally pulls in client-only / Relay-suspending logic; SSR
or pre-data render fails, or the skeleton bundle balloons.

**How to avoid.** Extract `tv()` styles into a sibling `variants.ts`. Both `*Root` and `*Skeleton`
import from `variants.ts`. Skeleton must be self-contained.

**Source.** `packages-ui.json` profile.

---

### 11. `withQueryRef` has no guard if loader didn't return `{queryRef, dispose}`

**What goes wrong.** A legacy route whose loader was refactored to return raw data instead of a
queryRef bundle now crashes inside `withQueryRef` with a confusing destructure error.

**How to avoid.** When migrating off `loaderFromQueryLoader`, replace **the whole pair**
(`loaderFromQueryLoader` + `withQueryRef`) with the `*PageLoader` pattern. Don't half-migrate.

**Source.** `packages-routes.json` profile.

---

### 12. Removing the 1000 ms dispose delay in `withQueryRef`

**What goes wrong.** The queryRef is disposed mid-route-transition, causing the suspended new page
to read from a disposed store and throw.

**Why.** The 1000 ms delay is **intentional** — it gives React Router enough time to commit the new
route before the previous queryRef is released.

**How to avoid.** Do not "optimize" the timeout. The proper fix is migrating off the legacy helper
entirely.

**Source.** `packages-routes.json` profile.

---

### 13. Forgetting to update the `@probo/helpers` barrel `index.ts`

**What goes wrong.** Your new helper is invisible to importers (`apps/console`, `apps/trust`,
other packages). TypeScript may still resolve it via deep import paths, but the project convention
is barrel-only.

**How to avoid.** When you add `packages/helpers/src/foo/myHelper.ts`, edit
`packages/helpers/src/index.ts` to re-export it.

**Source.** `packages-helpers.json` profile.

---

### 14. `packages/i18n` loaders return `{}` — i18n is dormant

**What goes wrong.** You write strings expecting them to be translated; nothing is. Language is
hard-coded `"en"`.

**How to avoid.** Continue wrapping every user-facing string in `__(...)` even though it's
currently a pass-through — that is what makes the eventual translation switch a no-op. Don't
delete the `__` calls "because they don't do anything".

**Source.** `packages-i18n.json` profile.

---

### 15. `packages/emails` `dist/` stale at `go build` time

**What goes wrong.** Go's `//go:embed` in `pkg/mailer` picks up empty or out-of-date templates;
production sends the wrong email body.

**How to avoid.** Run the email package build (or `make build WITH_APPS=1`) before `go build`.
After editing a `.tsx` template, re-run `tsx scripts/build.ts`. Verify the embedded files in
`pkg/mailer` match `packages/emails/dist/` before merging.

**Source.** `packages-emails.json` profile.

---

### 16. Email-template placeholders typo only surfaces at Go runtime

**What goes wrong.** `<p>{`{{ .UserFullNmae }}`}</p>` (note typo) compiles fine — TS can't see
inside the template string — and the email send fails at runtime when Go's `text/template`
complains about an unknown field.

**How to avoid.** Cross-check placeholder names against the Go side's template data struct in
`pkg/mailer`. Treat templates as a strict contract; add a smoke test in Go that renders each
template with a fixture struct.

**Source.** `packages-emails.json` profile.

---

### 17. `packages/cookie-banner` sessionStorage key minification collision

**What goes wrong.** The cookie banner uses `importFunction.toString()` as the storage key. Under
production minification, two different lazy imports may produce the same minified function string
and share state.

**How to avoid.** Pass an explicit string key alongside the import function (or hash the import
path at build time). The same hazard applies to `packages/react-lazy`'s reload counter — same
fix.

**Source.** `packages-cookie-banner.json`, `packages-react-lazy.json` profiles.

---

### 18. `packages/vendors` `data.d.ts` references undefined `CountryCode` type

**What goes wrong.** TypeScript may infer `any` for the `country` field, weakening the consumer's
type safety in `apps/console`.

**How to avoid.** Define `CountryCode` (ISO 3166-1 alpha-2 union or a generic `string` brand) in
`packages/vendors/src/data.d.ts` or import from a shared types module.

**Source.** `packages-vendors.json` profile.

---

### 19. `packages/n8n-node` resource export name mismatch

**What goes wrong.** The operation defined in `actions/<resource>/<operation>.operation.ts` is
re-exported under the wrong name in `actions/<resource>/index.ts`. n8n looks up operations by the
**string value** in the properties array; mismatch means the operation simply isn't found at
runtime — no error, just an inactive operation.

**How to avoid.** The export name in `index.ts` **must equal** the operation's `value` string used
in `Probo.node.ts` properties. Reference: [`contrib/claude/n8n.md`](../../../contrib/claude/n8n.md).

**Source.** `packages-n8n-node.json` profile.

---

### 20. `packages/n8n-node` IAM operations using `proboApiRequest`

**What goes wrong.** The request hits `/api/console/v1/graphql` with the wrong auth flow; IAM
operations fail with 401 or schema-mismatch errors.

**How to avoid.** IAM (Connect) operations **must** use `proboConnectApiRequest`. Console / Trust
operations use `proboApiRequest`. Pick the helper that matches the schema you're targeting.

**Source.** `packages-n8n-node.json` profile,
[`contrib/claude/n8n.md`](../../../contrib/claude/n8n.md).

---

## Low Severity

### 21. Missing ISC license header

**What goes wrong.** Reviewers flag it; CI does not (no automated linter for the header today).
Easy to forget.

**How to avoid.** Top of every new `.ts` / `.tsx` / `.css`: ISC block with current year. When
editing an existing file, **expand the year range**; never overwrite the original year. See
[shared.md § 6](../shared.md#6-license-headersisc-on-every-source-file).

---

### 22. `*Tab.tsx` naming for child-route components

**What goes wrong.** Diverges from the arborescence convention; new contributors copy the wrong
suffix.

**How to avoid.** Child-route components use `*Page.tsx`, even when visually presented as a tab.
Rename when you touch them. Ref: `contrib/claude/app-arborescence.md`.

---

### 23. Inline SVG in a component

**What goes wrong.** Reviewers block: PR-mining frequency 4 — *"all SVGs should be in a react
component"*.

**How to avoid.** Use Phosphor icons (`@phosphor-icons/react`) first. If a custom asset is
unavoidable, extract it to its own `<MyIcon />` component file. Never inline `<svg>` markup in a
page component.

**Source.** [conventions.md § 3](./conventions.md#3-icons), PR #957.

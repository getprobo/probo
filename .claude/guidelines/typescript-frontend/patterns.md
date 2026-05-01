# Probo — TypeScript Frontend — Patterns

> Cross-cutting principles (license headers, GID layout, four-surface API rule, PII-free logging,
> commit conventions) live in [shared.md](../shared.md). Stack-specific rules below.

---

## 1. App Layout (`apps/console`, `apps/trust`)

**Target arborescence:** `src/pages/` IS the route tree — feature slices under `src/pages/<feature>/`
mirror the URL path. Each leaf folder owns its own `routes.ts`, `*PageLoader.tsx`, `*Page.tsx`,
`*PageSkeleton.tsx`, optional `*Layout.tsx`, and a `_components/` subfolder for page-private bits.
Shared cross-page UI lives in `apps/console/src/components/`. See
[`contrib/claude/app-arborescence.md`](../../../contrib/claude/app-arborescence.md).

**Current state — known divergence:** The doc explicitly admits the codebase has not finished
migrating. Legacy routes still live under `apps/console/src/routes/` (e.g. `vendorRoutes.ts`,
`findingRoutes.ts`). New work must follow the colocated pattern; touching a legacy file is a good
opportunity to migrate it.

**Canonical example (target):** `apps/console/src/pages/iam/organizations/people/routes.ts` is a
colocated route file that gets spread into the parent route tree from
`apps/console/src/routes.tsx`.

**Path alias:** `#` maps to `src/` in both apps (set in each app's `vite.config.ts`). Always import
with `#/hooks/useOrganizationId`, never relative `../../../hooks/...`.

---

## 2. Route Definitions

The route tree is built with `react-router` v7's `createBrowserRouter`. Every route array is typed
with the `AppRoute` type from `@probo/routes` and asserted via `satisfies AppRoute[]` so additional
properties (e.g. `ErrorBoundary`, `Fallback`) keep their narrow types.

### Target pattern (new code)

Inside the page folder:

```ts
// apps/console/src/pages/organizations/findings/routes.ts
import type { AppRoute } from "@probo/routes";
import { lazy } from "@probo/react-lazy";

const FindingsPageLoader = lazy(() => import("./FindingsPageLoader"));

export const findingsRoutes = [
  {
    path: "findings",
    Component: FindingsPageLoader,
  },
] satisfies AppRoute[];
```

Spread into the parent in `apps/console/src/routes.tsx`. The lazy import means the bundle is split
per page; `@probo/react-lazy` adds retry-on-failure with a sessionStorage-backed reload counter.

### Legacy pattern (DEPRECATED — do not extend)

```ts
// apps/console/src/routes/vendorRoutes.ts
loaderFromQueryLoader(query, (params) => ({ organizationId: params.organizationId! }))
```

`loaderFromQueryLoader` and `withQueryRef` (from `@probo/routes`) are marked deprecated. They wrap
the loader return as `{ queryRef, dispose }` and then `withQueryRef` reads it back via
`useLoaderData`. The dispose path uses an intentional **1000 ms delay** before disposing the queryRef
to survive route transitions — do not "optimize" that timeout.

**New routes must use the `*PageLoader` component pattern instead.**

---

## 3. Relay Data Flow

> Source: [`contrib/claude/relay.md`](../../../contrib/claude/relay.md). All paths below are
> universal across `apps/console` and `apps/trust`.

### `*PageLoader` — the lazy-bundle entry point

The `*PageLoader` is the only component a route definition ever points at. It owns three
responsibilities:

1. Mount the correct **Relay environment provider** (`CoreRelayProvider` or `IAMRelayProvider`).
2. Call `useQueryLoader(query)` and kick off `loadQuery(...)` in a `useEffect`. The variables come
   from `useParams()` / `useSearchParams()` / `useOrganizationId()`.
3. Render `*PageSkeleton` while `queryRef === null`; once a `queryRef` is available, wrap the real
   `*Page` in `<Suspense fallback={<PageSkeleton />}>`.

Canonical:
`apps/console/src/pages/organizations/findings/FindingsPageLoader.tsx`. Apply the same shape for
mutation-only pages — see `apps/console/src/pages/iam/organizations/NewOrganizationPage.tsx`. A
mutation-only page **still needs** the Relay provider, even with no query.

### `*Page` — receives `queryRef`, calls `usePreloadedQuery`

```tsx
const data = usePreloadedQuery(query, queryRef);
```

Sub-components consume **fragments** via `useFragment(fragmentNode, fragmentKey)`. Pagination uses
`usePaginationFragment`. Refetchable lists use `useRefetchableFragment` with `@refetchable`.

### Mutations

Always `useMutation` (never `commitMutation`). Use store-update directives instead of refetching:

- `@appendEdge(connections: [...])` for adds at the end
- `@prependEdge(connections: [...])` for adds at the top
- `@deleteEdge(connections: [...])` for removes (server returns the deleted node id)

**Never name the result tuple `[commitMutation, isInFlight]`** — name after the action verb
(`[createCookieBanner, isCreating]`). PR-mining: PR #1073 reviewer comment *"not fan of `commit*`
naming."*

For multi-connection updates, use the `updater` callback with
`ConnectionHandler.getConnection` / `getConnectionID` / `deleteNode` / `insertEdgeAfter`. See
`FindingsPage.tsx` for the canonical `updater` shape.

**Avoid `useLazyLoadQuery`** — it defers the network request until render, producing
fetch-then-render. Use `useQueryLoader` + `usePreloadedQuery`. Legacy: `apps/console/src/hooks/graph/VendorGraph.ts`
still uses `useLazyLoadQuery` and is on the migration list.

### `@connection` and `filters: []` — non-negotiable

Every paginated `@connection` directive **must** specify `filters` explicitly:

```graphql
@connection(key: "FindingsPage_findings", filters: [])
```

If you omit `filters`, Relay includes **every** non-pagination argument in the connection store
key. `ConnectionHandler.getConnection(...)` lookups, `@appendEdge` and `@deleteEdge` then silently
miss the connection and the UI fails to update. Use `filters: []` for fixed-argument connections,
`filters: ["status"]` for user-selectable filters that should reset the connection.

### `@required` instead of GraphQL `!`

> PR-mining (frequency 4, high): PR #720 — *"cannot be bang there as resolver can fail. We have to
> use @require in relay for type issue."*

GraphQL fields whose resolvers can fail must **not** be declared non-null (`!`) on the schema side.
Mark them nullable in the schema and add Relay's `@required(action: LOG | THROW)` in the fragment.
Relay then surfaces a typed null at the right boundary instead of crashing the entire query.

### Two environments — Vite/Babel split

> Universal across `apps/console`. Universal in spirit in `apps/trust` though it has a single
> environment.

`apps/console/vite.config.ts` configures **two** `babel-plugin-relay` invocations:

| Source path | Compiles fragments against |
| --- | --- |
| `apps/console/src/pages/iam/**` | `apps/console/src/__generated__/iam/` |
| Everything else | `apps/console/src/__generated__/core/` |

`environments.ts` exports `coreEnvironment` and `iamEnvironment` — each a Relay `Environment` with a
`Store` (1-minute query cache, GC buffer) and a `Network.create` wrapping
`makeFetchQuery('/api/console/v1/graphql' | '/api/connect/v1/graphql')` from `@probo/relay`.

**Crossing the boundary fails silently.** A fragment under `pages/iam/` referring to a Core type
won't be found by the IAM compiler. Always: IAM pages → `IAMRelayProvider` + types from
`__generated__/iam/`; everything else → `CoreRelayProvider` + types from `__generated__/core/`.

### Use Relay-generated types — never declare local TS types

> PR-mining (frequency 4, high — highest-frequency frontend block): PR #800 *"We don't declare local
> types anymore, we should use relay generated types directly"*

Import generated types from `src/__generated__/{core,iam}/<Operation>.graphql.ts`. The shared
helpers `NodeOf<T>` and `ItemOf<T>` (in `apps/console/src/types.ts`) extract node / element types
from connection edges so you do not need to redeclare anything.

---

## 4. Component Shape

> Source: [`contrib/claude/react-components.md`](../../../contrib/claude/react-components.md).
> The doc explicitly notes older components do not yet match — refactor opportunistically.

### Function declarations, not arrow `const`

```tsx
// Good
export function FindingsList({ findings }: FindingsListProps) { ... }

// Bad — legacy
export const FindingsList = ({ findings }: FindingsListProps) => { ... };
```

`packages/ui/src/atoms/Button/Button.tsx` is a known legacy holdout still using
`export const Button = forwardRef(...)`. The current convention is `export function`.

### Props split: configure vs data

- **Configure props** — URL params, IDs, mode flags. Required to render at all.
- **Data props** — Relay fragment keys. Read via `useFragment` inside the component.

A component reading data from props and a hook should have **two** props interfaces (or one
union-discriminated by a mode flag). Configure-only and data-only hooks are kept separate so a
skeleton can use the configure hook without the data hook (and therefore without Relay).

### `useOrganizationId()` is called inside the component, not drilled

`apps/console/src/hooks/useOrganizationId.ts` reads `:organizationId` from React Router params.
Never accept `organizationId` as a prop — it's globally available; the prop drilling is noise.

### Skeleton must NOT import Root

Skeletons must be self-contained. **Never** have `*Skeleton` import `*Root` to "reuse the
container", because `*Root` may pull in client-only logic (Relay hooks, suspending boundaries) and
the skeleton must render synchronously on the server / before data is loaded.

The right pattern: extract `tv()` variants into a sibling `variants.ts`, then both `*Root` and
`*Skeleton` import from `variants.ts`. See `packages/ui/src/atoms/Button/` for the modern shape.

---

## 5. Component Library — `@probo/ui`

> Source: [`contrib/claude/ui.md`](../../../contrib/claude/ui.md).

### Single styling API: `tailwind-variants`

```ts
import { tv } from "tailwind-variants";

export const buttonVariants = tv({
  base: "inline-flex items-center justify-center …",
  variants: {
    variant: { primary: "...", secondary: "..." },
    size: { sm: "...", md: "...", lg: "..." },
  },
  defaultVariants: { variant: "primary", size: "md" },
});
```

`tv()` is the **single** styling API in `@probo/ui`. Do not mix raw Tailwind className strings into
the same file as a `tv()` definition. `tailwind-merge` is available transitively through `tv` so
caller-supplied `className` overrides cleanly.

Folder layout under `packages/ui/src/`:
- `atoms/` — primitives (Button, Input, Badge, Avatar, Icon, Field, …)
- `molecules/` — composed (Card, Dialog, Dropdown, Select, Tabs, Combobox, Table, PageHeader, Toast, …)

### Compound components: FLAT named exports

> Universal in `@probo/ui`.

Compound components export **flat** named members, not nested namespaces:

```tsx
// Good
import { DialogRoot, DialogShell, DialogTitle, DialogSkeleton } from "@probo/ui";

// Bad — Probo does not use this style
import { Dialog } from "@probo/ui";
<Dialog.Root>...</Dialog.Root>
```

Naming suffix conventions:
- `*Root` — the controlled boundary (Radix `Root` wrapper or equivalent)
- `*Shell` — the visual chrome / outer container
- `*Skeleton` — synchronous loading state
- `*Trigger`, `*Content`, `*Item`, … as Radix dictates

### `asChild` via custom `Slot`

`packages/ui/src/atoms/Slot.tsx` is a **custom** `Slot` (not the Radix `@radix-ui/react-slot`
one) that merges props/refs onto the single child element. Use `<Button asChild><Link to="…">…</Link></Button>`
to compose a button styled-anchor without nesting `<a><button></button></a>`.

### Underlying primitives

| Component | Library |
| --- | --- |
| Dialog, Dropdown, Select, Tabs, Label, Popover | Radix UI |
| Combobox | Ariakit |
| Command palette | `cmdk` |
| Rich text editor | Tiptap + `@tiptap/pm` (via `@probo/prosemirror`) |
| Diagrams | mermaid (lazy-loaded) |
| Markdown rendering | `react-markdown` |
| File drop | `react-dropzone` |

---

## 6. Forms — `react-hook-form` + Zod resolver

The intended pattern (universal target — current state is mixed):

```tsx
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { Field, Input, Button } from "@probo/ui";

const schema = z.object({
  name: z.string().min(1),
});

const { register, handleSubmit, formState: { errors } } = useForm({
  resolver: zodResolver(schema),
  defaultValues: { name: "" },
});
```

`@probo/ui`'s `Field`, `Input`, `Select` (with `Controller` for controlled libs), `Combobox`, and
form primitives integrate with `react-hook-form`. **Mixing plain `useState` with a form is a
known divergence** — refactor opportunistically.

---

## 7. Mutation Handler Naming

> PR-mining (frequency 2): PR #1073 — *"not fan of `commit*` naming."*

The destructured tuple from `useMutation` must be named after the **action verb**, not `commitFoo`:

```tsx
// Good
const [createFinding, isCreating] = useMutation(CreateFindingMutation);
const [deleteCookieBanner, isDeleting] = useMutation(DeleteCookieBannerMutation);

// Bad
const [commitMutation, isInFlight] = useMutation(...);
const [commitCreateFinding] = useMutation(...);
```

Likewise, prefer **store updates** to refetch:

> PR-mining (frequency 3): PR #1000 — *"Why should we refetch instead of just updating the store?"*

If the mutation response carries the new node, use `@appendEdge` / `@prependEdge` /
`@deleteEdge` or an `updater` callback. `refetch()` is reserved for cases where the update would
require querying additional fields the mutation didn't return.

---

## 8. Error Boundary Chain

> See [shared.md § 11 Error-Handling Principles](../shared.md#11-error-handling-principles) for the
> project-wide philosophy. Stack-specific implementation:

`@probo/relay` (in `packages/relay/src/`) exports six typed error classes:
`UnAuthenticatedError`, `ForbiddenError`, `FullNameRequiredError`, `AssumptionRequiredError`,
`NDASignatureRequiredError`, plus a generic `GraphQLError`. The fetch wrapper (`makeFetchQuery`)
inspects each GraphQL response's `errors[].extensions.code` and throws the matching subclass.

**Order of `instanceof` checks is significant** in `RootErrorBoundary`:

1. `UnAuthenticatedError` → redirect to `/auth/login`
2. `FullNameRequiredError` → redirect to `/auth/onboarding`
3. `AssumptionRequiredError` → re-prompt org assumption
4. `NDASignatureRequiredError` → redirect to NDA flow
5. `ForbiddenError` → render forbidden page
6. Default → render generic `PageError`

If you reorder this chain, more-specific errors get caught by less-specific handlers. Each subclass
**must** call `Object.setPrototypeOf(this, NewError.prototype)` after `super()` — without it,
TypeScript's class transpilation breaks the prototype chain and `instanceof` returns `false`.

`OrganizationErrorBoundary` is layered inside org-scoped routes so org-deletion / cross-org access
fail there before bubbling to root.

---

## 9. Email Templates — `@probo/emails`

`packages/emails` builds React Email JSX into pre-rendered HTML/text Go templates:

```
src/<Template>.tsx          → React Email JSX with placeholders
scripts/build.ts            → tsx + render() → dist/<Template>.html.tmpl
                                              dist/<Template>.txt.tmpl
pkg/mailer (Go)             → //go:embed dist/*.tmpl
```

**Placeholders are Go template strings inside JSX literals**: `<p>Hello {`{{.UserFullName}}`}</p>`.
There is **no TS-side type-check** of placeholder names — a typo only surfaces at Go runtime when
`text/template` complains about an unknown field.

**`dist/` MUST be pre-built before `go build`**, otherwise `//go:embed` picks up stale or empty
templates. `make build WITH_APPS=1` and the email package's own build step take care of it; if you
edit a template manually, run the package build before `go build`.

Branding tokens (color palette, header logo) live in `packages/emails/src/EmailLayout.tsx` — they
are the source of truth for emails, separate from the Tailwind theme used by the apps.

---

## 10. n8n Node Feature Slice — `@probo/n8n-node`

> Source: [`contrib/claude/n8n.md`](../../../contrib/claude/n8n.md).

Every resource is a folder under `packages/n8n-node/nodes/Probo/actions/<resource>/` with one file
per operation plus an `index.ts` barrel:

```
actions/
  framework/
    create.operation.ts
    delete.operation.ts
    list.operation.ts
    index.ts                ← exports a `framework` resource module
  index.ts                  ← resources map
```

**The export name in the resource `index.ts` MUST match the operation value string** that the n8n
properties array uses. A typo here makes the operation invisible at runtime.

Two GraphQL helpers — pick the right one for the API surface:

| Helper | Surface |
| --- | --- |
| `proboApiRequest` | Console / Trust GraphQL |
| `proboConnectApiRequest` | IAM (`/api/connect/v1/graphql`) |

Using `proboApiRequest` for an IAM operation hits the wrong endpoint with the wrong auth flow.

The four-surface API rule (see [shared.md § 3](../shared.md#3-the-four-surface-api-rule)) requires
that any backend operation also lands here — `npx n8n-node lint` must pass.

---

## 11. Observability

The frontend has **no structured logging framework**. Per [shared.md § 8](../shared.md#8-logging-principles-cross-stack):

- `console.*` is acceptable in browser code; PII rules still apply (no emails, names, tokens).
- User-facing errors flow through `useToast` from `@probo/ui` after `formatError(__, err)` from
  `@probo/helpers` translates them to a localized message.
- No `console.log` in production paths in `apps/console` (verified across the 437 files); reserve
  it for dev-time debugging.

# Probo — TypeScript Frontend — Conventions

> For commit format, branching, license headers, GIDs, four-surface API rule, configuration
> propagation, and PII-free logging: see [shared.md](../shared.md).
> Authoritative source: [`contrib/claude/ts-style.md`](../../../contrib/claude/ts-style.md),
> [`contrib/claude/react-components.md`](../../../contrib/claude/react-components.md),
> [`contrib/claude/ui.md`](../../../contrib/claude/ui.md).

---

## 1. File and Symbol Naming

| Kind | Convention | Examples |
| --- | --- | --- |
| Component file (`.tsx`) | PascalCase, matches default export | `FindingsPage.tsx`, `BannerSettingsForm.tsx`, `RootErrorBoundary.tsx` |
| Page-loader file | `<Feature>PageLoader.tsx` | `FindingsPageLoader.tsx`, `NewOrganizationPageLoader.tsx` |
| Skeleton | `<Feature>PageSkeleton.tsx` co-located with the Page | `FindingsPageSkeleton.tsx` |
| Layout | `<Feature>Layout.tsx` | `OrganizationLayout.tsx` |
| Hook | `useXxx.ts`, camelCase | `useOrganizationId.ts`, `useToggle.ts`, `usePageTitle.ts` |
| Route file | `routes.ts` (colocated) — never `Routes.tsx` | `pages/iam/organizations/people/routes.ts` |
| Generated Relay types | `__generated__/{core,iam}/<Op>.graphql.ts` | (auto) |
| Sub-component folder | `_components/` (underscore-prefixed) | `pages/organizations/findings/_components/FindingRow.tsx` |
| `tv()` variants | `variants.ts` next to the component | `packages/ui/src/atoms/Button/variants.ts` |

**Child-route components use `*Page` suffix, not `*Tab`.** The arborescence guide is explicit:
child routes are full pages, not tabs. Existing `*Tab.tsx` (e.g. `VendorOverviewTab.tsx`) is
legacy — rename when you touch them.

**Path alias `#` → `src/`.** Always use `#/hooks/...`, `#/components/...`, `#/pages/...`. Never
relative `../../../`.

---

## 2. Component Code Style

> Source: [`contrib/claude/react-components.md`](../../../contrib/claude/react-components.md).

```tsx
// Good — function declaration, props destructured in parameters
export function FindingRow({ finding, onDelete }: FindingRowProps) {
  const data = useFragment(FindingFragment, finding);
  return <tr>...</tr>;
}

// Bad — arrow const + React.FC + body destructure
export const FindingRow: React.FC<FindingRowProps> = (props) => {
  const { finding, onDelete } = props;
  ...
};
```

Rules:

- **Function declarations**, not arrow `const` (legacy: `packages/ui/src/atoms/Button/Button.tsx`).
- **No `React.FC`**; type props on the parameter.
- **Destructure props in parameters**. Accept the bare `props` object only when the destructure
  line would exceed 100 chars.
- **Non-callback props before callback props** in interface declarations:
  ```ts
  interface FindingRowProps {
    finding: FindingRowFragment$key;
    isSelected: boolean;
    onDelete: (id: string) => void;
    onSelect: (id: string) => void;
  }
  ```
- Props interface name: `<ComponentName>Props`. Use `interface`, not `type`, unless the shape
  needs unions / mapped types.
- **Named exports for everything**, except *route-target* `*PageLoader` / `*Page` files which use
  `export default` so `lazy(() => import(...))` resolves cleanly.

---

## 3. Icons

> PR-mining (frequency 4, high): PR #957 — *"all SVGs should be in a react component"*.

- **Phosphor icons first.** `import { CookieIcon, TrashIcon } from "@phosphor-icons/react"`.
  Import named icons, not `*`.
- **`@probo/ui` Icon set** only when Phosphor has no equivalent (audit before adding).
- **Never inline SVG markup** in a component. Extract as a React component or — preferably — pick a
  Phosphor icon.
- **Never emoji** as iconography.

---

## 4. Types — Use Relay-Generated, Don't Redeclare

> PR-mining (frequency 4, high — highest-frequency frontend block): PR #800 *"We don't declare
> local types anymore, we should use relay generated types directly"*.

When a component reads GraphQL data, type it from the Relay-generated artifact:

```tsx
import type { FindingRowFragment$key, FindingRowFragment$data } from "#/__generated__/core/FindingRowFragment.graphql";

type Finding = FindingRowFragment$data; // do NOT redeclare
```

Use the shared helpers in `apps/console/src/types.ts`:

- `NodeOf<T>` — extracts the node type from a Relay connection edges array.
- `ItemOf<T>` — extracts the element type of any array.

Local interfaces are appropriate only for **non-GraphQL** shapes (form state, UI-only flags, route
params). The IAM environment's types live under `__generated__/iam/`; pages under `src/pages/iam/`
must import from there, not from `__generated__/core/`.

`AppRoute` (from `@probo/routes`) is asserted with `satisfies AppRoute[]` on every route array.

---

## 5. Translator Injection

`@probo/i18n` exports `useTranslate()` which returns a function commonly bound to `__`:

```tsx
const { __ } = useTranslate();
return <Button>{__("Delete finding")}</Button>;
```

**Universal rule:** every helper that returns user-facing text **takes `__: Translator` as its
first argument**:

```ts
// packages/helpers/src/format/formatError.ts
import type { Translator } from "@probo/i18n";

export function formatError(__: Translator, err: unknown): string { ... }
```

This includes `formatError`, date/time formatters that produce labels, and any string-builder
shared across modules. The reason: `@probo/helpers` cannot call hooks; passing `__` keeps the
helper pure and unit-testable.

> Current state: i18n is **dormant**. Loaders return `{}` and the language is hard-coded `"en"`.
> The injection pattern still applies — when translations land later, helpers are already
> compatible.

---

## 6. URL Construction

> Source: [`contrib/claude/ts-style.md`](../../../contrib/claude/ts-style.md). Cross-stack rule —
> see also [shared.md § 12](../shared.md#12-security-baseline-cross-stack).

```ts
// Good
const url = new URL("/api/console/v1/graphql", window.location.origin);
url.searchParams.set("organizationId", organizationId);

// Or for in-app navigation paths:
const path = `/organizations/${encodeURIComponent(organizationId)}/findings`;
```

```ts
// Bad
const url = "/api/console/v1/graphql?organizationId=" + organizationId;
const url = `/api/console/v1/graphql?organizationId=${organizationId}`;
```

Never template-literal-concat URLs. Use `URL`, `URLSearchParams`, and `encodeURIComponent` for any
caller-controlled path segment.

---

## 7. Imports, Barrels, Module Structure

- **`@probo/helpers` uses a barrel `index.ts`.** When you add a new helper, **update
  `packages/helpers/src/index.ts`** or the symbol is invisible to callers. Same for any
  `packages/*` that exports through a barrel.
- Barrel re-exports are `export * from './foo'` or `export { foo } from './foo'`. Don't introduce
  a default export.
- **Import generated Relay types** with `import type { ... }` so they don't appear in the runtime
  bundle.

---

## 8. Reusing `@probo/ui` — Don't Re-implement

> PR-mining (frequency 2): PR #957 — *"nothing to reuse from @probo/ui here instead?"*.

Before writing a new local component in `apps/console/src/components/` or a `_components/` folder,
search `packages/ui/src/`. The library already covers Button, Card, Field, Dialog, Dropdown,
Select, Combobox, Tabs, Badge, Avatar, Toast, Confirm, PageHeader, Table primitives,
Tooltip, Popover, Skeleton primitives, etc.

Local components should:

- Compose `@probo/ui` primitives — never duplicate styles for "just a slightly different button".
- If a variant is missing in `@probo/ui`, add the variant **in `@probo/ui`** rather than fork it
  in `apps/console`.

---

## 9. Mutation Result Naming and Store Updates

(Repeated here because reviewers enforce it on every PR — full detail in
[patterns.md § Mutation Handler Naming](./patterns.md#7-mutation-handler-naming).)

```tsx
// Good
const [createFinding, isCreating] = useMutation(CreateFindingMutation);

// Bad
const [commit, isInFlight] = useMutation(...);
```

Avoid `refetch()` after a mutation — use `@appendEdge` / `@deleteEdge` / `updater`. PR-mining
(frequency 3): PR #1000.

---

## 10. Branding Tokens

| Surface | Source of truth |
| --- | --- |
| App UI (Console, Trust) | `packages/ui` Tailwind theme + tokens; consumed via `tailwindcss` v4 |
| Email templates | `packages/emails/src/EmailLayout.tsx` (color palette, fonts, header logo) |

These two are **deliberately separate** — emails ship as static HTML rendered at build time and
must not depend on Tailwind runtime. Update both when you change brand colors.

---

## 11. tsconfig — Per-Workspace, No Root

There is **no root `tsconfig.json`**. Each workspace owns its own `tsconfig.json`, typically
extending one of the presets in `packages/tsconfig/`.

> Known gap: **`@probo/helpers` does not enable `strict` mode.** Other packages do. Treat strict as
> the default for new packages; if `strict` is off, that is a debt to flag, not a license to relax.

---

## 12. License Headers

See [shared.md § 6](../shared.md#6-license-headersisc-on-every-source-file). Apply the
ISC header to every `.ts` / `.tsx` / `.css` / `.js` file with `//` (or `/* */` for CSS) comments,
expanding the year range when editing an existing file.

---

## 13. Review-Enforced Standards (Frontend Subset)

These come straight from PR mining (see [shared.md § 13](../shared.md#13-code-review-enforced-standards)
for the full table). They are **de-facto blockers**:

1. **Use Relay-generated types**, never local TS types duplicating GraphQL output (frequency 4).
2. **GraphQL fields whose resolvers can fail must use `@required`**, not non-null `!` (frequency 4).
3. **Inline SVGs are forbidden** — Phosphor icons or extracted React components only (frequency 4).
4. **Prefer `@probo/ui` primitives** over duplicated local UI (frequency 2).
5. **Mutations should update the Relay store**, not refetch when the mutation response carries the
   data (frequency 3).
6. **`commit*` is not a good name** for a mutation handler — use the action verb (frequency 2).

---

## 14. Git & Workflow

See [shared.md § 5](../shared.md#5-git--workflow) for branching (`{author}/{kebab-desc}`),
commit format (free-form, imperative, never Conventional Commits, never `Co-Authored-By` for AI),
the dual-sign requirement (`git commit -s -S`), and the rebase-only merge strategy.

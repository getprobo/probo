# React component conventions

This document describes **how to define and shape** React components in Probo frontends (`apps/compliance-portal`, [`packages/ui`](../../packages/ui), and related apps). It complements styling and package layout in [`contrib/claude/ui.md`](ui.md) and data loading in [`contrib/claude/relay.md`](relay.md).

These rules are the **source of truth**. Where existing code (e.g. `apps/console` or the legacy `@probo/ui` `Atoms/`/`Molecules/` tree) disagrees, the code is non-compliant and should be migrated — it is not precedent.

## Related guides

| Topic | Guide |
|-------|--------|
| `@probo/ui`, Tailwind, `tailwind-variants`, folders, skeletons, compound modules | [`contrib/claude/ui.md`](ui.md) |
| Relay queries, fragments, loaders, `queryRef` | [`contrib/claude/relay.md`](relay.md) |
| App folder layout, route segments, special folders | [`contrib/claude/app-arborescence.md`](app-arborescence.md) |
| Error boundaries, error/fallback props, async `try`/`catch` | [`contrib/claude/error-handling.md`](error-handling.md) |
| i18next, `_locales`, translation keys | [`contrib/claude/i18n.md`](i18n.md) |
| Forms and validation (Base UI Field/Form, zod, react-hook-form) | [`contrib/claude/forms.md`](forms.md) |
| Routing, navigation, URL state, auth | [`contrib/claude/routing.md`](routing.md) |
| Client state (Relay / URL / local / context / zustand) | [`contrib/claude/state-management.md`](state-management.md) |
| Permission-gated UI | [`contrib/claude/permissions.md`](permissions.md) |

## Destructuring

**Never destructure a value you do not use.** If only one element of a tuple or object is needed, stop destructuring at that element or omit the unused keys. Do not assign to `_`-prefixed throwaway names.

### Do / don't: unused destructured values

```tsx
// Bad — _isMoving is never read
const [moveCookie, _isMoving] =
  useMutation<MoveCookieMutation>(moveCookieMutation);
```

```tsx
// Good — stop at the last element you need
const [moveCookie] =
  useMutation<MoveCookieMutation>(moveCookieMutation);
```

## Component shape

| Rule | Convention |
|------|------------|
| Paradigm | **Functional components only.** Class components are not used except in rare cases that require lifecycle methods unavailable as hooks (e.g. `ErrorBoundary`). |
| Syntax | **Traditional `function` declarations**, not `const` storing arrow functions. |
| Typing | **Do not use `React.FC` (or `FC`).** Props are typed via the function's parameter; the return type is inferred. |
| Props | **Always destructure props.** Prefer destructuring in the function parameters. When that would make the declaration line exceed the lint line-length limit, accept `props` as the parameter and destructure in the function body. |

### Do / don't: component syntax

```tsx
// Bad — arrow function assigned to const + FC
const UserCard: FC<UserCardProps> = ({ name }) => {
  return <div>{name}</div>;
};
```

```tsx
// Bad — arrow function without FC
const UserCard = ({ name }: UserCardProps) => {
  return <div>{name}</div>;
};
```

```tsx
// Good — traditional function declaration, no FC
export function UserCard({ name }: UserCardProps) {
  return <div>{name}</div>;
}
```

### Do / don't: props destructuring

```tsx
// Bad — accessing props without destructuring
export function UserCard(props: UserCardProps) {
  return <div>{props.name}</div>;
}
```

```tsx
// Good — destructure in function parameters (preferred)
export function UserCard({ name }: UserCardProps) {
  return <div>{name}</div>;
}
```

```tsx
// Good — destructure in body when parameter-level destructuring would exceed the line-length limit
export function ThirdPartyComplianceOverviewPanel(
  props: ThirdPartyComplianceOverviewPanelProps,
) {
  const { className, thirdPartyKey, onStatusChange } = props;
  // …
}
```

## File and export

| Rule | Convention |
|------|------------|
| Components per file | **One primary component per file.** Colocate non-UI modules separately (`variants.ts`, `graphql` template strings, tiny helpers). |
| File name | **Matches the component name** in PascalCase (e.g. `UserProfileHeader.tsx` → `UserProfileHeader`). |
| Export | **Named export** (`export function UserProfileHeader`). **Exception:** route or lazy bundle **entry** components may use `export default` when the router or `lazy()` requires it (see relay.md route pages). |
| Props type | **`ComponentNameProps`**. Prefer **`interface`**; use **`type`** when you need unions, mapped types, or **`PropsWithChildren<…>`** (e.g. wrappers whose props are only `children`). |

### Do / don’t: file, name, and export

```tsx
// Bad — two components in one file (split into Panel.tsx and PanelSection.tsx)
export function Panel({ children }: PanelProps) {
  return <div>{children}</div>;
}
export function PanelSection({ children }: PanelSectionProps) {
  return <section>{children}</section>;
}
```

```tsx
// Good — one component per file: PanelSection.tsx
import type { PropsWithChildren } from "react";

export type PanelSectionProps = PropsWithChildren;

export function PanelSection({ children }: PanelSectionProps) {
  return <section>{children}</section>;
}
```

```tsx
// Bad — file UserThing.tsx exports Thing; name should match
export function Thing({ label }: ThingProps) {
  return null;
}
```

```tsx
// Good — file Thing.tsx
export interface ThingProps {
  label: string;
}

export function Thing({ label }: ThingProps) {
  return <span>{label}</span>;
}
```

```tsx
// Good — rare exception: route entry default export (names still clear in module)
type ThirdPartiesPageProps = {
  queryRef: PreloadedQuery<ThirdPartiesQuery>;
};

export default function ThirdPartiesPage({ queryRef }: ThirdPartiesPageProps) {
  // …
}
```

## Naming and suffixes

Component names are built from a **base** (the resource or concept) plus a **role suffix**. The suffix tells you what kind of component it is at a glance, and mirrors the file's role in the route tree. UI-kit primitives are the exception — they use bare names (see [`contrib/claude/ui.md`](ui.md)).

### Route / tree-level suffixes

| Suffix | Role |
|--------|------|
| `*Page` | A leaf page rendering final content (no `<Outlet />`). Default export when it is the `lazy()` entry. |
| `*Loader` | The lazy bundle entry: sets up providers, triggers the Relay query, renders a skeleton, then mounts the page. |
| `*Layout` | A layout route that renders shared chrome and an `<Outlet />`. |
| `*Skeleton` | The loading placeholder for a page, section, or component. |
| `*Error` | The error UI rendered by a boundary at this level (see [`contrib/claude/error-handling.md`](error-handling.md)). |
| `*Provider` | A context / Relay provider wrapper. |

### Content / section-level suffixes

| Suffix | Role |
|--------|------|
| `*Section` | A logical section of a page (owns its own fragment). |
| `*List` | The component that renders a **collection** (the index/list region, including its empty + loading states). |
| `*ListItem` | A **single item** within a `*List`. This is the canonical connection-item suffix. |
| `*Form` | A form (owns its fields + submit wiring). |
| `*Field` | A single form field. |
| `*Empty` | An empty-state placeholder. |

### Overlay suffixes

`*Dialog`, `*Drawer`, `*Menu` — overlay surfaces, styled from the headless primitive (see [`contrib/claude/ui.md`](ui.md)).

### Do / don't: `List` / `ListItem` over `Table` / `Row`

A collection's identity is "a list of things"; whether it is laid out as a table, cards, or rows is a **presentation detail** that can change. Name after the data, not the current layout.

```text
// Bad — named after the current visual treatment
ThirdPartiesTable.tsx
ThirdPartyRow.tsx
ThirdPartyCard.tsx

// Good — named after the collection and its item
ThirdPartyList.tsx
ThirdPartyListItem.tsx
```

`*ListItem` replaces both the old `*Row` (table) and `*Card` (card list) connection-item suffixes. See [Connection items are components](#connection-items-are-components).

## Props ordering

Within `ComponentNameProps`, order members as follows:

1. **Non-callback props first** — DOM/React attributes (`className`, `style`, …), `ref` (or `forwardRef` typing), static UI configuration (`title`, `hideSidebar`), initial UX state (`defaultOpen`, `initialTab`).
2. **Callback props last** — `onClose`, `onSave`, `onOpenChange`, etc.

### Do / don’t: prop order

```tsx
// Bad — callbacks mixed before configuration
interface FormActionsProps {
  onSave: () => void;
  title: string;
  className?: string;
  onCancel: () => void;
}
```

```tsx
// Good — configure first, then callbacks
interface FormActionsProps {
  className?: string;
  title: string;
  initialOpen?: boolean;
  onOpenChange?: (open: boolean) => void;
  onSave: () => void;
  onCancel: () => void;
}
```

## Props are for configuration and composition, not data

**Do not use props to pass data** in the broad sense: not fetched domain records, not lists of DTOs, and not identifiers that the component (or a dedicated hook) could read from the URL via React Router’s `useParams` or a hook built on it.

Props **configure** how a component behaves or looks, or **compose** it with UI fragments. **Everything else belongs in hooks** (Relay, router, local state, context, etc.) inside the component or an immediate parent that owns real wiring.

### Configure

Use props for:

- Standard HTML element attributes and React patterns: `className`, `style`, `id`, `aria-*`, `role`, and **`ref`** (including forwarded refs).
- **Static UI parameters:** `title`, `variant`, `hideSidebar`, `align`.
- **Initial client state** (parent does not own the live state): `defaultOpen`, `initialValue` — paired with `on*` if the parent must react.
- **Parent coordination callbacks** so the parent can update **its** state: `onCloseDropdown`, `onSubmit`, `onSelectionChange`.

### Compose

Use props for:

- **`children`** and other **`ReactNode` slots** (`header`, `footer`, `icon`) that are **UI building blocks**, not serialized API payloads.
- Render props or slot components when they express **layout or UI variation**, not “here is the loaded entity.”

### Hooks for data and URL-derived identity

- **Fetched data:** Colocate Relay fragments and queries per [`contrib/claude/relay.md`](relay.md) (`useFragment`, `useLazyLoadQuery`, `usePreloadedQuery`, etc.) in the component that needs the data.
- **Route parameters:** Call `useParams()` (or a small `useOrganizationId()`-style hook) **inside** the component that needs the id — avoid drilling `organizationId` / `thirdPartyId` from a parent that only read the URL to pass them down.

### Relay: framework wiring is not “business data props”

Relay sometimes requires **opaque handles** on props: e.g. **`queryRef`** for `usePreloadedQuery` on route pages, or a **fragment key** (`SomeFragment$key`) for `useFragment`. Those are **GraphQL/Relay wiring**, not passing arbitrary loaded objects through the tree. Keep using the patterns in [`contrib/claude/relay.md`](relay.md). Do not use those exceptions as a reason to pass plain domain objects or URL ids as props when a hook could read them instead.

### Do / don’t: URL params

```tsx
// Bad — parent only needed the param to pass it down
function ThirdPartyLayout() {
  const { thirdPartyId } = useParams();
  return (
    <main>
      <ThirdPartySummary thirdPartyId={thirdPartyId!} />
    </main>
  );
}

function ThirdPartySummary({ thirdPartyId }: { thirdPartyId: string }) {
  return <div>{/* … */}</div>;
}
```

```tsx
// Good — component that needs the id reads it (or uses a dedicated hook)
function ThirdPartyLayout() {
  return (
    <main>
      <ThirdPartySummary />
    </main>
  );
}

function ThirdPartySummary() {
  const { thirdPartyId } = useParams();
  if (thirdPartyId == null) {
    return null;
  }
  return <div>{/* use thirdPartyId in a hook / query … */}</div>;
}
```

### Do / don’t: fetched data

```tsx
// Bad — parent loaded data and passes fields as props
function ThirdPartyPage() {
  const thirdParty = useLazyLoadQuery(/* … */);
  return (
    <ThirdPartyHeader
      name={thirdParty.name}
      riskScore={thirdParty.riskScore}
      updatedAt={thirdParty.updatedAt}
    />
  );
}
```

```tsx
// Good — header colocates its fragment and reads via useFragment
const thirdPartyHeaderFragment = graphql`
  fragment ThirdPartyHeader_thirdParty on ThirdParty {
    name
    riskScore
    updatedAt
  }
`;

interface ThirdPartyHeaderProps {
  className?: string;
  thirdPartyKey: ThirdPartyHeader_thirdParty$key;
}

export function ThirdPartyHeader({ className, thirdPartyKey }: ThirdPartyHeaderProps) {
  const thirdParty = useFragment(thirdPartyHeaderFragment, thirdPartyKey);
  return (
    <header className={className}>
      {/* render from thirdParty … */}
    </header>
  );
}
```

### Do / don’t: composition vs data-as-props

```tsx
// Bad — every display field is a prop filled from fetched data elsewhere
interface ContactCardProps {
  fullName: string;
  email: string;
  role: string;
  createdAt: string;
}

export function ContactCard({ fullName, email }: ContactCardProps) {
  return (
    <article>
      <h2>{fullName}</h2>
      <p>{email}</p>
    </article>
  );
}
```

```tsx
// Good — props configure layout / slots; content is composed or read via hooks
import type { ReactNode } from "react";

interface PageSectionProps {
  className?: string;
  title: string;
  icon?: ReactNode;
  children: ReactNode;
}

export function PageSection({ className, title, icon, children }: PageSectionProps) {
  return (
    <section className={className}>
      <h2>
        {icon}
        {title}
      </h2>
      {children}
    </section>
  );
}
```

## Error and fallback props

Error handling is not reserved for route boundaries. A component that performs work which can fail (a query, a parse, a risky render) should be **wrappable in a boundary at any level**, and components that own a fallible region may expose error-handling props so the surrounding subtree can render its own error UI instead of taking down the whole page.

- `fallback` / `errorFallback` — a `ReactNode` (or render function receiving the error) shown when the wrapped content fails.
- `onError` — a callback invoked when the boundary catches, for logging / toasts.

These are **configuration / composition** props (UI slots and callbacks), so they obey the same rules as the rest of this guide — they never carry fetched domain data. The full pattern (the reusable `ErrorBoundary`, where to place it, and the `async` event-handler `try`/`catch` that boundaries cannot catch) lives in [`contrib/claude/error-handling.md`](error-handling.md).

```tsx
// Good — a section that can fail accepts a fallback slot and an onError callback
interface RiskSummarySectionProps {
  fallback?: ReactNode;
  onError?: (error: Error) => void;
}

export function RiskSummarySection({ fallback, onError }: RiskSummarySectionProps) {
  return (
    <ErrorBoundary fallback={fallback} onError={onError}>
      <RiskSummarySectionContent />
    </ErrorBoundary>
  );
}
```

## Interaction-triggered data

Data that is only needed after a user interaction (opening a dropdown, clicking a button, hovering) must **not** be fetched at page load. Instead, the parent component owns the query lifecycle with `useQueryLoader`, triggers `loadQuery` in the interaction event handler, and passes `queryRef` to a child component that reads data with `usePreloadedQuery`.

This applies to any data displayed after interaction, not at page load: dropdown menus with server-sourced options, hover cards, expandable panels, dialogs, etc.

### Pattern

1. The **parent** component calls `useQueryLoader` and triggers `loadQuery` on the interaction event (e.g. `onOpenChange` for a dropdown).
2. A **child** component exports its query, receives `queryRef` as a prop, and reads data with `usePreloadedQuery`.
3. The parent wraps the child in a `Suspense` boundary and only renders it when `queryRef` is available.
4. IDs needed for the query come from `useParams` in the parent — they are **not** passed as data props from a grandparent.

### Do / don't: dropdown with server-sourced options

```tsx
// Bad — page fetches categories at load and drills them as a data prop
function DetectionPage() {
  const data = usePreloadedQuery(/* query that includes consentCategories */);
  const categories = data.consentCategories.edges.map(e => e.node);
  return <PatternRow categories={categories} />;
}
```

```tsx
// Good — parent owns useQueryLoader, child reads with usePreloadedQuery

// PatternRow.tsx (parent — owns query lifecycle)
import { moveToCategoryQuery, MoveToCategoryMenu } from "./MoveToCategoryMenu";

function PatternRow() {
  const { cookieBannerId } = useParams<{ cookieBannerId: string }>();
  const [queryRef, loadQuery] = useQueryLoader<Query>(moveToCategoryQuery);

  const handleOpenChange = useCallback((open: boolean) => {
    if (open && cookieBannerId) {
      loadQuery({ cookieBannerId });
    }
  }, [loadQuery, cookieBannerId]);

  return (
    <Dropdown onOpenChange={handleOpenChange} toggle={/* … */}>
      {queryRef && (
        <Suspense>
          <MoveToCategoryMenu queryRef={queryRef} onMove={handleMove} />
        </Suspense>
      )}
    </Dropdown>
  );
}
```

```tsx
// MoveToCategoryMenu.tsx (child — reads data)
export const moveToCategoryQuery = graphql`
  query MoveToCategoryMenuQuery($cookieBannerId: ID!) { /* … */ }
`;

export function MoveToCategoryMenu({ queryRef, onMove }: Props) {
  const data = usePreloadedQuery(moveToCategoryQuery, queryRef);
  return /* render DropdownItems */;
}
```

See also the "Interaction-triggered queries" section in [`contrib/claude/relay.md`](relay.md).

(Snippet names and GraphQL types are illustrative; align with real schema and fragment names in the app.)

## Page sections are components

When a detail page has multiple distinct sections (e.g. a properties card and a paginated list), extract each section into its own component in `_components/`. Each section owns a colocated Relay fragment so that field additions never modify the parent page's query.

The page spreads the section fragments on the shared node and passes the fragment key:

```tsx
// Page query — spreads section fragments
export const detailPageQuery = graphql`
  query DetailPageQuery($nodeId: ID!) {
    node(id: $nodeId) {
      ... on MyType {
        id
        displayName
        ...MyTypePropertiesSection_myType
        ...MyTypeListSection_myType
      }
    }
  }
`;

// Page JSX:
<MyTypePropertiesSection myTypeKey={node} />
<MyTypeListSection myTypeKey={node} />
```

```tsx
// _components/MyTypePropertiesSection.tsx
const fragment = graphql`
  fragment MyTypePropertiesSection_myType on MyType {
    field1
    field2
  }
`;

interface MyTypePropertiesSectionProps {
  myTypeKey: MyTypePropertiesSection_myType$key;
}

export function MyTypePropertiesSection({ myTypeKey }: MyTypePropertiesSectionProps) {
  const data = useFragment(fragment, myTypeKey);
  return <Card padded>{/* PropertyRows */}</Card>;
}
```

For sections that own a paginated connection, use `usePaginationFragment` with a `@refetchable` fragment — the same pattern as a standalone page, but scoped to a section component.

## Connection items are components

When rendering items from a Relay connection (e.g. `edges.map(…)`), each item **must** be a dedicated component with its own colocated fragment — never inline the rendering of node fields directly in the parent's `.map()` body.

Place the item component in `_components/` adjacent to the page. Name it `<Type>ListItem` after the GraphQL type it renders (e.g. `DetectedTrackerListItem.tsx`, `ThirdPartyListItem.tsx`) — never `*Row` or `*Card` (see [Naming and suffixes](#naming-and-suffixes)). The component receives a single fragment key prop (e.g. `detectedTrackerKey: DetectedTrackerListItem_detectedTracker$key`) and calls `useFragment` internally.

```tsx
// Parent (the *List component) — spreads the item fragment in the connection:
edges { node { id ...DetectedTrackerListItem_detectedTracker } }

// Parent JSX:
{trackers.map(tracker => (
  <DetectedTrackerListItem key={tracker.id} detectedTrackerKey={tracker} />
))}
```

This ensures field additions/removals in the item never modify the parent's fragment, and keeps the item independently testable. The layout the item renders (a table row, a card, a plain `<li>`) is internal to the component and does not affect its name.

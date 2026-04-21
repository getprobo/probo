# React component conventions

This document describes **how to define and shape** React components in Probo frontends (`apps/console`, [`packages/ui`](../../packages/ui), and related apps). It complements styling and package layout in [`contrib/claude/ui.md`](ui.md) and data loading in [`contrib/claude/relay.md`](relay.md).

**The codebase does not fully match these rules yet.** Treat this guide as the target for new work and refactors.

## Related guides

| Topic | Guide |
|-------|--------|
| `@probo/ui`, Tailwind, `tailwind-variants`, folders, skeletons, compound modules | [`contrib/claude/ui.md`](ui.md) |
| Relay queries, fragments, loaders, `queryRef` | [`contrib/claude/relay.md`](relay.md) |

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
export function VendorComplianceOverviewPanel(
  props: VendorComplianceOverviewPanelProps,
) {
  const { className, vendorKey, onStatusChange } = props;
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
type VendorsPageProps = {
  queryRef: PreloadedQuery<VendorsQuery>;
};

export default function VendorsPage({ queryRef }: VendorsPageProps) {
  // …
}
```

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
- **Route parameters:** Call `useParams()` (or a small `useOrganizationId()`-style hook) **inside** the component that needs the id — avoid drilling `organizationId` / `vendorId` from a parent that only read the URL to pass them down.

### Relay: framework wiring is not “business data props”

Relay sometimes requires **opaque handles** on props: e.g. **`queryRef`** for `usePreloadedQuery` on route pages, or a **fragment key** (`SomeFragment$key`) for `useFragment`. Those are **GraphQL/Relay wiring**, not passing arbitrary loaded objects through the tree. Keep using the patterns in [`contrib/claude/relay.md`](relay.md). Do not use those exceptions as a reason to pass plain domain objects or URL ids as props when a hook could read them instead.

### Do / don’t: URL params

```tsx
// Bad — parent only needed the param to pass it down
function VendorLayout() {
  const { vendorId } = useParams();
  return (
    <main>
      <VendorSummary vendorId={vendorId!} />
    </main>
  );
}

function VendorSummary({ vendorId }: { vendorId: string }) {
  return <div>{/* … */}</div>;
}
```

```tsx
// Good — component that needs the id reads it (or uses a dedicated hook)
function VendorLayout() {
  return (
    <main>
      <VendorSummary />
    </main>
  );
}

function VendorSummary() {
  const { vendorId } = useParams();
  if (vendorId == null) {
    return null;
  }
  return <div>{/* use vendorId in a hook / query … */}</div>;
}
```

### Do / don’t: fetched data

```tsx
// Bad — parent loaded data and passes fields as props
function VendorPage() {
  const vendor = useLazyLoadQuery(/* … */);
  return (
    <VendorHeader
      name={vendor.name}
      riskScore={vendor.riskScore}
      updatedAt={vendor.updatedAt}
    />
  );
}
```

```tsx
// Good — header colocates its fragment and reads via useFragment
const vendorHeaderFragment = graphql`
  fragment VendorHeader_vendor on Vendor {
    name
    riskScore
    updatedAt
  }
`;

interface VendorHeaderProps {
  className?: string;
  vendorKey: VendorHeader_vendor$key;
}

export function VendorHeader({ className, vendorKey }: VendorHeaderProps) {
  const vendor = useFragment(vendorHeaderFragment, vendorKey);
  return (
    <header className={className}>
      {/* render from vendor … */}
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

(Snippet names and GraphQL types are illustrative; align with real schema and fragment names in the app.)

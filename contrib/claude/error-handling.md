# Error handling (frontend)

Errors must be **containable at any level** of the tree, not only at the route root. A failure in one section, list, or widget should be able to render a local fallback without taking down the rest of the page. This guide covers the reusable `ErrorBoundary`, the error/fallback props that let any subtree opt in, and how to handle the errors boundaries **cannot** catch (async work and event handlers).

## Related guides

| Topic | Guide |
|-------|--------|
| Error/fallback props as configuration | [`contrib/claude/react-components.md`](react-components.md#error-and-fallback-props) |
| Where `*Error` files live in the tree | [`contrib/claude/app-arborescence.md`](app-arborescence.md) |
| UI for error states (`ErrorLayout`, …) | [`contrib/claude/ui.md`](ui.md) |

## Two kinds of errors

React error boundaries only catch errors thrown **during rendering, in lifecycle methods, and in the constructors of the tree below them**. They do **not** catch:

- errors in **event handlers** (`onClick`, `onSubmit`, …),
- errors in **async** code (`await`, `.then`, `setTimeout`),
- errors thrown in the boundary itself.

So there are two complementary tools:

1. **`ErrorBoundary`** — for render-time failures (including Relay/Suspense errors thrown while reading data). Place it at the level where you want the blast radius to stop.
2. **`try`/`catch`** — for event handlers and async work. Surface the result through a toast and/or by storing the error in state.

## `ErrorBoundary`

A single reusable class component (the sanctioned use of a class — see [`react-components.md`](react-components.md#component-shape)) is the only error boundary primitive. It is generic and works at route, section, or component level.

```tsx
// packages/ui/src/v2/ErrorBoundary/ErrorBoundary.tsx
import { Component, type ErrorInfo, type ReactNode } from "react";

export interface ErrorBoundaryProps {
  children: ReactNode;
  // A node, or a render function that receives the caught value + a reset fn.
  // `unknown` because anything can be thrown, not just an Error.
  fallback?: ReactNode | ((error: unknown, reset: () => void) => ReactNode);
  onError?: (error: unknown, info: ErrorInfo) => void;
}

interface ErrorBoundaryState {
  // Tracked separately from `error` so a falsy thrown value (null, 0, "") still
  // renders the fallback instead of looping back into the failing subtree.
  hasError: boolean;
  error: unknown;
}

export class ErrorBoundary extends Component<ErrorBoundaryProps, ErrorBoundaryState> {
  state: ErrorBoundaryState = { hasError: false, error: null };

  static getDerivedStateFromError(error: unknown): ErrorBoundaryState {
    return { hasError: true, error };
  }

  componentDidCatch(error: unknown, info: ErrorInfo) {
    this.props.onError?.(error, info);
  }

  reset = () => this.setState({ hasError: false, error: null });

  render() {
    if (this.state.hasError) {
      const { fallback } = this.props;
      if (typeof fallback === "function") {
        return fallback(this.state.error, this.reset);
      }
      return fallback ?? null;
    }
    return this.props.children;
  }
}
```

### Trigger at any level

The same boundary wraps a whole route or a single widget — only the placement and the `fallback` differ.

```tsx
// Route level — a page's *Error file is the fallback
<ErrorBoundary fallback={<ThirdPartiesPageError />}>
  <ThirdPartiesPage queryRef={queryRef} />
</ErrorBoundary>
```

```tsx
// Section level — one failing section, the rest of the page survives
<ErrorBoundary
  fallback={(error, reset) => (
    <RiskSummarySectionError error={error} onRetry={reset} />
  )}
  onError={reportError}
>
  <RiskSummarySectionContent />
</ErrorBoundary>
```

In a router context, route boundaries are wired through the router's `ErrorBoundary` slot (e.g. `RootErrorBoundary` reading `useRouteError()`); `ErrorBoundary` above is for **in-page** boundaries below the route level.

### Components expose error/fallback props

Any component that owns a fallible region should let the surrounding subtree decide the fallback, by accepting `fallback` / `onError` and wrapping its risky content itself. These are configuration/composition props (a slot + a callback) — they never carry fetched data.

```tsx
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

## `try`/`catch` for events and async work

Boundaries will not catch a rejected promise in a submit handler. Wrap the risky call in `try`/`catch`, report via toast, and keep the UI responsive.

```tsx
// Good — async event handler guards itself; the boundary above can't help here
function PublishButton() {
  const { t } = useTranslation();
  const toast = Toast.useToastManager();
  const [isPublishing, setIsPublishing] = useState(false);

  async function onPublish() {
    setIsPublishing(true);
    try {
      await publishReport();
      toast.add({ title: t("reports.published"), type: "success" });
    } catch (error) {
      toast.add({
        title: t("reports.publishFailed"),
        description: error instanceof Error ? error.message : t("common.unknownError"),
        type: "error",
      });
    } finally {
      setIsPublishing(false);
    }
  }

  return <Button disabled={isPublishing} onClick={onPublish}>{t("reports.publish")}</Button>;
}
```

```tsx
// Bad — relying on an ErrorBoundary to catch an async rejection (it never will)
function PublishButton() {
  async function onPublish() {
    await publishReport(); // throws → unhandled rejection, boundary does not fire
  }
  return <Button onClick={onPublish}>Publish</Button>;
}
```

For Relay mutations, prefer the built-in `onCompleted` / `onError` callbacks (see [`relay.md`](relay.md)) over a manual `try`/`catch`; use `try`/`catch` for non-Relay async work (fetch, parsing, third-party SDKs).

## Relay field errors and fragment-level boundaries

A GraphQL response can be **partial**: `data` is present but one field carries an
error (with a `path`). To contain such a failure to the component that reads the
bad field — instead of collapsing the whole page — two pieces cooperate:

1. **The fetch layer only throws request-level errors.** A request-level error
   has **no `path`** (auth, malformed request, transport) and applies to the
   whole operation, so it throws and propagates to the nearest boundary.
   Field-level errors (those with a `path`) are **left in the response** so Relay
   can attribute them to the reading field.

   The compliance-portal wires this in its own
   [`apps/compliance-portal/src/lib/relay/fetch.ts`](../../apps/compliance-portal/src/lib/relay/fetch.ts)
   (it does **not** use `@probo/relay`'s `makeFetchQuery`, which throws for the
   whole operation on any known code).

2. **`@throwOnFieldError` on the query/fragment that reads the field.** With the
   directive set, a field error throws **at the read site** (`usePreloadedQuery`
   for a query, `useFragment` for a fragment). Put it on the **fragment** to
   isolate a section/row, and on the **query** to route page-level field errors
   to the route boundary.

Because `useFragment` throws in the component body (not in a child), the boundary
must be an **ancestor**. Split the component into a thin wrapper (holds the
`ErrorBoundary`) and a `*Content` child (reads the fragment):

```tsx
export function RecentUpdatesSection({ trustCenterKey }: Props) {
  const { t } = useTranslation();
  return (
    <ErrorBoundary
      // This section reads a field of the preloaded page query — there is no
      // local refetch, so recover with a reload, not the boundary's `reset`
      // (see "Retrying: reset vs refetch vs reload" below).
      fallback={(
        <InlineError
          message={t("errors.inline.message")}
          retryLabel={t("errors.inline.retry")}
          onRetry={() => window.location.reload()}
        />
      )}
    >
      <RecentUpdatesSectionContent trustCenterKey={trustCenterKey} />
    </ErrorBoundary>
  );
}

function RecentUpdatesSectionContent({ trustCenterKey }: Props) {
  const data = useFragment(fragment, trustCenterKey); // throws here on a field error
  // ...
}

const fragment = graphql`
  fragment RecentUpdatesSection_trustCenter on TrustCenter @throwOnFieldError { ... }
`;
```

The portal ships three fallback tiers, all backed by the same `ErrorBoundary`:

| Tier | Placement | Fallback |
|------|-----------|----------|
| Global | root route (`RootErrorBoundary`) | `ErrorState` full page (standalone) |
| Page | pathless child route inside the layout | `ErrorState` inside the shell (TopBar/footer survive) |
| Section / row | around a fragment-reading subtree | `InlineError` (vertical for sections, horizontal for rows) with a retry |

`ErrorState` and `InlineError` are presentational v2 kit components (see
[`ui.md`](ui.md)); the app maps the caught error to copy/actions and passes them
in.

### Retrying: reset vs refetch vs reload

A boundary's `reset` alone does **not** clear a Relay field error — at **any**
level, not just lists. `reset` only re-renders the subtree; the read hits the
**same errored record** still cached in the store and throws again. `reset` is
therefore only a real recovery for *transient render errors* (e.g. a non-Relay
render crash). To recover a Relay field error you must go back to the network
first, then clear the boundary. Pick the mechanism by what owns the data:

| Context | Recovery |
|---------|----------|
| Route / page boundary | `window.location.reload()` (or router revalidation) |
| Refetchable list/section (`useRefetchableFragment`) | `refetch(..., { fetchPolicy: "network-only" })`, then reset the boundary once it settles |
| Section reading a preloaded query (no local refetch) | reload the owning query via the loader's `loadQuery(..., { fetchPolicy: "network-only" })`, or fall back to `window.location.reload()` |
| Transient / non-Relay render error | `reset` |

In all the network cases, reset the boundary **after** the fetch settles (not
before), or the remount races the in-flight request straight back into the same
error.

`ListErrorBoundary` encapsulates the refetchable-list case: it owns a reset key
and exposes `onRetry(done)`, where the caller (which holds `refetch`, above the
boundary) refetches `network-only` and passes the `onComplete` callback as
`done`.

```tsx
// The page owns refetch; item fragments carry @throwOnFieldError, so a row's
// field error throws below the boundary while refetch survives above it.
<ListErrorBoundary
  onRetry={done => startTransition(() => {
    refetch(variables, { fetchPolicy: "network-only", onComplete: done });
  })}
>
  {rows}
</ListErrorBoundary>
```

Because `useRefetchableFragment` throws at its own read site, put
`@throwOnFieldError` on the **item** fragments (so the throw lands below the
boundary), not on the refetchable list fragment (whose read is above it — a
whole-connection failure there is a page-level error via `@required`).

## Custom errors for node-type mismatches

When a page fetches `node(id:)` and the resolved `__typename` is not the type the
view expects, throw a dedicated error, not a bare `Error`, so the boundary can
render the correct state (404):

```tsx
// Good — a typed error the boundary maps to the not-found page
import { NotFoundError } from "#/lib/relay/errors";

if (data.node?.__typename !== "MailingListUpdate") {
  throw new NotFoundError("Update not found");
}
```

```tsx
// Bad — an untyped error the boundary can only show as a generic failure
if (data.node?.__typename !== "MailingListUpdate") {
  throw new Error("Update not found");
}
```

See [`relay.md`](relay.md) (Node type guards).

## Placement guidance

- **Route root** — one boundary so an unhandled failure shows a full-page error instead of a blank screen.
- **Section / list / widget** — add a boundary around any independently-loaded region (especially Relay `Suspense` subtrees) so one failure degrades gracefully.
- **Interaction surfaces** (dropdowns, dialogs that load data on open) — wrap the lazily-loaded content so opening a broken menu doesn't crash the page.
- **Event/async paths** — `try`/`catch` + toast, never a boundary.

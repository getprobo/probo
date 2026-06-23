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
  // A node, or a render function that receives the caught error + a reset fn.
  fallback?: ReactNode | ((error: Error, reset: () => void) => ReactNode);
  onError?: (error: Error, info: ErrorInfo) => void;
}

interface ErrorBoundaryState {
  error: Error | null;
}

export class ErrorBoundary extends Component<ErrorBoundaryProps, ErrorBoundaryState> {
  state: ErrorBoundaryState = { error: null };

  static getDerivedStateFromError(error: Error): ErrorBoundaryState {
    return { error };
  }

  componentDidCatch(error: Error, info: ErrorInfo) {
    this.props.onError?.(error, info);
  }

  reset = () => this.setState({ error: null });

  render() {
    const { error } = this.state;
    if (error) {
      const { fallback } = this.props;
      if (typeof fallback === "function") {
        return fallback(error, this.reset);
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
  const { __ } = useTranslate();
  const { toast } = useToast();
  const [isPublishing, setIsPublishing] = useState(false);

  async function onPublish() {
    setIsPublishing(true);
    try {
      await publishReport();
      toast({ title: __("Published"), variant: "success" });
    } catch (error) {
      toast({
        title: __("Publish failed"),
        description: error instanceof Error ? error.message : __("Unknown error"),
        variant: "error",
      });
    } finally {
      setIsPublishing(false);
    }
  }

  return <Button disabled={isPublishing} onClick={onPublish}>{__("Publish")}</Button>;
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

## Placement guidance

- **Route root** — one boundary so an unhandled failure shows a full-page error instead of a blank screen.
- **Section / list / widget** — add a boundary around any independently-loaded region (especially Relay `Suspense` subtrees) so one failure degrades gracefully.
- **Interaction surfaces** (dropdowns, dialogs that load data on open) — wrap the lazily-loaded content so opening a broken menu doesn't crash the page.
- **Event/async paths** — `try`/`catch` + toast, never a boundary.

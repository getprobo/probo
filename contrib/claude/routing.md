# Routing, navigation, and auth

Probo frontends route with [React Router](https://reactrouter.com/) (`react-router` v8), wrapped by the **`@probo/routes`** helpers and lazy-loaded with **`@probo/react-lazy`**. This guide covers how routes are declared, how to navigate and read params, how to use the URL as state, and how authenticated/protected routes are composed. Folder placement of route files is covered in [`app-arborescence.md`](app-arborescence.md); this guide is about the routing API itself.

## Related guides

| Topic | Guide |
|-------|--------|
| Where `routes.ts` lives, route-segment folders | [`contrib/claude/app-arborescence.md`](app-arborescence.md) |
| Loaders, `queryRef`, preloading | [`contrib/claude/relay.md`](relay.md) |
| Route error boundaries | [`contrib/claude/error-handling.md`](error-handling.md) |
| Permission-gated UI within a route | [`contrib/claude/permissions.md`](permissions.md) |

## `AppRoute` and the route tree

Routes are declared as `AppRoute[]` and converted with `routeFromAppRoute` before being handed to `createBrowserRouter`. `AppRoute` extends React Router's `RouteObject` with a `Fallback` component; `routeFromAppRoute` wraps the `Component` in a `Suspense` boundary using that `Fallback` automatically.

```ts
// routes.ts — one per resource folder (see app-arborescence.md)
import { lazy } from "@probo/react-lazy";
import { type AppRoute } from "@probo/routes";

import { MeasuresPageSkeleton } from "./MeasuresPageSkeleton";

export const measureRoutes = [
  {
    path: "measures",
    Fallback: MeasuresPageSkeleton,
    Component: lazy(() => import("./MeasuresPageLoader")),
  },
  {
    path: "measures/:measureId",
    Component: lazy(() => import("./MeasureDetailLayoutLoader")),
    children: [
      { path: "overview", Component: lazy(() => import("./overview/MeasureOverviewPage")) },
    ],
  },
] satisfies AppRoute[];
```

The app root spreads each resource's routes and maps them once:

```tsx
import { routeFromAppRoute } from "@probo/routes";
import { measureRoutes } from "./pages/organizations/measures/routes";

const routes = [
  { path: "/", Component: lazy(() => import("./pages/MainLayout")), children: [...measureRoutes] },
] satisfies AppRoute[];

export const router = createBrowserRouter(routes.map(routeFromAppRoute));
```

Rules:
- Routes declare only `path`, `Fallback`, `Component` (a lazy loader), `ErrorBoundary`, and `children` — **no Relay logic in the route object** (that lives in the `*Loader`; see [`relay.md`](relay.md)).
- Use `lazy()` from `@probo/react-lazy` for the `Component` so every page is code-split.
- `Fallback` is the route-level skeleton; reuse the page's `*Skeleton`.

## Navigation

Navigate **declaratively** with `Link` for anything the user clicks, and **imperatively** with `useNavigate` only after an effect (e.g. post-submit redirect).

```tsx
import { Link, useNavigate } from "react-router";

// Declarative — preferred
<Link to={`measures/${measureId}`}>{t("measures.view")}</Link>

// Imperative — only when navigation follows an action
const navigate = useNavigate();
navigate(`measures/${newId}`);
```

Build paths from segments; never hand-concatenate query strings (see [`ts-style.md`](ts-style.md) — use `URL` / `URLSearchParams`).

## Route params

Read params with `useParams` **inside the component that needs them** — do not drill them as props from a parent that only read the URL to pass them down (see [`react-components.md`](react-components.md#props-are-for-configuration-and-composition-not-data)). Params are always `string | undefined`; narrow before use.

```tsx
const { measureId } = useParams<{ measureId: string }>();
if (measureId == null) {
  return null;
}
```

Prefer a small dedicated hook (`useOrganizationId()`-style) when the same param is read across many components.

## URL as state (search params)

State that should survive reload, be shareable, or be linkable — the active tab, a filter, a sort, a search term, pagination cursors — belongs in the **URL**, not `useState`. Use `useSearchParams`.

```tsx
import { useSearchParams } from "react-router";

const [searchParams, setSearchParams] = useSearchParams();
const status = searchParams.get("status") ?? "OPEN";

function onStatusChange(next: string) {
  setSearchParams((prev) => {
    prev.set("status", next);
    return prev;
  });
}
```

See [`state-management.md`](state-management.md) for when to choose the URL over local/global state.

## Redirects

Redirect from a route `loader` (throwing `redirect`) for canonical/index redirects, and with `<Navigate>` for render-time redirects (e.g. role-based landing).

```tsx
// Index redirect from a loader
{ index: true, loader: () => { throw redirect("general"); } }

// Render-time redirect
<Navigate to="login" replace />
```

## Auth and protected routes

Authentication state lives in a **provider** near the root (a viewer / current-user context), not in route objects. Protected subtrees are composed by nesting routes under a layout/provider that loads the viewer; unauthenticated access is handled by the **route error boundary**, which redirects to login.

```tsx
// RootErrorBoundary — redirect to login on an auth error, render the error page otherwise
export function RootErrorBoundary() {
  const error = useRouteError();
  if (error instanceof UnAuthenticatedError) {
    const search = new URLSearchParams({ continue: window.location.href });
    return <Navigate to={{ pathname: "/auth/login", search: `?${search}` }} />;
  }
  return <PageError error={error instanceof Error ? error : new Error("unknown error")} />;
}
```

```tsx
// Role-based landing — redirect at render time based on the viewer's role
function OrganizationIndex() {
  const { role } = use(CurrentUser);
  switch (role) {
    case Role.EMPLOYEE: return <Navigate to="employee" />;
    case Role.AUDITOR:  return <Navigate to="measures" />;
    default:            return <Navigate to="tasks" />;
  }
}
```

Rules:
- The **server** authorizes every request; the client redirects on `UnAuthenticatedError` purely for UX.
- Per-action gating inside an authenticated page uses `permission(action:)` booleans (see [`permissions.md`](permissions.md)), not role checks.
- Attach `ErrorBoundary` at the boundary you want auth failures to bubble to (root for whole-app, a section boundary for embedded widgets — see [`error-handling.md`](error-handling.md)).

## No domain data through `Outlet` context

A layout **must not** pass fetched domain data to child routes via `useOutletContext`. Each child page that needs data follows the Loader + Page pattern with its **own** query — the same as a sibling page. This keeps every page independently loadable and avoids hidden coupling to the parent's query.

```tsx
// Bad — layout fetches and forwards domain data through Outlet context
<Outlet context={{ showBranding: banner.showBranding }} />;
const { showBranding } = useOutletContext<{ showBranding: boolean }>();

// Good — the child page owns its loader + query (see relay.md)
const [queryRef, loadQuery] = useQueryLoader(snippetPageQuery);
useEffect(() => { loadQuery({ cookieBannerId }); }, [loadQuery, cookieBannerId]);
```

`Outlet` context is fine for **non-domain** UI coordination (a layout-owned callback, a `ref`), never for loaded entities or URL ids a child can read itself.

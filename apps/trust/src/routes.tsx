import {
  createBrowserRouter,
  Navigate,
  redirect,
  type RouteObject,
  useLoaderData,
  useRouteError,
} from "react-router";
import { type ComponentType, Fragment, Suspense } from "react";
import {
  relayEnvironment,
  UnAuthenticatedError,
} from "./providers/RelayProviders";
import { loadQuery, type PreloadedQuery } from "react-relay";
import { useCleanup } from "./hooks/useDelayedEffect";
import { PageError } from "./components/PageError";
import { MainLayout } from "/layouts/MainLayout";
import {
  currentTrustGraphQuery,
  currentTrustDocumentsQuery,
  currentTrustVendorsQuery,
} from "/queries/TrustGraph";
import { OverviewPage } from "/pages/OverviewPage";
import { DocumentsPage } from "/pages/DocumentsPage";
import { SubprocessorsPage } from "/pages/SubprocessorsPage";
import { AccessPage } from "./pages/AccessPage.tsx";
import { TabSkeleton } from "./components/Skeletons/TabSkeleton";
import { MainSkeleton } from "./components/Skeletons/MainSkeleton";

export type AppRoute = Omit<RouteObject, "Component" | "children"> & {
  Component?: ComponentType<any>;
  children?: AppRoute[];
  fallback?: ComponentType;
  queryLoader?: (params: any) => PreloadedQuery<any>;
};

/**
 * Top level error boundary
 */
function ErrorBoundary({ error: propsError }: { error?: string }) {
  const error = useRouteError() ?? propsError;

  if (error instanceof UnAuthenticatedError) {
    return <Navigate to="/auth/login" />;
  }

  return <PageError error={error?.toString()} />;
}

const routes = [
  {
    path: "/",
    loader: async () => {
      throw redirect("/overview");
    },
    Component: Fragment,
    ErrorBoundary: ErrorBoundary,
  },
  // Custom domain routes (subdomain-based)
  {
    path: "/overview",
    queryLoader: () => loadQuery(relayEnvironment, currentTrustGraphQuery, {}),
    Component: MainLayout,
    fallback: MainSkeleton,
    ErrorBoundary: ErrorBoundary,
    children: [
      {
        path: "",
        fallback: TabSkeleton,
        Component: OverviewPage,
      },
    ],
  },
  {
    path: "/documents",
    queryLoader: () => loadQuery(relayEnvironment, currentTrustGraphQuery, {}),
    Component: MainLayout,
    fallback: MainSkeleton,
    ErrorBoundary: ErrorBoundary,
    children: [
      {
        path: "",
        queryLoader: () =>
          loadQuery(relayEnvironment, currentTrustDocumentsQuery, {}),
        fallback: TabSkeleton,
        Component: DocumentsPage,
      },
    ],
  },
  {
    path: "/subprocessors",
    queryLoader: () => loadQuery(relayEnvironment, currentTrustGraphQuery, {}),
    Component: MainLayout,
    fallback: MainSkeleton,
    ErrorBoundary: ErrorBoundary,
    children: [
      {
        path: "",
        queryLoader: () =>
          loadQuery(relayEnvironment, currentTrustVendorsQuery, {}),
        fallback: TabSkeleton,
        Component: SubprocessorsPage,
      },
    ],
  },
  {
    path: "/access",
    Component: AccessPage,
    ErrorBoundary: ErrorBoundary,
  },
  // Fallback URL to the NotFound Page
  {
    path: "*",
    Component: PageError,
  },
] satisfies AppRoute[];

/**
 * Wrap components with suspense to handle lazy loading & relay loading states
 */
function routeTransformer({
  fallback: FallbackComponent,
  queryLoader,
  ...route
}: AppRoute): RouteObject {
  let result = { ...route };
  if (FallbackComponent && route.Component) {
    const OriginalComponent = route.Component;
    result = {
      ...result,
      Component: (props) => (
        <Suspense fallback={<FallbackComponent />}>
          <OriginalComponent {...props} />
        </Suspense>
      ),
    };
  }
  if (queryLoader && route.Component) {
    const OriginalComponent = route.Component;
    result = {
      ...result,
      loader: ({ params }) => {
        const query = queryLoader(params as Record<string, string>);
        return {
          queryRef: query,
          dispose: query.dispose,
        };
      },
      Component: () => {
        const { queryRef, dispose } = useLoaderData();

        useCleanup(dispose, 1000);

        return (
          <Suspense fallback={FallbackComponent ? <FallbackComponent /> : null}>
            <OriginalComponent queryRef={queryRef} />
          </Suspense>
        );
      },
    };
  }
  return {
    ...result,
    children: route.children?.map(routeTransformer),
  } as RouteObject;
}

// Detect basename from current URL path
// If URL starts with /trust/{slug}, extract that as the basename
// Otherwise, use "/" for custom domains
function getBasename(): string {
  const path = window.location.pathname;
  const trustMatch = path.match(/^\/trust\/[^/]+/);
  return trustMatch ? trustMatch[0] : "/";
}

export const router = createBrowserRouter(routes.map(routeTransformer), {
  basename: getBasename(),
});

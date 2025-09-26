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
  trustGraphQuery,
  trustDocumentsQuery,
  trustVendorsQuery,
} from "/queries/TrustGraph";
import { Overview } from "/pages/Overview";
import { Documents } from "/pages/Documents";
import { Subprocessors } from "/pages/Subprocessors";
import { Access } from "./pages/Access";
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
    Component: Fragment,
  },
  {
    path: "/trust/:slug/access",
    Component: Access,
  },
  {
    path: "/trust/:slug",
    queryLoader: ({ slug }) =>
      loadQuery(relayEnvironment, trustGraphQuery, { slug: slug }),
    Component: MainLayout,
    fallback: MainSkeleton,
    children: [
      {
        path: "",
        loader: ({ params }) => {
          throw redirect(`/trust/${params.slug}/overview`);
        },
      },
      {
        path: "overview",
        fallback: TabSkeleton,
        Component: Overview,
      },
      {
        path: "documents",
        fallback: TabSkeleton,
        Component: Documents,
        queryLoader: ({ slug }) =>
          loadQuery(relayEnvironment, trustDocumentsQuery, { slug: slug }),
      },
      {
        path: "subprocessors",
        fallback: TabSkeleton,
        Component: Subprocessors,
        queryLoader: ({ slug }) =>
          loadQuery(relayEnvironment, trustVendorsQuery, { slug: slug }),
      },
    ],
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

export const router = createBrowserRouter(routes.map(routeTransformer));

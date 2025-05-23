import {
  createBrowserRouter,
  useLoaderData,
  useRouteError,
  type RouteObject,
} from "react-router";
import { MainLayout } from "./layouts/MainLayout";
import { AuthLayout, CenteredLayout, CenteredLayoutSkeleton } from "@probo/ui";
import {
  lazy,
  Suspense,
  useEffect,
  type FC,
  type LazyExoticComponent,
} from "react";
import {
  relayEnvironment,
  UnAuthenticatedError,
} from "./providers/RelayProviders";
import { RisksPageSkeleton } from "./components/skeletons/RisksPageSkeleton.tsx";
import { PageSkeleton } from "./components/skeletons/PageSkeleton.tsx";
import { loadQuery, type PreloadedQuery } from "react-relay";
import { riskNodeQuery, risksQuery } from "./hooks/graph/RiskGraph.ts";
import { useCleanup } from "./hooks/useDelayedEffect.ts";

function ErrorBoundary() {
  const error = useRouteError();
  if (error instanceof UnAuthenticatedError) {
    return <div></div>;
  }
  console.log(error);
  return <div>error</div>;
}

type Route = {
  Component: FC<any> | LazyExoticComponent<FC<any>>;
  children?: Route[];
  fallback?: FC;
  queryLoader?: (params: Record<string, string>) => PreloadedQuery<any>;
} & Omit<RouteObject, "Component" | "children">;

const routes = [
  {
    path: "/auth",
    Component: AuthLayout,
    children: [
      {
        path: "login",
        Component: lazy(() => import("./pages/auth/LoginPage")),
      },
    ],
  },
  {
    path: "/",
    Component: CenteredLayout,
    fallback: CenteredLayoutSkeleton,
    children: [
      {
        path: "",
        Component: lazy(() => import("./pages/OrganizationsPage")),
      },
      {
        path: "organizations/new",
        Component: lazy(
          () => import("./pages/organizations/NewOrganizationPage"),
        ),
      },
    ],
  },
  {
    path: "/organizations/:organizationId",
    Component: MainLayout,
    ErrorBoundary: ErrorBoundary,
    children: [
      {
        path: "vendors",
        fallback: PageSkeleton,
        Component: lazy(
          () => import("./pages/organizations/vendors/VendorsPage"),
        ),
      },
      {
        path: "risks",
        fallback: RisksPageSkeleton,
        queryLoader: ({ organizationId }) =>
          loadQuery(relayEnvironment, risksQuery, { organizationId }),
        Component: lazy(() => import("./pages/organizations/risks/RisksPage")),
      },
      {
        path: "risks/:riskId",
        fallback: PageSkeleton,
        queryLoader: ({ riskId }) =>
          loadQuery(relayEnvironment, riskNodeQuery, { riskId }),
        Component: lazy(
          () => import("./pages/organizations/risks/RiskDetailPage"),
        ),
        children: [
          {
            path: "",
            Component: lazy(
              () => import("./pages/organizations/risks/RiskOverviewTab"),
            ),
          },
        ],
      },
    ],
  },
] satisfies Route[];

/**
 * Wrap component with a suspense to handle lazy loading & relay loading states
 */
function routeTransformer({
  fallback: FallbackComponent,
  queryLoader,
  ...route
}: Route): RouteObject {
  let result = { ...route };
  if (FallbackComponent) {
    result = {
      ...result,
      Component: (props) => (
        <Suspense fallback={<FallbackComponent />}>
          <route.Component {...props} />
        </Suspense>
      ),
    };
  }
  if (queryLoader) {
    result = {
      ...result,
      loader: ({ params }) => {
        const query = queryLoader(params as Record<string, string>);
        return {
          queryRef: query,
          dispose: () => {
            console.log("cleaning up query");
            query.dispose();
          },
        };
      },
      Component: () => {
        const { queryRef, dispose } = useLoaderData();

        useCleanup(dispose, 1000);

        return <route.Component queryRef={queryRef} />;
      },
    };
  }
  return {
    ...result,
    children: route.children?.map(routeTransformer),
  } as RouteObject;
}

export const router = createBrowserRouter(routes.map(routeTransformer));

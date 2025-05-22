import {
  createBrowserRouter,
  useRouteError,
  type RouteObject,
} from "react-router";
import { MainLayout } from "./layouts/MainLayout";
import { AuthLayout, CenteredLayout, CenteredLayoutSkeleton } from "@probo/ui";
import { lazy, Suspense, type FC } from "react";
import { UnAuthenticatedError } from "./providers/RelayProviders";
import { RisksPageSkeleton } from "./components/skeletons/RisksPageSkeleton.tsx";
import { PageSkeleton } from "./components/skeletons/PageSkeleton.tsx";

function ErrorBoundary() {
  const error = useRouteError();
  if (error instanceof UnAuthenticatedError) {
    return <div></div>;
  }
  console.log(error);
  return <div>error</div>;
}

type Route = {
  path: string;
  Component: FC;
  children?: Route[];
  ErrorBoundary?: FC;
  fallback?: FC;
};

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
        path: "risks",
        fallback: RisksPageSkeleton,
        Component: lazy(() => import("./pages/organizations/risks/RisksPage")),
      },
      {
        path: "risks/:riskId",
        fallback: PageSkeleton,
        Component: lazy(
          () => import("./pages/organizations/risks/RiskDetailPage"),
        ),
      },
    ],
  },
] satisfies Route[];

/**
 * Wrap component with a suspense to handle lazy loading & relay loading states
 */
function routeTransformer(route: Route): RouteObject {
  if ("fallback" in route && route.fallback) {
    const fallback = <route.fallback />;
    return {
      ...route,
      children: route.children?.map(routeTransformer),
      Component: () => (
        <Suspense fallback={fallback}>
          <route.Component />
        </Suspense>
      ),
    } as RouteObject;
  }
  return {
    ...route,
    children: route.children?.map(routeTransformer),
  };
}

export const router = createBrowserRouter(routes.map(routeTransformer));

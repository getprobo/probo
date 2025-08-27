import {
  createBrowserRouter,
  Navigate,
  redirect,
  useLoaderData,
  useRouteError,
  type RouteObject,
} from "react-router";
import { MainLayout } from "./layouts/MainLayout";
import { AuthLayout, CenteredLayout, CenteredLayoutSkeleton } from "@probo/ui";
import { Fragment, Suspense } from "react";
import {
  relayEnvironment,
  UnAuthenticatedError,
} from "./providers/RelayProviders";
import { PageSkeleton } from "./components/skeletons/PageSkeleton.tsx";
import { loadQuery, type PreloadedQuery } from "react-relay";
import { useCleanup } from "./hooks/useDelayedEffect.ts";
import { riskRoutes } from "./routes/riskRoutes.ts";
import { measureRoutes } from "./routes/measureRoutes.ts";
import { documentsRoutes } from "./routes/documentsRoutes.ts";
import { vendorRoutes } from "./routes/vendorRoutes.ts";
import { organizationViewQuery } from "./hooks/graph/OrganizationGraph.ts";
import { peopleRoutes } from "./routes/peopleRoutes.ts";
import { frameworkRoutes } from "./routes/frameworkRoutes.ts";
import { PageError } from "./components/PageError.tsx";
import { taskRoutes } from "./routes/taskRoutes.ts";
import { dataRoutes } from "./routes/dataRoutes.ts";
import { assetRoutes } from "./routes/assetRoutes.ts";
import { auditRoutes } from "./routes/auditRoutes.ts";
import { trustCenterRoutes } from "./routes/trustCenterRoutes.ts";
import { nonconformityRegistryRoutes } from "./routes/nonconformityRegistryRoutes.ts";
import { complianceRegistryRoutes } from "./routes/complianceRegistryRoutes.ts";
import { snapshotsRoutes } from "./routes/snapshotsRoutes.ts";
import { continualImprovementRegistryRoutes } from "./routes/continualImprovementRegistryRoutes.ts";
import { lazy } from "@probo/react-lazy";

export type AppRoute = Omit<RouteObject, "Component" | "children"> & {
  Component?: React.ComponentType<any>;
  children?: AppRoute[];
  fallback?: React.ComponentType;
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
    path: "/auth",
    Component: AuthLayout,
    children: [
      {
        path: "login",
        Component: lazy(() => import("./pages/auth/LoginPage")),
      },
      {
        path: "register",
        Component: lazy(() => import("./pages/auth/RegisterPage")),
      },
      {
        path: "confirm-email",
        Component: lazy(() => import("./pages/auth/ConfirmEmailPage")),
      },
      {
        path: "confirm-invitation",
        Component: lazy(() => import("./pages/auth/ConfirmInvitationPage")),
      },
      {
        path: "forgot-password",
        Component: lazy(() => import("./pages/auth/ForgotPasswordPage")),
      },
      {
        path: "reset-password",
        Component: lazy(() => import("./pages/auth/ResetPasswordPage")),
      },
    ],
  },
  {
    path: "/",
    Component: CenteredLayout,
    fallback: CenteredLayoutSkeleton,
    ErrorBoundary: ErrorBoundary,
    children: [
      {
        path: "",
        Component: lazy(() => import("./pages/OrganizationsPage")),
      },
      {
        path: "organizations/new",
        Component: lazy(
          () => import("./pages/organizations/NewOrganizationPage")
        ),
      },
      {
        path: "documents/signing-requests",
        Component: lazy(
          () => import("./pages/DocumentSigningRequestsPage.tsx")
        ),
      },
    ],
  },
  {
    path: "/trust/:slug",
    ErrorBoundary: ErrorBoundary,
    fallback: PageSkeleton,
    Component: lazy(() => import("./trust/pages/PublicTrustCenterPage")),
  },
  {
    path: "/trust/:slug/access",
    ErrorBoundary: ErrorBoundary,
    fallback: PageSkeleton,
    Component: lazy(() => import("./pages/TrustCenterAccessPage")),
  },
  {
    path: "/organizations/:organizationId",
    Component: MainLayout,
    ErrorBoundary: ErrorBoundary,
    children: [
      {
        path: "",
        loader: () => {
          throw redirect(`tasks`);
        },
        Component: Fragment,
      },
      {
        path: "settings",
        fallback: PageSkeleton,
        queryLoader: ({ organizationId }) =>
          loadQuery(relayEnvironment, organizationViewQuery, {
            organizationId,
          }),
        Component: lazy(() => import("./pages/organizations/SettingsPage")),
      },
      ...riskRoutes,
      ...measureRoutes,
      ...documentsRoutes,
      ...peopleRoutes,
      ...vendorRoutes,
      ...frameworkRoutes,
      ...taskRoutes,
      ...assetRoutes,
      ...dataRoutes,
      ...auditRoutes,
      ...nonconformityRegistryRoutes,
      ...complianceRegistryRoutes,
      ...snapshotsRoutes,
      ...continualImprovementRegistryRoutes,
      ...trustCenterRoutes,
      {
        path: "*",
        Component: PageError,
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

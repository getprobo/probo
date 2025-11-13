import {
  createBrowserRouter,
  Navigate,
  redirect,
  useLoaderData,
  useRouteError,
  type RouteObject,
} from "react-router";
import { Component as ReactComponent, type ErrorInfo } from "react";
import { MainLayout } from "./layouts/MainLayout";
import { AuthLayout, CenteredLayout, CenteredLayoutSkeleton } from "@probo/ui";
import { Fragment, Suspense } from "react";
import {
  relayEnvironment,
  UnAuthenticatedError,
  UnauthorizedError,
  ForbiddenError,
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
import { meetingsRoutes } from "./routes/meetingsRoutes.ts";
import { trustCenterRoutes } from "./routes/trustCenterRoutes.ts";
import { nonconformityRoutes } from "./routes/nonconformityRoutes.ts";
import { obligationRoutes } from "./routes/obligationRoutes.ts";
import { snapshotsRoutes } from "./routes/snapshotsRoutes.ts";
import { continualImprovementRoutes } from "./routes/continualImprovementRoutes.ts";
import { processingActivityRoutes } from "./routes/processingActivityRoutes.ts";
import { lazy } from "@probo/react-lazy";

export type AppRoute = Omit<RouteObject, "Component" | "children"> & {
  Component?: React.ComponentType<any>;
  children?: AppRoute[];
  fallback?: React.ComponentType;
  queryLoader?: (params: any) => PreloadedQuery<any>;
};

/**
 * Common error handling logic
 */
function renderError(error: unknown): React.ReactElement {
  if (error instanceof UnAuthenticatedError) {
    return <Navigate to="/auth/login" replace />;
  }

  if (error instanceof UnauthorizedError) {
    return <PageError error="UNAUTHORIZED" />;
  }

  if (error instanceof ForbiddenError) {
    return <PageError error="FORBIDDEN" />;
  }

  return <PageError error={error?.toString()} />;
}

/**
 * React Error Boundary for catching errors in Suspense
 */
class ReactErrorBoundary extends ReactComponent<
  { children: React.ReactNode },
  { error: Error | null }
> {
  constructor(props: { children: React.ReactNode }) {
    super(props);
    this.state = { error: null };
  }

  static getDerivedStateFromError(error: Error) {
    return { error };
  }

  componentDidCatch(error: Error, errorInfo: ErrorInfo) {
    console.error("ReactErrorBoundary caught error:", error, errorInfo);
  }

  render() {
    if (this.state.error) {
      return renderError(this.state.error);
    }

    return this.props.children;
  }
}

/**
 * Top level error boundary for Router errors
 */
function ErrorBoundary({ error: propsError }: { error?: string }) {
  const error = useRouteError() ?? propsError;
  return renderError(error);
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
        path: "signup-from-invitation",
        Component: lazy(() => import("./pages/auth/SignupFromInvitationPage")),
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
      {
        path: "api-keys",
        Component: lazy(() => import("./pages/APIKeysPage")),
      },
    ],
  },
  {
    path: "/organizations/:organizationId",
    Component: MainLayout,
    ErrorBoundary: ErrorBoundary,
    children: [
      {
        path: "",
        loader: () => {
          throw redirect(`documents`);
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
        children: [
          {
            path: "",
            loader: () => {
              throw redirect("general");
            },
          },
          {
            path: "general",
            Component: lazy(() => import("./pages/organizations/settings/GeneralSettingsTab")),
          },
          {
            path: "members",
            Component: lazy(() => import("./pages/organizations/settings/MembersSettingsTab")),
          },
          {
            path: "domain",
            Component: lazy(() => import("./pages/organizations/settings/DomainSettingsTab")),
          },
          {
            path: "saml-sso",
            Component: lazy(() => import("./pages/organizations/settings/SAMLSettingsTab")),
          },
        ],
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
      ...meetingsRoutes,
      ...nonconformityRoutes,
      ...obligationRoutes,
      ...continualImprovementRoutes,
      ...processingActivityRoutes,
      ...snapshotsRoutes,
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
        <ReactErrorBoundary>
          <Suspense fallback={<FallbackComponent />}>
            <OriginalComponent {...props} />
          </Suspense>
        </ReactErrorBoundary>
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
          <ReactErrorBoundary>
            <Suspense fallback={FallbackComponent ? <FallbackComponent /> : null}>
              <OriginalComponent queryRef={queryRef} />
            </Suspense>
          </ReactErrorBoundary>
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

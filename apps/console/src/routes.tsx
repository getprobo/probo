import {
  createBrowserRouter,
  Navigate,
  redirect,
  useLoaderData,
  useRouteError,
  type LoaderFunction,
  type LoaderFunctionArgs,
  type RouteObject,
} from "react-router";
import { MainLayout } from "./layouts/MainLayout";
import { AuthLayout, CenteredLayout, CenteredLayoutSkeleton } from "@probo/ui";
import { Fragment, Suspense, type ComponentType, type LazyExoticComponent } from "react";
import {
  relayEnvironment,
  UnAuthenticatedError,
  UnauthorizedError,
  ForbiddenError,
} from "./providers/RelayProviders";
import { PageSkeleton } from "./components/skeletons/PageSkeleton.tsx";
import { loadQuery, type EnvironmentProviderOptions, type PreloadedQuery } from "react-relay";
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
import type { OperationType } from "relay-runtime";

export function withQueryRef<
  TQuery extends OperationType,
  TEnvironmentProviderOptions = EnvironmentProviderOptions
>(
  Component: LazyExoticComponent<ComponentType<{ queryRef: PreloadedQuery<TQuery, TEnvironmentProviderOptions> }>>,
) {
  return () => {
    const { queryRef, dispose } = useLoaderData();

    useCleanup(dispose, 1000);

    return <Component queryRef={queryRef} />
  }
}

export function loaderFromQueryLoader<
  TQuery extends OperationType,
  TEnvironmentProviderOptions = EnvironmentProviderOptions
>(
  queryLoader: (params: Record<string, string>) => PreloadedQuery<TQuery, TEnvironmentProviderOptions>
): LoaderFunction {
  return ({ params }: LoaderFunctionArgs) => {
    const query = queryLoader(params as Record<string, string>);
    return {
      queryRef: query,
      dispose: query.dispose,
    };
  }
}

export type AppRoute = Omit<RouteObject, "children"> & {
  children?: AppRoute[];
  fallback?: ComponentType;
}

/**
 * Top level error boundary
 */
function ErrorBoundary({ error: propsError }: { error?: string }) {
  const error = useRouteError() ?? propsError;

  if (error instanceof UnAuthenticatedError) {
    return <Navigate to="/auth/login" />;
  }

  if (error instanceof UnauthorizedError) {
    return <PageError error="UNAUTHORIZED" />;
  }

  if (error instanceof ForbiddenError) {
    return <PageError error="FORBIDDEN" />;
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
          throw redirect(`tasks`);
        },
        Component: Fragment,
      },
      {
        path: "settings",
        fallback: PageSkeleton,
        loader: loaderFromQueryLoader(
          ({ organizationId }) =>
            loadQuery(relayEnvironment, organizationViewQuery, {
              organizationId: organizationId!,
            })
        ),
        Component: withQueryRef(lazy(() => import("./pages/organizations/SettingsPage"))),
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
  Component,
  fallback: FallbackComponent,
  children,
  ...rest
}: AppRoute): RouteObject {
  let route: RouteObject = { ...rest };

  if (Component && FallbackComponent) {
    const OriginalComponent = Component as ComponentType;

    route = {
      ...route,
      Component: () => (
        <Suspense fallback={<FallbackComponent />}>
          <OriginalComponent />
        </Suspense>
      ),
    };
  }

  return {
    ...route,
    children: children?.map(routeTransformer),
  } as RouteObject;
}

export const router = createBrowserRouter(routes.map(routeTransformer));

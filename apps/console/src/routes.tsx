import {
  createBrowserRouter,
  Navigate,
  redirect,
  useRouteError,
} from "react-router";
import { PageSkeleton } from "./components/skeletons/PageSkeleton.tsx";
import { riskRoutes } from "./routes/riskRoutes.ts";
import { measureRoutes } from "./routes/measureRoutes.ts";
import { documentsRoutes } from "./routes/documentsRoutes.ts";
import { vendorRoutes } from "./routes/vendorRoutes.ts";
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
import { rightsRequestRoutes } from "./routes/rightsRequestRoutes.ts";
import { processingActivityRoutes } from "./routes/processingActivityRoutes.ts";
import { statesOfApplicabilityRoutes } from "./routes/statesOfApplicabilityRoutes.ts";
import { lazy } from "@probo/react-lazy";
import { routeFromAppRoute, type AppRoute } from "@probo/routes";
import { Role } from "@probo/helpers";
import {
  ForbiddenError,
  UnAuthenticatedError,
  UnauthorizedError,
} from "@probo/relay";
import { CurrentUser } from "./providers/CurrentUser.tsx";
import { use } from "react";
import { ViewerLayoutLoading } from "./pages/iam/memberships/ViewerLayoutLoading.tsx";
import { CenteredLayout } from "@probo/ui";

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
    Component: lazy(() => import("./pages/iam/auth/AuthLayout")),
    children: [
      {
        path: "login",
        Component: lazy(() => import("./pages/iam/auth/sign-in/SignInPage")),
      },
      {
        path: "password-login",
        Component: lazy(
          () => import("./pages/iam/auth/sign-in/PasswordSignInPage"),
        ),
      },
      {
        path: "sso-login",
        Component: lazy(() => import("./pages/iam/auth/sign-in/SSOSignInPage")),
      },
      {
        path: "register",
        Component: lazy(() => import("./pages/iam/auth/sign-up/SignUpPage")),
      },
      {
        path: "verify-email",
        Component: lazy(() => import("./pages/iam/auth/VerifyEmailPage")),
      },
      {
        path: "signup-from-invitation",
        Component: lazy(
          () => import("./pages/iam/auth/sign-up/SignUpFromInvitationPage"),
        ),
      },
      {
        path: "forgot-password",
        Component: lazy(() => import("./pages/iam/auth/ForgotPasswordPage")),
      },
      {
        path: "reset-password",
        Component: lazy(() => import("./pages/iam/auth/ResetPasswordPage")),
      },
    ],
  },
  {
    path: "/",
    ErrorBoundary: ErrorBoundary,
    children: [
      {
        Component: lazy(() => import("./pages/iam/memberships/ViewerLayoutLoader")),
        Fallback: ViewerLayoutLoading,
        children: [
          {
            index: true,
            Component: lazy(
              () => import("./pages/iam/memberships/MembershipsPageLoader"),
            ),
          },
          {
            path: "me/api-keys",
            Component: lazy(
              () => import("./pages/iam/apiKeys/APIKeysPageLoader"),
            ),
          },
          {
            Component: CenteredLayout,
            children: [
              {
                path: "organizations/new",
                Component: lazy(
                  () => import("./pages/iam/organizations/NewOrganizationPage"),
                ),
              },
            ],
          },
        ]
      },
    ],
  },
  {
    path: "documents/signing-requests",
    ErrorBoundary: ErrorBoundary,
    Component: lazy(
      () => import("./pages/DocumentSigningRequestsPage"),
    ),
  },
  {
    path: "/organizations/:organizationId/employee",
    Component: lazy(
      () => import("./pages/organizations/employee/EmployeeLayoutLoader"),
    ),
    ErrorBoundary: ErrorBoundary,
    children: [
      {
        index: true,
        // Component: () => "hello world",
        Component: lazy(
          () =>
            import("./pages/organizations/employee/EmployeeDocumentsPageLoader"),
        ),
      },
      {
        path: ":documentId",
        ErrorBoundary: ErrorBoundary,
        Component: lazy(
          () =>
            import("./pages/organizations/employee/EmployeeDocumentSignaturePageLoader"),
        ),
      },
    ],
  },
  {
    path: "/organizations/:organizationId",
    Component: lazy(
      () => import("./pages/iam/organizations/ViewerMembershipLayoutLoader"),
    ),
    ErrorBoundary: ErrorBoundary,
    children: [
      {
        path: "",
        Component: () => {
          const { role } = use(CurrentUser);
          switch (role) {
            case Role.EMPLOYEE:
              return <Navigate to="employee" />;
            case Role.AUDITOR:
              return <Navigate to="measures" />;
            default:
              return <Navigate to="tasks" />;
          }
        },
      },
      {
        path: "settings",
        Fallback: PageSkeleton,
        Component: lazy(
          () => import("./pages/iam/organizations/settings/SettingsLayout"),
        ),
        children: [
          {
            index: true,
            loader: () => {
              throw redirect("general");
            },
          },
          {
            path: "general",
            Component: lazy(
              () =>
                import("./pages/iam/organizations/settings/GeneralSettingsPageLoader"),
            ),
          },
          {
            path: "members",
            Component: lazy(
              () =>
                import("./pages/iam/organizations/settings/MembersPageLoader"),
            ),
          },
          {
            path: "domain",
            Component: lazy(
              () =>
                import("./pages/organizations/settings/DomainSettingsPageLoader"),
            ),
          },
          {
            path: "saml-sso",
            Component: lazy(
              () =>
                import("./pages/iam/organizations/settings/SAMLSettingsPageLoader"),
            ),
          },
          {
            path: "scim",
            Component: lazy(
              () =>
                import("./pages/iam/organizations/settings/SCIMSettingsPageLoader"),
            ),
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
      ...rightsRequestRoutes,
      ...processingActivityRoutes,
      ...statesOfApplicabilityRoutes,
      ...trustCenterRoutes,
      ...snapshotsRoutes,
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

export const router = createBrowserRouter(routes.map(routeFromAppRoute));

import {
  createBrowserRouter,
  Navigate,
  redirect,
  useRouteError,
} from "react-router";
import { EmployeeLayout } from "./layouts/EmployeeLayout";
import { CenteredLayout, CenteredLayoutSkeleton } from "@probo/ui";
import { PageSkeleton } from "./components/skeletons/PageSkeleton.tsx";
import { loadQuery } from "react-relay";
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
import {
  loaderFromQueryLoader,
  routeFromAppRoute,
  withQueryRef,
  type AppRoute,
} from "@probo/routes";
import { employeeDocumentsQuery } from "./pages/organizations/employee/EmployeeDocumentsPage";
import { employeeDocumentSignatureQuery } from "./pages/organizations/employee/EmployeeDocumentSignaturePage";
import { Role } from "@probo/helpers";
import { PermissionsContext } from "./providers/PermissionsContext";
import { use } from "react";
import { coreEnvironment } from "./environments.ts";
import {
  ForbiddenError,
  UnAuthenticatedError,
  UnauthorizedError,
} from "@probo/relay";

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
        Component: lazy(() => import("./pages/iam/auth/SignInPage")),
      },
      {
        path: "register",
        Component: lazy(() => import("./pages/auth/RegisterPage")),
      },
      // {
      //   path: "confirm-email",
      //   Component: lazy(() => import("./pages/auth/ConfirmEmailPage")),
      // },
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
    Fallback: CenteredLayoutSkeleton,
    ErrorBoundary: ErrorBoundary,
    children: [
      {
        index: true,
        Component: lazy(
          () => import("./pages/iam/memberships/MembershipsPageLoader"),
        ),
      },
      {
        path: "organizations/new",
        Component: lazy(
          () => import("./pages/iam/organizations/NewOrganizationPage"),
        ),
      },
      {
        path: "documents/signing-requests",
        Component: lazy(
          () => import("./pages/DocumentSigningRequestsPage.tsx"),
        ),
      },
      {
        path: "api-keys",
        Component: lazy(() => import("./pages/APIKeysPage")),
      },
    ],
  },
  {
    path: "/organizations/:organizationId/employee",
    Component: EmployeeLayout,
    ErrorBoundary: ErrorBoundary,
    children: [
      {
        path: "",
        Fallback: PageSkeleton,
        loader: loaderFromQueryLoader(({ organizationId }) =>
          loadQuery(coreEnvironment, employeeDocumentsQuery, {
            organizationId: organizationId!,
          }),
        ),
        Component: withQueryRef(
          lazy(
            () =>
              import("./pages/organizations/employee/EmployeeDocumentsPage"),
          ),
        ),
      },
      {
        path: ":documentId",
        Fallback: PageSkeleton,
        ErrorBoundary: ErrorBoundary,
        loader: loaderFromQueryLoader(({ documentId }) =>
          loadQuery(coreEnvironment, employeeDocumentSignatureQuery, {
            documentId: documentId!,
          }),
        ),
        Component: withQueryRef(
          lazy(
            () =>
              import("./pages/organizations/employee/EmployeeDocumentSignaturePage"),
          ),
        ),
      },
    ],
  },
  {
    path: "/organizations/:organizationId",
    Component: lazy(
      () => import("./pages/iam/memberships/MembershipLayoutLoader.tsx"),
    ),
    ErrorBoundary: ErrorBoundary,
    children: [
      {
        path: "",
        Component: () => {
          const { role } = use(PermissionsContext);
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
          () => import("./pages/iam/organizations/settings/SettingsLayout.tsx"),
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
                import("./pages/iam/organizations/settings/GeneralSettingsPageLoader.tsx"),
            ),
          },
          // {
          //   path: "members",
          //   Component: lazy(
          //     () => import("./pages/organizations/settings/MembersSettingsTab")
          //   ),
          // },
          {
            path: "domain",
            Component: lazy(
              () => import("./pages/organizations/settings/DomainSettingsTab"),
            ),
          },
          // {
          //   path: "saml-sso",
          //   Component: lazy(
          //     () => import("./pages/organizations/settings/SAMLSettingsTab")
          //   ),
          // },
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

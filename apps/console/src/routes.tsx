import { Role } from "@probo/helpers";
import { lazy } from "@probo/react-lazy";
import { type AppRoute, routeFromAppRoute } from "@probo/routes";
import { CenteredLayout } from "@probo/ui";
import { use } from "react";
import {
  createBrowserRouter,
  Navigate,
  redirect,
} from "react-router";

import { OrganizationErrorBoundary } from "./components/OrganizationErrorBoundary";
import { PageError } from "./components/PageError";
import { RootErrorBoundary } from "./components/RootErrorBoundary";
import { PageSkeleton } from "./components/skeletons/PageSkeleton";
import { ViewerLayoutLoading } from "./pages/iam/memberships/ViewerLayoutLoading";
import { peopleRoutes } from "./pages/iam/organizations/people/routes";
import { compliancePageRoutes } from "./pages/organizations/compliance-page/routes";
import { CurrentUser } from "./providers/CurrentUser";
import { assetRoutes } from "./routes/assetRoutes";
import { auditRoutes } from "./routes/auditRoutes";
import { continualImprovementRoutes } from "./routes/continualImprovementRoutes";
import { dataRoutes } from "./routes/dataRoutes";
import { documentsRoutes } from "./routes/documentsRoutes";
import { frameworkRoutes } from "./routes/frameworkRoutes";
import { measureRoutes } from "./routes/measureRoutes";
import { meetingsRoutes } from "./routes/meetingsRoutes";
import { nonconformityRoutes } from "./routes/nonconformityRoutes";
import { obligationRoutes } from "./routes/obligationRoutes";
import { processingActivityRoutes } from "./routes/processingActivityRoutes";
import { rightsRequestRoutes } from "./routes/rightsRequestRoutes";
import { riskRoutes } from "./routes/riskRoutes";
import { snapshotsRoutes } from "./routes/snapshotsRoutes";
import { statesOfApplicabilityRoutes } from "./routes/statesOfApplicabilityRoutes";
import { taskRoutes } from "./routes/taskRoutes";
import { vendorRoutes } from "./routes/vendorRoutes";

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
    ErrorBoundary: RootErrorBoundary,
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
        ],
      },
    ],
  },
  {
    path: "documents/signing-requests",
    ErrorBoundary: RootErrorBoundary,
    Component: lazy(
      () => import("./pages/DocumentSigningRequestsPage"),
    ),
  },
  {
    path: "/organizations/:organizationId",
    children: [
      {
        path: "assume",
        Component: lazy(() => import("./pages/iam/organizations/AssumePage")),
      },
      {
        path: "employee",
        ErrorBoundary: OrganizationErrorBoundary,
        Component: lazy(
          () => import("./pages/organizations/employee/EmployeeLayoutLoader"),
        ),
        children: [
          {
            index: true,
            Component: lazy(
              () =>
                import("./pages/organizations/employee/EmployeeDocumentsPageLoader"),
            ),
          },
          {
            path: ":documentId",
            Component: lazy(
              () =>
                import("./pages/organizations/employee/EmployeeDocumentSignaturePageLoader"),
            ),
          },
        ],
      },
      {
        Component: lazy(
          () => import("./pages/iam/organizations/ViewerMembershipLayoutLoader"),
        ),
        ErrorBoundary: OrganizationErrorBoundary,
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
                  // eslint-disable-next-line
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
              {
                path: "webhooks",
                Component: lazy(
                  () =>
                    import("./pages/iam/organizations/settings/WebhooksSettingsPageLoader"),
                ),
              },
            ],
          },
          ...peopleRoutes,
          ...riskRoutes,
          ...measureRoutes,
          ...documentsRoutes,
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
          ...compliancePageRoutes,
          ...snapshotsRoutes,
          {
            path: "*",
            Component: PageError,
          },
        ],
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

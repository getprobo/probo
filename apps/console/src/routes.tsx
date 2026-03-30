// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

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
import { accessReviewRoutes } from "./routes/accessReviewRoutes";
import { assetRoutes } from "./routes/assetRoutes";
import { auditRoutes } from "./routes/auditRoutes";
import { contextRoutes } from "./routes/contextRoutes";
import { dataRoutes } from "./routes/dataRoutes";
import { documentsRoutes } from "./routes/documentsRoutes";
import { findingRoutes } from "./routes/findingRoutes";
import { frameworkRoutes } from "./routes/frameworkRoutes";
import { measureRoutes } from "./routes/measureRoutes";
import { obligationRoutes } from "./routes/obligationRoutes";
import { processingActivityRoutes } from "./routes/processingActivityRoutes";
import { rightsRequestRoutes } from "./routes/rightsRequestRoutes";
import { riskRoutes } from "./routes/riskRoutes";
import { snapshotsRoutes } from "./routes/snapshotsRoutes";
import { statementsOfApplicabilityRoutes } from "./routes/statementsOfApplicabilityRoutes";
import { taskRoutes } from "./routes/taskRoutes";
import { vendorRoutes } from "./routes/vendorRoutes";

const routes = [
  {
    path: "/auth",
    Component: lazy(() => import("./pages/iam/auth/AuthLayout")),
    children: [
      {
        path: "login",
        Component: lazy(
          () => import("./pages/iam/auth/sign-in/SignInPageLoader"),
        ),
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
        Component: lazy(() => import("./pages/iam/auth/SignUpPage")),
      },
      {
        path: "verify-email",
        Component: lazy(() => import("./pages/iam/auth/VerifyEmailPage")),
      },
      {
        path: "activate-account",
        Component: lazy(
          () => import("./pages/iam/auth/ActivateAccountPage"),
        ),
      },
      {
        path: "create-password",
        Component: lazy(
          () => import("./pages/iam/auth/CreatePasswordPage"),
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
      {
        path: "device",
        ErrorBoundary: RootErrorBoundary,
        Component: lazy(
          () => import("./pages/iam/auth/DeviceActivationPageLoader"),
        ),
      },
      {
        path: "consent",
        ErrorBoundary: RootErrorBoundary,
        Component: lazy(
          () => import("./pages/iam/auth/ConsentPageLoader"),
        ),
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
    path: "/organizations/:organizationId",
    children: [
      {
        path: "assume",
        Component: lazy(() => import("./pages/iam/organizations/AssumePageLoader")),
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
            loader: ({ params: { organizationId } }) => {
              // eslint-disable-next-line
              throw redirect(`/organizations/${organizationId}/employee/signatures`);
            },
            Component: () => null,
          },
          {
            Component: lazy(
              () => import("./pages/organizations/employee/EmployeeTabsLayout"),
            ),
            children: [
              {
                path: "signatures",
                Component: lazy(
                  () =>
                    import("./pages/organizations/employee/EmployeeDocumentsPageLoader"),
                ),
              },
              {
                path: "approvals",
                Component: lazy(
                  () =>
                    import("./pages/organizations/employee/EmployeeApprovalsPageLoader"),
                ),
              },
            ],
          },
          {
            path: ":documentId",
            loader: ({ params: { organizationId, documentId } }) => {
              // eslint-disable-next-line
              throw redirect(`/organizations/${organizationId}/employee/signatures/${documentId}`);
            },
            Component: () => null,
          },
          {
            path: "signatures/:documentId",
            Component: lazy(
              () =>
                import("./pages/organizations/employee/EmployeeDocumentSignaturePageLoader"),
            ),
          },
          {
            path: "approvals/:documentId",
            Component: lazy(
              () =>
                import("./pages/organizations/documents/approve/DocumentApprovePageLoader"),
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
              {
                path: "audit-log",
                Component: lazy(
                  () =>
                    import("./pages/iam/organizations/settings/AuditLogSettingsPageLoader"),
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
          ...contextRoutes,
          ...findingRoutes,
          ...obligationRoutes,
          ...rightsRequestRoutes,
          ...processingActivityRoutes,
          ...statementsOfApplicabilityRoutes,
          ...accessReviewRoutes,
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

// Copyright (c) 2026 Probo Inc <hello@probo.com>.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

import { lazy } from "@probo/react-lazy";
import type { AppRoute } from "@probo/routes";
import { redirect } from "react-router";

import { LinkCardSkeleton } from "#/components/skeletons/LinkCardSkeleton";
import { PageSkeleton } from "#/components/skeletons/PageSkeleton";

function redirectTo(to: string) {
  return () => {
    // eslint-disable-next-line @typescript-eslint/only-throw-error
    throw redirect(to);
  };
}

export const compliancePortalRoutes = [
  {
    path: "compliance-portals",
    Fallback: PageSkeleton,
    Component: lazy(() => import("#/pages/organizations/compliance-portals/CompliancePortalsIndexPageLoader")),
  },
  {
    path: "compliance-portals/new",
    Fallback: PageSkeleton,
    Component: lazy(() => import("#/pages/organizations/compliance-portals/NewCompliancePortalPage")),
  },
  {
    path: "compliance-page",
    loader: redirectTo("compliance-portals"),
  },
  {
    path: "compliance-page/*",
    loader: redirectTo("../compliance-portals"),
  },
  {
    path: "compliance-pages",
    loader: redirectTo("compliance-portals"),
  },
  {
    path: "compliance-pages/*",
    loader: redirectTo("../compliance-portals"),
  },
  {
    path: "compliance-portals/:compliancePortalId",
    Fallback: PageSkeleton,
    Component: lazy(() => import("#/pages/organizations/compliance-portals/CompliancePortalLayoutLoader")),
    children: [
      {
        path: "hosting",
        Fallback: LinkCardSkeleton,
        Component: lazy(() => import("#/pages/organizations/compliance-portals/hosting/CompliancePortalHostingPageLoader")),
      },
      {
        path: "permissions",
        Fallback: LinkCardSkeleton,
        Component: lazy(() => import("#/pages/organizations/compliance-portals/permissions/CompliancePortalPermissionsPageLoader")),
      },
      {
        path: "integrations",
        Fallback: LinkCardSkeleton,
        Component: lazy(() => import("#/pages/organizations/compliance-portals/integrations/CompliancePortalIntegrationsPageLoader")),
      },
      {
        path: "landing",
        Fallback: LinkCardSkeleton,
        Component: lazy(() => import("#/pages/organizations/compliance-portals/landing/CompliancePortalLandingLayoutLoader")),
        children: [
          {
            index: true,
            Fallback: LinkCardSkeleton,
            Component: lazy(() => import("#/pages/organizations/compliance-portals/landing/branding/CompliancePortalBrandingPageLoader")),
          },
          {
            path: "content",
            Fallback: LinkCardSkeleton,
            Component: lazy(() => import("#/pages/organizations/compliance-portals/landing/content/CompliancePortalContentPageLoader")),
          },
        ],
      },
      {
        path: "documents",
        Fallback: LinkCardSkeleton,
        Component: lazy(() => import("#/pages/organizations/compliance-portals/documents/CompliancePortalDocumentsLayoutLoader")),
        children: [
          {
            index: true,
            Fallback: LinkCardSkeleton,
            Component: lazy(() => import("#/pages/organizations/compliance-portals/documents/CompliancePortalDocumentsPageLoader")),
          },
          {
            path: "audits",
            Fallback: LinkCardSkeleton,
            Component: lazy(() => import("#/pages/organizations/compliance-portals/documents/audits/CompliancePortalAuditsPageLoader")),
          },
          {
            path: "files",
            Fallback: LinkCardSkeleton,
            Component: lazy(() => import("#/pages/organizations/compliance-portals/documents/files/CompliancePortalFilesPageLoader")),
          },
        ],
      },
      {
        path: "subprocessors",
        Fallback: LinkCardSkeleton,
        Component: lazy(() => import("#/pages/organizations/compliance-portals/subprocessors/CompliancePortalSubprocessorsPageLoader")),
      },
      {
        path: "updates",
        Fallback: LinkCardSkeleton,
        Component: lazy(() => import("#/pages/organizations/compliance-portals/updates/CompliancePortalUpdatesPageLoader")),
      },
      {
        path: "right-requests",
        Fallback: LinkCardSkeleton,
        Component: lazy(() => import("#/pages/organizations/compliance-portals/right-requests/CompliancePortalRightRequestsPageLoader")),
      },
    ],
  },
] satisfies AppRoute[];

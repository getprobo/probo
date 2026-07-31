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

import { ContentErrorBoundary } from "#/components/ContentErrorBoundary";
import { LinkCardSkeleton } from "#/components/skeletons/LinkCardSkeleton";
import { PageSkeleton } from "#/components/skeletons/PageSkeleton";

export const compliancePortalRoutes = [
  {
    path: "compliance-portals",
    Fallback: PageSkeleton,
    Component: lazy(() => import("#/pages/organizations/compliance-portals/overview/CompliancePortalsOverviewPageLoader")),
  },
  {
    path: "compliance-portals/new",
    Fallback: PageSkeleton,
    Component: lazy(() => import("#/pages/organizations/compliance-portals/NewCompliancePortalPage")),
  },
  {
    path: "compliance-page",
    loader: () => {
      // eslint-disable-next-line @typescript-eslint/only-throw-error
      throw redirect("compliance-portals");
    },
  },
  {
    path: "compliance-page/*",
    loader: () => {
      // eslint-disable-next-line @typescript-eslint/only-throw-error
      throw redirect("../compliance-portals");
    },
  },
  {
    path: "compliance-pages",
    loader: () => {
      // eslint-disable-next-line @typescript-eslint/only-throw-error
      throw redirect("compliance-portals");
    },
  },
  {
    path: "compliance-pages/*",
    loader: () => {
      // eslint-disable-next-line @typescript-eslint/only-throw-error
      throw redirect("../compliance-portals");
    },
  },
  {
    path: "compliance-portals/:compliancePortalId",
    Fallback: PageSkeleton,
    Component: lazy(() => import("#/pages/organizations/compliance-portals/configuration/CompliancePortalConfigLayoutLoader")),
    children: [
      {
        index: true,
        Fallback: LinkCardSkeleton,
        ErrorBoundary: ContentErrorBoundary,
        Component: lazy(() => import("#/pages/organizations/compliance-portals/configuration/overview/CompliancePortalOverviewPageLoader")),
      },
      {
        path: "brand",
        Fallback: LinkCardSkeleton,
        ErrorBoundary: ContentErrorBoundary,
        Component: lazy(() => import("#/pages/organizations/compliance-portals/configuration/brand/CompliancePortalBrandPageLoader")),
      },
      {
        path: "references",
        Fallback: LinkCardSkeleton,
        ErrorBoundary: ContentErrorBoundary,
        Component: lazy(() => import("#/pages/organizations/compliance-portals/configuration/references/CompliancePortalReferencesPageLoader")),
      },
      {
        path: "commitments",
        Fallback: LinkCardSkeleton,
        ErrorBoundary: ContentErrorBoundary,
        Component: lazy(() => import("#/pages/organizations/compliance-portals/configuration/commitments/CompliancePortalCommitmentsPageLoader")),
      },
      {
        path: "audits",
        Fallback: LinkCardSkeleton,
        ErrorBoundary: ContentErrorBoundary,
        Component: lazy(() => import("#/pages/organizations/compliance-portals/configuration/audits/CompliancePortalAuditsPageLoader")),
      },
      {
        path: "documents",
        Fallback: LinkCardSkeleton,
        ErrorBoundary: ContentErrorBoundary,
        Component: lazy(() => import("#/pages/organizations/compliance-portals/configuration/documents/CompliancePortalDocumentsPageLoader")),
      },
      {
        path: "files",
        Fallback: LinkCardSkeleton,
        ErrorBoundary: ContentErrorBoundary,
        Component: lazy(() => import("#/pages/organizations/compliance-portals/configuration/files/CompliancePortalFilesPageLoader")),
      },
      {
        path: "third-parties",
        Fallback: LinkCardSkeleton,
        ErrorBoundary: ContentErrorBoundary,
        Component: lazy(() => import("#/pages/organizations/compliance-portals/configuration/third-parties/CompliancePortalThirdPartiesPageLoader")),
      },
      {
        path: "access",
        Fallback: LinkCardSkeleton,
        ErrorBoundary: ContentErrorBoundary,
        Component: lazy(() => import("#/pages/organizations/compliance-portals/configuration/access/CompliancePortalAccessPageLoader")),
      },
      {
        path: "mailing-list",
        Fallback: LinkCardSkeleton,
        ErrorBoundary: ContentErrorBoundary,
        Component: lazy(() => import("#/pages/organizations/compliance-portals/configuration/mailing-list/CompliancePortalMailingListPageLoader")),
      },
    ],
  },
] satisfies AppRoute[];

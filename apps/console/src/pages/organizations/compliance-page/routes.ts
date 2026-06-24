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

import { LinkCardSkeleton } from "#/components/skeletons/LinkCardSkeleton";
import { PageSkeleton } from "#/components/skeletons/PageSkeleton";

export const compliancePageRoutes = [
  {
    path: "compliance-page",
    Fallback: PageSkeleton,
    Component: lazy(() => import("#/pages/organizations/compliance-page/CompliancePageLayoutLoader")),
    children: [
      {
        index: true,
        Fallback: LinkCardSkeleton,
        Component: lazy(() => import("#/pages/organizations/compliance-page/overview/CompliancePageOverviewPageLoader")),
      },
      {
        path: "domain",
        Fallback: LinkCardSkeleton,
        Component: lazy(
          () =>
            import("#/pages/organizations/compliance-page/domain/CompliancePageDomainPageLoader"),
        ),
      },
      {
        path: "brand",
        Fallback: LinkCardSkeleton,
        Component: lazy(() => import("#/pages/organizations/compliance-page/brand/CompliancePageBrandPageLoader")),
      },
      {
        path: "references",
        Fallback: LinkCardSkeleton,
        Component: lazy(() => import("#/pages/organizations/compliance-page/references/CompliancePageReferencesPageLoader")),
      },
      {
        path: "audits",
        Fallback: LinkCardSkeleton,
        Component: lazy(() => import("#/pages/organizations/compliance-page/audits/CompliancePageAuditsPageLoader")),
      },
      {
        path: "documents",
        Fallback: LinkCardSkeleton,
        Component: lazy(() => import("#/pages/organizations/compliance-page/documents/CompliancePageDocumentsPageLoader")),
      },
      {
        path: "files",
        Fallback: LinkCardSkeleton,
        Component: lazy(() => import("#/pages/organizations/compliance-page/files/CompliancePageFilesPageLoader")),
      },
      {
        path: "third-parties",
        Fallback: LinkCardSkeleton,
        Component: lazy(() => import("#/pages/organizations/compliance-page/third-parties/CompliancePageThirdPartiesPageLoader")),
      },
      {
        path: "access",
        Fallback: LinkCardSkeleton,
        Component: lazy(() => import("#/pages/organizations/compliance-page/access/CompliancePageAccessPageLoader")),
      },
      {
        path: "mailing-list",
        Fallback: LinkCardSkeleton,
        Component: lazy(() => import("#/pages/organizations/compliance-page/mailing-list/CompliancePageMailingListPageLoader")),
      },
    ],
  },
];

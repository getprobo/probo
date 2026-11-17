// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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
import { type AppRoute } from "@probo/routes";
import { Fragment } from "react";
import { type LoaderFunctionArgs, redirect } from "react-router";

import { LinkCardSkeleton } from "#/components/skeletons/LinkCardSkeleton";
import { PageSkeleton } from "#/components/skeletons/PageSkeleton";
import { detailLayout } from "#/pages/iam/organizations/_lib/shellLayout";

import { DocumentDetailsPageSkeleton } from "./DocumentDetailsPageSkeleton";

function documentTabs(prefix: string, includeIndexRedirect = true): AppRoute[] {
  const tabs: AppRoute[] = [
    {
      path: `${prefix}description`,
      Fallback: LinkCardSkeleton,
      Component: lazy(
        () =>
          import("./description/DocumentDescriptionPageLoader"),
      ),
    },
    {
      path: `${prefix}controls`,
      Fallback: LinkCardSkeleton,
      Component: lazy(
        () =>
          import("./controls/DocumentControlsPageLoader"),
      ),
    },
    {
      path: `${prefix}approvals`,
      Fallback: LinkCardSkeleton,
      Component: lazy(
        () =>
          import("./approvals/DocumentApprovalsPageLoader"),
      ),
    },
    {
      path: `${prefix}signatures`,
      Fallback: LinkCardSkeleton,
      Component: lazy(
        () =>
          import("./signatures/DocumentSignaturesPageLoader"),
      ),
    },
  ];

  if (includeIndexRedirect) {
    tabs.unshift({
      path: prefix,
      loader: ({
        params: { organizationId, documentId, versionId },
      }: LoaderFunctionArgs) => {
        const basePath = `/organizations/${organizationId}/governance/documents/${documentId}`;
        const redirectPath = versionId
          ? `${basePath}/versions/${versionId}/description`
          : `${basePath}/description`;
        // eslint-disable-next-line
        throw redirect(redirectPath);
      },
      Component: Fragment,
    });
  }

  return tabs;
}

export const documentsRoutes = [
  {
    path: "documents",
    Fallback: PageSkeleton,
    Component: lazy(() => import("./DocumentsPageLoader")),
  },
  {
    path: "documents/:documentId",
    children: [
      {
        index: true,
        handle: detailLayout,
        Fallback: DocumentDetailsPageSkeleton,
        Component: lazy(() => import("./DocumentDetailsPageLoader")),
      },
      {
        Fallback: PageSkeleton,
        Component: lazy(() => import("./DocumentLayoutLoader")),
        children: [
          ...documentTabs("", false),
          ...documentTabs("versions/:versionId/"),
        ],
      },
    ],
  },
] satisfies AppRoute[];

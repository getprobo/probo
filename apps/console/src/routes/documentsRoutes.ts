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

import { lazy } from "@probo/react-lazy";
import {
  type AppRoute,
} from "@probo/routes";
import { Fragment } from "react";
import { type LoaderFunctionArgs, redirect } from "react-router";

import { LinkCardSkeleton } from "#/components/skeletons/LinkCardSkeleton";
import { PageSkeleton } from "#/components/skeletons/PageSkeleton";

const documentTabs = (prefix: string) => {
  return [
    {
      path: `${prefix}`,
      loader: ({
        params: { organizationId, documentId, versionId },
      }: LoaderFunctionArgs) => {
        const basePath = `/organizations/${organizationId}/documents/${documentId}`;
        const redirectPath = versionId
          ? `${basePath}/versions/${versionId}/description`
          : `${basePath}/description`;
        // eslint-disable-next-line
        throw redirect(redirectPath);
      },
      Component: Fragment,
    },
    {
      path: `${prefix}description`,
      Fallback: LinkCardSkeleton,
      Component: lazy(
        () =>
          import("#/pages/organizations/documents/description/DocumentDescriptionPageLoader"),
      ),
    },
    {
      path: `${prefix}controls`,
      Fallback: LinkCardSkeleton,
      Component: lazy(
        () =>
          import("#/pages/organizations/documents/controls/DocumentControlsPageLoader"),
      ),
    },
    {
      path: `${prefix}approvals`,
      Fallback: LinkCardSkeleton,
      Component: lazy(
        () =>
          import("#/pages/organizations/documents/approvals/DocumentApprovalsPageLoader"),
      ),
    },
    {
      path: `${prefix}signatures`,
      Fallback: LinkCardSkeleton,
      Component: lazy(
        () =>
          import("#/pages/organizations/documents/signatures/DocumentSignaturesPageLoader"),
      ),
    },
  ];
};

export const documentsRoutes = [
  {
    path: "documents",
    Fallback: PageSkeleton,
    Component: lazy(() => import("#/pages/organizations/documents/DocumentsPageLoader")),
  },
  {
    path: "documents/:documentId",
    Fallback: PageSkeleton,
    Component: lazy(() => import("#/pages/organizations/documents/DocumentLayoutLoader")),
    children: [
      ...documentTabs(""),
      ...documentTabs("versions/:versionId/"),
    ],
  },
] satisfies AppRoute[];

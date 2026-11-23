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

import { PageSkeleton } from "#/components/skeletons/PageSkeleton";
import { detailLayout } from "#/pages/iam/organizations/_lib/shellLayout";

import { DocumentDetailsPageSkeleton } from "./DocumentDetailsPageSkeleton";

// Old document subpaths (tabs and versions/:versionId/...) are no longer
// generated, but bookmarks to them should land on the document page.
function redirectOldDocumentPath({ params }: LoaderFunctionArgs) {
  const { organizationId = "", documentId = "" } = params;
  const path = [
    "",
    "organizations",
    encodeURIComponent(organizationId),
    "governance",
    "documents",
    encodeURIComponent(documentId),
  ].join("/");
  const versionId = params["*"]?.match(/^versions\/([^/]+)/)?.[1];
  const search = versionId == null
    ? ""
    : `?${new URLSearchParams({ version: versionId }).toString()}`;
  // eslint-disable-next-line
  throw redirect(`${path}${search}`);
}

export const documentsRoutes = [
  {
    path: "documents",
    Fallback: PageSkeleton,
    Component: lazy(() => import("./DocumentsPageLoader")),
  },
  {
    path: "documents/:documentId",
    handle: detailLayout,
    Fallback: DocumentDetailsPageSkeleton,
    Component: lazy(() => import("./DocumentDetailsPageLoader")),
  },
  {
    path: "documents/:documentId/*",
    loader: redirectOldDocumentPath,
    Component: Fragment,
  },
] satisfies AppRoute[];

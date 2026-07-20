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
import {
  type AppRoute,
  loaderFromQueryLoader,
  routeFromAppRoute,
  withQueryRef,
} from "@probo/routes";
import { Fragment } from "react";
import { loadQuery } from "react-relay";
import { createBrowserRouter, redirect } from "react-router";

import { MainLayout } from "#/layouts/MainLayout";
import { DocumentsPage } from "#/pages/DocumentsPage";
import { OverviewPage } from "#/pages/OverviewPage";
import { SubprocessorsPage } from "#/pages/SubprocessorsPage";
import { currentTrustUpdatesQuery, UpdatesPage } from "#/pages/UpdatesPage";
import {
  currentTrustDocumentsQuery,
  currentCompliancePortalGraphQuery,
  currentTrustSubprocessorsQuery,
} from "#/queries/CompliancePortalGraph";

import { DocumentPageErrorBoundary } from "./components/DocumentPageErrorBoundary";
import { PageError } from "./components/PageError";
import { RootErrorBoundary } from "./components/RootErrorBoundary";
import { MainSkeleton } from "./components/Skeletons/MainSkeleton";
import { TabSkeleton } from "./components/Skeletons/TabSkeleton";
import { consoleEnvironment } from "./providers/RelayProviders";

const routes = [
  {
    Component: lazy(() => import("#/pages/auth/AuthLayoutLoader")),
    children: [
      {
        path: "/connect",
        Component: lazy(() => import("#/pages/auth/ConnectPageLoader")),
      },
      {
        path: "/full-name",
        Component: lazy(() => import("#/pages/auth/FullNamePage")),
      },
    ],
  },
  {
    path: "/",
    loader: () => {
      // eslint-disable-next-line
      throw redirect("/overview");
    },
    Component: Fragment,
    ErrorBoundary: RootErrorBoundary,
  },
  {
    path: "/nda",
    Component: lazy(() => import("#/pages/NDAPageLoader")),
    ErrorBoundary: RootErrorBoundary,
  },
  // Custom domain routes (subdomain-based)
  {
    path: "/overview",
    loader: loaderFromQueryLoader(() =>
      loadQuery(consoleEnvironment, currentCompliancePortalGraphQuery, {}),
    ),
    Component: withQueryRef(MainLayout),
    Fallback: MainSkeleton,
    ErrorBoundary: RootErrorBoundary,
    children: [
      {
        path: "",
        Fallback: TabSkeleton,
        Component: OverviewPage,
      },
    ],
  },
  {
    path: "/documents/:documentId",
    Component: lazy(() => import("#/pages/DocumentPageLoader")),
    ErrorBoundary: DocumentPageErrorBoundary,
  },
  {
    path: "/documents",
    loader: loaderFromQueryLoader(() =>
      loadQuery(consoleEnvironment, currentCompliancePortalGraphQuery, {}),
    ),
    Component: withQueryRef(MainLayout),
    Fallback: MainSkeleton,
    ErrorBoundary: RootErrorBoundary,
    children: [
      {
        path: "",
        loader: loaderFromQueryLoader(() =>
          loadQuery(consoleEnvironment, currentTrustDocumentsQuery, {}),
        ),
        Fallback: TabSkeleton,
        Component: withQueryRef(DocumentsPage),
      },
    ],
  },
  {
    path: "/subprocessors",
    loader: loaderFromQueryLoader(() =>
      loadQuery(consoleEnvironment, currentCompliancePortalGraphQuery, {}),
    ),
    Component: withQueryRef(MainLayout),
    Fallback: MainSkeleton,
    ErrorBoundary: RootErrorBoundary,
    children: [
      {
        path: "",
        loader: loaderFromQueryLoader(() =>
          loadQuery(consoleEnvironment, currentTrustSubprocessorsQuery, {}),
        ),
        Fallback: TabSkeleton,
        Component: withQueryRef(SubprocessorsPage),
      },
    ],
  },
  {
    path: "/updates",
    loader: loaderFromQueryLoader(() =>
      loadQuery(consoleEnvironment, currentCompliancePortalGraphQuery, {}),
    ),
    Component: withQueryRef(MainLayout),
    Fallback: MainSkeleton,
    ErrorBoundary: RootErrorBoundary,
    children: [
      {
        path: "",
        loader: loaderFromQueryLoader(() =>
          loadQuery(consoleEnvironment, currentTrustUpdatesQuery, {}),
        ),
        Fallback: TabSkeleton,
        Component: withQueryRef(UpdatesPage),
      },
    ],
  },
  // Fallback URL to the NotFound Page
  {
    path: "*",
    Component: PageError,
  },
] satisfies AppRoute[];

export const router = createBrowserRouter(routes.map(routeFromAppRoute), {});

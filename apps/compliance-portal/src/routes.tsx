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
import { type AppRoute, routeFromAppRoute } from "@probo/routes";
import { createBrowserRouter } from "react-router";

import { PageErrorBoundary } from "#/components/errors/PageErrorBoundary";
import { RootErrorBoundary } from "#/components/errors/RootErrorBoundary";
import { getPathPrefix } from "#/lib/http/pathPrefix";
import { documentRoutes } from "#/pages/documents/routes";
import { HomePageSkeleton } from "#/pages/HomePageSkeleton";
import { MainLayoutSkeleton } from "#/pages/MainLayoutSkeleton";
import { subprocessorRoutes } from "#/pages/subprocessors/routes";
import { updateRoutes } from "#/pages/updates/routes";

const routes = [
  {
    path: "/",
    Fallback: MainLayoutSkeleton,
    Component: lazy(() => import("#/pages/MainLayoutLoader")),
    // A layout failure takes down the shell, so it shows a standalone full page.
    ErrorBoundary: RootErrorBoundary,
    children: [
      {
        // Pathless layout route: page failures bubble here and render inside the
        // MainLayout Outlet, keeping the TopBar and footer chrome.
        ErrorBoundary: PageErrorBoundary,
        children: [
          {
            index: true,
            Fallback: HomePageSkeleton,
            Component: lazy(() => import("#/pages/HomePageLoader")),
          },
          ...documentRoutes,
          ...subprocessorRoutes,
          ...updateRoutes,
          {
            path: "requests",
            Component: lazy(() => import("#/pages/RequestsPage")),
          },
          {
            path: "*",
            Component: lazy(() => import("#/pages/NotFoundPage")),
          },
        ],
      },
    ],
  },
] satisfies AppRoute[];

// The portal is served under a /trust/{slug} path prefix (or a bare custom
// domain). Match the router basename to that prefix so the routes resolve.
export const router = createBrowserRouter(routes.map(routeFromAppRoute), {
  basename: getPathPrefix() || "/",
});

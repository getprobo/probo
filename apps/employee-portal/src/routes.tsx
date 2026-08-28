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
import { approvalsRoutes } from "#/pages/approvals/routes";
import { devicesRoutes } from "#/pages/devices/routes";
import { HomePageSkeleton } from "#/pages/HomePageSkeleton";
import { MainLayoutSkeleton } from "#/pages/iam/MainLayoutSkeleton";
import { OrganizationsPageSkeleton } from "#/pages/iam/OrganizationsPageSkeleton";
import { signaturesRoutes } from "#/pages/signatures/routes";

const routes = [
  {
    index: true,
    Fallback: OrganizationsPageSkeleton,
    Component: lazy(() => import("#/pages/iam/OrganizationsPageLoader")),
    ErrorBoundary: RootErrorBoundary,
  },
  {
    path: ":organizationId",
    Fallback: MainLayoutSkeleton,
    Component: lazy(() => import("#/pages/iam/MainLayoutLoader")),
    ErrorBoundary: RootErrorBoundary,
    children: [
      {
        ErrorBoundary: PageErrorBoundary,
        children: [
          {
            index: true,
            Fallback: HomePageSkeleton,
            Component: lazy(() => import("#/pages/HomePageLoader")),
          },
          ...signaturesRoutes,
          ...approvalsRoutes,
          ...devicesRoutes,
          {
            path: "*",
            Component: lazy(() => import("#/pages/NotFoundPage")),
          },
        ],
      },
    ],
  },
  {
    path: "*",
    Component: lazy(() => import("#/pages/NotFoundPage")),
    ErrorBoundary: RootErrorBoundary,
  },
] satisfies AppRoute[];

export const router = createBrowserRouter(routes.map(routeFromAppRoute), {
  basename: "/employee-portal",
});

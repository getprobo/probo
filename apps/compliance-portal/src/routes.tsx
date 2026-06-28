// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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
import { type AppRoute, routeFromAppRoute } from "@probo/routes";
import { createBrowserRouter } from "react-router";

import { getPathPrefix } from "#/lib/http/pathPrefix";
import { HomePageSkeleton } from "#/pages/HomePageSkeleton";
import { MainLayoutSkeleton } from "#/pages/MainLayoutSkeleton";

const routes = [
  {
    path: "/",
    Fallback: MainLayoutSkeleton,
    Component: lazy(() => import("#/pages/MainLayoutLoader")),
    children: [
      {
        index: true,
        Fallback: HomePageSkeleton,
        Component: lazy(() => import("#/pages/HomePageLoader")),
      },
      {
        path: "documents",
        Component: lazy(() => import("#/pages/DocumentsPage")),
      },
      {
        path: "subprocessors",
        Component: lazy(() => import("#/pages/SubprocessorsPage")),
      },
      {
        path: "updates",
        Component: lazy(() => import("#/pages/UpdatesPage")),
      },
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
] satisfies AppRoute[];

// The portal is served under a /trust/{slug} path prefix (or a bare custom
// domain). Match the router basename to that prefix so the routes resolve.
export const router = createBrowserRouter(routes.map(routeFromAppRoute), {
  basename: getPathPrefix() || "/",
});

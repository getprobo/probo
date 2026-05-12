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

import { type ComponentType, Suspense } from "react";
import { type RouteObject } from "react-router";

export type AppRoute = Omit<RouteObject, "children"> & {
  children?: AppRoute[];
  Fallback?: ComponentType;
}

export function routeFromAppRoute(appRoute: AppRoute): RouteObject {
  const { Component, Fallback, children, ...rest } = appRoute;
  let route = { ...rest } as RouteObject;

  if (Component && Fallback) {
    route = {
      ...route,
      Component: () => (
        <Suspense fallback={<Fallback />}>
          <Component />
        </Suspense>
      ),
    };
  } else if (Component) {
    route = {
      ...route,
      Component,
    };
  }

  return {
    ...route,
    children: children?.map(routeFromAppRoute),
  } as RouteObject;
}

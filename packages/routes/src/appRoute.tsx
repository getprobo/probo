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
    const OriginalComponent = Component as ComponentType;

    route = {
      ...route,
      Component: () => (
        <Suspense fallback={<Fallback />}>
          <OriginalComponent />
        </Suspense>
      ),
    };
  }

  return {
    ...route,
    children: children?.map(routeFromAppRoute),
  } as RouteObject;
}

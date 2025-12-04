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

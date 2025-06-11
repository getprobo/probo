import { loadQuery } from "react-relay";
import { PageSkeleton } from "/components/skeletons/PageSkeleton";
import { relayEnvironment } from "/providers/RelayProviders";
import {
  frameworksQuery,
  frameworkNodeQuery,
} from "/hooks/graph/FrameworkGraph";
import type { AppRoute } from "/routes";
import { Fragment } from "react";
import { lazy } from "@probo/react-lazy";

export const frameworkRoutes = [
  {
    path: "frameworks",
    fallback: PageSkeleton,
    queryLoader: ({ organizationId }) =>
      loadQuery(relayEnvironment, frameworksQuery, { organizationId }),
    Component: lazy(
      () => import("/pages/organizations/frameworks/FrameworksPage")
    ),
  },
  {
    path: "frameworks/:frameworkId",
    fallback: PageSkeleton,
    queryLoader: ({ frameworkId }) =>
      loadQuery(relayEnvironment, frameworkNodeQuery, { frameworkId }),
    Component: lazy(
      () => import("/pages/organizations/frameworks/FrameworkDetailPage")
    ),
    children: [
      {
        path: "",
        Component: Fragment,
      },
      {
        path: "controls/:controlId",
        Component: Fragment,
      },
    ],
  },
] satisfies AppRoute[];

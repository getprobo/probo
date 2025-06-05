import { loadQuery } from "react-relay";
import { PageSkeleton } from "/components/skeletons/PageSkeleton";
import { relayEnvironment } from "/providers/RelayProviders";
import {
  frameworksQuery,
  frameworkNodeQuery,
} from "/hooks/graph/FrameworkGraph";
import type { AppRoute } from "/routes";
import { lazy } from "react";

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
    path: "frameworks/:frameworkId/:controlId?",
    fallback: PageSkeleton,
    queryLoader: ({ frameworkId }) =>
      loadQuery(relayEnvironment, frameworkNodeQuery, { frameworkId }),
    Component: lazy(
      () => import("/pages/organizations/frameworks/FrameworkDetailPage")
    ),
  },
] satisfies AppRoute[];

import { loadQuery } from "react-relay";
import { PageSkeleton } from "/components/skeletons/PageSkeleton";
import { relayEnvironment } from "/providers/RelayProviders";
import { frameworksQuery } from "/hooks/graph/FrameworkGraph";
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
] satisfies AppRoute[];

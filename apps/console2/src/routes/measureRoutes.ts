import { lazy } from "react";
import { loadQuery } from "react-relay";
import { RisksPageSkeleton } from "/components/skeletons/RisksPageSkeleton.tsx";
import type { AppRoute } from "/routes.tsx";
import { relayEnvironment } from "/providers/RelayProviders";
import { measuresQuery } from "/hooks/graph/MeasureGraph";

export const measureRoutes = [
  {
    path: "measures",
    fallback: RisksPageSkeleton,
    queryLoader: ({ organizationId }) =>
      loadQuery(relayEnvironment, measuresQuery, { organizationId }),
    Component: lazy(() => import("/pages/organizations/measures/MeasuresPage")),
  },
] satisfies AppRoute[];

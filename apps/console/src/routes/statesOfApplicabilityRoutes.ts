import { lazy } from "@probo/react-lazy";
import type { AppRoute } from "@probo/routes";

import { PageSkeleton } from "#/components/skeletons/PageSkeleton";

export const statesOfApplicabilityRoutes = [
  {
    path: "states-of-applicability",
    Fallback: PageSkeleton,
    Component: lazy(
      () => import("#/pages/organizations/states-of-applicability/StatesOfApplicabilityPageLoader"),
    ),
  },
  {
    path: "snapshots/:snapshotId/states-of-applicability",
    Fallback: PageSkeleton,
    Component: lazy(
      () => import("#/pages/organizations/states-of-applicability/StatesOfApplicabilityPageLoader"),
    ),
  },
  {
    path: "states-of-applicability/:stateOfApplicabilityId",
    Fallback: PageSkeleton,
    Component: lazy(
      () => import("#/pages/organizations/states-of-applicability/StateOfApplicabilityDetailPageLoader"),
    ),
  },
  {
    path: "snapshots/:snapshotId/states-of-applicability/:stateOfApplicabilityId",
    Fallback: PageSkeleton,
    Component: lazy(
      () => import("#/pages/organizations/states-of-applicability/StateOfApplicabilityDetailPageLoader"),
    ),
  },
] satisfies AppRoute[];

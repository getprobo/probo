import { lazy } from "@probo/react-lazy";
import type { AppRoute } from "@probo/routes";

import { PageSkeleton } from "#/components/skeletons/PageSkeleton";

export const findingRoutes = [
  {
    path: "findings",
    Fallback: PageSkeleton,
    Component: lazy(
      () =>
        import("#/pages/organizations/findings/FindingsPageLoader"),
    ),
  },
  {
    path: "snapshots/:snapshotId/findings",
    Fallback: PageSkeleton,
    Component: lazy(
      () =>
        import("#/pages/organizations/findings/FindingsPageLoader"),
    ),
  },
  {
    path: "findings/:findingId",
    Fallback: PageSkeleton,
    Component: lazy(
      () =>
        import("#/pages/organizations/findings/FindingDetailsPageLoader"),
    ),
  },
  {
    path: "snapshots/:snapshotId/findings/:findingId",
    Fallback: PageSkeleton,
    Component: lazy(
      () =>
        import("#/pages/organizations/findings/FindingDetailsPageLoader"),
    ),
  },
] satisfies AppRoute[];

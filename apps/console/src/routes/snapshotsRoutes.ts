import { loadQuery } from "react-relay";
import { PageSkeleton } from "/components/skeletons/PageSkeleton";
import { relayEnvironment } from "/providers/RelayProviders";
import { snapshotsQuery, snapshotNodeQuery } from "/hooks/graph/SnapshotGraph";
import type { AppRoute } from "/routes";
import { lazy } from "@probo/react-lazy";

export const snapshotsRoutes = [
  {
    path: "snapshots",
    fallback: PageSkeleton,
    queryLoader: ({ organizationId }) =>
      loadQuery(relayEnvironment, snapshotsQuery, { organizationId }),
    Component: lazy(
      () => import("/pages/organizations/snapshots/SnapshotsPage")
    ),
  },
  {
    path: "snapshots/:snapshotId",
    fallback: PageSkeleton,
    queryLoader: ({ snapshotId }) =>
      loadQuery(relayEnvironment, snapshotNodeQuery, { snapshotId }),
    Component: lazy(
      () => import("/pages/organizations/snapshots/SnapshotDetailPage")
    ),
  },
] satisfies AppRoute[];

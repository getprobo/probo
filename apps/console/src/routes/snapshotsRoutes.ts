import { loadQuery } from "react-relay";
import { PageSkeleton } from "/components/skeletons/PageSkeleton";
import { relayEnvironment } from "/providers/RelayProviders";
import { snapshotsQuery } from "/hooks/graph/SnapshotGraph";
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
] satisfies AppRoute[];

import { loadQuery } from "react-relay";
import { PageSkeleton } from "/components/skeletons/PageSkeleton";
import { relayEnvironment } from "/providers/RelayProviders";
import { snapshotsQuery, snapshotNodeQuery } from "/hooks/graph/SnapshotGraph";
import { lazy } from "@probo/react-lazy";
import type { SnapshotGraphListQuery } from "/hooks/graph/__generated__/SnapshotGraphListQuery.graphql";
import type { SnapshotGraphNodeQuery } from "/hooks/graph/__generated__/SnapshotGraphNodeQuery.graphql";
import { loaderFromQueryLoader, withQueryRef, type AppRoute } from "/routes";

export const snapshotsRoutes = [
  {
    path: "snapshots",
    fallback: PageSkeleton,
    loader: loaderFromQueryLoader(({ organizationId }) =>
      loadQuery<SnapshotGraphListQuery>(relayEnvironment, snapshotsQuery, { organizationId }),
    ),
    Component: withQueryRef(lazy(
      () => import("/pages/organizations/snapshots/SnapshotsPage")
    )),
  },
  {
    path: "snapshots/:snapshotId",
    fallback: PageSkeleton,
    loader: loaderFromQueryLoader(({ snapshotId }) =>
      loadQuery<SnapshotGraphNodeQuery>(relayEnvironment, snapshotNodeQuery, { snapshotId }),
    ),
    Component: withQueryRef(lazy(
      () => import("/pages/organizations/snapshots/SnapshotDetailPage")
    )),
  },
] satisfies AppRoute[];

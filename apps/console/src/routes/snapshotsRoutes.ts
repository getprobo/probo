import { lazy } from "@probo/react-lazy";
import {
  type AppRoute,
  loaderFromQueryLoader,
  withQueryRef,
} from "@probo/routes";
import { loadQuery } from "react-relay";

import type { SnapshotGraphListQuery } from "#/__generated__/core/SnapshotGraphListQuery.graphql";
import type { SnapshotGraphNodeQuery } from "#/__generated__/core/SnapshotGraphNodeQuery.graphql";
import { PageSkeleton } from "#/components/skeletons/PageSkeleton";
import { coreEnvironment } from "#/environments";
import { snapshotNodeQuery, snapshotsQuery } from "#/hooks/graph/SnapshotGraph";

export const snapshotsRoutes = [
  {
    path: "snapshots",
    Fallback: PageSkeleton,
    loader: loaderFromQueryLoader(({ organizationId }) =>
      loadQuery<SnapshotGraphListQuery>(coreEnvironment, snapshotsQuery, {
        organizationId,
      }),
    ),
    Component: withQueryRef(
      lazy(() => import("#/pages/organizations/snapshots/SnapshotsPage")),
    ),
  },
  {
    path: "snapshots/:snapshotId",
    Fallback: PageSkeleton,
    loader: loaderFromQueryLoader(({ snapshotId }) =>
      loadQuery<SnapshotGraphNodeQuery>(coreEnvironment, snapshotNodeQuery, {
        snapshotId,
      }),
    ),
    Component: withQueryRef(
      lazy(() => import("#/pages/organizations/snapshots/SnapshotDetailPage")),
    ),
  },
] satisfies AppRoute[];

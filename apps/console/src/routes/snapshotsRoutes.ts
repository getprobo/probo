import { loadQuery } from "react-relay";
import { PageSkeleton } from "/components/skeletons/PageSkeleton";
import { coreEnvironment } from "/environments";
import { snapshotsQuery, snapshotNodeQuery } from "/hooks/graph/SnapshotGraph";
import { lazy } from "@probo/react-lazy";
import type { SnapshotGraphListQuery } from "/__generated__/core/SnapshotGraphListQuery.graphql";
import type { SnapshotGraphNodeQuery } from "/__generated__/core/SnapshotGraphNodeQuery.graphql";
import {
  loaderFromQueryLoader,
  withQueryRef,
  type AppRoute,
} from "@probo/routes";

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
      lazy(() => import("/pages/organizations/snapshots/SnapshotsPage")),
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
      lazy(() => import("/pages/organizations/snapshots/SnapshotDetailPage")),
    ),
  },
] satisfies AppRoute[];

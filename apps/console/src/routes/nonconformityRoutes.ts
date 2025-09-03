import { loadQuery } from "react-relay";
import { relayEnvironment } from "/providers/RelayProviders";
import { PageSkeleton } from "/components/skeletons/PageSkeleton";
import { lazy } from "@probo/react-lazy";
import { nonconformitiesQuery, nonconformityNodeQuery } from "../hooks/graph/NonconformityGraph";
import type { AppRoute } from "/routes";

export const nonconformityRoutes= [
  {
    path: "nonconformities",
    fallback: PageSkeleton,
    queryLoader: (params: Record<string, string>) =>
      loadQuery(relayEnvironment, nonconformitiesQuery, {
        organizationId: params.organizationId,
        snapshotId: null
      }),
    Component: lazy(
      () => import("../pages/organizations/nonconformities/NonconformitiesPage")
    ),
  },
  {
    path: "snapshots/:snapshotId/nonconformities",
    fallback: PageSkeleton,
    queryLoader: (params: Record<string, string>) =>
      loadQuery(relayEnvironment, nonconformitiesQuery, {
        organizationId: params.organizationId,
        snapshotId: params.snapshotId
      }),
    Component: lazy(
      () => import("../pages/organizations/nonconformities/NonconformitiesPage")
    ),
  },
  {
    path: "nonconformities/:nonconformityId",
    fallback: PageSkeleton,
    queryLoader: (params: Record<string, string>) =>
      loadQuery(relayEnvironment, nonconformityNodeQuery, {
        nonconformityId: params.nonconformityId
      }),
    Component: lazy(
      () => import("../pages/organizations/nonconformities/NonconformityDetailsPage")
    ),
  },
  {
    path: "snapshots/:snapshotId/nonconformities/:nonconformityId",
    fallback: PageSkeleton,
    queryLoader: (params: Record<string, string>) =>
      loadQuery(relayEnvironment, nonconformityNodeQuery, {
        nonconformityId: params.nonconformityId
      }),
    Component: lazy(
      () => import("../pages/organizations/nonconformities/NonconformityDetailsPage")
    ),
  },
] satisfies AppRoute[];

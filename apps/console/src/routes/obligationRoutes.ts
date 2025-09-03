import { loadQuery } from "react-relay";
import { relayEnvironment } from "/providers/RelayProviders";
import { PageSkeleton } from "/components/skeletons/PageSkeleton";
import { lazy } from "@probo/react-lazy";
import { obligationsQuery, obligationNodeQuery } from "/hooks/graph/ObligationGraph";
import type { AppRoute } from "/routes";

export const obligationRoutes = [
  {
    path: "obligations",
    fallback: PageSkeleton,
    queryLoader: (params: Record<string, string>) =>
      loadQuery(relayEnvironment, obligationsQuery, {
        organizationId: params.organizationId,
        snapshotId: null
      }),
    Component: lazy(
      () => import("/pages/organizations/obligations/ObligationsPage")
    ),
  },
  {
    path: "snapshots/:snapshotId/obligations",
    fallback: PageSkeleton,
    queryLoader: (params: Record<string, string>) =>
      loadQuery(relayEnvironment, obligationsQuery, {
        organizationId: params.organizationId,
        snapshotId: params.snapshotId
      }),
    Component: lazy(
      () => import("/pages/organizations/obligations/ObligationsPage")
    ),
  },
  {
    path: "obligations/:obligationId",
    fallback: PageSkeleton,
    queryLoader: (params: Record<string, string>) =>
      loadQuery(relayEnvironment, obligationNodeQuery, {
        obligationId: params.obligationId
      }),
    Component: lazy(
      () => import("/pages/organizations/obligations/ObligationDetailsPage")
    ),
  },
  {
    path: "snapshots/:snapshotId/obligations/:obligationId",
    fallback: PageSkeleton,
    queryLoader: (params: Record<string, string>) =>
      loadQuery(relayEnvironment, obligationNodeQuery, {
        obligationId: params.obligationId
      }),
    Component: lazy(
      () => import("/pages/organizations/obligations/ObligationDetailsPage")
    ),
  },
] satisfies AppRoute[];

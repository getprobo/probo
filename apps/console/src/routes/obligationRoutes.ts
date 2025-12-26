import { loadQuery } from "react-relay";
import { coreEnvironment } from "/environments";
import { PageSkeleton } from "/components/skeletons/PageSkeleton";
import { lazy } from "@probo/react-lazy";
import {
  obligationsQuery,
  obligationNodeQuery,
} from "/hooks/graph/ObligationGraph";
import type { ObligationGraphListQuery } from "/hooks/graph/__generated__/ObligationGraphListQuery.graphql";
import type { ObligationGraphNodeQuery } from "/hooks/graph/__generated__/ObligationGraphNodeQuery.graphql";
import {
  loaderFromQueryLoader,
  withQueryRef,
  type AppRoute,
} from "@probo/routes";

export const obligationRoutes = [
  {
    path: "obligations",
    Fallback: PageSkeleton,
    loader: loaderFromQueryLoader(({ organizationId }) =>
      loadQuery<ObligationGraphListQuery>(coreEnvironment, obligationsQuery, {
        organizationId,
        snapshotId: null,
      }),
    ),
    Component: withQueryRef(
      lazy(() => import("/pages/organizations/obligations/ObligationsPage")),
    ),
  },
  {
    path: "snapshots/:snapshotId/obligations",
    Fallback: PageSkeleton,
    loader: loaderFromQueryLoader(({ organizationId, snapshotId }) =>
      loadQuery<ObligationGraphListQuery>(coreEnvironment, obligationsQuery, {
        organizationId,
        snapshotId,
      }),
    ),
    Component: withQueryRef(
      lazy(() => import("/pages/organizations/obligations/ObligationsPage")),
    ),
  },
  {
    path: "obligations/:obligationId",
    Fallback: PageSkeleton,
    loader: loaderFromQueryLoader(({ obligationId }) =>
      loadQuery<ObligationGraphNodeQuery>(
        coreEnvironment,
        obligationNodeQuery,
        {
          obligationId,
        },
      ),
    ),
    Component: withQueryRef(
      lazy(
        () => import("/pages/organizations/obligations/ObligationDetailsPage"),
      ),
    ),
  },
  {
    path: "snapshots/:snapshotId/obligations/:obligationId",
    Fallback: PageSkeleton,
    loader: loaderFromQueryLoader(({ obligationId }) =>
      loadQuery<ObligationGraphNodeQuery>(
        coreEnvironment,
        obligationNodeQuery,
        {
          obligationId,
        },
      ),
    ),
    Component: withQueryRef(
      lazy(
        () => import("/pages/organizations/obligations/ObligationDetailsPage"),
      ),
    ),
  },
] satisfies AppRoute[];

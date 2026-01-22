import { loadQuery } from "react-relay";
import { lazy } from "@probo/react-lazy";
import {
  loaderFromQueryLoader,
  withQueryRef,
  type AppRoute,
} from "@probo/routes";

import { coreEnvironment } from "/environments";
import { PageSkeleton } from "/components/skeletons/PageSkeleton";
import {
  obligationsQuery,
  obligationNodeQuery,
} from "/hooks/graph/ObligationGraph";
import type { ObligationGraphListQuery } from "/__generated__/core/ObligationGraphListQuery.graphql";
import type { ObligationGraphNodeQuery } from "/__generated__/core/ObligationGraphNodeQuery.graphql";

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

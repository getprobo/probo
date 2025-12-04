import { loadQuery } from "react-relay";
import { relayEnvironment } from "/providers/RelayProviders";
import { PageSkeleton } from "/components/skeletons/PageSkeleton";
import { lazy } from "@probo/react-lazy";
import { obligationsQuery, obligationNodeQuery } from "/hooks/graph/ObligationGraph";
import type { ObligationGraphListQuery } from "/hooks/graph/__generated__/ObligationGraphListQuery.graphql";
import type { ObligationGraphNodeQuery } from "/hooks/graph/__generated__/ObligationGraphNodeQuery.graphql";
import { loaderFromQueryLoader, withQueryRef } from "/routes";

export const obligationRoutes = [
  {
    path: "obligations",
    fallback: PageSkeleton,
    loader: loaderFromQueryLoader(({ organizationId }) =>
      loadQuery<ObligationGraphListQuery>(relayEnvironment, obligationsQuery, {
        organizationId,
        snapshotId: null
      }),
    ),
    Component: withQueryRef(lazy(
      () => import("/pages/organizations/obligations/ObligationsPage")
    )),
  },
  {
    path: "snapshots/:snapshotId/obligations",
    fallback: PageSkeleton,
    loader: loaderFromQueryLoader(({ organizationId, snapshotId }) =>
      loadQuery<ObligationGraphListQuery>(relayEnvironment, obligationsQuery, {
        organizationId,
        snapshotId,
      }),
    ),
    Component: withQueryRef(lazy(
      () => import("/pages/organizations/obligations/ObligationsPage")
    )),
  },
  {
    path: "obligations/:obligationId",
    fallback: PageSkeleton,
    loader: loaderFromQueryLoader(({ obligationId }) =>
      loadQuery<ObligationGraphNodeQuery>(relayEnvironment, obligationNodeQuery, {
        obligationId,
      }),
    ),
    Component: withQueryRef(lazy(
      () => import("/pages/organizations/obligations/ObligationDetailsPage")
    )),
  },
  {
    path: "snapshots/:snapshotId/obligations/:obligationId",
    fallback: PageSkeleton,
    loader: loaderFromQueryLoader(({ obligationId }) =>
      loadQuery<ObligationGraphNodeQuery>(relayEnvironment, obligationNodeQuery, {
        obligationId,
      }),
    ),
    Component: withQueryRef(lazy(
      () => import("/pages/organizations/obligations/ObligationDetailsPage")
    )),
  },
];

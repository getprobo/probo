import { loadQuery } from "react-relay";
import { relayEnvironment } from "/providers/RelayProviders";
import { PageSkeleton } from "/components/skeletons/PageSkeleton";
import { lazy } from "@probo/react-lazy";
import { nonconformitiesQuery, nonconformityNodeQuery } from "../hooks/graph/NonconformityGraph";
import type { NonconformityGraphListQuery } from "/hooks/graph/__generated__/NonconformityGraphListQuery.graphql";
import type { NonconformityGraphNodeQuery } from "/hooks/graph/__generated__/NonconformityGraphNodeQuery.graphql";
import { loaderFromQueryLoader, withQueryRef, type AppRoute } from "/routes";

export const nonconformityRoutes = [
  {
    path: "nonconformities",
    fallback: PageSkeleton,
    loader: loaderFromQueryLoader(({ organizationId }) =>
      loadQuery<NonconformityGraphListQuery>(relayEnvironment, nonconformitiesQuery, {
        organizationId,
        snapshotId: null
      }),
    ),
    Component: withQueryRef(lazy(
      () => import("../pages/organizations/nonconformities/NonconformitiesPage")
    )),
  },
  {
    path: "snapshots/:snapshotId/nonconformities",
    fallback: PageSkeleton,
    loader: loaderFromQueryLoader(({ organizationId, snapshotId }) =>
      loadQuery<NonconformityGraphListQuery>(relayEnvironment, nonconformitiesQuery, {
        organizationId,
        snapshotId,
      }),
    ),
    Component: withQueryRef(lazy(
      () => import("../pages/organizations/nonconformities/NonconformitiesPage")
    )),
  },
  {
    path: "nonconformities/:nonconformityId",
    fallback: PageSkeleton,
    loader: loaderFromQueryLoader(({ nonconformityId }) =>
      loadQuery<NonconformityGraphNodeQuery>(relayEnvironment, nonconformityNodeQuery, {
        nonconformityId,
      }),
    ),
    Component: withQueryRef(lazy(
      () => import("../pages/organizations/nonconformities/NonconformityDetailsPage")
    )),
  },
  {
    path: "snapshots/:snapshotId/nonconformities/:nonconformityId",
    fallback: PageSkeleton,
    loader: loaderFromQueryLoader(({ nonconformityId }) =>
      loadQuery<NonconformityGraphNodeQuery>(relayEnvironment, nonconformityNodeQuery, {
        nonconformityId: nonconformityId
      }),
    ),
    Component: withQueryRef(lazy(
      () => import("../pages/organizations/nonconformities/NonconformityDetailsPage")
    )),
  },
] satisfies AppRoute[];

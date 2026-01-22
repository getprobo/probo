import { lazy } from "@probo/react-lazy";
import {
  type AppRoute,
  loaderFromQueryLoader,
  withQueryRef,
} from "@probo/routes";
import { loadQuery } from "react-relay";

import type { NonconformityGraphListQuery } from "#/__generated__/core/NonconformityGraphListQuery.graphql";
import type { NonconformityGraphNodeQuery } from "#/__generated__/core/NonconformityGraphNodeQuery.graphql";
import { PageSkeleton } from "#/components/skeletons/PageSkeleton";
import { coreEnvironment } from "#/environments";

import {
  nonconformitiesQuery,
  nonconformityNodeQuery,
} from "../hooks/graph/NonconformityGraph";

export const nonconformityRoutes = [
  {
    path: "nonconformities",
    Fallback: PageSkeleton,
    loader: loaderFromQueryLoader(({ organizationId }) =>
      loadQuery<NonconformityGraphListQuery>(
        coreEnvironment,
        nonconformitiesQuery,
        {
          organizationId,
          snapshotId: null,
        },
      ),
    ),
    Component: withQueryRef(
      lazy(
        () =>
          import("../pages/organizations/nonconformities/NonconformitiesPage"),
      ),
    ),
  },
  {
    path: "snapshots/:snapshotId/nonconformities",
    Fallback: PageSkeleton,
    loader: loaderFromQueryLoader(({ organizationId, snapshotId }) =>
      loadQuery<NonconformityGraphListQuery>(
        coreEnvironment,
        nonconformitiesQuery,
        {
          organizationId,
          snapshotId,
        },
      ),
    ),
    Component: withQueryRef(
      lazy(
        () =>
          import("../pages/organizations/nonconformities/NonconformitiesPage"),
      ),
    ),
  },
  {
    path: "nonconformities/:nonconformityId",
    Fallback: PageSkeleton,
    loader: loaderFromQueryLoader(({ nonconformityId }) =>
      loadQuery<NonconformityGraphNodeQuery>(
        coreEnvironment,
        nonconformityNodeQuery,
        {
          nonconformityId,
        },
      ),
    ),
    Component: withQueryRef(
      lazy(
        () =>
          import("../pages/organizations/nonconformities/NonconformityDetailsPage"),
      ),
    ),
  },
  {
    path: "snapshots/:snapshotId/nonconformities/:nonconformityId",
    Fallback: PageSkeleton,
    loader: loaderFromQueryLoader(({ nonconformityId }) =>
      loadQuery<NonconformityGraphNodeQuery>(
        coreEnvironment,
        nonconformityNodeQuery,
        {
          nonconformityId: nonconformityId,
        },
      ),
    ),
    Component: withQueryRef(
      lazy(
        () =>
          import("../pages/organizations/nonconformities/NonconformityDetailsPage"),
      ),
    ),
  },
] satisfies AppRoute[];

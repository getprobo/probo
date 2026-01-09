import { loadQuery } from "react-relay";
import { coreEnvironment } from "/environments";
import { PageSkeleton } from "/components/skeletons/PageSkeleton";
import { lazy } from "@probo/react-lazy";
import {
  nonconformitiesQuery,
  nonconformityNodeQuery,
} from "../hooks/graph/NonconformityGraph";
import type { NonconformityGraphListQuery } from "/__generated__/core/NonconformityGraphListQuery.graphql";
import type { NonconformityGraphNodeQuery } from "/__generated__/core/NonconformityGraphNodeQuery.graphql";
import {
  loaderFromQueryLoader,
  withQueryRef,
  type AppRoute,
} from "@probo/routes";

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

import { lazy } from "@probo/react-lazy";
import {
  type AppRoute,
  loaderFromQueryLoader,
  withQueryRef,
} from "@probo/routes";
import { loadQuery } from "react-relay";

import type { DatumGraphListQuery } from "#/__generated__/core/DatumGraphListQuery.graphql";
import type { DatumGraphNodeQuery } from "#/__generated__/core/DatumGraphNodeQuery.graphql";
import { PageSkeleton } from "#/components/skeletons/PageSkeleton";
import { coreEnvironment } from "#/environments";

import { dataQuery, datumNodeQuery } from "../hooks/graph/DatumGraph";

export const dataRoutes = [
  {
    path: "data",
    Fallback: PageSkeleton,
    loader: loaderFromQueryLoader(({ organizationId }) =>
      loadQuery<DatumGraphListQuery>(coreEnvironment, dataQuery, {
        organizationId: organizationId,
        snapshotId: null,
      }),
    ),
    Component: withQueryRef(
      lazy(() => import("#/pages/organizations/data/DataPage")),
    ),
  },
  {
    path: "snapshots/:snapshotId/data",
    Fallback: PageSkeleton,
    loader: loaderFromQueryLoader(({ organizationId, snapshotId }) =>
      loadQuery<DatumGraphListQuery>(coreEnvironment, dataQuery, {
        organizationId,
        snapshotId,
      }),
    ),
    Component: withQueryRef(
      lazy(() => import("#/pages/organizations/data/DataPage")),
    ),
  },
  {
    path: "data/:dataId",
    Fallback: PageSkeleton,
    loader: loaderFromQueryLoader(({ dataId }) =>
      loadQuery<DatumGraphNodeQuery>(coreEnvironment, datumNodeQuery, {
        dataId,
      }),
    ),
    Component: withQueryRef(
      lazy(() => import("../pages/organizations/data/DatumDetailsPage")),
    ),
  },
  {
    path: "snapshots/:snapshotId/data/:dataId",
    Fallback: PageSkeleton,
    loader: loaderFromQueryLoader(({ dataId }) =>
      loadQuery<DatumGraphNodeQuery>(coreEnvironment, datumNodeQuery, {
        dataId,
      }),
    ),
    Component: withQueryRef(
      lazy(() => import("../pages/organizations/data/DatumDetailsPage")),
    ),
  },
] satisfies AppRoute[];

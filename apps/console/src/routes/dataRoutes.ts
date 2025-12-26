import { loadQuery } from "react-relay";
import { coreEnvironment } from "/environments";
import { PageSkeleton } from "/components/skeletons/PageSkeleton";
import { lazy } from "@probo/react-lazy";
import { dataQuery, datumNodeQuery } from "../hooks/graph/DatumGraph";
import type { DatumGraphListQuery } from "/hooks/graph/__generated__/DatumGraphListQuery.graphql";
import type { DatumGraphNodeQuery } from "/hooks/graph/__generated__/DatumGraphNodeQuery.graphql";
import {
  loaderFromQueryLoader,
  withQueryRef,
  type AppRoute,
} from "@probo/routes";

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
      lazy(() => import("/pages/organizations/data/DataPage")),
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
      lazy(() => import("/pages/organizations/data/DataPage")),
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

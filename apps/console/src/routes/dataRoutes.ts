import { loadQuery } from "react-relay";
import { relayEnvironment } from "/providers/RelayProviders";
import { PageSkeleton } from "/components/skeletons/PageSkeleton";
import { lazy } from "@probo/react-lazy";
import { dataQuery, datumNodeQuery } from "../hooks/graph/DatumGraph";
import type { DatumGraphListQuery } from "/hooks/graph/__generated__/DatumGraphListQuery.graphql";
import type { DatumGraphNodeQuery } from "/hooks/graph/__generated__/DatumGraphNodeQuery.graphql";
import { loaderFromQueryLoader, withQueryRef, type AppRoute } from "/routes";

export const dataRoutes = [
  {
    path: "data",
    fallback: PageSkeleton,
    loader: loaderFromQueryLoader(({ organizationId }) =>
      loadQuery<DatumGraphListQuery>(relayEnvironment, dataQuery, {
        organizationId: organizationId,
        snapshotId: null
      }),
    ),
    Component: withQueryRef(lazy(
      () => import("/pages/organizations/data/DataPage")
    )),
  },
  {
    path: "snapshots/:snapshotId/data",
    fallback: PageSkeleton,
    loader: loaderFromQueryLoader(({ organizationId, snapshotId }) =>
      loadQuery<DatumGraphListQuery>(relayEnvironment, dataQuery, {
        organizationId,
        snapshotId,
      }),
    ),
    Component: withQueryRef(lazy(
      () => import("/pages/organizations/data/DataPage")
    )),
  },
  {
    path: "data/:dataId",
    fallback: PageSkeleton,
    loader: loaderFromQueryLoader(({ dataId }) =>
      loadQuery<DatumGraphNodeQuery>(relayEnvironment, datumNodeQuery, { dataId }),
    ),
    Component: withQueryRef(lazy(
      () => import("../pages/organizations/data/DatumDetailsPage")
    )),
  },
  {
    path: "snapshots/:snapshotId/data/:dataId",
    fallback: PageSkeleton,
    loader: loaderFromQueryLoader(({ dataId }) =>
      loadQuery<DatumGraphNodeQuery>(relayEnvironment, datumNodeQuery, { dataId }),
    ),
    Component: withQueryRef(lazy(
      () => import("../pages/organizations/data/DatumDetailsPage")
    )),
  },
] satisfies AppRoute[];

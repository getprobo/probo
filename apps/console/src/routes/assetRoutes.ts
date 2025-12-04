import { loadQuery } from "react-relay";
import { relayEnvironment } from "/providers/RelayProviders";
import { PageSkeleton } from "/components/skeletons/PageSkeleton";
import { lazy } from "@probo/react-lazy";
import { assetsQuery, assetNodeQuery } from "../hooks/graph/AssetGraph";
import type { AssetGraphListQuery } from "/hooks/graph/__generated__/AssetGraphListQuery.graphql";
import type { AssetGraphNodeQuery } from "/hooks/graph/__generated__/AssetGraphNodeQuery.graphql";
import { loaderFromQueryLoader, withQueryRef, type AppRoute } from "@probo/routes";

export const assetRoutes = [
  {
    path: "assets",
    Fallback: PageSkeleton,
    loader: loaderFromQueryLoader(({ organizationId }) =>
      loadQuery<AssetGraphListQuery>(relayEnvironment, assetsQuery, {
        organizationId: organizationId,
        snapshotId: null,
      }),
    ),
    Component: withQueryRef(lazy(() => import("/pages/organizations/assets/AssetsPage"))),
  },
  {
    path: "snapshots/:snapshotId/assets",
    Fallback: PageSkeleton,
    loader: loaderFromQueryLoader(({ organizationId, snapshotId }) =>
      loadQuery<AssetGraphListQuery>(relayEnvironment, assetsQuery, {
        organizationId: organizationId,
        snapshotId: snapshotId,
      }),
    ),
    Component: withQueryRef(lazy(() => import("/pages/organizations/assets/AssetsPage"))),
  },
  {
    path: "assets/:assetId",
    Fallback: PageSkeleton,
    loader: loaderFromQueryLoader(({ assetId }) =>
      loadQuery<AssetGraphNodeQuery>(relayEnvironment, assetNodeQuery, { assetId }),
    ),
    Component: withQueryRef(lazy(
      () => import("/pages/organizations/assets/AssetDetailsPage"),
    )),
  },
  {
    path: "snapshots/:snapshotId/assets/:assetId",
    Fallback: PageSkeleton,
    loader: loaderFromQueryLoader(({ assetId }) =>
      loadQuery<AssetGraphNodeQuery>(relayEnvironment, assetNodeQuery, { assetId }),
    ),
    Component: withQueryRef(lazy(
      () => import("/pages/organizations/assets/AssetDetailsPage"),
    )),
  },
] satisfies AppRoute[];

import { lazy } from "@probo/react-lazy";
import {
  type AppRoute,
  loaderFromQueryLoader,
  withQueryRef,
} from "@probo/routes";
import { loadQuery } from "react-relay";

import type { AssetGraphListQuery } from "#/__generated__/core/AssetGraphListQuery.graphql";
import type { AssetGraphNodeQuery } from "#/__generated__/core/AssetGraphNodeQuery.graphql";
import { PageSkeleton } from "#/components/skeletons/PageSkeleton";
import { coreEnvironment } from "#/environments";

import { assetNodeQuery, assetsQuery } from "../hooks/graph/AssetGraph";

export const assetRoutes = [
  {
    path: "assets",
    Fallback: PageSkeleton,
    loader: loaderFromQueryLoader(({ organizationId }) =>
      loadQuery<AssetGraphListQuery>(coreEnvironment, assetsQuery, {
        organizationId: organizationId,
        snapshotId: null,
      }),
    ),
    Component: withQueryRef(
      lazy(() => import("#/pages/organizations/assets/AssetsPage")),
    ),
  },
  {
    path: "snapshots/:snapshotId/assets",
    Fallback: PageSkeleton,
    loader: loaderFromQueryLoader(({ organizationId, snapshotId }) =>
      loadQuery<AssetGraphListQuery>(coreEnvironment, assetsQuery, {
        organizationId: organizationId,
        snapshotId: snapshotId,
      }),
    ),
    Component: withQueryRef(
      lazy(() => import("#/pages/organizations/assets/AssetsPage")),
    ),
  },
  {
    path: "assets/:assetId",
    Fallback: PageSkeleton,
    loader: loaderFromQueryLoader(({ assetId }) =>
      loadQuery<AssetGraphNodeQuery>(coreEnvironment, assetNodeQuery, {
        assetId,
      }),
    ),
    Component: withQueryRef(
      lazy(() => import("#/pages/organizations/assets/AssetDetailsPage")),
    ),
  },
  {
    path: "snapshots/:snapshotId/assets/:assetId",
    Fallback: PageSkeleton,
    loader: loaderFromQueryLoader(({ assetId }) =>
      loadQuery<AssetGraphNodeQuery>(coreEnvironment, assetNodeQuery, {
        assetId,
      }),
    ),
    Component: withQueryRef(
      lazy(() => import("#/pages/organizations/assets/AssetDetailsPage")),
    ),
  },
] satisfies AppRoute[];

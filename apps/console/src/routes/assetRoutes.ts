// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

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

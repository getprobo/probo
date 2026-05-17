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

import type { DeviceGraphEnrollmentTokensQuery } from "#/__generated__/core/DeviceGraphEnrollmentTokensQuery.graphql";
import type { DeviceGraphListQuery } from "#/__generated__/core/DeviceGraphListQuery.graphql";
import type { DeviceGraphNodeQuery } from "#/__generated__/core/DeviceGraphNodeQuery.graphql";
import { PageSkeleton } from "#/components/skeletons/PageSkeleton";
import { coreEnvironment } from "#/environments";
import {
  deviceEnrollmentTokensQuery,
  deviceNodeQuery,
  devicesQuery,
} from "#/hooks/graph/DeviceGraph";

export const deviceRoutes = [
  {
    path: "devices",
    Fallback: PageSkeleton,
    loader: loaderFromQueryLoader(({ organizationId }) =>
      loadQuery<DeviceGraphListQuery>(coreEnvironment, devicesQuery, {
        organizationId: organizationId,
      }),
    ),
    Component: withQueryRef(
      lazy(() => import("#/pages/organizations/devices/DevicesPage")),
    ),
  },
  {
    path: "devices/enroll",
    Fallback: PageSkeleton,
    loader: loaderFromQueryLoader(({ organizationId }) =>
      loadQuery<DeviceGraphEnrollmentTokensQuery>(
        coreEnvironment,
        deviceEnrollmentTokensQuery,
        { organizationId: organizationId },
      ),
    ),
    Component: withQueryRef(
      lazy(() => import("#/pages/organizations/devices/DeviceEnrollPage")),
    ),
  },
  {
    path: "devices/:deviceId",
    Fallback: PageSkeleton,
    loader: loaderFromQueryLoader(({ deviceId }) =>
      loadQuery<DeviceGraphNodeQuery>(coreEnvironment, deviceNodeQuery, {
        deviceId: deviceId,
      }),
    ),
    Component: withQueryRef(
      lazy(() => import("#/pages/organizations/devices/DeviceDetailPage")),
    ),
  },
] satisfies AppRoute[];

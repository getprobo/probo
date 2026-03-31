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

import type { RightsRequestGraphListQuery } from "#/__generated__/core/RightsRequestGraphListQuery.graphql";
import type { RightsRequestGraphNodeQuery } from "#/__generated__/core/RightsRequestGraphNodeQuery.graphql";
import { PageSkeleton } from "#/components/skeletons/PageSkeleton";
import { coreEnvironment } from "#/environments";
import {
  rightsRequestNodeQuery,
  rightsRequestsQuery,
} from "#/hooks/graph/RightsRequestGraph";

export const rightsRequestRoutes = [
  {
    path: "rights-requests",
    Fallback: PageSkeleton,
    loader: loaderFromQueryLoader(({ organizationId }) =>
      loadQuery<RightsRequestGraphListQuery>(
        coreEnvironment,
        rightsRequestsQuery,
        {
          organizationId,
        },
      ),
    ),
    Component: withQueryRef(
      lazy(
        () => import("#/pages/organizations/rightsRequests/RightsRequestsPage"),
      ),
    ),
  },
  {
    path: "rights-requests/:requestId",
    Fallback: PageSkeleton,
    loader: loaderFromQueryLoader(({ requestId }) =>
      loadQuery<RightsRequestGraphNodeQuery>(
        coreEnvironment,
        rightsRequestNodeQuery,
        {
          rightsRequestId: requestId,
        },
      ),
    ),
    Component: withQueryRef(
      lazy(
        () =>
          import("#/pages/organizations/rightsRequests/RightsRequestDetailsPage"),
      ),
    ),
  },
] satisfies AppRoute[];

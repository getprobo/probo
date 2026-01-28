import { lazy } from "@probo/react-lazy";
import {
  type AppRoute,
  loaderFromQueryLoader,
  withQueryRef,
} from "@probo/routes";
import { loadQuery } from "react-relay";

import type { TrustCenterGraphQuery } from "#/__generated__/core/TrustCenterGraphQuery.graphql";
import { PageSkeleton } from "#/components/skeletons/PageSkeleton";
import { coreEnvironment } from "#/environments";

import { trustCenterQuery } from "../hooks/graph/TrustCenterGraph";

export const trustCenterRoutes = [
  {
    path: "trust-center",
    Fallback: PageSkeleton,
    loader: loaderFromQueryLoader(({ organizationId }) =>
      loadQuery<TrustCenterGraphQuery>(
        coreEnvironment,
        trustCenterQuery,
        { organizationId },
        { fetchPolicy: "network-only" },
      ),
    ),
    Component: withQueryRef(
      lazy(() => import("#/pages/organizations/trustCenter/TrustCenterPage")),
    ),
  },
] satisfies AppRoute[];

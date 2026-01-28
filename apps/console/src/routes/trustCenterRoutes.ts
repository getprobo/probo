import { lazy } from "@probo/react-lazy";
import {
  type AppRoute,
  loaderFromQueryLoader,
  withQueryRef,
} from "@probo/routes";
import { loadQuery } from "react-relay";

import type { TrustCenterGraphQuery } from "#/__generated__/core/TrustCenterGraphQuery.graphql";
import { LinkCardSkeleton } from "#/components/skeletons/LinkCardSkeleton";
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
    children: [
      {
        path: "audits",
        Fallback: LinkCardSkeleton,
        Component: lazy(
          () => import("#/pages/organizations/trustCenter/TrustCenterAuditsTab"),
        ),
      },
      {
        path: "documents",
        Fallback: LinkCardSkeleton,
        Component: lazy(
          () =>
            import("#/pages/organizations/trustCenter/TrustCenterDocumentsTab"),
        ),
      },
      {
        path: "files",
        Fallback: LinkCardSkeleton,
        Component: lazy(
          () => import("#/pages/organizations/trustCenter/TrustCenterFilesTab"),
        ),
      },
      {
        path: "vendors",
        Fallback: LinkCardSkeleton,
        Component: lazy(
          () =>
            import("#/pages/organizations/trustCenter/TrustCenterVendorsTab"),
        ),
      },
      {
        path: "access",
        Fallback: LinkCardSkeleton,
        Component: lazy(
          () =>
            import("#/pages/organizations/trustCenter/TrustCenterAccessTab/TrustCenterAccessTab"),
        ),
      },
    ],
  },
] satisfies AppRoute[];

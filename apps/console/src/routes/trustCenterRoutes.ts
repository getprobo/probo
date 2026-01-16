import { loadQuery } from "react-relay";
import { coreEnvironment } from "/environments";
import { PageSkeleton } from "/components/skeletons/PageSkeleton";
import { LinkCardSkeleton } from "/components/skeletons/LinkCardSkeleton";
import { lazy } from "@probo/react-lazy";
import { trustCenterQuery } from "../hooks/graph/TrustCenterGraph";
import type { TrustCenterGraphQuery } from "/__generated__/core/TrustCenterGraphQuery.graphql";
import {
  loaderFromQueryLoader,
  withQueryRef,
  type AppRoute,
} from "@probo/routes";

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
      lazy(() => import("/pages/organizations/trustCenter/TrustCenterPage")),
    ),
    children: [
      {
        index: true,
        Fallback: LinkCardSkeleton,
        Component: lazy(
          () =>
            import("/pages/organizations/trustCenter/TrustCenterOverviewTab"),
        ),
      },
      {
        path: "trust-by",
        Fallback: LinkCardSkeleton,
        Component: lazy(
          () =>
            import("/pages/organizations/trustCenter/TrustCenterReferencesTab"),
        ),
      },
      {
        path: "audits",
        Fallback: LinkCardSkeleton,
        Component: lazy(
          () => import("/pages/organizations/trustCenter/TrustCenterAuditsTab"),
        ),
      },
      {
        path: "documents",
        Fallback: LinkCardSkeleton,
        Component: lazy(
          () =>
            import("/pages/organizations/trustCenter/TrustCenterDocumentsTab"),
        ),
      },
      {
        path: "files",
        Fallback: LinkCardSkeleton,
        Component: lazy(
          () => import("/pages/organizations/trustCenter/TrustCenterFilesTab"),
        ),
      },
      {
        path: "vendors",
        Fallback: LinkCardSkeleton,
        Component: lazy(
          () =>
            import("/pages/organizations/trustCenter/TrustCenterVendorsTab"),
        ),
      },
      {
        path: "access",
        Fallback: LinkCardSkeleton,
        Component: lazy(
          () =>
            import("/pages/organizations/trustCenter/TrustCenterAccessTab/TrustCenterAccessTab"),
        ),
      },
    ],
  },
] satisfies AppRoute[];

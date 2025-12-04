import { loadQuery } from "react-relay";
import { relayEnvironment } from "/providers/RelayProviders";
import { PageSkeleton } from "/components/skeletons/PageSkeleton";
import { LinkCardSkeleton } from "/components/skeletons/LinkCardSkeleton";
import { lazy } from "@probo/react-lazy";
import { trustCenterQuery } from "../hooks/graph/TrustCenterGraph";
import type { TrustCenterGraphQuery } from "/hooks/graph/__generated__/TrustCenterGraphQuery.graphql";
import { loaderFromQueryLoader, withQueryRef, type AppRoute } from "/routes";

export const trustCenterRoutes = [
  {
    path: "trust-center",
    fallback: PageSkeleton,
    loader: loaderFromQueryLoader(({ organizationId }) =>
      loadQuery<TrustCenterGraphQuery>(relayEnvironment, trustCenterQuery, { organizationId }, { fetchPolicy: "network-only" }),
    ),
    Component: withQueryRef(lazy(
      () => import("/pages/organizations/trustCenter/TrustCenterPage")
    )),
    children: [
      {
        index: true,
        fallback: LinkCardSkeleton,
        Component: lazy(
          () => import("/pages/organizations/trustCenter/TrustCenterOverviewTab")
        ),
      },
      {
        path: "trust-by",
        fallback: LinkCardSkeleton,
        Component: lazy(
          () => import("/pages/organizations/trustCenter/TrustCenterReferencesTab")
        ),
      },
      {
        path: "audits",
        fallback: LinkCardSkeleton,
        Component: lazy(
          () => import("/pages/organizations/trustCenter/TrustCenterAuditsTab")
        ),
      },
      {
        path: "documents",
        fallback: LinkCardSkeleton,
        Component: lazy(
          () => import("/pages/organizations/trustCenter/TrustCenterDocumentsTab")
        ),
      },
      {
        path: "files",
        fallback: LinkCardSkeleton,
        Component: lazy(
          () => import("/pages/organizations/trustCenter/TrustCenterFilesTab")
        ),
      },
      {
        path: "vendors",
        fallback: LinkCardSkeleton,
        Component: lazy(
          () => import("/pages/organizations/trustCenter/TrustCenterVendorsTab")
        ),
      },
      {
        path: "access",
        fallback: LinkCardSkeleton,
        Component: lazy(
          () => import("/pages/organizations/trustCenter/TrustCenterAccessTab")
        ),
      },
    ],
  },
] satisfies AppRoute[];

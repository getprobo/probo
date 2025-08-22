import { loadQuery } from "react-relay";
import { relayEnvironment } from "/providers/RelayProviders";
import { PageSkeleton } from "/components/skeletons/PageSkeleton";
import { LinkCardSkeleton } from "/components/skeletons/LinkCardSkeleton";
import { lazy } from "@probo/react-lazy";
import { trustCenterQuery } from "../hooks/graph/TrustCenterGraph";
import type { AppRoute } from "/routes";

export const trustCenterRoutes = [
  {
    path: "trust-center",
    fallback: PageSkeleton,
    queryLoader: ({ organizationId }) =>
      loadQuery(relayEnvironment, trustCenterQuery, { organizationId }),
    Component: lazy(
      () => import("/pages/organizations/trustCenter/TrustCenterPage")
    ),
    children: [
      {
        index: true,
        fallback: LinkCardSkeleton,
        Component: lazy(
          () => import("/pages/organizations/trustCenter/TrustCenterAuditsTab")
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
        path: "vendors",
        fallback: LinkCardSkeleton,
        Component: lazy(
          () => import("/pages/organizations/trustCenter/TrustCenterVendorsTab")
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
        path: "access",
        fallback: LinkCardSkeleton,
        Component: lazy(
          () => import("/pages/organizations/trustCenter/TrustCenterAccessTab")
        ),
      },
    ],
  },
] satisfies AppRoute[];

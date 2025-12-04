import { lazy } from "@probo/react-lazy";
import { loadQuery } from "react-relay";
import { relayEnvironment } from "/providers/RelayProviders";
import { PageSkeleton } from "/components/skeletons/PageSkeleton";
import { vendorNodeQuery, vendorsQuery } from "/hooks/graph/VendorGraph";
import { LinkCardSkeleton } from "/components/skeletons/LinkCardSkeleton";
import { loaderFromQueryLoader, withQueryRef, type AppRoute } from "/routes";
import type { VendorGraphListQuery } from "/hooks/graph/__generated__/VendorGraphListQuery.graphql";
import type { VendorGraphNodeQuery } from "/hooks/graph/__generated__/VendorGraphNodeQuery.graphql";

export const vendorRoutes = [
  {
    path: "vendors",
    fallback: PageSkeleton,
    loader: loaderFromQueryLoader(({ organizationId }) =>
      loadQuery<VendorGraphListQuery>(relayEnvironment, vendorsQuery, {
        organizationId: organizationId!,
        snapshotId: null
      }),
    ),
    Component: withQueryRef(lazy(() => import("/pages/organizations/vendors/VendorsPage"))),
  },
  {
    path: "snapshots/:snapshotId/vendors",
    fallback: PageSkeleton,
    loader: loaderFromQueryLoader(({ organizationId, snapshotId }) =>
      loadQuery<VendorGraphListQuery>(relayEnvironment, vendorsQuery, {
        organizationId: organizationId!,
        snapshotId
      }),
    ),
    Component: withQueryRef(lazy(() => import("/pages/organizations/vendors/VendorsPage"))),
  },
  {
    path: "vendors/:vendorId",
    fallback: PageSkeleton,
    loader: loaderFromQueryLoader(({ vendorId }) =>
      loadQuery<VendorGraphNodeQuery>(relayEnvironment, vendorNodeQuery, {
        vendorId: vendorId!,
      }),
    ),
    Component: withQueryRef(lazy(
      () => import("../pages/organizations/vendors/VendorDetailPage")
    )),
    children: [
      {
        path: "overview",
        fallback: LinkCardSkeleton,
        Component: lazy(
          () => import("../pages/organizations/vendors/tabs/VendorOverviewTab")
        ),
      },
      {
        path: "certifications",
        fallback: LinkCardSkeleton,
        Component: lazy(
          () =>
            import(
              "../pages/organizations/vendors/tabs/VendorCertificationsTab"
            )
        ),
      },
      {
        path: "compliance",
        fallback: LinkCardSkeleton,
        Component: lazy(
          () =>
            import("../pages/organizations/vendors/tabs/VendorComplianceTab")
        ),
      },
      {
        path: "risks",
        fallback: LinkCardSkeleton,
        Component: lazy(
          () =>
            import(
              "../pages/organizations/vendors/tabs/VendorRiskAssessmentTab"
            )
        ),
      },
      {
        path: "contacts",
        fallback: LinkCardSkeleton,
        Component: lazy(
          () =>
            import("../pages/organizations/vendors/tabs/VendorContactsTab")
        ),
      },
      {
        path: "services",
        fallback: LinkCardSkeleton,
        Component: lazy(
          () =>
            import("../pages/organizations/vendors/tabs/VendorServicesTab")
        ),
      },
    ],
  },
  {
    path: "snapshots/:snapshotId/vendors/:vendorId",
    fallback: PageSkeleton,
    loader: loaderFromQueryLoader(({ vendorId }) =>
      loadQuery<VendorGraphNodeQuery>(relayEnvironment, vendorNodeQuery, {
        vendorId: vendorId!,
      }),
    ),
    Component: withQueryRef(lazy(
      () => import("../pages/organizations/vendors/VendorDetailPage")
    )),
    children: [
      {
        path: "overview",
        fallback: LinkCardSkeleton,
        Component: lazy(
          () => import("../pages/organizations/vendors/tabs/VendorOverviewTab")
        ),
      },
      {
        path: "certifications",
        fallback: LinkCardSkeleton,
        Component: lazy(
          () =>
            import(
              "../pages/organizations/vendors/tabs/VendorCertificationsTab"
            )
        ),
      },
      {
        path: "compliance",
        fallback: LinkCardSkeleton,
        Component: lazy(
          () =>
            import("../pages/organizations/vendors/tabs/VendorComplianceTab")
        ),
      },
      {
        path: "risks",
        fallback: LinkCardSkeleton,
        Component: lazy(
          () =>
            import(
              "../pages/organizations/vendors/tabs/VendorRiskAssessmentTab"
            )
        ),
      },
      {
        path: "contacts",
        fallback: LinkCardSkeleton,
        Component: lazy(
          () =>
            import("../pages/organizations/vendors/tabs/VendorContactsTab")
        ),
      },
      {
        path: "services",
        fallback: LinkCardSkeleton,
        Component: lazy(
          () =>
            import("../pages/organizations/vendors/tabs/VendorServicesTab")
        ),
      },
    ],
  },
] satisfies AppRoute[];

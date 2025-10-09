import { lazy } from "@probo/react-lazy";
import { loadQuery } from "react-relay";
import type { AppRoute } from "/routes.tsx";
import { relayEnvironment } from "/providers/RelayProviders";
import { PageSkeleton } from "/components/skeletons/PageSkeleton";
import { vendorNodeQuery, vendorsQuery } from "/hooks/graph/VendorGraph";
import { LinkCardSkeleton } from "/components/skeletons/LinkCardSkeleton";

export const vendorRoutes = [
  {
    path: "vendors",
    fallback: PageSkeleton,
    queryLoader: ({ organizationId }) =>
      loadQuery(relayEnvironment, vendorsQuery, {
        organizationId,
        snapshotId: null
      }),
    Component: lazy(() => import("/pages/organizations/vendors/VendorsPage")),
  },
  {
    path: "snapshots/:snapshotId/vendors",
    fallback: PageSkeleton,
    queryLoader: ({ organizationId, snapshotId }) =>
      loadQuery(relayEnvironment, vendorsQuery, {
        organizationId,
        snapshotId
      }),
    Component: lazy(() => import("/pages/organizations/vendors/VendorsPage")),
  },
  {
    path: "vendors/:vendorId",
    fallback: PageSkeleton,
    queryLoader: ({ vendorId }) =>
      loadQuery(relayEnvironment, vendorNodeQuery, {
        vendorId,
      }),
    Component: lazy(
      () => import("../pages/organizations/vendors/VendorDetailPage")
    ),
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
    queryLoader: ({ vendorId }) =>
      loadQuery(relayEnvironment, vendorNodeQuery, {
        vendorId,
      }),
    Component: lazy(
      () => import("../pages/organizations/vendors/VendorDetailPage")
    ),
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

import { lazy } from "@probo/react-lazy";
import { loadQuery } from "react-relay";
import { coreEnvironment } from "/environments";
import { PageSkeleton } from "/components/skeletons/PageSkeleton";
import { vendorNodeQuery, vendorsQuery } from "/hooks/graph/VendorGraph";
import { LinkCardSkeleton } from "/components/skeletons/LinkCardSkeleton";
import {
  loaderFromQueryLoader,
  withQueryRef,
  type AppRoute,
} from "@probo/routes";
import type { VendorGraphListQuery } from "/__generated__/core/VendorGraphListQuery.graphql";
import type { VendorGraphNodeQuery } from "/__generated__/core/VendorGraphNodeQuery.graphql";

export const vendorRoutes = [
  {
    path: "vendors",
    Fallback: PageSkeleton,
    loader: loaderFromQueryLoader(({ organizationId }) =>
      loadQuery<VendorGraphListQuery>(coreEnvironment, vendorsQuery, {
        organizationId: organizationId!,
        snapshotId: null,
      }),
    ),
    Component: withQueryRef(
      lazy(() => import("/pages/organizations/vendors/VendorsPage")),
    ),
  },
  {
    path: "snapshots/:snapshotId/vendors",
    Fallback: PageSkeleton,
    loader: loaderFromQueryLoader(({ organizationId, snapshotId }) =>
      loadQuery<VendorGraphListQuery>(coreEnvironment, vendorsQuery, {
        organizationId: organizationId!,
        snapshotId,
      }),
    ),
    Component: withQueryRef(
      lazy(() => import("/pages/organizations/vendors/VendorsPage")),
    ),
  },
  {
    path: "vendors/:vendorId",
    Fallback: PageSkeleton,
    loader: loaderFromQueryLoader(({ vendorId }) =>
      loadQuery<VendorGraphNodeQuery>(coreEnvironment, vendorNodeQuery, {
        vendorId: vendorId!,
      }),
    ),
    Component: withQueryRef(
      lazy(() => import("../pages/organizations/vendors/VendorDetailPage")),
    ),
    children: [
      {
        path: "overview",
        Fallback: LinkCardSkeleton,
        Component: lazy(
          () => import("../pages/organizations/vendors/tabs/VendorOverviewTab"),
        ),
      },
      {
        path: "certifications",
        Fallback: LinkCardSkeleton,
        Component: lazy(
          () =>
            import("../pages/organizations/vendors/tabs/VendorCertificationsTab"),
        ),
      },
      {
        path: "compliance",
        Fallback: LinkCardSkeleton,
        Component: lazy(
          () =>
            import("../pages/organizations/vendors/tabs/VendorComplianceTab"),
        ),
      },
      {
        path: "risks",
        Fallback: LinkCardSkeleton,
        Component: lazy(
          () =>
            import("../pages/organizations/vendors/tabs/VendorRiskAssessmentTab"),
        ),
      },
      {
        path: "contacts",
        Fallback: LinkCardSkeleton,
        Component: lazy(
          () => import("../pages/organizations/vendors/tabs/VendorContactsTab"),
        ),
      },
      {
        path: "services",
        Fallback: LinkCardSkeleton,
        Component: lazy(
          () => import("../pages/organizations/vendors/tabs/VendorServicesTab"),
        ),
      },
    ],
  },
  {
    path: "snapshots/:snapshotId/vendors/:vendorId",
    Fallback: PageSkeleton,
    loader: loaderFromQueryLoader(({ vendorId }) =>
      loadQuery<VendorGraphNodeQuery>(coreEnvironment, vendorNodeQuery, {
        vendorId: vendorId!,
      }),
    ),
    Component: withQueryRef(
      lazy(() => import("../pages/organizations/vendors/VendorDetailPage")),
    ),
    children: [
      {
        path: "overview",
        Fallback: LinkCardSkeleton,
        Component: lazy(
          () => import("../pages/organizations/vendors/tabs/VendorOverviewTab"),
        ),
      },
      {
        path: "certifications",
        Fallback: LinkCardSkeleton,
        Component: lazy(
          () =>
            import("../pages/organizations/vendors/tabs/VendorCertificationsTab"),
        ),
      },
      {
        path: "compliance",
        Fallback: LinkCardSkeleton,
        Component: lazy(
          () =>
            import("../pages/organizations/vendors/tabs/VendorComplianceTab"),
        ),
      },
      {
        path: "risks",
        Fallback: LinkCardSkeleton,
        Component: lazy(
          () =>
            import("../pages/organizations/vendors/tabs/VendorRiskAssessmentTab"),
        ),
      },
      {
        path: "contacts",
        Fallback: LinkCardSkeleton,
        Component: lazy(
          () => import("../pages/organizations/vendors/tabs/VendorContactsTab"),
        ),
      },
      {
        path: "services",
        Fallback: LinkCardSkeleton,
        Component: lazy(
          () => import("../pages/organizations/vendors/tabs/VendorServicesTab"),
        ),
      },
    ],
  },
] satisfies AppRoute[];

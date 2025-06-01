import { lazy } from "react";
import { loadQuery } from "react-relay";
import type { AppRoute } from "/routes.tsx";
import { relayEnvironment } from "/providers/RelayProviders";
import { PageSkeleton } from "/components/skeletons/PageSkeleton";
import { vendorNodeQuery, vendorsQuery } from "/hooks/graph/VendorGraph";

export const vendorRoutes = [
  {
    path: "vendors",
    fallback: PageSkeleton,
    queryLoader: ({ organizationId }) =>
      loadQuery(relayEnvironment, vendorsQuery, { organizationId }),
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
        Component: lazy(
          () => import("../pages/organizations/vendors/tabs/VendorOverviewTab")
        ),
      },
      {
        path: "certifications",
        Component: lazy(
          () =>
            import(
              "../pages/organizations/vendors/tabs/VendorCertificationsTab"
            )
        ),
      },
      {
        path: "compliance",
        Component: lazy(
          () =>
            import("../pages/organizations/vendors/tabs/VendorComplianceTab")
        ),
      },
    ],
  },
] satisfies AppRoute[];

import { lazy } from "@probo/react-lazy";

import { LinkCardSkeleton } from "#/components/skeletons/LinkCardSkeleton";
import { PageSkeleton } from "#/components/skeletons/PageSkeleton";

export const compliancePageRoutes = [
  {
    path: "compliance-page",
    Fallback: PageSkeleton,
    Component: lazy(() => import("#/pages/organizations/compliance-page/CompliancePageLayoutLoader")),
    children: [
      {
        index: true,
        Fallback: LinkCardSkeleton,
        Component: lazy(() => import("#/pages/organizations/compliance-page/overview/CompliancePageOverviewPageLoader")),
      },
      {
        path: "domain",
        Fallback: LinkCardSkeleton,
        Component: lazy(
          () =>
            import("#/pages/organizations/compliance-page/domain/CompliancePageDomainPageLoader"),
        ),
      },
      {
        path: "brand",
        Fallback: LinkCardSkeleton,
        Component: lazy(() => import("#/pages/organizations/compliance-page/brand/CompliancePageBrandPageLoader")),
      },
      {
        path: "references",
        Fallback: LinkCardSkeleton,
        Component: lazy(() => import("#/pages/organizations/compliance-page/references/CompliancePageReferencesPageLoader")),
      },
      {
        path: "audits",
        Fallback: LinkCardSkeleton,
        Component: lazy(() => import("#/pages/organizations/compliance-page/audits/CompliancePageAuditsPageLoader")),
      },
      {
        path: "documents",
        Fallback: LinkCardSkeleton,
        Component: lazy(() => import("#/pages/organizations/compliance-page/documents/CompliancePageDocumentsPageLoader")),
      },
      {
        path: "files",
        Fallback: LinkCardSkeleton,
        Component: lazy(() => import("#/pages/organizations/compliance-page/files/CompliancePageFilesPageLoader")),
      },
      {
        path: "vendors",
        Fallback: LinkCardSkeleton,
        Component: lazy(() => import("#/pages/organizations/compliance-page/vendors/CompliancePageVendorsPageLoader")),
      },
      {
        path: "access",
        Fallback: LinkCardSkeleton,
        Component: lazy(() => import("#/pages/organizations/compliance-page/access/CompliancePageAccessPageLoader")),
      },
    ],
  },
];

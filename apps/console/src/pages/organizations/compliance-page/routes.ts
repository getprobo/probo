import { lazy } from "@probo/react-lazy";

import { PageSkeleton } from "#/components/skeletons/PageSkeleton";

export const compliancePageRoutes = [
  {
    path: "compliance-page",
    Fallback: PageSkeleton,
    Component: lazy(() => import("#/pages/organizations/compliance-page/CompliancePageLayoutLoader")),
  },
];

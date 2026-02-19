import { lazy } from "@probo/react-lazy";
import type { AppRoute } from "@probo/routes";

import { PageSkeleton } from "#/components/skeletons/PageSkeleton";

export const reportRoutes = [
  {
    path: "reports",
    Fallback: PageSkeleton,
    Component: lazy(() => import("#/pages/organizations/reports/ReportsPageLoader")),
  },
  {
    path: "reports/:reportId",
    Fallback: PageSkeleton,
    Component: lazy(() => import("#/pages/organizations/reports/ReportDetailsPageLoader")),
  },
] satisfies AppRoute[];

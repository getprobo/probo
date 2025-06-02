import { Fragment, lazy } from "react";
import { loadQuery } from "react-relay";
import type { AppRoute } from "/routes.tsx";
import { relayEnvironment } from "/providers/RelayProviders";
import { measuresQuery } from "/hooks/graph/MeasureGraph";
import { PageSkeleton } from "/components/skeletons/PageSkeleton";

export const measureRoutes = [
  {
    path: "measures",
    fallback: PageSkeleton,
    queryLoader: ({ organizationId }) =>
      loadQuery(relayEnvironment, measuresQuery, { organizationId }),
    Component: lazy(() => import("/pages/organizations/measures/MeasuresPage")),
    children: [
      {
        path: "category/:categoryId",
        Component: Fragment,
      },
    ],
  },
  {
    path: "measures/:measureId",
    fallback: PageSkeleton,
    Component: Fragment,
  },
] satisfies AppRoute[];

import { Fragment, lazy } from "react";
import { loadQuery } from "react-relay";
import type { AppRoute } from "/routes.tsx";
import { relayEnvironment } from "/providers/RelayProviders";
import { measureNodeQuery, measuresQuery } from "/hooks/graph/MeasureGraph";
import { PageSkeleton } from "/components/skeletons/PageSkeleton";
import { redirect } from "react-router";

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
    queryLoader: ({ measureId }) =>
      loadQuery(relayEnvironment, measureNodeQuery, { measureId }),
    Component: lazy(
      () => import("/pages/organizations/measures/MeasureDetailPage")
    ),
    children: [
      {
        path: "",
        loader: () => {
          throw redirect("evidences");
        },
        Component: Fragment,
      },
      {
        path: "risks",
        Component: lazy(
          () => import("/pages/organizations/measures/tabs/MeasureRisksTab.tsx")
        ),
      },
      {
        path: "controls",
        Component: lazy(
          () =>
            import("/pages/organizations/measures/tabs/MeasureControlsTab.tsx")
        ),
      },
      {
        path: "evidences",
        Component: lazy(
          () =>
            import("/pages/organizations/measures/tabs/MeasureEvidencesTab.tsx")
        ),
      },
    ],
  },
] satisfies AppRoute[];

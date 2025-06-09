import { Fragment, lazy } from "react";
import { loadQuery } from "react-relay";
import { RisksPageSkeleton } from "/components/skeletons/RisksPageSkeleton.tsx";
import type { AppRoute } from "/routes.tsx";
import { relayEnvironment } from "/providers/RelayProviders";
import { riskNodeQuery, risksQuery } from "/hooks/graph/RiskGraph";
import { PageSkeleton } from "/components/skeletons/PageSkeleton.tsx";
import { redirect } from "react-router";

export const riskRoutes = [
  {
    path: "risks",
    fallback: RisksPageSkeleton,
    queryLoader: ({ organizationId }) =>
      loadQuery(relayEnvironment, risksQuery, { organizationId }),
    Component: lazy(() => import("/pages/organizations/risks/RisksPage")),
  },
  {
    path: "risks/:riskId",
    fallback: PageSkeleton,
    queryLoader: ({ riskId }) =>
      loadQuery(relayEnvironment, riskNodeQuery, { riskId }),
    Component: lazy(() => import("/pages/organizations/risks/RiskDetailPage")),
    children: [
      {
        path: "",
        loader: () => {
          throw redirect("overview");
        },
        Component: Fragment,
      },
      {
        path: "overview",
        Component: lazy(
          () => import("/pages/organizations/risks/tabs/RiskOverviewTab.tsx")
        ),
      },
      {
        path: "measures",
        Component: lazy(
          () => import("/pages/organizations/risks/tabs/RiskMeasuresTab.tsx")
        ),
      },
      {
        path: "documents",
        Component: lazy(
          () => import("/pages/organizations/risks/tabs/RiskDocumentsTab.tsx")
        ),
      },
    ],
  },
] satisfies AppRoute[];

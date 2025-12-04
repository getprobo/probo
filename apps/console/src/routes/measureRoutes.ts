import { Fragment } from "react";
import { loadQuery } from "react-relay";
import { relayEnvironment } from "/providers/RelayProviders";
import { measureNodeQuery, measuresQuery } from "/hooks/graph/MeasureGraph";
import { PageSkeleton } from "/components/skeletons/PageSkeleton";
import { redirect } from "react-router";
import { lazy } from "@probo/react-lazy";
import { LinkCardSkeleton } from "/components/skeletons/LinkCardSkeleton";
import { loaderFromQueryLoader, withQueryRef, type AppRoute } from "@probo/routes";
import type { MeasureGraphListQuery } from "/hooks/graph/__generated__/MeasureGraphListQuery.graphql";
import type { MeasureGraphNodeQuery } from "/hooks/graph/__generated__/MeasureGraphNodeQuery.graphql";

export const measureRoutes = [
  {
    path: "measures",
    Fallback: PageSkeleton,
    loader: loaderFromQueryLoader(({ organizationId }) =>
      loadQuery<MeasureGraphListQuery>(relayEnvironment, measuresQuery, { organizationId: organizationId! }),
    ),
    Component: withQueryRef(lazy(() => import("/pages/organizations/measures/MeasuresPage"))),
    children: [
      {
        path: "category/:categoryId",
        Component: Fragment,
      },
    ],
  },
  {
    path: "measures/:measureId",
    Fallback: PageSkeleton,
    loader: loaderFromQueryLoader(({ measureId }) =>
      loadQuery<MeasureGraphNodeQuery>(relayEnvironment, measureNodeQuery, { measureId: measureId! }),
    ),
    Component: withQueryRef(lazy(() => import("/pages/organizations/measures/MeasureDetailPage"))),
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
        Fallback: LinkCardSkeleton,
        Component: lazy(
          () => import("/pages/organizations/measures/tabs/MeasureRisksTab.tsx")
        ),
      },
      {
        path: "tasks",
        Fallback: LinkCardSkeleton,
        Component: lazy(
          () => import("/pages/organizations/measures/tabs/MeasureTasksTab.tsx")
        ),
      },
      {
        path: "controls",
        Fallback: LinkCardSkeleton,
        Component: lazy(
          () =>
            import("/pages/organizations/measures/tabs/MeasureControlsTab.tsx")
        ),
      },
      {
        path: "evidences/:evidenceId?",
        Fallback: LinkCardSkeleton,
        Component: lazy(
          () =>
            import("/pages/organizations/measures/tabs/MeasureEvidencesTab.tsx")
        ),
      },
    ],
  },
] satisfies AppRoute[];

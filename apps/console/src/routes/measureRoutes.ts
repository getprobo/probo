import { Fragment } from "react";
import { loadQuery } from "react-relay";
import { coreEnvironment } from "/environments";
import { measureNodeQuery, measuresQuery } from "/hooks/graph/MeasureGraph";
import { PageSkeleton } from "/components/skeletons/PageSkeleton";
import { redirect } from "react-router";
import { lazy } from "@probo/react-lazy";
import { LinkCardSkeleton } from "/components/skeletons/LinkCardSkeleton";
import {
  loaderFromQueryLoader,
  withQueryRef,
  type AppRoute,
} from "@probo/routes";
import type { MeasureGraphListQuery } from "/__generated__/core/MeasureGraphListQuery.graphql";
import type { MeasureGraphNodeQuery } from "/__generated__/core/MeasureGraphNodeQuery.graphql";

export const measureRoutes = [
  {
    path: "measures",
    Fallback: PageSkeleton,
    loader: loaderFromQueryLoader(({ organizationId }) =>
      loadQuery<MeasureGraphListQuery>(coreEnvironment, measuresQuery, {
        organizationId: organizationId!,
      }),
    ),
    Component: withQueryRef(
      lazy(() => import("/pages/organizations/measures/MeasuresPage")),
    ),
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
      loadQuery<MeasureGraphNodeQuery>(coreEnvironment, measureNodeQuery, {
        measureId: measureId!,
      }),
    ),
    Component: withQueryRef(
      lazy(() => import("/pages/organizations/measures/MeasureDetailPage")),
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
        Fallback: LinkCardSkeleton,
        Component: lazy(
          () =>
            import("/pages/organizations/measures/tabs/MeasureRisksTab.tsx"),
        ),
      },
      {
        path: "tasks",
        Fallback: LinkCardSkeleton,
        Component: lazy(
          () =>
            import("/pages/organizations/measures/tabs/MeasureTasksTab.tsx"),
        ),
      },
      {
        path: "controls",
        Fallback: LinkCardSkeleton,
        Component: lazy(
          () =>
            import("/pages/organizations/measures/tabs/MeasureControlsTab.tsx"),
        ),
      },
      {
        path: "evidences/:evidenceId?",
        Fallback: LinkCardSkeleton,
        Component: lazy(
          () =>
            import("/pages/organizations/measures/tabs/MeasureEvidencesTab.tsx"),
        ),
      },
    ],
  },
] satisfies AppRoute[];

import { lazy } from "@probo/react-lazy";
import {
  type AppRoute,
  loaderFromQueryLoader,
  withQueryRef,
} from "@probo/routes";
import { Fragment } from "react";
import { loadQuery } from "react-relay";
import { redirect } from "react-router";

import type { MeasureGraphListQuery } from "#/__generated__/core/MeasureGraphListQuery.graphql";
import type { MeasureGraphNodeQuery } from "#/__generated__/core/MeasureGraphNodeQuery.graphql";
import { LinkCardSkeleton } from "#/components/skeletons/LinkCardSkeleton";
import { PageSkeleton } from "#/components/skeletons/PageSkeleton";
import { coreEnvironment } from "#/environments";
import { measureNodeQuery, measuresQuery } from "#/hooks/graph/MeasureGraph";

export const measureRoutes = [
  {
    path: "measures",
    Fallback: PageSkeleton,
    loader: loaderFromQueryLoader(({ organizationId }) =>
      loadQuery<MeasureGraphListQuery>(coreEnvironment, measuresQuery, {
        organizationId: organizationId,
      }),
    ),
    Component: withQueryRef(
      lazy(() => import("#/pages/organizations/measures/MeasuresPage")),
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
        measureId: measureId,
      }),
    ),
    Component: withQueryRef(
      lazy(() => import("#/pages/organizations/measures/MeasureDetailPage")),
    ),
    children: [
      {
        path: "",
        loader: () => {
          // eslint-disable-next-line
          throw redirect("evidences");
        },
        Component: Fragment,
      },
      {
        path: "risks",
        Fallback: LinkCardSkeleton,
        Component: lazy(
          () =>
            import("#/pages/organizations/measures/tabs/MeasureRisksTab"),
        ),
      },
      {
        path: "tasks",
        Fallback: LinkCardSkeleton,
        Component: lazy(
          () =>
            import("#/pages/organizations/measures/tabs/MeasureTasksTab"),
        ),
      },
      {
        path: "controls",
        Fallback: LinkCardSkeleton,
        Component: lazy(
          () =>
            import("#/pages/organizations/measures/tabs/MeasureControlsTab"),
        ),
      },
      {
        path: "evidences/:evidenceId?",
        Fallback: LinkCardSkeleton,
        Component: lazy(
          () =>
            import("#/pages/organizations/measures/tabs/MeasureEvidencesTab"),
        ),
      },
    ],
  },
] satisfies AppRoute[];

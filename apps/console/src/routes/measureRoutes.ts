import { Fragment } from "react";
import { loadQuery } from "react-relay";
import type { AppRoute } from "/routes.tsx";
import { relayEnvironment } from "/providers/RelayProviders";
import { measureNodeQuery, measuresQuery } from "/hooks/graph/MeasureGraph";
import { PageSkeleton } from "/components/skeletons/PageSkeleton";
import { redirect } from "react-router";
import { lazy } from "@probo/react-lazy";
import { LinkCardSkeleton } from "/components/skeletons/LinkCardSkeleton";

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
        fallback: LinkCardSkeleton,
        Component: lazy(
          () => import("/pages/organizations/measures/tabs/MeasureRisksTab.tsx")
        ),
      },
      {
        path: "tasks",
        fallback: LinkCardSkeleton,
        Component: lazy(
          () => import("/pages/organizations/measures/tabs/MeasureTasksTab.tsx")
        ),
      },
      {
        path: "controls",
        fallback: LinkCardSkeleton,
        Component: lazy(
          () =>
            import("/pages/organizations/measures/tabs/MeasureControlsTab.tsx")
        ),
      },
      {
        path: "evidences/:evidenceId?",
        fallback: LinkCardSkeleton,
        Component: lazy(
          () =>
            import("/pages/organizations/measures/tabs/MeasureEvidencesTab.tsx")
        ),
      },
    ],
  },
  // Snapshot measure routes
  {
    path: "snapshots/:snapshotId/risks/measures/:measureId",
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
        path: "evidences/:evidenceId?",
        fallback: LinkCardSkeleton,
        Component: lazy(
          () =>
            import("/pages/organizations/measures/tabs/MeasureEvidencesTab.tsx")
        ),
      },
    ],
  },
] satisfies AppRoute[];

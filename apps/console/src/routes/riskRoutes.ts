import { Fragment } from "react";
import { loadQuery } from "react-relay";
import { RisksPageSkeleton } from "/components/skeletons/RisksPageSkeleton.tsx";
import type { AppRoute } from "/routes.tsx";
import { relayEnvironment } from "/providers/RelayProviders";
import { riskNodeQuery, risksQuery } from "/hooks/graph/RiskGraph";
import { PageSkeleton } from "/components/skeletons/PageSkeleton";
import { redirect } from "react-router";
import { lazy } from "@probo/react-lazy";
import { LinkCardSkeleton } from "/components/skeletons/LinkCardSkeleton";

export const riskRoutes = [
  {
    path: "risks",
    fallback: RisksPageSkeleton,
    queryLoader: ({ organizationId }) =>
      loadQuery(relayEnvironment, risksQuery, {
        organizationId,
        snapshotId: null
      }),
    Component: lazy(() => import("/pages/organizations/risks/RisksPage")),
  },
  {
    path: "snapshots/:snapshotId/risks",
    fallback: RisksPageSkeleton,
    queryLoader: ({ organizationId, snapshotId }) =>
      loadQuery(relayEnvironment, risksQuery, {
        organizationId,
        snapshotId
      }),
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
        fallback: LinkCardSkeleton,
        Component: lazy(
          () => import("/pages/organizations/risks/tabs/RiskOverviewTab.tsx")
        ),
      },
      {
        path: "measures",
        fallback: LinkCardSkeleton,
        Component: lazy(
          () => import("/pages/organizations/risks/tabs/RiskMeasuresTab.tsx")
        ),
      },
      {
        path: "documents",
        fallback: LinkCardSkeleton,
        Component: lazy(
          () => import("/pages/organizations/risks/tabs/RiskDocumentsTab.tsx")
        ),
      },
      {
        path: "controls",
        fallback: LinkCardSkeleton,
        Component: lazy(
          () => import("/pages/organizations/risks/tabs/RiskControlsTab.tsx")
        ),
      },
    ],
  },
  {
    path: "snapshots/:snapshotId/risks/:riskId",
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
        fallback: LinkCardSkeleton,
        Component: lazy(
          () => import("/pages/organizations/risks/tabs/RiskOverviewTab.tsx")
        ),
      },
      {
        path: "measures",
        fallback: LinkCardSkeleton,
        Component: lazy(
          () => import("/pages/organizations/risks/tabs/RiskMeasuresTab.tsx")
        ),
      },
      {
        path: "documents",
        fallback: LinkCardSkeleton,
        Component: lazy(
          () => import("/pages/organizations/risks/tabs/RiskDocumentsTab.tsx")
        ),
      },
    ],
  },
] satisfies AppRoute[];

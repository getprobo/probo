import { Fragment } from "react";
import { loadQuery } from "react-relay";
import { RisksPageSkeleton } from "/components/skeletons/RisksPageSkeleton.tsx";
import { coreEnvironment } from "/environments";
import { riskNodeQuery, risksQuery } from "/hooks/graph/RiskGraph";
import { PageSkeleton } from "/components/skeletons/PageSkeleton";
import { redirect } from "react-router";
import { lazy } from "@probo/react-lazy";
import { LinkCardSkeleton } from "/components/skeletons/LinkCardSkeleton";
import type { RiskGraphNodeQuery } from "/__generated__/core/RiskGraphNodeQuery.graphql";
import {
  loaderFromQueryLoader,
  withQueryRef,
  type AppRoute,
} from "@probo/routes";
import type { RiskGraphListQuery } from "/__generated__/core/RiskGraphListQuery.graphql";

export const riskRoutes = [
  {
    path: "risks",
    Fallback: RisksPageSkeleton,
    loader: loaderFromQueryLoader(({ organizationId }) =>
      loadQuery<RiskGraphListQuery>(coreEnvironment, risksQuery, {
        organizationId: organizationId,
        snapshotId: null,
      }),
    ),
    Component: withQueryRef(
      lazy(() => import("/pages/organizations/risks/RisksPage")),
    ),
  },
  {
    path: "snapshots/:snapshotId/risks",
    Fallback: RisksPageSkeleton,
    loader: loaderFromQueryLoader(({ organizationId, snapshotId }) =>
      loadQuery<RiskGraphListQuery>(coreEnvironment, risksQuery, {
        organizationId: organizationId,
        snapshotId: snapshotId,
      }),
    ),
    Component: withQueryRef(
      lazy(() => import("/pages/organizations/risks/RisksPage")),
    ),
  },
  {
    path: "risks/:riskId",
    Fallback: PageSkeleton,
    loader: loaderFromQueryLoader(({ riskId }) =>
      loadQuery<RiskGraphNodeQuery>(coreEnvironment, riskNodeQuery, {
        riskId: riskId,
      }),
    ),
    Component: withQueryRef(
      lazy(() => import("/pages/organizations/risks/RiskDetailPage")),
    ),
    children: [
      {
        path: "",
        loader: () => {
          // eslint-disable-next-line
          throw redirect("overview");
        },
        Component: Fragment,
      },
      {
        path: "overview",
        Fallback: LinkCardSkeleton,
        Component: lazy(
          () => import("/pages/organizations/risks/tabs/RiskOverviewTab.tsx"),
        ),
      },
      {
        path: "measures",
        Fallback: LinkCardSkeleton,
        Component: lazy(
          () => import("/pages/organizations/risks/tabs/RiskMeasuresTab.tsx"),
        ),
      },
      {
        path: "documents",
        Fallback: LinkCardSkeleton,
        Component: lazy(
          () => import("/pages/organizations/risks/tabs/RiskDocumentsTab.tsx"),
        ),
      },
      {
        path: "controls",
        Fallback: LinkCardSkeleton,
        Component: lazy(
          () => import("/pages/organizations/risks/tabs/RiskControlsTab.tsx"),
        ),
      },
      {
        path: "obligations",
        Fallback: LinkCardSkeleton,
        Component: lazy(
          () =>
            import("/pages/organizations/risks/tabs/RiskObligationsTab.tsx"),
        ),
      },
    ],
  },
  {
    path: "snapshots/:snapshotId/risks/:riskId",
    Fallback: PageSkeleton,
    loader: loaderFromQueryLoader(({ riskId }) =>
      loadQuery<RiskGraphNodeQuery>(coreEnvironment, riskNodeQuery, {
        riskId: riskId,
      }),
    ),
    Component: withQueryRef(
      lazy(() => import("/pages/organizations/risks/RiskDetailPage")),
    ),
    children: [
      {
        path: "",
        loader: () => {
          // eslint-disable-next-line
          throw redirect("overview");
        },
        Component: Fragment,
      },
      {
        path: "overview",
        Fallback: LinkCardSkeleton,
        Component: lazy(
          () => import("/pages/organizations/risks/tabs/RiskOverviewTab.tsx"),
        ),
      },
    ],
  },
] satisfies AppRoute[];

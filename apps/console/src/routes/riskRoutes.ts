import { lazy } from "@probo/react-lazy";
import {
  type AppRoute,
  loaderFromQueryLoader,
  withQueryRef,
} from "@probo/routes";
import { Fragment } from "react";
import { loadQuery } from "react-relay";
import { redirect } from "react-router";

import type { RiskGraphListQuery } from "#/__generated__/core/RiskGraphListQuery.graphql";
import type { RiskGraphNodeQuery } from "#/__generated__/core/RiskGraphNodeQuery.graphql";
import { LinkCardSkeleton } from "#/components/skeletons/LinkCardSkeleton";
import { PageSkeleton } from "#/components/skeletons/PageSkeleton";
import { RisksPageSkeleton } from "#/components/skeletons/RisksPageSkeleton";
import { coreEnvironment } from "#/environments";
import { riskNodeQuery, risksQuery } from "#/hooks/graph/RiskGraph";

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
      lazy(() => import("#/pages/organizations/risks/RisksPage")),
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
      lazy(() => import("#/pages/organizations/risks/RisksPage")),
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
      lazy(() => import("#/pages/organizations/risks/RiskDetailPage")),
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
          () => import("#/pages/organizations/risks/tabs/RiskOverviewTab"),
        ),
      },
      {
        path: "measures",
        Fallback: LinkCardSkeleton,
        Component: lazy(
          () => import("#/pages/organizations/risks/tabs/RiskMeasuresTab"),
        ),
      },
      {
        path: "documents",
        Fallback: LinkCardSkeleton,
        Component: lazy(
          () => import("#/pages/organizations/risks/tabs/RiskDocumentsTab"),
        ),
      },
      {
        path: "controls",
        Fallback: LinkCardSkeleton,
        Component: lazy(
          () => import("#/pages/organizations/risks/tabs/RiskControlsTab"),
        ),
      },
      {
        path: "obligations",
        Fallback: LinkCardSkeleton,
        Component: lazy(
          () =>
            import("#/pages/organizations/risks/tabs/RiskObligationsTab"),
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
      lazy(() => import("#/pages/organizations/risks/RiskDetailPage")),
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
          () => import("#/pages/organizations/risks/tabs/RiskOverviewTab"),
        ),
      },
    ],
  },
] satisfies AppRoute[];

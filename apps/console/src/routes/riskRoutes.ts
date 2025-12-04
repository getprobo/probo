import { Fragment } from "react";
import { loadQuery } from "react-relay";
import { RisksPageSkeleton } from "/components/skeletons/RisksPageSkeleton.tsx";
import { relayEnvironment } from "/providers/RelayProviders";
import { riskNodeQuery, risksQuery } from "/hooks/graph/RiskGraph";
import { PageSkeleton } from "/components/skeletons/PageSkeleton";
import { redirect } from "react-router";
import { lazy } from "@probo/react-lazy";
import { LinkCardSkeleton } from "/components/skeletons/LinkCardSkeleton";
import type { RiskGraphNodeQuery } from "/hooks/graph/__generated__/RiskGraphNodeQuery.graphql";
import { loaderFromQueryLoader, withQueryRef, type AppRoute } from "/routes";
import type { RiskGraphListQuery } from "/hooks/graph/__generated__/RiskGraphListQuery.graphql";

export const riskRoutes = [
  {
    path: "risks",
    fallback: RisksPageSkeleton,
    loader: loaderFromQueryLoader(({ organizationId }) =>
      loadQuery<RiskGraphListQuery>(relayEnvironment, risksQuery, {
        organizationId: organizationId!,
        snapshotId: null
      }),
    ),
    Component: withQueryRef(lazy(() => import("/pages/organizations/risks/RisksPage"))),
  },
  {
    path: "snapshots/:snapshotId/risks",
    fallback: RisksPageSkeleton,
    loader: loaderFromQueryLoader(({ organizationId, snapshotId }) =>
      loadQuery<RiskGraphListQuery>(relayEnvironment, risksQuery, {
        organizationId: organizationId!,
        snapshotId: snapshotId!
      }),
    ),
    Component: withQueryRef(lazy(() => import("/pages/organizations/risks/RisksPage"))),
  },
  {
    path: "risks/:riskId",
    fallback: PageSkeleton,
    loader: loaderFromQueryLoader(({ riskId }) =>
      loadQuery<RiskGraphNodeQuery>(relayEnvironment, riskNodeQuery, { riskId: riskId! }),
    ),
    Component: withQueryRef(lazy(() => import("/pages/organizations/risks/RiskDetailPage"))),
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
      {
        path: "obligations",
        fallback: LinkCardSkeleton,
        Component: lazy(
          () => import("/pages/organizations/risks/tabs/RiskObligationsTab.tsx")
        ),
      },
    ],
  },
  {
    path: "snapshots/:snapshotId/risks/:riskId",
    fallback: PageSkeleton,
    loader: loaderFromQueryLoader(({ riskId }) =>
      loadQuery<RiskGraphNodeQuery>(relayEnvironment, riskNodeQuery, { riskId: riskId! }),
    ),
    Component: withQueryRef(lazy(() => import("/pages/organizations/risks/RiskDetailPage"))),
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
    ],
  },
] satisfies AppRoute[];

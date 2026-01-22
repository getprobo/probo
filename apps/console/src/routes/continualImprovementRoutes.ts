import { lazy } from "@probo/react-lazy";
import {
  type AppRoute,
  loaderFromQueryLoader,
  withQueryRef,
} from "@probo/routes";
import { loadQuery } from "react-relay";

import type { ContinualImprovementGraphListQuery } from "#/__generated__/core/ContinualImprovementGraphListQuery.graphql";
import type { ContinualImprovementGraphNodeQuery } from "#/__generated__/core/ContinualImprovementGraphNodeQuery.graphql";
import { PageSkeleton } from "#/components/skeletons/PageSkeleton";
import { coreEnvironment } from "#/environments";
import {
  continualImprovementNodeQuery,
  continualImprovementsQuery,
} from "#/hooks/graph/ContinualImprovementGraph";

export const continualImprovementRoutes = [
  {
    path: "continual-improvements",
    Fallback: PageSkeleton,
    loader: loaderFromQueryLoader(({ organizationId }) =>
      loadQuery<ContinualImprovementGraphListQuery>(
        coreEnvironment,
        continualImprovementsQuery,
        {
          organizationId,
          snapshotId: null,
        },
      ),
    ),
    Component: withQueryRef(
      lazy(
        () =>
          import("#/pages/organizations/continualImprovements/ContinualImprovementsPage"),
      ),
    ),
  },
  {
    path: "snapshots/:snapshotId/continual-improvements",
    Fallback: PageSkeleton,
    loader: loaderFromQueryLoader(({ organizationId, snapshotId }) =>
      loadQuery<ContinualImprovementGraphListQuery>(
        coreEnvironment,
        continualImprovementsQuery,
        {
          organizationId,
          snapshotId,
        },
      ),
    ),
    Component: withQueryRef(
      lazy(
        () =>
          import("#/pages/organizations/continualImprovements/ContinualImprovementsPage"),
      ),
    ),
  },
  {
    path: "continual-improvements/:improvementId",
    Fallback: PageSkeleton,
    loader: loaderFromQueryLoader(({ improvementId }) =>
      loadQuery<ContinualImprovementGraphNodeQuery>(
        coreEnvironment,
        continualImprovementNodeQuery,
        {
          continualImprovementId: improvementId,
        },
      ),
    ),
    Component: withQueryRef(
      lazy(
        () =>
          import("#/pages/organizations/continualImprovements/ContinualImprovementDetailsPage"),
      ),
    ),
  },
  {
    path: "snapshots/:snapshotId/continual-improvements/:improvementId",
    Fallback: PageSkeleton,
    loader: loaderFromQueryLoader(({ improvementId }) =>
      loadQuery<ContinualImprovementGraphNodeQuery>(
        coreEnvironment,
        continualImprovementNodeQuery,
        {
          continualImprovementId: improvementId,
        },
      ),
    ),
    Component: withQueryRef(
      lazy(
        () =>
          import("#/pages/organizations/continualImprovements/ContinualImprovementDetailsPage"),
      ),
    ),
  },
] satisfies AppRoute[];

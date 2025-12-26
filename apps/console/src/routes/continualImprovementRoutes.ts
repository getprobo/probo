import { loadQuery } from "react-relay";
import { coreEnvironment } from "/environments";
import { PageSkeleton } from "/components/skeletons/PageSkeleton";
import { lazy } from "@probo/react-lazy";
import {
  continualImprovementsQuery,
  continualImprovementNodeQuery,
} from "/hooks/graph/ContinualImprovementGraph";
import type { ContinualImprovementGraphListQuery } from "/hooks/graph/__generated__/ContinualImprovementGraphListQuery.graphql";
import type { ContinualImprovementGraphNodeQuery } from "/hooks/graph/__generated__/ContinualImprovementGraphNodeQuery.graphql";
import {
  loaderFromQueryLoader,
  withQueryRef,
  type AppRoute,
} from "@probo/routes";

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
          import("/pages/organizations/continualImprovements/ContinualImprovementsPage"),
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
          import("/pages/organizations/continualImprovements/ContinualImprovementsPage"),
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
          continualImprovementId: improvementId!,
        },
      ),
    ),
    Component: withQueryRef(
      lazy(
        () =>
          import("/pages/organizations/continualImprovements/ContinualImprovementDetailsPage"),
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
          continualImprovementId: improvementId!,
        },
      ),
    ),
    Component: withQueryRef(
      lazy(
        () =>
          import("/pages/organizations/continualImprovements/ContinualImprovementDetailsPage"),
      ),
    ),
  },
] satisfies AppRoute[];

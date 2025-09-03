import { loadQuery } from "react-relay";
import { relayEnvironment } from "/providers/RelayProviders";
import { PageSkeleton } from "/components/skeletons/PageSkeleton";
import { lazy } from "@probo/react-lazy";
import { continualImprovementsQuery, continualImprovementNodeQuery } from "/hooks/graph/ContinualImprovementGraph";
import type { AppRoute } from "/routes";

export const continualImprovementRoutes = [
  {
    path: "continual-improvements",
    fallback: PageSkeleton,
    queryLoader: (params: Record<string, string>) =>
      loadQuery(relayEnvironment, continualImprovementsQuery, {
        organizationId: params.organizationId,
        snapshotId: null
      }),
    Component: lazy(
      () => import("/pages/organizations/continualImprovements/ContinualImprovementsPage")
    ),
  },
  {
    path: "snapshots/:snapshotId/continual-improvements",
    fallback: PageSkeleton,
    queryLoader: (params: Record<string, string>) =>
      loadQuery(relayEnvironment, continualImprovementsQuery, {
        organizationId: params.organizationId,
        snapshotId: params.snapshotId
      }),
    Component: lazy(
      () => import("/pages/organizations/continualImprovements/ContinualImprovementsPage")
    ),
  },
  {
    path: "continual-improvements/:improvementId",
    fallback: PageSkeleton,
    queryLoader: (params: Record<string, string>) =>
      loadQuery(relayEnvironment, continualImprovementNodeQuery, {
        continualImprovementId: params.improvementId
      }),
    Component: lazy(
      () => import("/pages/organizations/continualImprovements/ContinualImprovementDetailsPage")
    ),
  },
  {
    path: "snapshots/:snapshotId/continual-improvements/:improvementId",
    fallback: PageSkeleton,
    queryLoader: (params: Record<string, string>) =>
      loadQuery(relayEnvironment, continualImprovementNodeQuery, {
        continualImprovementId: params.improvementId
      }),
    Component: lazy(
      () => import("/pages/organizations/continualImprovements/ContinualImprovementDetailsPage")
    ),
  },
] satisfies AppRoute[];

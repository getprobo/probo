import { loadQuery } from "react-relay";
import { relayEnvironment } from "/providers/RelayProviders";
import { PageSkeleton } from "/components/skeletons/PageSkeleton";
import { lazy } from "@probo/react-lazy";
import { continualImprovementRegistriesQuery, continualImprovementRegistryNodeQuery } from "/hooks/graph/ContinualImprovementRegistryGraph";
import type { AppRoute } from "/routes";

export const continualImprovementRegistryRoutes = [
  {
    path: "continual-improvement-registries",
    fallback: PageSkeleton,
    queryLoader: (params: Record<string, string>) =>
      loadQuery(relayEnvironment, continualImprovementRegistriesQuery, {
        organizationId: params.organizationId,
        snapshotId: null
      }),
    Component: lazy(
      () => import("/pages/organizations/continualImprovementRegistries/ContinualImprovementRegistriesPage")
    ),
  },
  {
    path: "snapshots/:snapshotId/continual-improvement-registries",
    fallback: PageSkeleton,
    queryLoader: (params: Record<string, string>) =>
      loadQuery(relayEnvironment, continualImprovementRegistriesQuery, {
        organizationId: params.organizationId,
        snapshotId: params.snapshotId
      }),
    Component: lazy(
      () => import("/pages/organizations/continualImprovementRegistries/ContinualImprovementRegistriesPage")
    ),
  },
  {
    path: "continual-improvement-registries/:registryId",
    fallback: PageSkeleton,
    queryLoader: (params: Record<string, string>) =>
      loadQuery(relayEnvironment, continualImprovementRegistryNodeQuery, {
        continualImprovementRegistryId: params.registryId
      }),
    Component: lazy(
      () => import("/pages/organizations/continualImprovementRegistries/ContinualImprovementRegistryDetailsPage")
    ),
  },
  {
    path: "snapshots/:snapshotId/continual-improvement-registries/:registryId",
    fallback: PageSkeleton,
    queryLoader: (params: Record<string, string>) =>
      loadQuery(relayEnvironment, continualImprovementRegistryNodeQuery, {
        continualImprovementRegistryId: params.registryId
      }),
    Component: lazy(
      () => import("/pages/organizations/continualImprovementRegistries/ContinualImprovementRegistryDetailsPage")
    ),
  },
] satisfies AppRoute[];

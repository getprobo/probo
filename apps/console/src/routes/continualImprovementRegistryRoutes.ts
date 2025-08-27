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
    queryLoader: ({ organizationId }: { organizationId: string }) =>
      loadQuery(relayEnvironment, continualImprovementRegistriesQuery, { organizationId }),
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
] satisfies AppRoute[];

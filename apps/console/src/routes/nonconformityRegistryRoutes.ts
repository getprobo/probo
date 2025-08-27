import { loadQuery } from "react-relay";
import { relayEnvironment } from "/providers/RelayProviders";
import { PageSkeleton } from "/components/skeletons/PageSkeleton";
import { lazy } from "@probo/react-lazy";
import { nonconformityRegistriesQuery, nonconformityRegistryNodeQuery } from "/hooks/graph/NonconformityRegistryGraph";
import type { AppRoute } from "/routes";

export const nonconformityRegistryRoutes= [
  {
    path: "nonconformity-registries",
    fallback: PageSkeleton,
    queryLoader: ({ organizationId }: { organizationId: string }) =>
      loadQuery(relayEnvironment, nonconformityRegistriesQuery, { organizationId }),
    Component: lazy(
      () => import("/pages/organizations/nonconformityRegistries/NonconformityRegistriesPage")
    ),
  },
  {
    path: "nonconformity-registries/:registryId",
    fallback: PageSkeleton,
    queryLoader: (params: Record<string, string>) =>
      loadQuery(relayEnvironment, nonconformityRegistryNodeQuery, {
        nonconformityRegistryId: params.registryId
      }),
    Component: lazy(
      () => import("/pages/organizations/nonconformityRegistries/NonconformityRegistryDetailsPage")
    ),
  },
] satisfies AppRoute[];

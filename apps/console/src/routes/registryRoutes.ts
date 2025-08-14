import { loadQuery } from "react-relay";
import { relayEnvironment } from "/providers/RelayProviders";
import { PageSkeleton } from "/components/skeletons/PageSkeleton";
import { lazy } from "@probo/react-lazy";
import { nonconformityRegistriesQuery } from "/hooks/graph/NonconformityRegistryGraph";
import type { AppRoute } from "/routes";

export const registryRoutes = [
  {
    path: "registries",
    fallback: PageSkeleton,
    queryLoader: ({ organizationId }: { organizationId: string }) =>
      loadQuery(relayEnvironment, nonconformityRegistriesQuery, { organizationId }),
    Component: lazy(
      () => import("/pages/organizations/registries/RegistriesPage")
    ),
  },
] satisfies AppRoute[];;

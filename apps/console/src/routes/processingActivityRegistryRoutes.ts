import { loadQuery } from "react-relay";
import { relayEnvironment } from "/providers/RelayProviders";
import { PageSkeleton } from "/components/skeletons/PageSkeleton";
import { lazy } from "@probo/react-lazy";
import { processingActivityRegistriesQuery, processingActivityRegistryNodeQuery } from "/hooks/graph/ProcessingActivityRegistryGraph";
import type { AppRoute } from "/routes";

export const processingActivityRegistryRoutes = [
  {
    path: "processing-activity-registries",
    fallback: PageSkeleton,
    queryLoader: ({ organizationId }: { organizationId: string }) =>
      loadQuery(relayEnvironment, processingActivityRegistriesQuery, { organizationId }),
    Component: lazy(
      () => import("/pages/organizations/processingActivityRegistries/ProcessingActivityRegistriesPage")
    ),
  },
  {
    path: "processing-activity-registries/:registryId",
    fallback: PageSkeleton,
    queryLoader: (params: Record<string, string>) =>
      loadQuery(relayEnvironment, processingActivityRegistryNodeQuery, {
        processingActivityRegistryId: params.registryId
      }),
    Component: lazy(
      () => import("/pages/organizations/processingActivityRegistries/ProcessingActivityRegistryDetailsPage")
    ),
  },
] satisfies AppRoute[];

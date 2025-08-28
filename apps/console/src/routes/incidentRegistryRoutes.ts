import { lazy } from "@probo/react-lazy";
import { PageSkeleton } from "/components/skeletons/PageSkeleton";
import type { AppRoute } from "/routes";

export const incidentRegistryRoutes = [
  {
    path: "incident-registries",
    fallback: PageSkeleton,
    Component: lazy(
      () => import("/pages/organizations/incidentRegistries/IncidentRegistriesPage")
    ),
  },
  {
    path: "incident-registries/:incidentRegistryId",
    fallback: PageSkeleton,
    Component: lazy(
      () => import("/pages/organizations/incidentRegistries/IncidentRegistryDetailsPage")
    ),
  },
  // Legacy route support for backward compatibility
  {
    path: "incident-register",
    fallback: PageSkeleton,
    Component: lazy(
      () => import("/pages/organizations/incidentRegistries/IncidentRegistriesPage")
    ),
  },
  {
    path: "incident-register/:incidentRegistryId",
    fallback: PageSkeleton,
    Component: lazy(
      () => import("/pages/organizations/incidentRegistries/IncidentRegistryDetailsPage")
    ),
  },
] satisfies AppRoute[];
import { loadQuery } from "react-relay";
import { relayEnvironment } from "/providers/RelayProviders";
import { PageSkeleton } from "/components/skeletons/PageSkeleton";
import { lazy } from "@probo/react-lazy";
import { complianceRegistriesQuery, complianceRegistryNodeQuery } from "/hooks/graph/ComplianceRegistryGraph";
import type { AppRoute } from "/routes";

export const complianceRegistryRoutes = [
  {
    path: "compliance-registries",
    fallback: PageSkeleton,
    queryLoader: (params: Record<string, string>) =>
      loadQuery(relayEnvironment, complianceRegistriesQuery, {
        organizationId: params.organizationId,
        snapshotId: null
      }),
    Component: lazy(
      () => import("/pages/organizations/complianceRegistries/ComplianceRegistriesPage")
    ),
  },
  {
    path: "snapshots/:snapshotId/compliance-registries",
    fallback: PageSkeleton,
    queryLoader: (params: Record<string, string>) =>
      loadQuery(relayEnvironment, complianceRegistriesQuery, {
        organizationId: params.organizationId,
        snapshotId: params.snapshotId
      }),
    Component: lazy(
      () => import("/pages/organizations/complianceRegistries/ComplianceRegistriesPage")
    ),
  },
  {
    path: "compliance-registries/:registryId",
    fallback: PageSkeleton,
    queryLoader: (params: Record<string, string>) =>
      loadQuery(relayEnvironment, complianceRegistryNodeQuery, {
        complianceRegistryId: params.registryId
      }),
    Component: lazy(
      () => import("/pages/organizations/complianceRegistries/ComplianceRegistryDetailsPage")
    ),
  },
  {
    path: "snapshots/:snapshotId/compliance-registries/:registryId",
    fallback: PageSkeleton,
    queryLoader: (params: Record<string, string>) =>
      loadQuery(relayEnvironment, complianceRegistryNodeQuery, {
        complianceRegistryId: params.registryId
      }),
    Component: lazy(
      () => import("/pages/organizations/complianceRegistries/ComplianceRegistryDetailsPage")
    ),
  },
] satisfies AppRoute[];

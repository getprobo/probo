import { lazy } from "react";
import { loadQuery } from "react-relay";
import type { AppRoute } from "/routes.tsx";
import { relayEnvironment } from "/providers/RelayProviders";
import { policiesQuery } from "/hooks/graph/PolicyGraph";
import { policyNodeQuery } from "/hooks/graph/PolicyGraph";
import { PageSkeleton } from "/components/skeletons/PageSkeleton";

export const policiesRoutes = [
  {
    path: "policies",
    fallback: PageSkeleton,
    queryLoader: ({ organizationId }) =>
      loadQuery(relayEnvironment, policiesQuery, { organizationId }),
    Component: lazy(() => import("/pages/organizations/policies/PoliciesPage")),
  },
  {
    path: "policies/:policyId",
    fallback: PageSkeleton,
    queryLoader: ({ policyId }) =>
      loadQuery(relayEnvironment, policyNodeQuery, { policyId }),
    Component: lazy(() => import("/pages/organizations/policies/PolicyPage")),
  },
] satisfies AppRoute[];

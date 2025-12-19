import { lazy } from "@probo/react-lazy";
import { loadQuery } from "react-relay";
import { consoleEnvironment } from "/environments";
import { PageSkeleton } from "/components/skeletons/PageSkeleton.tsx";
import {
  paginatedStateOfApplicabilityQuery,
  stateOfApplicabilityNodeQuery,
} from "/hooks/graph/StateOfApplicabilityGraph";
import type { StateOfApplicabilityGraphPaginatedQuery } from "/hooks/graph/__generated__/StateOfApplicabilityGraphPaginatedQuery.graphql";
import type { StateOfApplicabilityGraphNodeQuery } from "/hooks/graph/__generated__/StateOfApplicabilityGraphNodeQuery.graphql";
import { loaderFromQueryLoader, withQueryRef, type AppRoute } from "@probo/routes";

export const statesOfApplicabilityRoutes = [
  {
    path: "states-of-applicability",
    Fallback: PageSkeleton,
    loader: loaderFromQueryLoader(({ organizationId }) =>
      loadQuery<StateOfApplicabilityGraphPaginatedQuery>(
        consoleEnvironment,
        paginatedStateOfApplicabilityQuery,
        { organizationId: organizationId! },
      ),
    ),
    Component: withQueryRef(
      lazy(() => import("/pages/organizations/states-of-applicability/StatesOfApplicabilityPage")),
    ),
  },
  {
    path: "snapshots/:snapshotId/states-of-applicability",
    Fallback: PageSkeleton,
    loader: loaderFromQueryLoader(({ organizationId }) =>
      loadQuery<StateOfApplicabilityGraphPaginatedQuery>(
        consoleEnvironment,
        paginatedStateOfApplicabilityQuery,
        { organizationId: organizationId! },
      ),
    ),
    Component: withQueryRef(
      lazy(() => import("/pages/organizations/states-of-applicability/StatesOfApplicabilityPage")),
    ),
  },
  {
    path: "states-of-applicability/:stateOfApplicabilityId",
    Fallback: PageSkeleton,
    loader: loaderFromQueryLoader(({ stateOfApplicabilityId }) =>
      loadQuery<StateOfApplicabilityGraphNodeQuery>(
        consoleEnvironment,
        stateOfApplicabilityNodeQuery,
        { stateOfApplicabilityId: stateOfApplicabilityId! },
      ),
    ),
    Component: withQueryRef(
      lazy(() => import("/pages/organizations/states-of-applicability/StateOfApplicabilityDetailPage")),
    ),
  },
  {
    path: "snapshots/:snapshotId/states-of-applicability/:stateOfApplicabilityId",
    Fallback: PageSkeleton,
    loader: loaderFromQueryLoader(({ stateOfApplicabilityId }) =>
      loadQuery<StateOfApplicabilityGraphNodeQuery>(
        consoleEnvironment,
        stateOfApplicabilityNodeQuery,
        { stateOfApplicabilityId: stateOfApplicabilityId! },
      ),
    ),
    Component: withQueryRef(
      lazy(() => import("/pages/organizations/states-of-applicability/StateOfApplicabilityDetailPage")),
    ),
  },
] satisfies AppRoute[];

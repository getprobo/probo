import { loadQuery } from "react-relay";
import { coreEnvironment } from "/environments";
import { PageSkeleton } from "/components/skeletons/PageSkeleton";
import { lazy } from "@probo/react-lazy";
import { rightsRequestsQuery, rightsRequestNodeQuery } from "/hooks/graph/RightsRequestGraph";
import type { RightsRequestGraphListQuery } from "/hooks/graph/__generated__/RightsRequestGraphListQuery.graphql";
import type { RightsRequestGraphNodeQuery } from "/hooks/graph/__generated__/RightsRequestGraphNodeQuery.graphql";
import { loaderFromQueryLoader, withQueryRef, type AppRoute } from "@probo/routes";

export const rightsRequestRoutes = [
  {
    path: "rights-requests",
    Fallback: PageSkeleton,
    loader: loaderFromQueryLoader(({ organizationId }) =>
      loadQuery<RightsRequestGraphListQuery>(coreEnvironment, rightsRequestsQuery, {
        organizationId,
      }),
    ),
    Component: withQueryRef(lazy(
      () => import("/pages/organizations/rightsRequests/RightsRequestsPage")
    )),
  },
  {
    path: "rights-requests/:requestId",
    Fallback: PageSkeleton,
    loader: loaderFromQueryLoader(({ requestId }) =>
      loadQuery<RightsRequestGraphNodeQuery>(coreEnvironment, rightsRequestNodeQuery, {
        rightsRequestId: requestId!,
      }),
    ),
    Component: withQueryRef(lazy(
      () => import("/pages/organizations/rightsRequests/RightsRequestDetailsPage")
    )),
  },
] satisfies AppRoute[];

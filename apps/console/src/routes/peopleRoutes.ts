import { lazy } from "@probo/react-lazy";
import { loadQuery } from "react-relay";
import { relayEnvironment } from "/providers/RelayProviders";
import { PageSkeleton } from "/components/skeletons/PageSkeleton.tsx";
import {
  paginatedPeopleQuery,
  peopleNodeQuery,
} from "/hooks/graph/PeopleGraph";
import { LinkCardSkeleton } from "/components/skeletons/LinkCardSkeleton";
import { loaderFromQueryLoader, withQueryRef, type AppRoute } from "@probo/routes";
import type { PeopleGraphPaginatedQuery } from "/hooks/graph/__generated__/PeopleGraphPaginatedQuery.graphql";
import type { PeopleGraphNodeQuery } from "/hooks/graph/__generated__/PeopleGraphNodeQuery.graphql";

export const peopleRoutes = [
  {
    path: "people",
    Fallback: PageSkeleton,
    loader: loaderFromQueryLoader(({ organizationId }) =>
      loadQuery<PeopleGraphPaginatedQuery>(relayEnvironment, paginatedPeopleQuery, { organizationId: organizationId! }),
    ),
    Component: withQueryRef(lazy(() => import("/pages/organizations/people/PeopleListPage"))),
  },
  {
    path: "people/:peopleId",
    Fallback: PageSkeleton,
    loader: loaderFromQueryLoader(({ peopleId }) =>
      loadQuery<PeopleGraphNodeQuery>(relayEnvironment, peopleNodeQuery, { peopleId: peopleId! }),
    ),
    Component: withQueryRef(lazy(
      () => import("/pages/organizations/people/PeopleDetailPage")
    )),
    children: [
      {
        path: "tasks",
        Fallback: LinkCardSkeleton,
        Component: lazy(
          () => import("/pages/organizations/people/tabs/PeopleTasksTab")
        ),
      },
      {
        path: "role",
        Fallback: LinkCardSkeleton,
        Component: lazy(
          () => import("/pages/organizations/people/tabs/PeopleRoleTab")
        ),
      },
      {
        path: "profile",
        Fallback: LinkCardSkeleton,
        Component: lazy(
          () => import("/pages/organizations/people/tabs/PeopleProfileTab")
        ),
      },
    ],
  },
] satisfies AppRoute[];

import { lazy } from "@probo/react-lazy";
import { loadQuery } from "react-relay";
import {
  loaderFromQueryLoader,
  withQueryRef,
  type AppRoute,
} from "@probo/routes";

import { coreEnvironment } from "/environments";
import { PageSkeleton } from "/components/skeletons/PageSkeleton.tsx";
import {
  paginatedPeopleQuery,
  peopleNodeQuery,
} from "/hooks/graph/PeopleGraph";
import { LinkCardSkeleton } from "/components/skeletons/LinkCardSkeleton";
import type { PeopleGraphPaginatedQuery } from "/__generated__/core/PeopleGraphPaginatedQuery.graphql";
import type { PeopleGraphNodeQuery } from "/__generated__/core/PeopleGraphNodeQuery.graphql";

export const peopleRoutes = [
  {
    path: "people",
    Fallback: PageSkeleton,
    loader: loaderFromQueryLoader(({ organizationId }) =>
      loadQuery<PeopleGraphPaginatedQuery>(
        coreEnvironment,
        paginatedPeopleQuery,
        { organizationId: organizationId },
      ),
    ),
    Component: withQueryRef(
      lazy(() => import("/pages/organizations/people/PeopleListPage")),
    ),
  },
  {
    path: "people/:peopleId",
    Fallback: PageSkeleton,
    loader: loaderFromQueryLoader(({ peopleId }) =>
      loadQuery<PeopleGraphNodeQuery>(coreEnvironment, peopleNodeQuery, {
        peopleId: peopleId,
      }),
    ),
    Component: withQueryRef(
      lazy(() => import("/pages/organizations/people/PeopleDetailPage")),
    ),
    children: [
      {
        path: "tasks",
        Fallback: LinkCardSkeleton,
        Component: lazy(
          () => import("/pages/organizations/people/tabs/PeopleTasksTab"),
        ),
      },
      {
        path: "role",
        Fallback: LinkCardSkeleton,
        Component: lazy(
          () => import("/pages/organizations/people/tabs/PeopleRoleTab"),
        ),
      },
      {
        path: "profile",
        Fallback: LinkCardSkeleton,
        Component: lazy(
          () => import("/pages/organizations/people/tabs/PeopleProfileTab"),
        ),
      },
    ],
  },
] satisfies AppRoute[];

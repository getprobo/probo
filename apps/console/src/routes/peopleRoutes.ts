import { lazy } from "@probo/react-lazy";
import { loadQuery } from "react-relay";
import type { AppRoute } from "/routes.tsx";
import { relayEnvironment } from "/providers/RelayProviders";
import { PageSkeleton } from "/components/skeletons/PageSkeleton.tsx";
import {
  paginatedPeopleQuery,
  peopleNodeQuery,
} from "/hooks/graph/PeopleGraph";
import { LinkCardSkeleton } from "/components/skeletons/LinkCardSkeleton";

export const peopleRoutes = [
  {
    path: "people",
    fallback: PageSkeleton,
    queryLoader: ({ organizationId }) =>
      loadQuery(relayEnvironment, paginatedPeopleQuery, { organizationId }),
    Component: lazy(() => import("/pages/organizations/people/PeopleListPage")),
  },
  {
    path: "people/:peopleId",
    fallback: PageSkeleton,
    queryLoader: ({ peopleId }) =>
      loadQuery(relayEnvironment, peopleNodeQuery, { peopleId }),
    Component: lazy(
      () => import("/pages/organizations/people/PeopleDetailPage")
    ),
    children: [
      {
        path: "tasks",
        fallback: LinkCardSkeleton,
        Component: lazy(
          () => import("/pages/organizations/people/tabs/PeopleTasksTab")
        ),
      },
      {
        path: "role",
        fallback: LinkCardSkeleton,
        Component: lazy(
          () => import("/pages/organizations/people/tabs/PeopleRoleTab")
        ),
      },
      {
        path: "profile",
        fallback: LinkCardSkeleton,
        Component: lazy(
          () => import("/pages/organizations/people/tabs/PeopleProfileTab")
        ),
      },
    ],
  },
] satisfies AppRoute[];

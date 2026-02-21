import { lazy } from "@probo/react-lazy";

import { LinkCardSkeleton } from "#/components/skeletons/LinkCardSkeleton";

export const peopleRoutes = [
  {
    path: "people",
    children: [
      {
        index: true,
        Component: lazy(() => import("#/pages/iam/organizations/people/PeoplePageLoader")),
        Fallback: LinkCardSkeleton,
      },
      {
        path: ":personId",
        Component: lazy(() => import("#/pages/iam/organizations/people/PersonPageLoader")),
        Fallback: LinkCardSkeleton,
      },
    ],
  },
];

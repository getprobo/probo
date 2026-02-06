import { lazy } from "@probo/react-lazy";

import { LinkCardSkeleton } from "#/components/skeletons/LinkCardSkeleton";

export const userRoutes = [
  {
    path: "users",
    children: [
      {
        index: true,
        Component: lazy(() => import("#/pages/iam/organizations/users/MembersPageLoader")),
        Fallback: LinkCardSkeleton,
      },
      {
        path: ":userId",
        Component: lazy(() => import("#/pages/iam/organizations/users/UserPageLoader")),
        Fallback: LinkCardSkeleton,
      },
    ],
  },
];

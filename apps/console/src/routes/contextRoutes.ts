import { lazy } from "@probo/react-lazy";
import type { AppRoute } from "@probo/routes";

import { LinkCardSkeleton } from "#/components/skeletons/LinkCardSkeleton";
import { PageSkeleton } from "#/components/skeletons/PageSkeleton";

export const contextRoutes = [
  {
    path: "context",
    Fallback: PageSkeleton,
    Component: lazy(
      () => import("#/pages/organizations/context/ContextLayoutLoader"),
    ),
    children: [
      {
        index: true,
        Fallback: LinkCardSkeleton,
        Component: lazy(
          () => import("#/pages/organizations/context/ContextPageLoader"),
        ),
      },
      {
        path: "meetings",
        Fallback: LinkCardSkeleton,
        Component: lazy(
          () => import("#/pages/organizations/meetings/MeetingsPageLoader"),
        ),
      },
      {
        path: "meetings/:meetingId",
        Fallback: PageSkeleton,
        Component: lazy(
          () => import("#/pages/organizations/meetings/MeetingDetailPageLoader"),
        ),
        children: [
          {
            index: true,
            Fallback: LinkCardSkeleton,
            Component: lazy(
              () => import("#/pages/organizations/meetings/tabs/MeetingMinutesTab"),
            ),
          },
        ],
      },
    ],
  },
] satisfies AppRoute[];

import { lazy } from "@probo/react-lazy";
import type { AppRoute } from "@probo/routes";

import { PageSkeleton } from "#/components/skeletons/PageSkeleton";

export const accessReviewRoutes = [
  {
    path: "access-reviews",
    Fallback: PageSkeleton,
    Component: lazy(
      () => import("#/pages/organizations/access-reviews/AccessReviewLayoutLoader"),
    ),
    children: [
      {
        index: true,
        Fallback: PageSkeleton,
        Component: lazy(
          () => import("#/pages/organizations/access-reviews/campaigns/AccessReviewCampaignsTab"),
        ),
      },
      {
        path: "sources",
        Fallback: PageSkeleton,
        Component: lazy(
          () => import("#/pages/organizations/access-reviews/sources/AccessReviewSourcesTab"),
        ),
      },
    ],
  },
  {
    path: "access-reviews/campaigns/:campaignId",
    Fallback: PageSkeleton,
    Component: lazy(
      () => import("#/pages/organizations/access-reviews/campaigns/CampaignDetailPageLoader"),
    ),
  },
  {
    path: "access-reviews/sources/new/csv",
    Fallback: PageSkeleton,
    Component: lazy(
      () => import("#/pages/organizations/access-reviews/CreateCsvAccessSourcePageLoader"),
    ),
  },
] satisfies AppRoute[];

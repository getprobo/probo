import { lazy } from "@probo/react-lazy";
import type { AppRoute } from "@probo/routes";

import { PageSkeleton } from "#/components/skeletons/PageSkeleton";

export const accessReviewRoutes = [
  {
    path: "access-reviews",
    Fallback: PageSkeleton,
    Component: lazy(
      () => import("#/pages/organizations/access-reviews/AccessReviewPageLoader"),
    ),
  },
  {
    path: "access-reviews/sources/new",
    Fallback: PageSkeleton,
    Component: lazy(
      () => import("#/pages/organizations/access-reviews/CreateAccessSourcePageLoader"),
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

import { lazy } from "@probo/react-lazy";
import {
  type AppRoute,
  loaderFromQueryLoader,
  withQueryRef,
} from "@probo/routes";
import { loadQuery } from "react-relay";

import type { CookieBannerDetailPageQuery } from "#/__generated__/core/CookieBannerDetailPageQuery.graphql";
import type { CookieBannersPageQuery } from "#/__generated__/core/CookieBannersPageQuery.graphql";
import { PageSkeleton } from "#/components/skeletons/PageSkeleton";
import { coreEnvironment } from "#/environments";

import {
  cookieBannerNodeQuery,
} from "../pages/organizations/cookie-banners/CookieBannerDetailPage";
import {
  cookieBannersQuery,
} from "../pages/organizations/cookie-banners/CookieBannersPage";

export const cookieBannerRoutes = [
  {
    path: "cookie-banners",
    Fallback: PageSkeleton,
    loader: loaderFromQueryLoader(({ organizationId }) =>
      loadQuery<CookieBannersPageQuery>(
        coreEnvironment,
        cookieBannersQuery,
        {
          organizationId: organizationId,
        },
      ),
    ),
    Component: withQueryRef(
      lazy(
        () =>
          import(
            "#/pages/organizations/cookie-banners/CookieBannersPage"
          ),
      ),
    ),
  },
  {
    path: "cookie-banners/:cookieBannerId",
    Fallback: PageSkeleton,
    loader: loaderFromQueryLoader(({ cookieBannerId }) =>
      loadQuery<CookieBannerDetailPageQuery>(
        coreEnvironment,
        cookieBannerNodeQuery,
        {
          cookieBannerId,
        },
      ),
    ),
    Component: withQueryRef(
      lazy(
        () =>
          import(
            "#/pages/organizations/cookie-banners/CookieBannerDetailPage"
          ),
      ),
    ),
    children: [
      {
        path: "overview",
        Component: lazy(
          () =>
            import(
              "#/pages/organizations/cookie-banners/tabs/CookieBannerOverviewTab"
            ),
        ),
      },
      {
        path: "appearance",
        Component: lazy(
          () =>
            import(
              "#/pages/organizations/cookie-banners/tabs/CookieBannerAppearanceTab"
            ),
        ),
      },
      {
        path: "categories",
        Component: lazy(
          () =>
            import(
              "#/pages/organizations/cookie-banners/tabs/CookieBannerCategoriesTab"
            ),
        ),
      },
      {
        path: "consent-records",
        Component: lazy(
          () =>
            import(
              "#/pages/organizations/cookie-banners/tabs/CookieBannerConsentRecordsTab"
            ),
        ),
      },
      {
        path: "embed",
        Component: lazy(
          () =>
            import(
              "#/pages/organizations/cookie-banners/tabs/CookieBannerEmbedTab"
            ),
        ),
      },
    ],
  },
] satisfies AppRoute[];

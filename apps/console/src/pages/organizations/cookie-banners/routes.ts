// Copyright (c) 2026 Probo Inc <hello@probo.com>.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

import { lazy } from "@probo/react-lazy";
import type { AppRoute } from "@probo/routes";
import { Fragment } from "react";
import { redirect } from "react-router";

import { LinkCardSkeleton } from "#/components/skeletons/LinkCardSkeleton";
import { PageSkeleton } from "#/components/skeletons/PageSkeleton";

export const cookieBannerRoutes = [
  {
    path: "cookie-banners",
    Fallback: PageSkeleton,
    Component: lazy(() => import("#/pages/organizations/cookie-banners/overview/CookieBannersOverviewPageLoader")),
  },
  {
    path: "cookie-banners/new",
    Fallback: PageSkeleton,
    Component: lazy(() => import("#/pages/organizations/cookie-banners/NewCookieBannerPage")),
  },
  {
    path: "cookie-banners/:cookieBannerId",
    Fallback: PageSkeleton,
    Component: lazy(() => import("#/pages/organizations/cookie-banners/configuration/CookieBannerConfigLayoutLoader")),
    children: [
      {
        path: "",
        loader: () => {
          // eslint-disable-next-line @typescript-eslint/only-throw-error
          throw redirect("settings");
        },
        Component: Fragment,
      },
      {
        path: "settings",
        Fallback: LinkCardSkeleton,
        Component: lazy(() => import("#/pages/organizations/cookie-banners/configuration/settings/CookieBannerSettingsPageLoader")),
      },
      {
        path: "translations",
        Fallback: LinkCardSkeleton,
        Component: lazy(() => import("#/pages/organizations/cookie-banners/configuration/translations/CookieBannerTranslationsPageLoader")),
      },
      {
        path: "display",
        Fallback: LinkCardSkeleton,
        Component: lazy(() => import("#/pages/organizations/cookie-banners/configuration/display/CookieBannerDisplayPageLoader")),
      },
      {
        path: "consent-records",
        Fallback: LinkCardSkeleton,
        Component: lazy(() => import("#/pages/organizations/cookie-banners/configuration/consent-records/CookieBannerConsentRecordsPageLoader")),
      },
      {
        path: "trackers",
        Fallback: LinkCardSkeleton,
        Component: lazy(() => import("#/pages/organizations/cookie-banners/configuration/trackers/CookieBannerTrackersPageLoader")),
      },
      {
        path: "resources",
        Fallback: LinkCardSkeleton,
        Component: lazy(() => import("#/pages/organizations/cookie-banners/configuration/resources/CookieBannerResourcesPageLoader")),
      },
    ],
  },
  {
    path: "cookie-banners/:cookieBannerId/consent-records/:consentRecordId",
    Fallback: PageSkeleton,
    Component: lazy(() => import("#/pages/organizations/cookie-banners/configuration/consent-records/CookieBannerConsentRecordPageLoader")),
  },
  {
    path: "cookie-banners/:cookieBannerId/trackers/:trackerPatternId",
    Fallback: PageSkeleton,
    Component: lazy(() => import("#/pages/organizations/cookie-banners/configuration/trackers/TrackerPatternDetailPageLoader")),
  },
] satisfies AppRoute[];

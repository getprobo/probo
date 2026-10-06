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

import { PageSkeleton } from "#/components/skeletons/PageSkeleton";

import { CookieBannerConfigurePageSkeleton } from "./configuration/configure/CookieBannerConfigurePageSkeleton";
import { CookieBannerConsentRecordsPageSkeleton } from "./configuration/consent-records/CookieBannerConsentRecordsPageSkeleton";
import { CookieBannerConfigLayoutSkeleton } from "./configuration/CookieBannerConfigLayoutSkeleton";
import { CookieBannerInstallPageSkeleton } from "./configuration/install/CookieBannerInstallPageSkeleton";
import { CookieBannerResourcesPageSkeleton } from "./configuration/resources/CookieBannerResourcesPageSkeleton";
import { CookieBannerTCFPageSkeleton } from "./configuration/tcf/CookieBannerTCFPageSkeleton";
import { CookieBannerTrackersPageSkeleton } from "./configuration/trackers/CookieBannerTrackersPageSkeleton";
import { CookieBannerTranslationsPageSkeleton } from "./configuration/translations/CookieBannerTranslationsPageSkeleton";

export const cookieBannerRoutes = [
  {
    path: "cookie-banners/new",
    Fallback: PageSkeleton,
    Component: lazy(() => import("#/pages/organizations/cookie-banners/NewCookieBannerPage")),
  },
  {
    path: "cookie-banners/:cookieBannerId",
    Fallback: CookieBannerConfigLayoutSkeleton,
    Component: lazy(() => import("#/pages/organizations/cookie-banners/configuration/CookieBannerConfigLayoutLoader")),
    children: [
      {
        path: "install",
        Fallback: CookieBannerInstallPageSkeleton,
        Component: lazy(() => import("#/pages/organizations/cookie-banners/configuration/install/CookieBannerInstallPageLoader")),
      },
      {
        path: "configure",
        Fallback: CookieBannerConfigurePageSkeleton,
        Component: lazy(() => import("#/pages/organizations/cookie-banners/configuration/configure/CookieBannerConfigurePageLoader")),
      },
      {
        path: "translations",
        Fallback: CookieBannerTranslationsPageSkeleton,
        Component: lazy(() => import("#/pages/organizations/cookie-banners/configuration/translations/CookieBannerTranslationsPageLoader")),
      },
      {
        path: "trackers",
        Fallback: CookieBannerTrackersPageSkeleton,
        Component: lazy(() => import("#/pages/organizations/cookie-banners/configuration/trackers/CookieBannerTrackersPageLoader")),
      },
      {
        path: "resources",
        Fallback: CookieBannerResourcesPageSkeleton,
        Component: lazy(() => import("#/pages/organizations/cookie-banners/configuration/resources/CookieBannerResourcesPageLoader")),
      },
      {
        path: "trail",
        Fallback: CookieBannerConsentRecordsPageSkeleton,
        Component: lazy(() => import("#/pages/organizations/cookie-banners/configuration/consent-records/CookieBannerConsentRecordsPageLoader")),
      },
      {
        path: "tcf",
        Fallback: CookieBannerTCFPageSkeleton,
        Component: lazy(() => import("#/pages/organizations/cookie-banners/configuration/tcf/CookieBannerTCFPageLoader")),
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

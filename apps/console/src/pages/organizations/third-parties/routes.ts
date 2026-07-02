// Copyright (c) 2026 Probo Inc <hello@probo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

import { lazy } from "@probo/react-lazy";
import type { AppRoute } from "@probo/routes";
import { Fragment } from "react";
import { redirect } from "react-router";

import { LinkCardSkeleton } from "#/components/skeletons/LinkCardSkeleton";
import { PageSkeleton } from "#/components/skeletons/PageSkeleton";

export const thirdPartyRoutes = [
  {
    path: "third-parties",
    Fallback: PageSkeleton,
    Component: lazy(() => import("./ThirdPartiesPageLoader")),
  },
  {
    path: "third-parties/:thirdPartyId",
    Fallback: PageSkeleton,
    Component: lazy(() => import("./ThirdPartyDetailLayoutLoader")),
    children: [
      {
        path: "",
        loader: () => {
          // eslint-disable-next-line
          throw redirect("overview");
        },
        Component: Fragment,
      },
      {
        path: "overview",
        Fallback: LinkCardSkeleton,
        Component: lazy(() => import("./overview/ThirdPartyOverviewPageLoader")),
      },
      {
        path: "certifications",
        Fallback: LinkCardSkeleton,
        Component: lazy(() => import("./certifications/ThirdPartyCertificationsPageLoader")),
      },
      {
        path: "compliance",
        Fallback: LinkCardSkeleton,
        Component: lazy(() => import("./compliance/ThirdPartyCompliancePageLoader")),
      },
      {
        path: "risks",
        Fallback: LinkCardSkeleton,
        Component: lazy(() => import("./risks/ThirdPartyRiskAssessmentPageLoader")),
      },
      {
        path: "contacts",
        Fallback: LinkCardSkeleton,
        Component: lazy(() => import("./contacts/ThirdPartyContactsPageLoader")),
      },
      {
        path: "services",
        Fallback: LinkCardSkeleton,
        Component: lazy(() => import("./services/ThirdPartyServicesPageLoader")),
      },
      {
        path: "third-parties",
        Fallback: LinkCardSkeleton,
        Component: lazy(() => import("./third-parties/ThirdPartyThirdPartiesPageLoader")),
      },
      {
        path: "measures",
        Fallback: LinkCardSkeleton,
        Component: lazy(() => import("./measures/ThirdPartyMeasuresPageLoader")),
      },
    ],
  },
] satisfies AppRoute[];

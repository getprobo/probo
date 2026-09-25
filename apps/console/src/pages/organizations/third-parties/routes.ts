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

import { firstThirdPartySection } from "./_lib/thirdPartySections";

function redirectTo(path: string) {
  return () => {
    // eslint-disable-next-line
    throw redirect(path);
  };
}

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
        loader: redirectTo(firstThirdPartySection().path),
        Component: Fragment,
      },
      {
        path: "profile",
        Fallback: LinkCardSkeleton,
        Component: lazy(() => import("./profile/ThirdPartyProfilePageLoader")),
      },
      {
        path: "services",
        Fallback: LinkCardSkeleton,
        Component: lazy(() => import("./services/ThirdPartyServicesPageLoader")),
      },
      {
        path: "supply-chain",
        Fallback: LinkCardSkeleton,
        Component: lazy(() => import("./supply-chain/ThirdPartySupplyChainPageLoader")),
      },
      {
        path: "stakeholders",
        Fallback: LinkCardSkeleton,
        Component: lazy(() => import("./stakeholders/ThirdPartyStakeholdersPageLoader")),
      },
      {
        path: "assurance",
        Fallback: LinkCardSkeleton,
        Component: lazy(() => import("./assurance/ThirdPartyAssurancePageLoader")),
      },
      {
        path: "agreements",
        Fallback: LinkCardSkeleton,
        Component: lazy(() => import("./agreements/ThirdPartyAgreementsPageLoader")),
      },
      {
        path: "risks",
        Fallback: LinkCardSkeleton,
        Component: lazy(() => import("./risks/ThirdPartyRiskAssessmentPageLoader")),
      },
      {
        path: "measures",
        Fallback: LinkCardSkeleton,
        Component: lazy(() => import("./measures/ThirdPartyMeasuresPageLoader")),
      },
      {
        path: "overview",
        loader: redirectTo(`../${firstThirdPartySection().path}`),
        Component: Fragment,
      },
      {
        path: "certifications",
        loader: redirectTo("../assurance"),
        Component: Fragment,
      },
      {
        path: "compliance",
        loader: redirectTo("../assurance"),
        Component: Fragment,
      },
      {
        path: "contacts",
        loader: redirectTo("../stakeholders"),
        Component: Fragment,
      },
      {
        path: "third-parties",
        loader: redirectTo("../supply-chain"),
        Component: Fragment,
      },
    ],
  },
] satisfies AppRoute[];

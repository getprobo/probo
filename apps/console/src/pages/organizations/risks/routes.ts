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
import { RisksPageSkeleton } from "#/components/skeletons/RisksPageSkeleton";

export const riskRoutes = [
  {
    Fallback: PageSkeleton,
    Component: lazy(() => import("./RisksLayoutLoader")),
    children: [
      {
        path: "risks",
        Fallback: RisksPageSkeleton,
        Component: lazy(() => import("./RisksPageLoader")),
      },
      {
        path: "risk-assessments",
        Fallback: PageSkeleton,
        Component: lazy(
          () =>
            import("./risk-assessments/RiskAssessmentsPageLoader"),
        ),
      },
    ],
  },
  {
    path: "risks/:riskId",
    Fallback: PageSkeleton,
    Component: lazy(() => import("./RiskDetailLayoutLoader")),
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
        Component: lazy(() => import("./overview/RiskOverviewPageLoader")),
      },
      {
        path: "measures",
        Fallback: LinkCardSkeleton,
        Component: lazy(() => import("./measures/RiskMeasuresPageLoader")),
      },
      {
        path: "documents",
        Fallback: LinkCardSkeleton,
        Component: lazy(() => import("./documents/RiskDocumentsPageLoader")),
      },
      {
        path: "controls",
        Fallback: LinkCardSkeleton,
        Component: lazy(() => import("./controls/RiskControlsPageLoader")),
      },
      {
        path: "obligations",
        Fallback: LinkCardSkeleton,
        Component: lazy(() => import("./obligations/RiskObligationsPageLoader")),
      },
      {
        path: "scenarios",
        Fallback: LinkCardSkeleton,
        Component: lazy(() => import("./scenarios/RiskScenariosPageLoader")),
      },
    ],
  },
  {
    path: "risk-assessments/:riskAssessmentId",
    Fallback: PageSkeleton,
    Component: lazy(
      () =>
        import("./risk-assessments/RiskAssessmentDetailPageLoader"),
    ),
  },
] satisfies AppRoute[];

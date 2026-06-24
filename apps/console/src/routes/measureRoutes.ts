// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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
import {
  type AppRoute,
  loaderFromQueryLoader,
  withQueryRef,
} from "@probo/routes";
import { Fragment } from "react";
import { loadQuery } from "react-relay";
import { redirect } from "react-router";

import type { MeasureDetailPageNodeQuery } from "#/__generated__/core/MeasureDetailPageNodeQuery.graphql";
import { LinkCardSkeleton } from "#/components/skeletons/LinkCardSkeleton";
import { PageSkeleton } from "#/components/skeletons/PageSkeleton";
import { coreEnvironment } from "#/environments";
import { measureNodeQuery } from "#/pages/organizations/measures/MeasureDetailPage";

export const measureRoutes = [
  {
    path: "measures",
    Fallback: PageSkeleton,
    Component: lazy(
      () =>
        import("#/pages/organizations/measures/MeasuresPageLoader"),
    ),
  },
  {
    path: "measures/:measureId",
    Fallback: PageSkeleton,
    loader: loaderFromQueryLoader(({ measureId }) =>
      loadQuery<MeasureDetailPageNodeQuery>(coreEnvironment, measureNodeQuery, {
        measureId: measureId,
      }),
    ),
    Component: withQueryRef(
      lazy(() => import("#/pages/organizations/measures/MeasureDetailPage")),
    ),
    children: [
      {
        path: "",
        loader: () => {
          // eslint-disable-next-line
          throw redirect("evidences");
        },
        Component: Fragment,
      },
      {
        path: "risks",
        Fallback: LinkCardSkeleton,
        Component: lazy(
          () =>
            import("#/pages/organizations/measures/tabs/MeasureRisksTab"),
        ),
      },
      {
        path: "tasks",
        Fallback: LinkCardSkeleton,
        Component: lazy(
          () =>
            import("#/pages/organizations/measures/tabs/MeasureTasksTab"),
        ),
      },
      {
        path: "controls",
        Fallback: LinkCardSkeleton,
        Component: lazy(
          () =>
            import("#/pages/organizations/measures/tabs/MeasureControlsTab"),
        ),
      },
      {
        path: "evidences/:evidenceId?",
        Fallback: LinkCardSkeleton,
        Component: lazy(
          () =>
            import("#/pages/organizations/measures/tabs/MeasureEvidencesTab"),
        ),
      },
      {
        path: "documents",
        Fallback: LinkCardSkeleton,
        Component: lazy(
          () =>
            import("#/pages/organizations/measures/tabs/MeasureDocumentsTab"),
        ),
      },
      {
        path: "third-parties",
        Fallback: LinkCardSkeleton,
        Component: lazy(
          () =>
            import("#/pages/organizations/measures/third-parties/MeasureThirdPartiesPage"),
        ),
      },
    ],
  },
] satisfies AppRoute[];

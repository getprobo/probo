// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
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
import { RisksPageSkeleton } from "#/components/skeletons/RisksPageSkeleton";

export const riskRoutes = [
  {
    path: "risks",
    Fallback: RisksPageSkeleton,
    Component: lazy(() => import("./RisksPageLoader")),
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
        Component: lazy(
          () => import("./tabs/RiskOverviewTab"),
        ),
      },
      {
        path: "measures",
        Fallback: LinkCardSkeleton,
        Component: lazy(
          () => import("./tabs/RiskMeasuresTab"),
        ),
      },
      {
        path: "documents",
        Fallback: LinkCardSkeleton,
        Component: lazy(
          () => import("./tabs/RiskDocumentsTab"),
        ),
      },
      {
        path: "controls",
        Fallback: LinkCardSkeleton,
        Component: lazy(
          () => import("./tabs/RiskControlsTab"),
        ),
      },
      {
        path: "obligations",
        Fallback: LinkCardSkeleton,
        Component: lazy(
          () => import("./tabs/RiskObligationsTab"),
        ),
      },
      {
        path: "scenarios",
        Fallback: LinkCardSkeleton,
        Component: lazy(
          () => import("./scenarios/RiskScenariosPage"),
        ),
      },
    ],
  },
] satisfies AppRoute[];

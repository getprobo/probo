// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
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
import {
  type AppRoute,
  loaderFromQueryLoader,
  withQueryRef,
} from "@probo/routes";
import { loadQuery } from "react-relay";

import type { ThirdPartyGraphListQuery } from "#/__generated__/core/ThirdPartyGraphListQuery.graphql";
import type { ThirdPartyGraphNodeQuery } from "#/__generated__/core/ThirdPartyGraphNodeQuery.graphql";
import { LinkCardSkeleton } from "#/components/skeletons/LinkCardSkeleton";
import { PageSkeleton } from "#/components/skeletons/PageSkeleton";
import { coreEnvironment } from "#/environments";
import { thirdPartyNodeQuery, thirdPartiesQuery } from "#/hooks/graph/ThirdPartyGraph";

export const thirdPartyRoutes = [
  {
    path: "third-parties",
    Fallback: PageSkeleton,
    loader: loaderFromQueryLoader(({ organizationId }) =>
      loadQuery<ThirdPartyGraphListQuery>(coreEnvironment, thirdPartiesQuery, {
        organizationId: organizationId,
        snapshotId: null,
      }),
    ),
    Component: withQueryRef(
      lazy(() => import("#/pages/organizations/third-parties/ThirdPartiesPage")),
    ),
  },
  {
    path: "snapshots/:snapshotId/third-parties",
    Fallback: PageSkeleton,
    loader: loaderFromQueryLoader(({ organizationId, snapshotId }) =>
      loadQuery<ThirdPartyGraphListQuery>(coreEnvironment, thirdPartiesQuery, {
        organizationId: organizationId,
        snapshotId,
      }),
    ),
    Component: withQueryRef(
      lazy(() => import("#/pages/organizations/third-parties/ThirdPartiesPage")),
    ),
  },
  {
    path: "third-parties/:thirdPartyId",
    Fallback: PageSkeleton,
    loader: loaderFromQueryLoader(({ thirdPartyId }) =>
      loadQuery<ThirdPartyGraphNodeQuery>(coreEnvironment, thirdPartyNodeQuery, {
        thirdPartyId: thirdPartyId,
      }),
    ),
    Component: withQueryRef(
      lazy(() => import("../pages/organizations/third-parties/ThirdPartyDetailPage")),
    ),
    children: [
      {
        path: "overview",
        Fallback: LinkCardSkeleton,
        Component: lazy(
          () => import("../pages/organizations/third-parties/tabs/ThirdPartyOverviewTab"),
        ),
      },
      {
        path: "certifications",
        Fallback: LinkCardSkeleton,
        Component: lazy(
          () =>
            import("../pages/organizations/third-parties/tabs/ThirdPartyCertificationsTab"),
        ),
      },
      {
        path: "compliance",
        Fallback: LinkCardSkeleton,
        Component: lazy(
          () =>
            import("../pages/organizations/third-parties/tabs/ThirdPartyComplianceTab"),
        ),
      },
      {
        path: "risks",
        Fallback: LinkCardSkeleton,
        Component: lazy(
          () =>
            import("../pages/organizations/third-parties/tabs/ThirdPartyRiskAssessmentTab"),
        ),
      },
      {
        path: "contacts",
        Fallback: LinkCardSkeleton,
        Component: lazy(
          () => import("../pages/organizations/third-parties/tabs/ThirdPartyContactsTab"),
        ),
      },
      {
        path: "services",
        Fallback: LinkCardSkeleton,
        Component: lazy(
          () => import("../pages/organizations/third-parties/tabs/ThirdPartyServicesTab"),
        ),
      },
    ],
  },
  {
    path: "snapshots/:snapshotId/third-parties/:thirdPartyId",
    Fallback: PageSkeleton,
    loader: loaderFromQueryLoader(({ thirdPartyId }) =>
      loadQuery<ThirdPartyGraphNodeQuery>(coreEnvironment, thirdPartyNodeQuery, {
        thirdPartyId: thirdPartyId,
      }),
    ),
    Component: withQueryRef(
      lazy(() => import("../pages/organizations/third-parties/ThirdPartyDetailPage")),
    ),
    children: [
      {
        path: "overview",
        Fallback: LinkCardSkeleton,
        Component: lazy(
          () => import("../pages/organizations/third-parties/tabs/ThirdPartyOverviewTab"),
        ),
      },
      {
        path: "certifications",
        Fallback: LinkCardSkeleton,
        Component: lazy(
          () =>
            import("../pages/organizations/third-parties/tabs/ThirdPartyCertificationsTab"),
        ),
      },
      {
        path: "compliance",
        Fallback: LinkCardSkeleton,
        Component: lazy(
          () =>
            import("../pages/organizations/third-parties/tabs/ThirdPartyComplianceTab"),
        ),
      },
      {
        path: "risks",
        Fallback: LinkCardSkeleton,
        Component: lazy(
          () =>
            import("../pages/organizations/third-parties/tabs/ThirdPartyRiskAssessmentTab"),
        ),
      },
      {
        path: "contacts",
        Fallback: LinkCardSkeleton,
        Component: lazy(
          () => import("../pages/organizations/third-parties/tabs/ThirdPartyContactsTab"),
        ),
      },
      {
        path: "services",
        Fallback: LinkCardSkeleton,
        Component: lazy(
          () => import("../pages/organizations/third-parties/tabs/ThirdPartyServicesTab"),
        ),
      },
    ],
  },
] satisfies AppRoute[];

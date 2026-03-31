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

import type { VendorGraphListQuery } from "#/__generated__/core/VendorGraphListQuery.graphql";
import type { VendorGraphNodeQuery } from "#/__generated__/core/VendorGraphNodeQuery.graphql";
import { LinkCardSkeleton } from "#/components/skeletons/LinkCardSkeleton";
import { PageSkeleton } from "#/components/skeletons/PageSkeleton";
import { coreEnvironment } from "#/environments";
import { vendorNodeQuery, vendorsQuery } from "#/hooks/graph/VendorGraph";

export const vendorRoutes = [
  {
    path: "vendors",
    Fallback: PageSkeleton,
    loader: loaderFromQueryLoader(({ organizationId }) =>
      loadQuery<VendorGraphListQuery>(coreEnvironment, vendorsQuery, {
        organizationId: organizationId,
        snapshotId: null,
      }),
    ),
    Component: withQueryRef(
      lazy(() => import("#/pages/organizations/vendors/VendorsPage")),
    ),
  },
  {
    path: "snapshots/:snapshotId/vendors",
    Fallback: PageSkeleton,
    loader: loaderFromQueryLoader(({ organizationId, snapshotId }) =>
      loadQuery<VendorGraphListQuery>(coreEnvironment, vendorsQuery, {
        organizationId: organizationId,
        snapshotId,
      }),
    ),
    Component: withQueryRef(
      lazy(() => import("#/pages/organizations/vendors/VendorsPage")),
    ),
  },
  {
    path: "vendors/:vendorId",
    Fallback: PageSkeleton,
    loader: loaderFromQueryLoader(({ vendorId }) =>
      loadQuery<VendorGraphNodeQuery>(coreEnvironment, vendorNodeQuery, {
        vendorId: vendorId,
      }),
    ),
    Component: withQueryRef(
      lazy(() => import("../pages/organizations/vendors/VendorDetailPage")),
    ),
    children: [
      {
        path: "overview",
        Fallback: LinkCardSkeleton,
        Component: lazy(
          () => import("../pages/organizations/vendors/tabs/VendorOverviewTab"),
        ),
      },
      {
        path: "certifications",
        Fallback: LinkCardSkeleton,
        Component: lazy(
          () =>
            import("../pages/organizations/vendors/tabs/VendorCertificationsTab"),
        ),
      },
      {
        path: "compliance",
        Fallback: LinkCardSkeleton,
        Component: lazy(
          () =>
            import("../pages/organizations/vendors/tabs/VendorComplianceTab"),
        ),
      },
      {
        path: "risks",
        Fallback: LinkCardSkeleton,
        Component: lazy(
          () =>
            import("../pages/organizations/vendors/tabs/VendorRiskAssessmentTab"),
        ),
      },
      {
        path: "contacts",
        Fallback: LinkCardSkeleton,
        Component: lazy(
          () => import("../pages/organizations/vendors/tabs/VendorContactsTab"),
        ),
      },
      {
        path: "services",
        Fallback: LinkCardSkeleton,
        Component: lazy(
          () => import("../pages/organizations/vendors/tabs/VendorServicesTab"),
        ),
      },
    ],
  },
  {
    path: "snapshots/:snapshotId/vendors/:vendorId",
    Fallback: PageSkeleton,
    loader: loaderFromQueryLoader(({ vendorId }) =>
      loadQuery<VendorGraphNodeQuery>(coreEnvironment, vendorNodeQuery, {
        vendorId: vendorId,
      }),
    ),
    Component: withQueryRef(
      lazy(() => import("../pages/organizations/vendors/VendorDetailPage")),
    ),
    children: [
      {
        path: "overview",
        Fallback: LinkCardSkeleton,
        Component: lazy(
          () => import("../pages/organizations/vendors/tabs/VendorOverviewTab"),
        ),
      },
      {
        path: "certifications",
        Fallback: LinkCardSkeleton,
        Component: lazy(
          () =>
            import("../pages/organizations/vendors/tabs/VendorCertificationsTab"),
        ),
      },
      {
        path: "compliance",
        Fallback: LinkCardSkeleton,
        Component: lazy(
          () =>
            import("../pages/organizations/vendors/tabs/VendorComplianceTab"),
        ),
      },
      {
        path: "risks",
        Fallback: LinkCardSkeleton,
        Component: lazy(
          () =>
            import("../pages/organizations/vendors/tabs/VendorRiskAssessmentTab"),
        ),
      },
      {
        path: "contacts",
        Fallback: LinkCardSkeleton,
        Component: lazy(
          () => import("../pages/organizations/vendors/tabs/VendorContactsTab"),
        ),
      },
      {
        path: "services",
        Fallback: LinkCardSkeleton,
        Component: lazy(
          () => import("../pages/organizations/vendors/tabs/VendorServicesTab"),
        ),
      },
    ],
  },
] satisfies AppRoute[];

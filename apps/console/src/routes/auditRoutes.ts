import { lazy } from "@probo/react-lazy";
import {
  type AppRoute,
  loaderFromQueryLoader,
  withQueryRef,
} from "@probo/routes";
import { loadQuery } from "react-relay";

import type { AuditGraphListQuery } from "#/__generated__/core/AuditGraphListQuery.graphql";
import type { AuditGraphNodeQuery } from "#/__generated__/core/AuditGraphNodeQuery.graphql";
import { PageSkeleton } from "#/components/skeletons/PageSkeleton";
import { coreEnvironment } from "#/environments";

import { auditNodeQuery, auditsQuery } from "../hooks/graph/AuditGraph";

export const auditRoutes = [
  {
    path: "audits",
    Fallback: PageSkeleton,
    loader: loaderFromQueryLoader(({ organizationId }) =>
      loadQuery<AuditGraphListQuery>(coreEnvironment, auditsQuery, {
        organizationId,
      }),
    ),
    Component: withQueryRef(
      lazy(() => import("#/pages/organizations/audits/AuditsPage")),
    ),
  },
  {
    path: "audits/:auditId",
    Fallback: PageSkeleton,
    loader: loaderFromQueryLoader(({ auditId }) =>
      loadQuery<AuditGraphNodeQuery>(coreEnvironment, auditNodeQuery, {
        auditId,
      }),
    ),
    Component: withQueryRef(
      lazy(() => import("#/pages/organizations/audits/AuditDetailsPage")),
    ),
  },
] satisfies AppRoute[];

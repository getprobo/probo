import { loadQuery } from "react-relay";
import { coreEnvironment } from "/environments";
import { PageSkeleton } from "/components/skeletons/PageSkeleton";
import { lazy } from "@probo/react-lazy";
import { auditsQuery, auditNodeQuery } from "../hooks/graph/AuditGraph";
import type { AuditGraphListQuery } from "/hooks/graph/__generated__/AuditGraphListQuery.graphql";
import type { AuditGraphNodeQuery } from "/hooks/graph/__generated__/AuditGraphNodeQuery.graphql";
import {
  loaderFromQueryLoader,
  withQueryRef,
  type AppRoute,
} from "@probo/routes";

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
      lazy(() => import("/pages/organizations/audits/AuditsPage")),
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
      lazy(() => import("/pages/organizations/audits/AuditDetailsPage")),
    ),
  },
] satisfies AppRoute[];

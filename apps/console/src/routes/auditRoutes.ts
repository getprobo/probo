import { loadQuery } from "react-relay";
import { relayEnvironment } from "/providers/RelayProviders";
import { PageSkeleton } from "/components/skeletons/PageSkeleton";
import { lazy } from "@probo/react-lazy";
import { auditsQuery, auditNodeQuery } from "../hooks/graph/AuditGraph";

export const auditRoutes = [
  {
    path: "audits",
    fallback: PageSkeleton,
    queryLoader: (params: Record<string, string>) =>
      loadQuery(relayEnvironment, auditsQuery, { organizationId: params.organizationId }),
    Component: lazy(
      () => import("/pages/organizations/audits/AuditsPage")
    ),
  },
  {
    path: "audits/:auditId",
    fallback: PageSkeleton,
    queryLoader: (params: Record<string, string>) =>
      loadQuery(relayEnvironment, auditNodeQuery, { auditId: params.auditId }),
    Component: lazy(
      () => import("/pages/organizations/audits/AuditDetailsPage")
    ),
  },
];

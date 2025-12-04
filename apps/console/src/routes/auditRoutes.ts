import { loadQuery } from "react-relay";
import { relayEnvironment } from "/providers/RelayProviders";
import { PageSkeleton } from "/components/skeletons/PageSkeleton";
import { lazy } from "@probo/react-lazy";
import { auditsQuery, auditNodeQuery } from "../hooks/graph/AuditGraph";
import type { AuditGraphListQuery } from "/hooks/graph/__generated__/AuditGraphListQuery.graphql";
import type { AuditGraphNodeQuery } from "/hooks/graph/__generated__/AuditGraphNodeQuery.graphql";
import { loaderFromQueryLoader, withQueryRef } from "/routes";

export const auditRoutes = [
  {
    path: "audits",
    fallback: PageSkeleton,
    loader: loaderFromQueryLoader(({ organizationId }) =>
      loadQuery<AuditGraphListQuery>(relayEnvironment, auditsQuery, { organizationId }),
    ),
    Component: withQueryRef(lazy(
      () => import("/pages/organizations/audits/AuditsPage")
    )),
  },
  {
    path: "audits/:auditId",
    fallback: PageSkeleton,
    loader: loaderFromQueryLoader(({ auditId }) =>
      loadQuery<AuditGraphNodeQuery>(relayEnvironment, auditNodeQuery, { auditId }),
    ),
    Component: withQueryRef(lazy(
      () => import("/pages/organizations/audits/AuditDetailsPage")
    )),
  },
];

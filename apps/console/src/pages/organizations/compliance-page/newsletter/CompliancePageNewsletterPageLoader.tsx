import { useEffect } from "react";
import { useQueryLoader } from "react-relay";

import type { CompliancePageNewsletterPageQuery } from "#/__generated__/core/CompliancePageNewsletterPageQuery.graphql";
import { LinkCardSkeleton } from "#/components/skeletons/LinkCardSkeleton";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { CoreRelayProvider } from "#/providers/CoreRelayProvider";

import {
  CompliancePageNewsletterPage,
  compliancePageNewsletterPageQuery,
} from "./CompliancePageNewsletterPage";

function CompliancePageNewsletterPageQueryLoader() {
  const organizationId = useOrganizationId();
  const [queryRef, loadQuery] = useQueryLoader<CompliancePageNewsletterPageQuery>(
    compliancePageNewsletterPageQuery,
  );

  useEffect(() => {
    loadQuery({ organizationId });
  }, [loadQuery, organizationId]);

  if (!queryRef) return <LinkCardSkeleton />;

  return <CompliancePageNewsletterPage queryRef={queryRef} />;
}

export default function CompliancePageNewsletterPageLoader() {
  return (
    <CoreRelayProvider>
      <CompliancePageNewsletterPageQueryLoader />
    </CoreRelayProvider>
  );
}

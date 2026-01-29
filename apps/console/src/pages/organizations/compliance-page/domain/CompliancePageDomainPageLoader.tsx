import { useEffect } from "react";
import { useQueryLoader } from "react-relay";

import type { CompliancePageDomainPageQuery } from "#/__generated__/core/CompliancePageDomainPageQuery.graphql";
import { LinkCardSkeleton } from "#/components/skeletons/LinkCardSkeleton";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { CoreRelayProvider } from "#/providers/CoreRelayProvider";

import {
  CompliancePageDomainPage,
  compliancePageDomainPageQuery,
} from "./CompliancePageDomainPage";

function CompliancePageDomainPageQueryLoader() {
  const organizationId = useOrganizationId();
  const [queryRef, loadQuery] = useQueryLoader<CompliancePageDomainPageQuery>(
    compliancePageDomainPageQuery,
  );

  useEffect(() => {
    loadQuery({
      organizationId,
    });
  }, [loadQuery, organizationId]);

  if (!queryRef) {
    return <LinkCardSkeleton />;
  }

  return <CompliancePageDomainPage queryRef={queryRef} />;
}

export default function CompliancePageDomainPageLoader() {
  return (
    <CoreRelayProvider>
      <CompliancePageDomainPageQueryLoader />
    </CoreRelayProvider>
  );
}

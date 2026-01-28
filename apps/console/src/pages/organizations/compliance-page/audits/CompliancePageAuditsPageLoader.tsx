import { useEffect } from "react";
import { useQueryLoader } from "react-relay";

import type { CompliancePageAuditsPageQuery } from "#/__generated__/core/CompliancePageAuditsPageQuery.graphql";
import { LinkCardSkeleton } from "#/components/skeletons/LinkCardSkeleton";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { CoreRelayProvider } from "#/providers/CoreRelayProvider";

import { CompliancePageAuditsPage, compliancePageAuditsPageQuery } from "./CompliancePageAuditsPage";

function CompliancePageAuditsPageQueryLoader() {
  const organizationId = useOrganizationId();
  const [queryRef, loadQuery] = useQueryLoader<CompliancePageAuditsPageQuery>(compliancePageAuditsPageQuery);

  useEffect(() => {
    if (!queryRef) {
      loadQuery({ organizationId });
    }
  });

  if (!queryRef) return <LinkCardSkeleton />;

  return <CompliancePageAuditsPage queryRef={queryRef} />;
}

export default function CompliancePageAuditsPageLoader() {
  return (
    <CoreRelayProvider>
      <CompliancePageAuditsPageQueryLoader />
    </CoreRelayProvider>
  );
}

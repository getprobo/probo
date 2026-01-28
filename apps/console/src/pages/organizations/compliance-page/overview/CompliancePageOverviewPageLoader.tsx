import { useEffect } from "react";
import { useQueryLoader } from "react-relay";

import type { CompliancePageOverviewPageQuery } from "#/__generated__/core/CompliancePageOverviewPageQuery.graphql";
import { LinkCardSkeleton } from "#/components/skeletons/LinkCardSkeleton";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { CoreRelayProvider } from "#/providers/CoreRelayProvider";

import { CompliancePageOverviewPage, compliancePageOverviewPageQuery } from "./CompliancePageOverviewPage";

function CompliancePageOverviewPageQueryLoader() {
  const organizationId = useOrganizationId();
  const [queryRef, loadQuery] = useQueryLoader<CompliancePageOverviewPageQuery>(compliancePageOverviewPageQuery);

  useEffect(() => {
    if (!queryRef) {
      loadQuery({ organizationId });
    }
  });

  if (!queryRef) return <LinkCardSkeleton />;

  return <CompliancePageOverviewPage queryRef={queryRef} />;
}

export default function CompliancePageOverviewPageLoader() {
  return (
    <CoreRelayProvider>
      <CompliancePageOverviewPageQueryLoader />
    </CoreRelayProvider>
  );
}

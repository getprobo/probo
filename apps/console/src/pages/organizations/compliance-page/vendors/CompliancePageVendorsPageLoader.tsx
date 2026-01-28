import { useEffect } from "react";
import { useQueryLoader } from "react-relay";

import type { CompliancePageVendorsPageQuery } from "#/__generated__/core/CompliancePageVendorsPageQuery.graphql";
import { LinkCardSkeleton } from "#/components/skeletons/LinkCardSkeleton";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { CoreRelayProvider } from "#/providers/CoreRelayProvider";

import { CompliancePageVendorsPage, compliancePageVendorsPageQuery } from "./CompliancePageVendorsPage";

function CompliancePageVendorsPageQueryLoader() {
  const organizationId = useOrganizationId();
  const [queryRef, loadQuery] = useQueryLoader<CompliancePageVendorsPageQuery>(compliancePageVendorsPageQuery);

  useEffect(() => {
    if (!queryRef) {
      loadQuery({ organizationId });
    }
  });

  if (!queryRef) return <LinkCardSkeleton />;

  return <CompliancePageVendorsPage queryRef={queryRef} />;
}

export default function CompliancePageVendorsPageLoader() {
  return (
    <CoreRelayProvider>
      <CompliancePageVendorsPageQueryLoader />
    </CoreRelayProvider>
  );
}

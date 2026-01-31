import { useEffect } from "react";
import { useQueryLoader } from "react-relay";

import type { CompliancePageBrandPageQuery } from "#/__generated__/core/CompliancePageBrandPageQuery.graphql";
import { LinkCardSkeleton } from "#/components/skeletons/LinkCardSkeleton";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { CoreRelayProvider } from "#/providers/CoreRelayProvider";

import { CompliancePageBrandPage, compliancePageBrandPageQuery } from "./CompliancePageBrandPage";

function CompliancePageBrandPageQueryLoader() {
  const organizationId = useOrganizationId();
  const [queryRef, loadQuery] = useQueryLoader<CompliancePageBrandPageQuery>(compliancePageBrandPageQuery);

  useEffect(() => {
    if (!queryRef) {
      loadQuery({ organizationId });
    }
  });

  if (!queryRef) return <LinkCardSkeleton />;

  return <CompliancePageBrandPage queryRef={queryRef} />;
}

export default function CompliancePageBrandPageLoader() {
  return (
    <CoreRelayProvider>
      <CompliancePageBrandPageQueryLoader />
    </CoreRelayProvider>
  );
}

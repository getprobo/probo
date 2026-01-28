import { useEffect } from "react";
import { useQueryLoader } from "react-relay";

import type { CompliancePageAccessPageQuery } from "#/__generated__/core/CompliancePageAccessPageQuery.graphql";
import { LinkCardSkeleton } from "#/components/skeletons/LinkCardSkeleton";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { CoreRelayProvider } from "#/providers/CoreRelayProvider";

import { CompliancePageAccessPage, compliancePageAccessPageQuery } from "./CompliancePageAccessPage";

function CompliancePageAccessPageQueryLoader() {
  const organizationId = useOrganizationId();
  const [queryRef, loadQuery] = useQueryLoader<CompliancePageAccessPageQuery>(compliancePageAccessPageQuery);

  useEffect(() => {
    if (!queryRef) {
      loadQuery({ organizationId });
    }
  });

  if (!queryRef) return <LinkCardSkeleton />;

  return <CompliancePageAccessPage queryRef={queryRef} />;
}

export default function CompliancePageAccessPageLoader() {
  return (
    <CoreRelayProvider>
      <CompliancePageAccessPageQueryLoader />
    </CoreRelayProvider>
  );
}

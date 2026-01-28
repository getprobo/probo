import { useEffect } from "react";
import { useQueryLoader } from "react-relay";

import type { CompliancePageLayoutQuery } from "#/__generated__/core/CompliancePageLayoutQuery.graphql";
import { PageSkeleton } from "#/components/skeletons/PageSkeleton";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { CoreRelayProvider } from "#/providers/CoreRelayProvider";

import { CompliancePageLayout, compliancePageLayoutQuery } from "./CompliancePageLayout";

function CompliancePageLayoutQueryLoader() {
  const organizationId = useOrganizationId();
  const [queryRef, loadQuery] = useQueryLoader<CompliancePageLayoutQuery>(compliancePageLayoutQuery);

  useEffect(() => {
    if (!queryRef) {
      loadQuery({ organizationId });
    }
  });

  if (!queryRef) return <PageSkeleton />;

  return <CompliancePageLayout queryRef={queryRef} />;
}

export default function CompliancePageLayoutLoader() {
  return (
    <CoreRelayProvider>
      <CompliancePageLayoutQueryLoader />
    </CoreRelayProvider>
  );
}

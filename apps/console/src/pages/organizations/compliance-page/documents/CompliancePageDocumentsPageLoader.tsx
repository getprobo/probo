import { useEffect } from "react";
import { useQueryLoader } from "react-relay";

import type { CompliancePageDocumentsPageQuery } from "#/__generated__/core/CompliancePageDocumentsPageQuery.graphql";
import { LinkCardSkeleton } from "#/components/skeletons/LinkCardSkeleton";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { CoreRelayProvider } from "#/providers/CoreRelayProvider";

import { CompliancePageDocumentsPage, compliancePageDocumentsPageQuery } from "./CompliancePageDocumentsPage";

function CompliancePageDocumentsPageQueryLoader() {
  const organizationId = useOrganizationId();
  const [queryRef, loadQuery] = useQueryLoader<CompliancePageDocumentsPageQuery>(compliancePageDocumentsPageQuery);

  useEffect(() => {
    if (!queryRef) {
      loadQuery({ organizationId });
    }
  });

  if (!queryRef) return <LinkCardSkeleton />;

  return <CompliancePageDocumentsPage queryRef={queryRef} />;
}

export default function CompliancePageDocumentsPageLoader() {
  return (
    <CoreRelayProvider>
      <CompliancePageDocumentsPageQueryLoader />
    </CoreRelayProvider>
  );
}

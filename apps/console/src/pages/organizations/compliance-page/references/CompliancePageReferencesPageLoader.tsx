import { useEffect } from "react";
import { useQueryLoader } from "react-relay";

import type { CompliancePageReferencesPageQuery } from "#/__generated__/core/CompliancePageReferencesPageQuery.graphql";
import { LinkCardSkeleton } from "#/components/skeletons/LinkCardSkeleton";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { CoreRelayProvider } from "#/providers/CoreRelayProvider";

import { CompliancePageReferencesPage, compliancePageReferencesPageQuery } from "./CompliancePageReferencesPage";

function CompliancePageReferencesPageQueryLoader() {
  const organizationId = useOrganizationId();
  const [queryRef, loadQuery] = useQueryLoader<CompliancePageReferencesPageQuery>(compliancePageReferencesPageQuery);

  useEffect(() => {
    if (!queryRef) {
      loadQuery({ organizationId });
    }
  });

  if (!queryRef) return <LinkCardSkeleton />;

  return <CompliancePageReferencesPage queryRef={queryRef} />;
}

export default function CompliancePageReferencesPageLoader() {
  return (
    <CoreRelayProvider>
      <CompliancePageReferencesPageQueryLoader />
    </CoreRelayProvider>
  );
}

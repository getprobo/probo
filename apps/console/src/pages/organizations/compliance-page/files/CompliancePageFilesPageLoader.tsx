import { useEffect } from "react";
import { useQueryLoader } from "react-relay";

import type { CompliancePageFilesPageQuery } from "#/__generated__/core/CompliancePageFilesPageQuery.graphql";
import { LinkCardSkeleton } from "#/components/skeletons/LinkCardSkeleton";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { CoreRelayProvider } from "#/providers/CoreRelayProvider";

import { CompliancePageFilesPage, compliancePageFilesPageQuery } from "./CompliancePageFilesPage";

function CompliancePageFilesPageQueryLoader() {
  const organizationId = useOrganizationId();
  const [queryRef, loadQuery] = useQueryLoader<CompliancePageFilesPageQuery>(compliancePageFilesPageQuery);

  useEffect(() => {
    if (!queryRef) {
      loadQuery({ organizationId });
    }
  });

  if (!queryRef) return <LinkCardSkeleton />;

  return <CompliancePageFilesPage queryRef={queryRef} />;
}

export default function CompliancePageFilesPageLoader() {
  return (
    <CoreRelayProvider>
      <CompliancePageFilesPageQueryLoader />
    </CoreRelayProvider>
  );
}

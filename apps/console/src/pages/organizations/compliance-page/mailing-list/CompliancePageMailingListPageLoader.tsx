import { useEffect } from "react";
import { useQueryLoader } from "react-relay";

import type { CompliancePageMailingListPageQuery } from "#/__generated__/core/CompliancePageMailingListPageQuery.graphql";
import { LinkCardSkeleton } from "#/components/skeletons/LinkCardSkeleton";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { CoreRelayProvider } from "#/providers/CoreRelayProvider";

import {
  CompliancePageMailingListPage,
  compliancePageMailingListPageQuery,
} from "./CompliancePageMailingListPage";

function CompliancePageMailingListPageQueryLoader() {
  const organizationId = useOrganizationId();
  const [queryRef, loadQuery] = useQueryLoader<CompliancePageMailingListPageQuery>(
    compliancePageMailingListPageQuery,
  );

  useEffect(() => {
    if (!queryRef) {
      loadQuery({ organizationId });
    }
  });

  if (!queryRef) return <LinkCardSkeleton />;

  return <CompliancePageMailingListPage queryRef={queryRef} />;
}

export default function CompliancePageMailingListPageLoader() {
  return (
    <CoreRelayProvider>
      <CompliancePageMailingListPageQueryLoader />
    </CoreRelayProvider>
  );
}

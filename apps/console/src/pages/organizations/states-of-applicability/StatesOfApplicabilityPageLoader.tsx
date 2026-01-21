import { Suspense, useEffect } from "react";
import { useQueryLoader } from "react-relay";

import type { StatesOfApplicabilityPageQuery } from "#/__generated__/core/StatesOfApplicabilityPageQuery.graphql";
import { PageSkeleton } from "#/components/skeletons/PageSkeleton";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import StatesOfApplicabilityPage, { statesOfApplicabilityPageQuery } from "./StatesOfApplicabilityPage";

export default function StatesOfApplicabilityPageLoader() {
  const organizationId = useOrganizationId();
  const [queryRef, loadQuery]
    = useQueryLoader<StatesOfApplicabilityPageQuery>(statesOfApplicabilityPageQuery);

  useEffect(() => {
    loadQuery({ organizationId });
  }, [loadQuery, organizationId]);

  if (!queryRef) {
    return <PageSkeleton />;
  }

  return (
    <Suspense fallback={<PageSkeleton />}>
      <StatesOfApplicabilityPage queryRef={queryRef} />
    </Suspense>
  );
}

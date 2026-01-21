import { Suspense, useEffect } from "react";
import { useQueryLoader } from "react-relay";
import { useParams } from "react-router";

import type { StateOfApplicabilityDetailPageQuery } from "#/__generated__/core/StateOfApplicabilityDetailPageQuery.graphql";
import { PageSkeleton } from "#/components/skeletons/PageSkeleton";

import StateOfApplicabilityDetailPage, { stateOfApplicabilityDetailPageQuery } from "./StateOfApplicabilityDetailPage";

export default function StateOfApplicabilityDetailPageLoader() {
  const { stateOfApplicabilityId } = useParams<{ stateOfApplicabilityId: string }>();
  const [queryRef, loadQuery]
    = useQueryLoader<StateOfApplicabilityDetailPageQuery>(stateOfApplicabilityDetailPageQuery);

  useEffect(() => {
    if (stateOfApplicabilityId) {
      loadQuery({ stateOfApplicabilityId });
    }
  }, [loadQuery, stateOfApplicabilityId]);

  if (!queryRef) {
    return <PageSkeleton />;
  }

  return (
    <Suspense fallback={<PageSkeleton />}>
      <StateOfApplicabilityDetailPage queryRef={queryRef} />
    </Suspense>
  );
}

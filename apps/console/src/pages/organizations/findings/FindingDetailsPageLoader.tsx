import { Suspense, useEffect } from "react";
import { useQueryLoader } from "react-relay";
import { useParams } from "react-router";

import type { FindingDetailsPageQuery } from "#/__generated__/core/FindingDetailsPageQuery.graphql";
import { PageSkeleton } from "#/components/skeletons/PageSkeleton";

import FindingDetailsPage, { findingDetailsPageQuery } from "./FindingDetailsPage";

export default function FindingDetailsPageLoader() {
  const { findingId } = useParams<{ findingId: string }>();
  const [queryRef, loadQuery]
    = useQueryLoader<FindingDetailsPageQuery>(findingDetailsPageQuery);

  useEffect(() => {
    if (findingId) {
      loadQuery({ findingId });
    }
  }, [loadQuery, findingId]);

  if (!queryRef) {
    return <PageSkeleton />;
  }

  return (
    <Suspense fallback={<PageSkeleton />}>
      <FindingDetailsPage queryRef={queryRef} />
    </Suspense>
  );
}

import { Suspense, useEffect } from "react";
import { useQueryLoader } from "react-relay";

import type { AccessReviewPageQuery } from "#/__generated__/core/AccessReviewPageQuery.graphql";
import { PageSkeleton } from "#/components/skeletons/PageSkeleton";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import AccessReviewPage, { accessReviewPageQuery } from "./AccessReviewPage";

export default function AccessReviewPageLoader() {
  const organizationId = useOrganizationId();
  const [queryRef, loadQuery]
    = useQueryLoader<AccessReviewPageQuery>(accessReviewPageQuery);

  useEffect(() => {
    loadQuery({ organizationId });
  }, [loadQuery, organizationId]);

  if (!queryRef) {
    return <PageSkeleton />;
  }

  return (
    <Suspense fallback={<PageSkeleton />}>
      <AccessReviewPage queryRef={queryRef} />
    </Suspense>
  );
}

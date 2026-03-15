import { Suspense, useEffect } from "react";
import { useQueryLoader } from "react-relay";

import type { AccessReviewPageQuery } from "#/__generated__/core/AccessReviewPageQuery.graphql";
import { PageSkeleton } from "#/components/skeletons/PageSkeleton";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import { accessReviewPageQuery } from "./AccessReviewPage";
import CreateCsvAccessSourcePage from "./CreateCsvAccessSourcePage";

export default function CreateCsvAccessSourcePageLoader() {
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
      <CreateCsvAccessSourcePage queryRef={queryRef} />
    </Suspense>
  );
}

import { Suspense, useEffect } from "react";
import { useQueryLoader } from "react-relay";

import type { MeasuresPageListQuery } from "#/__generated__/core/MeasuresPageListQuery.graphql";
import { PageSkeleton } from "#/components/skeletons/PageSkeleton";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import MeasuresPage, { measuresPageQuery } from "./MeasuresPage";

export default function MeasuresPageLoader() {
  const organizationId = useOrganizationId();
  const [queryRef, loadQuery]
    = useQueryLoader<MeasuresPageListQuery>(measuresPageQuery);

  useEffect(() => {
    loadQuery({ organizationId });
  }, [loadQuery, organizationId]);

  if (!queryRef) {
    return <PageSkeleton />;
  }

  return (
    <Suspense fallback={<PageSkeleton />}>
      <MeasuresPage queryRef={queryRef} />
    </Suspense>
  );
}

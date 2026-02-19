import { Suspense, useEffect } from "react";
import { useQueryLoader } from "react-relay";

import type { ReportsPageQuery } from "#/__generated__/core/ReportsPageQuery.graphql";
import { PageSkeleton } from "#/components/skeletons/PageSkeleton";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import ReportsPage, { reportsPageQuery } from "./ReportsPage";

export default function ReportsPageLoader() {
  const organizationId = useOrganizationId();
  const [queryRef, loadQuery] = useQueryLoader<ReportsPageQuery>(reportsPageQuery);

  useEffect(() => {
    loadQuery({ organizationId });
  }, [loadQuery, organizationId]);

  if (!queryRef) {
    return <PageSkeleton />;
  }

  return (
    <Suspense fallback={<PageSkeleton />}>
      <ReportsPage queryRef={queryRef} />
    </Suspense>
  );
}

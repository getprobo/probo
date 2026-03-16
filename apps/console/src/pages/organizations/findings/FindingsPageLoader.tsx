import { Suspense, useEffect } from "react";
import { useQueryLoader } from "react-relay";
import { useParams } from "react-router";

import type { FindingsPageListQuery } from "#/__generated__/core/FindingsPageListQuery.graphql";
import { PageSkeleton } from "#/components/skeletons/PageSkeleton";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import FindingsPage, { findingsPageQuery } from "./FindingsPage";

export default function FindingsPageLoader() {
  const organizationId = useOrganizationId();
  const { snapshotId } = useParams<{ snapshotId?: string }>();
  const [queryRef, loadQuery]
    = useQueryLoader<FindingsPageListQuery>(findingsPageQuery);

  useEffect(() => {
    loadQuery({
      organizationId,
      snapshotId: snapshotId ?? null,
    });
  }, [loadQuery, organizationId, snapshotId]);

  if (!queryRef) {
    return <PageSkeleton />;
  }

  return (
    <Suspense fallback={<PageSkeleton />}>
      <FindingsPage queryRef={queryRef} />
    </Suspense>
  );
}

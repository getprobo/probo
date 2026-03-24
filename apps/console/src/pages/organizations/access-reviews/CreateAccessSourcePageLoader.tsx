import { Suspense, useEffect } from "react";
import { useQueryLoader } from "react-relay";

import type { CreateAccessSourcePageQuery } from "#/__generated__/core/CreateAccessSourcePageQuery.graphql";
import { PageSkeleton } from "#/components/skeletons/PageSkeleton";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import CreateAccessSourcePage, { createAccessSourcePageQuery } from "./CreateAccessSourcePage";

export default function CreateAccessSourcePageLoader() {
  const organizationId = useOrganizationId();
  const [queryRef, loadQuery]
    = useQueryLoader<CreateAccessSourcePageQuery>(createAccessSourcePageQuery);

  useEffect(() => {
    loadQuery({ organizationId });
  }, [loadQuery, organizationId]);

  if (!queryRef) {
    return <PageSkeleton />;
  }

  return (
    <Suspense fallback={<PageSkeleton />}>
      <CreateAccessSourcePage queryRef={queryRef} />
    </Suspense>
  );
}

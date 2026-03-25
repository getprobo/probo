import { useEffect } from "react";
import { useQueryLoader } from "react-relay";

import type { AccessReviewLayoutQuery } from "#/__generated__/core/AccessReviewLayoutQuery.graphql";
import { PageSkeleton } from "#/components/skeletons/PageSkeleton";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import AccessReviewLayout, { accessReviewLayoutQuery } from "./AccessReviewLayout";

export default function AccessReviewLayoutLoader() {
  const organizationId = useOrganizationId();
  const [queryRef, loadQuery] = useQueryLoader<AccessReviewLayoutQuery>(accessReviewLayoutQuery);

  useEffect(() => {
    if (!queryRef) {
      loadQuery({ organizationId });
    }
  }, [loadQuery, organizationId]);

  if (!queryRef) return <PageSkeleton />;

  return <AccessReviewLayout queryRef={queryRef} />;
}

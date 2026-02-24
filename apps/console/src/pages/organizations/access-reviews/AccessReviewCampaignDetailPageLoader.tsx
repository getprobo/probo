import { Suspense, useEffect } from "react";
import { useQueryLoader } from "react-relay";
import { useParams } from "react-router";

import type { AccessReviewCampaignDetailPageQuery } from "#/__generated__/core/AccessReviewCampaignDetailPageQuery.graphql";
import { PageSkeleton } from "#/components/skeletons/PageSkeleton";

import AccessReviewCampaignDetailPage, { accessReviewCampaignDetailPageQuery } from "./AccessReviewCampaignDetailPage";

export default function AccessReviewCampaignDetailPageLoader() {
  const { campaignId } = useParams<{ campaignId: string }>();
  const [queryRef, loadQuery]
    = useQueryLoader<AccessReviewCampaignDetailPageQuery>(accessReviewCampaignDetailPageQuery);

  useEffect(() => {
    if (campaignId) {
      loadQuery({ campaignId });
    }
  }, [loadQuery, campaignId]);

  if (!queryRef) {
    return <PageSkeleton />;
  }

  return (
    <Suspense fallback={<PageSkeleton />}>
      <AccessReviewCampaignDetailPage queryRef={queryRef} />
    </Suspense>
  );
}

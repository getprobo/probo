import { Suspense, useEffect } from "react";
import { useQueryLoader } from "react-relay";
import { useParams } from "react-router";

import type { CampaignDetailPageQuery } from "#/__generated__/core/CampaignDetailPageQuery.graphql";
import { PageSkeleton } from "#/components/skeletons/PageSkeleton";

import CampaignDetailPage, { campaignDetailPageQuery } from "./CampaignDetailPage";

export default function CampaignDetailPageLoader() {
  const { campaignId } = useParams<{ campaignId: string }>();
  const [queryRef, loadQuery] = useQueryLoader<CampaignDetailPageQuery>(campaignDetailPageQuery);

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
      <CampaignDetailPage queryRef={queryRef} />
    </Suspense>
  );
}

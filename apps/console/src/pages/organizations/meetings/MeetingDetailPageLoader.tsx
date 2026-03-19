import { useEffect } from "react";
import { useQueryLoader } from "react-relay";
import { useParams } from "react-router";

import type { MeetingDetailPageQuery } from "#/__generated__/core/MeetingDetailPageQuery.graphql";
import { PageSkeleton } from "#/components/skeletons/PageSkeleton";
import { CoreRelayProvider } from "#/providers/CoreRelayProvider";

import MeetingDetailPage, { meetingDetailPageQuery } from "./MeetingDetailPage";

function MeetingDetailPageQueryLoader() {
  const { meetingId } = useParams<{ meetingId: string }>();
  const [queryRef, loadQuery] = useQueryLoader<MeetingDetailPageQuery>(meetingDetailPageQuery);

  useEffect(() => {
    if (!queryRef && meetingId) {
      loadQuery({ meetingId });
    }
  });

  if (!queryRef) return <PageSkeleton />;

  return <MeetingDetailPage queryRef={queryRef} />;
}

export default function MeetingDetailPageLoader() {
  return (
    <CoreRelayProvider>
      <MeetingDetailPageQueryLoader />
    </CoreRelayProvider>
  );
}

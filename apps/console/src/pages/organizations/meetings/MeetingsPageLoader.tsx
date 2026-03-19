import { useEffect } from "react";
import { useQueryLoader } from "react-relay";

import type { MeetingsPageQuery } from "#/__generated__/core/MeetingsPageQuery.graphql";
import { LinkCardSkeleton } from "#/components/skeletons/LinkCardSkeleton";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { CoreRelayProvider } from "#/providers/CoreRelayProvider";

import MeetingsPage, { meetingsPageQuery } from "./MeetingsPage";

function MeetingsPageQueryLoader() {
  const organizationId = useOrganizationId();
  const [queryRef, loadQuery] = useQueryLoader<MeetingsPageQuery>(meetingsPageQuery);

  useEffect(() => {
    loadQuery({ organizationId });
  }, [organizationId, loadQuery]);

  if (!queryRef) return <LinkCardSkeleton />;

  return <MeetingsPage queryRef={queryRef} />;
}

export default function MeetingsPageLoader() {
  return (
    <CoreRelayProvider>
      <MeetingsPageQueryLoader />
    </CoreRelayProvider>
  );
}

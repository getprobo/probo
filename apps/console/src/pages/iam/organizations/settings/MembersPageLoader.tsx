import { useEffect } from "react";
import { useQueryLoader } from "react-relay";

import type { MembersPageQuery } from "#/__generated__/iam/MembersPageQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { IAMRelayProvider } from "#/providers/IAMRelayProvider";

import { MembersPage, membersPageQuery } from "./MembersPage";

function MembersPageLoader() {
  const organizationId = useOrganizationId();
  const [queryRef, loadQuery]
    = useQueryLoader<MembersPageQuery>(membersPageQuery);

  useEffect(() => {
    loadQuery({
      organizationId,
    });
  }, [loadQuery, organizationId]);

  if (!queryRef) {
    return null;
  }

  return <MembersPage queryRef={queryRef} />;
}

export default function () {
  return (
    <IAMRelayProvider>
      <MembersPageLoader />
    </IAMRelayProvider>
  );
}

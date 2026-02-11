import { useEffect } from "react";
import { useQueryLoader } from "react-relay";

import type { PeoplePageQuery } from "#/__generated__/iam/PeoplePageQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { IAMRelayProvider } from "#/providers/IAMRelayProvider";

import { PeoplePage, peoplePageQuery } from "./PeoplePage";

function PeoplePageQueryLoader() {
  const organizationId = useOrganizationId();
  const [queryRef, loadQuery]
    = useQueryLoader<PeoplePageQuery>(peoplePageQuery);

  useEffect(() => {
    loadQuery({
      organizationId,
    });
  }, [loadQuery, organizationId]);

  if (!queryRef) {
    return null;
  }

  return <PeoplePage queryRef={queryRef} />;
}

export default function PeoplePageLoader() {
  return (
    <IAMRelayProvider>
      <PeoplePageQueryLoader />
    </IAMRelayProvider>
  );
}

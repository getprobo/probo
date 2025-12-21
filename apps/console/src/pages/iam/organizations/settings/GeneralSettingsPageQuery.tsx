import { graphql } from "relay-runtime";
import { useOrganizationId } from "/hooks/useOrganizationId";
import { useQueryLoader } from "react-relay";
import { useEffect } from "react";
import { GeneralSettingsPage } from "./GeneralSettingsPage";
import type { GeneralSettingsPageQuery } from "./__generated__/GeneralSettingsPageQuery.graphql";
import { IAMRelayProvider } from "/providers/IAMRelayProvider";

export const generalSettingsPageQuery = graphql`
  query GeneralSettingsPageQuery($organizationId: ID!) {
    organization: node(id: $organizationId) @required(action: THROW) {
      __typename
      ... on Organization {
        id
        name @required(action: THROW)
        ...OrganizationFormFragment
      }
    }
  }
`;

function GeneralSettingsPageQuery() {
  const organizationId = useOrganizationId();
  const [queryRef, loadQuery] = useQueryLoader<GeneralSettingsPageQuery>(
    generalSettingsPageQuery,
  );

  useEffect(() => {
    loadQuery({
      organizationId,
    });
  }, [loadQuery, organizationId]);

  if (!queryRef) {
    return null;
  }

  return <GeneralSettingsPage queryRef={queryRef} />;
}

export default function () {
  return (
    <IAMRelayProvider>
      <GeneralSettingsPageQuery />
    </IAMRelayProvider>
  );
}

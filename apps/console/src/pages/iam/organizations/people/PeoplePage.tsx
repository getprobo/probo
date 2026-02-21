import { useTranslate } from "@probo/i18n";
import { Button, PageHeader } from "@probo/ui";
import { useState } from "react";
import { type PreloadedQuery, usePreloadedQuery } from "react-relay";
import { ConnectionHandler, type DataID, graphql } from "relay-runtime";

import type { PeoplePageQuery } from "#/__generated__/iam/PeoplePageQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import { AddPersonDialog } from "./_components/AddPersonDialog";
import { PeopleList } from "./_components/PeopleList";

export const peoplePageQuery = graphql`
  query PeoplePageQuery($organizationId: ID!) {
    organization: node(id: $organizationId) @required(action: THROW) {
      __typename
      ... on Organization {
        canCreateUser: permission(action: "iam:membership-profile:create")
        ...PeopleListFragment
          @arguments(first: 20, order: { direction: ASC, field: FULL_NAME })
      }
    }
  }
`;

export function PeoplePage(props: {
  queryRef: PreloadedQuery<PeoplePageQuery>;
}) {
  const { queryRef } = props;

  const organizationId = useOrganizationId();
  const { __ } = useTranslate();

  const [connectionId, setConnectionId] = useState<DataID>(
    ConnectionHandler.getConnectionID(
      organizationId,
      "PeopleListFragment_profiles",
      { orderBy: { direction: "ASC", field: "FULL_NAME" } },
    ),
  );

  const { organization } = usePreloadedQuery<PeoplePageQuery>(
    peoplePageQuery,
    queryRef,
  );
  if (organization.__typename !== "Organization") {
    throw new Error("node is of invalid type");
  }

  return (
    <div className="space-y-6">
      <PageHeader title={__("People")}>
        {organization.canCreateUser
          && (
            <AddPersonDialog connectionId={connectionId}>
              <Button variant="secondary">{__("Add Person")}</Button>
            </AddPersonDialog>
          )}
      </PageHeader>

      <div className="pb-6 pt-6">
        <PeopleList
          fKey={organization}
          onConnectionIdChange={setConnectionId}
        />
      </div>
    </div>
  );
}

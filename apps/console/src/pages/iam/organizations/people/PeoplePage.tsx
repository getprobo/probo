// Copyright (c) 2026 Probo Inc <hello@probo.com>.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

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

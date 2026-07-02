// Copyright (c) 2026 Probo Inc <hello@probo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

import { useTranslate } from "@probo/i18n";
import { Button } from "@probo/ui";
import { useState } from "react";
import { type PreloadedQuery, usePreloadedQuery } from "react-relay";
import { ConnectionHandler, type DataID, graphql } from "relay-runtime";

import type { MembersSettingsPageQuery } from "#/__generated__/iam/MembersSettingsPageQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import { AddMemberDialog } from "./_components/AddMemberDialog";
import { MembersList } from "./_components/MembersList";

export const membersSettingsPageQuery = graphql`
  query MembersSettingsPageQuery($organizationId: ID!) {
    organization: node(id: $organizationId) @required(action: THROW) {
      __typename
      ... on Organization {
        canCreateUser: permission(action: "iam:membership-profile:create")
        ...MembersListFragment
          @arguments(first: 20, order: { direction: ASC, field: EMAIL_ADDRESS })
      }
    }
  }
`;

export function MembersSettingsPage(props: {
  queryRef: PreloadedQuery<MembersSettingsPageQuery>;
}) {
  const { queryRef } = props;

  const organizationId = useOrganizationId();
  const { __ } = useTranslate();

  const [connectionId, setConnectionId] = useState<DataID>(
    ConnectionHandler.getConnectionID(
      organizationId,
      "MembersListFragment_profiles",
      { orderBy: { direction: "ASC", field: "EMAIL_ADDRESS" } },
    ),
  );

  const { organization } = usePreloadedQuery<MembersSettingsPageQuery>(
    membersSettingsPageQuery,
    queryRef,
  );
  if (organization.__typename !== "Organization") {
    throw new Error("node is of invalid type");
  }

  return (
    <div className="space-y-6">
      {organization.canCreateUser && (
        <div className="flex justify-end">
          <AddMemberDialog connectionId={connectionId}>
            <Button variant="secondary">{__("Add member")}</Button>
          </AddMemberDialog>
        </div>
      )}

      <MembersList
        fKey={organization}
        onConnectionIdChange={setConnectionId}
      />
    </div>
  );
}

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

import { getAssignableRoles } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import { Tbody, Td, Th, Thead, Tr } from "@probo/ui";
import type { ComponentProps } from "react";
import { use } from "react";
import { ConnectionHandler, graphql, usePaginationFragment } from "react-relay";

import type { MembersListFragment$key } from "#/__generated__/iam/MembersListFragment.graphql";
import type { MembersListFragment_RefetchQuery } from "#/__generated__/iam/MembersListFragment_RefetchQuery.graphql";
import { type Order, SortableTable, SortableTh } from "#/components/SortableTable";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { CurrentUser } from "#/providers/CurrentUser";

import { MembersListItem } from "./MembersListItem";

const fragment = graphql`
  fragment MembersListFragment on Organization
  @refetchable(queryName: "MembersListFragment_RefetchQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 20 }
    order: {
      type: "ProfileOrder"
      defaultValue: { direction: ASC, field: EMAIL_ADDRESS }
    }
    after: { type: "CursorKey", defaultValue: null }
    before: { type: "CursorKey", defaultValue: null }
    last: { type: "Int", defaultValue: null }
  ) {
    profiles(
      first: $first
      after: $after
      last: $last
      before: $before
      orderBy: $order
    ) @connection(key: "MembersListFragment_profiles", filters: ["orderBy"]) @required(action: THROW) {
      __id
      totalCount
      edges @required(action: THROW) {
        node {
          id
          ...MembersListItemFragment
        }
      }
    }
  }
`;

export function MembersList(props: {
  fKey: MembersListFragment$key;
  onConnectionIdChange: (connectionId: string) => void;
}) {
  const { fKey, onConnectionIdChange } = props;

  const organizationId = useOrganizationId();
  const { __ } = useTranslate();
  const { role } = use(CurrentUser);
  const canManageRoles = getAssignableRoles(role).length > 0;

  const membersPagination = usePaginationFragment<
    MembersListFragment_RefetchQuery,
    MembersListFragment$key
  >(fragment, fKey);

  const refetchMembers = () => {
    membersPagination.refetch({}, { fetchPolicy: "network-only" });
  };

  const handleOrderChange = (order: Order) => {
    onConnectionIdChange(
      ConnectionHandler.getConnectionID(
        organizationId,
        "MembersListFragment_profiles",
        { orderBy: order },
      ),
    );
  };

  const columnCount = canManageRoles ? 3 : 2;

  return (
    <SortableTable
      {...membersPagination}
      refetch={
        membersPagination.refetch as ComponentProps<
          typeof SortableTable
        >["refetch"]
      }
      pageSize={20}
    >
      <Thead>
        <Tr>
          <SortableTh field="EMAIL_ADDRESS" onOrderChange={handleOrderChange}>{__("Email")}</SortableTh>
          {canManageRoles && <Th>{__("Role")}</Th>}
          <Th></Th>
        </Tr>
      </Thead>
      <Tbody>
        {membersPagination.data.profiles.totalCount === 0
          ? (
              <Tr>
                <Td colSpan={columnCount} className="text-center text-txt-secondary">
                  {__("No members")}
                </Td>
              </Tr>
            )
          : (
              membersPagination.data.profiles.edges.map(({ node: profile }) => (
                <MembersListItem
                  connectionId={membersPagination.data.profiles.__id}
                  key={profile.id}
                  fKey={profile}
                  onRefetch={refetchMembers}
                />
              ))
            )}
      </Tbody>
    </SortableTable>
  );
}

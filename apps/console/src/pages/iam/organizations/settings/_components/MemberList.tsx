import { useTranslate } from "@probo/i18n";
import { Tbody, Td, Th, Thead, Tr } from "@probo/ui";
import type { ComponentProps } from "react";
import { graphql, usePaginationFragment } from "react-relay";

import type { MemberListFragment$key } from "#/__generated__/iam/MemberListFragment.graphql";
import type { MemberListFragment_RefetchQuery } from "#/__generated__/iam/MemberListFragment_RefetchQuery.graphql";
import { SortableTable, SortableTh } from "#/components/SortableTable";

import { MemberListItem } from "./MemberListItem";

const fragment = graphql`
  fragment MemberListFragment on Organization
  @refetchable(queryName: "MemberListFragment_RefetchQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 20 }
    order: {
      type: "MembershipOrder"
      defaultValue: { direction: ASC, field: FULL_NAME }
    }
    after: { type: "CursorKey", defaultValue: null }
    before: { type: "CursorKey", defaultValue: null }
    last: { type: "Int", defaultValue: null }
  ) {
    members(
      first: $first
      after: $after
      last: $last
      before: $before
      orderBy: $order
    ) @connection(key: "MemberListFragment_members") @required(action: THROW) {
      __id
      totalCount
      edges @required(action: THROW) {
        node {
          id
          ...MemberListItemFragment
        }
      }
    }
  }
`;

export function MemberList(props: { fKey: MemberListFragment$key }) {
  const { fKey } = props;

  const { __ } = useTranslate();

  const membersPagination = usePaginationFragment<
    MemberListFragment_RefetchQuery,
    MemberListFragment$key
  >(fragment, fKey);

  const refetchMemberships = () => {
    membersPagination.refetch({}, { fetchPolicy: "network-only" });
  };

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
          <SortableTh field="FULL_NAME">{__("Name")}</SortableTh>
          <SortableTh field="EMAIL_ADDRESS">{__("Email")}</SortableTh>
          <Th>{__("Type")}</Th>
          <SortableTh field="ROLE">{__("Role")}</SortableTh>
          <SortableTh field="CREATED_AT">{__("Joined")}</SortableTh>
          <Th>{__("Position")}</Th>
          <Th></Th>
        </Tr>
      </Thead>
      <Tbody>
        {membersPagination.data.members.totalCount === 0
          ? (
              <Tr>
                <Td colSpan={7} className="text-center text-txt-secondary">
                  {__("No members")}
                </Td>
              </Tr>
            )
          : (
              membersPagination.data.members.edges.map(({ node: membership }) => (
                <MemberListItem
                  connectionId={membersPagination.data.members.__id}
                  key={membership.id}
                  fKey={membership}
                  onRefetch={refetchMemberships}
                />
              ))
            )}
      </Tbody>
    </SortableTable>
  );
}

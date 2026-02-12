import { useTranslate } from "@probo/i18n";
import { Tbody, Td, Th, Thead, Tr } from "@probo/ui";
import type { ComponentProps } from "react";
import { graphql, usePaginationFragment } from "react-relay";

import type { PeopleListFragment$key } from "#/__generated__/iam/PeopleListFragment.graphql";
import type { PeopleListFragment_RefetchQuery } from "#/__generated__/iam/PeopleListFragment_RefetchQuery.graphql";
import { SortableTable, SortableTh } from "#/components/SortableTable";

import { MemberListItem } from "./PeopleListItem";

const fragment = graphql`
  fragment PeopleListFragment on Organization
  @refetchable(queryName: "PeopleListFragment_RefetchQuery")
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
    ) @connection(key: "PeopleListFragment_members", filters: ["orderBy"]) @required(action: THROW) {
      __id
      totalCount
      edges @required(action: THROW) {
        node {
          id
          ...PeopleListItemFragment
        }
      }
    }
  }
`;

export function PeopleList(props: { fKey: PeopleListFragment$key }) {
  const { fKey } = props;

  const { __ } = useTranslate();

  const peoplePagination = usePaginationFragment<
    PeopleListFragment_RefetchQuery,
    PeopleListFragment$key
  >(fragment, fKey);

  const refetchMemberships = () => {
    peoplePagination.refetch({}, { fetchPolicy: "network-only" });
  };

  return (
    <SortableTable
      {...peoplePagination}
      refetch={
        peoplePagination.refetch as ComponentProps<
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
        {peoplePagination.data.members.totalCount === 0
          ? (
              <Tr>
                <Td colSpan={7} className="text-center text-txt-secondary">
                  {__("No members")}
                </Td>
              </Tr>
            )
          : (
              peoplePagination.data.members.edges.map(({ node: membership }) => (
                <MemberListItem
                  connectionId={peoplePagination.data.members.__id}
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

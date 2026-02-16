import { useTranslate } from "@probo/i18n";
import { Tbody, Td, Th, Thead, Tr } from "@probo/ui";
import type { ComponentProps } from "react";
import { ConnectionHandler, graphql, usePaginationFragment } from "react-relay";

import type { PeopleListFragment$key } from "#/__generated__/iam/PeopleListFragment.graphql";
import type { PeopleListFragment_RefetchQuery } from "#/__generated__/iam/PeopleListFragment_RefetchQuery.graphql";
import { type Order, SortableTable, SortableTh } from "#/components/SortableTable";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import { PeopleListItem } from "./PeopleListItem";

const fragment = graphql`
  fragment PeopleListFragment on Organization
  @refetchable(queryName: "PeopleListFragment_RefetchQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 20 }
    order: {
      type: "ProfileOrder"
      defaultValue: { direction: ASC, field: FULL_NAME }
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
    ) @connection(key: "PeopleListFragment_profiles", filters: ["orderBy"]) @required(action: THROW) {
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

export function PeopleList(props: {
  fKey: PeopleListFragment$key;
  onConnectionIdChange: (connectionId: string) => void;
}) {
  const { fKey, onConnectionIdChange } = props;

  const organizationId = useOrganizationId();
  const { __ } = useTranslate();

  const peoplePagination = usePaginationFragment<
    PeopleListFragment_RefetchQuery,
    PeopleListFragment$key
  >(fragment, fKey);

  const refetchPeople = () => {
    peoplePagination.refetch({}, { fetchPolicy: "network-only" });
  };

  const handleOrderChange = (order: Order) => {
    onConnectionIdChange(
      ConnectionHandler.getConnectionID(
        organizationId,
        "PeopleListFragment_profiles",
        { orderBy: order },
      ),
    );
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
          <SortableTh field="FULL_NAME" onOrderChange={handleOrderChange}>{__("Name")}</SortableTh>
          <SortableTh field="STATUS">{__("Status")}</SortableTh>
          <SortableTh field="EMAIL_ADDRESS" onOrderChange={handleOrderChange}>{__("Email")}</SortableTh>
          <SortableTh field="ROLE" onOrderChange={handleOrderChange}>{__("Role")}</SortableTh>
          <SortableTh field="CREATED_AT" onOrderChange={handleOrderChange}>{__("Created on")}</SortableTh>
          <Th></Th>
        </Tr>
      </Thead>
      <Tbody>
        {peoplePagination.data.profiles.totalCount === 0
          ? (
              <Tr>
                <Td colSpan={7} className="text-center text-txt-secondary">
                  {__("No people")}
                </Td>
              </Tr>
            )
          : (
              peoplePagination.data.profiles.edges.map(({ node: profile }) => (
                <PeopleListItem
                  connectionId={peoplePagination.data.profiles.__id}
                  key={profile.id}
                  fKey={profile}
                  onRefetch={refetchPeople}
                />
              ))
            )}
      </Tbody>
    </SortableTable>
  );
}

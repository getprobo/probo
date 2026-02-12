import { useTranslate } from "@probo/i18n";
import { Tbody, Td, Th, Thead, Tr } from "@probo/ui";
import { type ComponentProps } from "react";
import { ConnectionHandler, graphql, usePaginationFragment } from "react-relay";

import type { InvitationListFragment$key } from "#/__generated__/iam/InvitationListFragment.graphql";
import type { InvitationListFragment_RefetchQuery } from "#/__generated__/iam/InvitationListFragment_RefetchQuery.graphql";
import type { PeoplePage_invitationsTotalCountFragment$key } from "#/__generated__/iam/PeoplePage_invitationsTotalCountFragment.graphql";
import {
  type Order,
  SortableTable,
  SortableTh,
} from "#/components/SortableTable";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import { InvitationListItem } from "./InvitationListItem";

const fragment = graphql`
  fragment InvitationListFragment on Organization
  @refetchable(queryName: "InvitationListFragment_RefetchQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 20 }
    order: {
      type: "InvitationOrder"
      defaultValue: { direction: DESC, field: CREATED_AT }
    }
    after: { type: "CursorKey", defaultValue: null }
    before: { type: "CursorKey", defaultValue: null }
    last: { type: "Int", defaultValue: null }
  ) {
    invitations(
      first: $first
      after: $after
      last: $last
      before: $before
      orderBy: $order
    )
      @connection(
        key: "InvitationListFragment_invitations"
        filters: ["orderBy"]
      )
      @required(action: THROW) {
      __id
      edges @required(action: THROW) {
        node {
          id
          ...InvitationListItemFragment
        }
      }
    }
  }
`;

export function InvitationList(props: {
  fKey: InvitationListFragment$key;
  onConnectionIdChange: (connectionId: string) => void;
  totalCount: number;
  totalCountFKey: PeoplePage_invitationsTotalCountFragment$key;
}) {
  const { fKey, onConnectionIdChange, totalCount, totalCountFKey } = props;

  const organizationId = useOrganizationId();
  const { __ } = useTranslate();

  const invitationsPagination = usePaginationFragment<
    InvitationListFragment_RefetchQuery,
    InvitationListFragment$key
  >(fragment, fKey);

  const refetchInvitations = () => {
    invitationsPagination.refetch({}, { fetchPolicy: "network-only" });
  };

  const handleOrderChange = (order: Order) => {
    onConnectionIdChange(
      ConnectionHandler.getConnectionID(
        organizationId,
        "InvitationListFragment_invitations",
        { orderBy: order },
      ),
    );
  };

  return (
    <SortableTable
      {...invitationsPagination}
      refetch={
        invitationsPagination.refetch as ComponentProps<
          typeof SortableTable
        >["refetch"]
      }
      pageSize={20}
    >
      <Thead>
        <Tr>
          <SortableTh field="FULL_NAME" onOrderChange={handleOrderChange}>
            {__("Name")}
          </SortableTh>
          <SortableTh field="EMAIL" onOrderChange={handleOrderChange}>
            {__("Email")}
          </SortableTh>
          <SortableTh field="ROLE" onOrderChange={handleOrderChange}>
            {__("Role")}
          </SortableTh>
          <SortableTh field="CREATED_AT" onOrderChange={handleOrderChange}>
            {__("Invited")}
          </SortableTh>
          <Th>{__("Status")}</Th>
          <SortableTh field="ACCEPTED_AT" onOrderChange={handleOrderChange}>
            {__("Accepted at")}
          </SortableTh>
          <Th></Th>
        </Tr>
      </Thead>
      <Tbody>
        {totalCount === 0
          ? (
              <Tr>
                <Td colSpan={7} className="text-center text-txt-secondary">
                  {__("No invitations")}
                </Td>
              </Tr>
            )
          : (
              invitationsPagination.data.invitations.edges.map(
                ({ node: invitation }) => (
                  <InvitationListItem
                    connectionId={invitationsPagination.data.invitations.__id}
                    key={invitation.id}
                    fKey={invitation}
                    onRefetch={refetchInvitations}
                    totalCountFKey={totalCountFKey}
                  />
                ),
              )
            )}
      </Tbody>
    </SortableTable>
  );
}

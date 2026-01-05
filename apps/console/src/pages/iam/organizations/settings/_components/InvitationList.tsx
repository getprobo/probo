import { Tbody, Td, Th, Thead, Tr } from "@probo/ui";
import { SortableTable, SortableTh } from "/components/SortableTable";
import { useTranslate } from "@probo/i18n";
import { graphql, usePaginationFragment } from "react-relay";
import { InvitationListItem } from "./InvitationListItem";
import type { InvitationListFragment$key } from "/__generated__/iam/InvitationListFragment.graphql";
import type { InvitationListFragment_RefetchQuery } from "/__generated__/iam/InvitationListFragment_RefetchQuery.graphql";
import type { ComponentProps } from "react";

const fragment = graphql`
  fragment InvitationListFragment on Organization
  @refetchable(queryName: "InvitationListFragment_RefetchQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 20 }
    order: {
      type: "InvitationOrder"
      defaultValue: { direction: ASC, field: CREATED_AT }
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
      @connection(key: "InvitationListFragment_invitations", filters: [])
      @required(action: THROW) {
      __id
      totalCount
      edges @required(action: THROW) {
        node {
          id
          ...InvitationListItemFragment
        }
      }
    }
  }
`;

export function InvitationList(props: { fKey: InvitationListFragment$key }) {
  const { fKey } = props;

  const { __ } = useTranslate();

  const invitationsPagination = usePaginationFragment<
    InvitationListFragment_RefetchQuery,
    InvitationListFragment$key
  >(fragment, fKey);

  const refetchInvitations = () => {
    invitationsPagination.refetch({}, { fetchPolicy: "network-only" });
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
          <SortableTh field="FULL_NAME">{__("Name")}</SortableTh>
          <SortableTh field="EMAIL">{__("Email")}</SortableTh>
          <SortableTh field="ROLE">{__("Role")}</SortableTh>
          <SortableTh field="CREATED_AT">{__("Invited")}</SortableTh>
          <Th>{__("Status")}</Th>
          <SortableTh field="ACCEPTED_AT">{__("Accepted at")}</SortableTh>
          <Th></Th>
        </Tr>
      </Thead>
      <Tbody>
        {invitationsPagination.data.invitations.totalCount === 0 ? (
          <Tr>
            <Td colSpan={7} className="text-center text-txt-secondary">
              {__("No invitations")}
            </Td>
          </Tr>
        ) : (
          invitationsPagination.data.invitations.edges.map(
            ({ node: invitation }) => (
              <InvitationListItem
                connectionId={invitationsPagination.data.invitations.__id}
                key={invitation.id}
                fKey={invitation}
                onRefetch={refetchInvitations}
              />
            ),
          )
        )}
      </Tbody>
    </SortableTable>
  );
}
